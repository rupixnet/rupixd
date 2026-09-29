package consensus_test

import (
	"errors"
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/model/testapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/consensus/utils/txscript"
	"github.com/rupixnet/rupixd/domain/dagconfig"
)

// TestCheckpointDAG prueba la regla de checkpoints para DAG (v0.6.1): todo bloque con
// blue score >= X + MergeDepth debe tener a H en su pasado.
//
// Se construye el DAG una vez en un nodo SIN checkpoint (para conocer el hash de H)
// y se reproduce bloque por bloque en un nodo CON checkpoint en H:
//
//	genesis - g1..g5 - H - h1..h15          (cadena honesta, pasa el umbral)
//	                 \- a1..a11              (atacante: nunca incluye a H)
//	                 \- S                    (hermano de H, llega tarde)
//
// Esperado en el nodo con checkpoint:
//   - la cadena honesta entra completa, tambien despues del umbral;
//   - el atacante entra mientras esta por debajo de X + MergeDepth, y su primer
//     bloque en el umbral se rechaza con ErrCheckpointMismatch;
//   - S, hermano de H con su mismo blue score, entra aunque llegue despues de que
//     la red paso el umbral (el caso que la version por DAA rompia).
func TestCheckpointDAG(t *testing.T) {
	params := dagconfig.DevnetParams
	params.MergeDepth = 10

	nuevo := func(cps []dagconfig.Checkpoint, name string) (testapi.TestConsensus, func(bool)) {
		p := params
		p.Checkpoints = cps
		cfg := consensus.Config{Params: p}
		cfg.SkipProofOfWork = true
		tc, teardown, err := consensus.NewFactory().NewTestConsensus(&cfg, name)
		if err != nil {
			t.Fatalf("NewTestConsensus: %+v", err)
		}
		return tc, teardown
	}

	// Nodo A: sin checkpoint. Aqui se construyen todos los bloques.
	tcA, teardownA := nuevo(nil, "TestCheckpointDAG_A")
	defer teardownA(false)

	type paso struct {
		nombre string
		block  *externalapi.DomainBlock
	}
	var orden []paso
	agregar := func(nombre string, parent *externalapi.DomainHash, extra string) *externalapi.DomainHash {
		var cb *externalapi.DomainCoinbaseData
		if extra != "" {
			script, err := txscript.PayToScriptHashScript([]byte{txscript.OpTrue})
			if err != nil {
				t.Fatalf("script: %+v", err)
			}
			cb = &externalapi.DomainCoinbaseData{
				ScriptPublicKey: &externalapi.ScriptPublicKey{Script: script, Version: 0},
				ExtraData:       []byte(extra),
			}
		}
		block, _, err := tcA.BuildBlockWithParents([]*externalapi.DomainHash{parent}, cb, nil)
		if err != nil {
			t.Fatalf("BuildBlockWithParents %s: %+v", nombre, err)
		}
		if err := tcA.ValidateAndInsertBlock(block, true); err != nil {
			t.Fatalf("nodo sin checkpoint rechazo %s: %+v", nombre, err)
		}
		orden = append(orden, paso{nombre, block})
		return consensushashing.BlockHash(block)
	}
	blueScore := func(tc testapi.TestConsensus, h *externalapi.DomainHash) uint64 {
		d, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), model.NewStagingArea(), h, false)
		if err != nil {
			t.Fatalf("GHOSTDAG %s: %+v", h, err)
		}
		return d.BlueScore()
	}

	tip := params.GenesisHash
	for i := 1; i <= 5; i++ {
		tip = agregar("g", tip, "")
	}
	g5 := tip
	H := agregar("H", g5, "")
	tip = H
	for i := 1; i <= 15; i++ {
		tip = agregar("h", tip, "")
	}
	honestoFinal := tip
	tip = g5
	for i := 1; i <= 11; i++ {
		tip = agregar("a", tip, "atacante")
	}
	S := agregar("S", g5, "tardio")

	X := blueScore(tcA, H)
	umbral := X + params.MergeDepth
	if blueScore(tcA, honestoFinal) < umbral {
		t.Fatalf("la cadena honesta debe pasar el umbral: %d < %d", blueScore(tcA, honestoFinal), umbral)
	}
	if blueScore(tcA, S) != X {
		t.Fatalf("S debe tener el mismo blue score que H: %d != %d", blueScore(tcA, S), X)
	}

	// Nodo B: con checkpoint en H. Se reproduce el mismo DAG.
	tcB, teardownB := nuevo([]dagconfig.Checkpoint{{BlueScore: X, Hash: H}}, "TestCheckpointDAG_B")
	defer teardownB(false)

	atacanteRechazado := false
	for _, p := range orden {
		hash := consensushashing.BlockHash(p.block)
		bs := blueScore(tcA, hash)
		err := tcB.ValidateAndInsertBlock(p.block, true)
		switch p.nombre {
		case "a":
			if atacanteRechazado {
				continue // sus hijos ya no pueden entrar; no es lo que se prueba aqui
			}
			if bs < umbral {
				if err != nil {
					t.Fatalf("bloque del atacante por debajo del umbral (blue score %d < %d) debe entrar, dio: %+v", bs, umbral, err)
				}
			} else {
				if !errors.Is(err, ruleerrors.ErrCheckpointMismatch) {
					t.Fatalf("bloque del atacante en el umbral (blue score %d) debe rechazarse con ErrCheckpointMismatch, dio: %+v", bs, err)
				}
				atacanteRechazado = true
			}
		default:
			if err != nil {
				t.Fatalf("bloque %s (blue score %d) debe entrar, dio: %+v", p.nombre, bs, err)
			}
		}
	}
	if !atacanteRechazado {
		t.Fatalf("el atacante nunca llego al umbral: la prueba no probo el rechazo")
	}
}
