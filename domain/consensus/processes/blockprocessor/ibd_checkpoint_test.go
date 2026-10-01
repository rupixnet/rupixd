package blockprocessor_test

import (
	"math"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/model/testapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/rupixnet/rupixd/domain/consensus/utils/testutils"
	"github.com/rupixnet/rupixd/domain/dagconfig"
)

// TestNodoNuevoContraPeerHostil (hueco #2 de ESPECIFICACION.md): un nodo que
// sincroniza desde cero recibe de un peer la lista de pruning points y arranca
// desde el ultimo. Si la red tiene un checkpoint H publicado y la lista del peer
// no lo contiene (una historia alterna), el nodo nuevo NO debe importar ese
// pruning point: ErrCheckpointMismatch en ValidateAndInsertImportedPruningPoint.
//
// El "peer" es un consenso sin checkpoints que mina hasta que su pruning point
// se mueve; el "nodo nuevo" repite la sincronizacion del test heredado
// (TestValidateAndInsertImportedPruningPoint) con distintos checkpoints
// configurados y se mira solo el resultado de importar el pruning point.
func TestNodoNuevoContraPeerHostil(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()

		// Igual que el test heredado: profundidad de poda de unos 6 bloques.
		finalityDepth := 5
		consensusConfig.FinalityDuration = time.Duration(finalityDepth) * consensusConfig.TargetTimePerBlock
		consensusConfig.K = 0
		consensusConfig.PruningProofM = 1
		// El peer no tiene checkpoints: es el que sirve la historia.
		consensusConfig.Checkpoints = nil
		consensusConfig.CheckpointsExpireDAAScore = 0

		peer, teardownPeer, err := factory.NewTestConsensus(consensusConfig, "TestNodoNuevoContraPeerHostilPeer")
		if err != nil {
			t.Fatalf("peer: %+v", err)
		}
		defer teardownPeer(false)

		// Cadena lineal hasta que el pruning point deje de ser el genesis.
		tip := consensusConfig.GenesisHash
		var intermedio *externalapi.DomainHash
		for i := 0; ; i++ {
			tip = addBlock(peer, []*externalapi.DomainHash{tip}, t)
			if i == 1 {
				intermedio = tip
			}
			pp, err := peer.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}
			if !pp.Equal(consensusConfig.GenesisHash) {
				break
			}
		}
		// Unos bloques mas para que el pruning point tenga futuro.
		for i := 0; i < 3; i++ {
			tip = addBlock(peer, []*externalapi.DomainHash{tip}, t)
		}
		pruningPoint, err := peer.PruningPoint()
		if err != nil {
			t.Fatalf("PruningPoint: %+v", err)
		}
		ppHeader, err := peer.GetBlockHeader(pruningPoint)
		if err != nil {
			t.Fatalf("GetBlockHeader: %+v", err)
		}
		ppBlue := ppHeader.BlueScore()
		if ppBlue == 0 {
			t.Fatalf("el pruning point del peer sigue siendo el genesis")
		}
		otraHistoria := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{0xba, 0xd0, 0xca, 0xfe})

		casos := []struct {
			nombre      string
			checkpoints []dagconfig.Checkpoint
			quiereError bool
		}{
			{"sin checkpoints (control heredado)", nil, false},
			{"checkpoint = el pruning point real (peer honesto)",
				[]dagconfig.Checkpoint{{BlueScore: ppBlue, Hash: pruningPoint}}, false},
			{"checkpoint = un ancestro del pruning point que no es punto de poda (peer honesto)",
				[]dagconfig.Checkpoint{{BlueScore: 2, Hash: intermedio}}, false},
			{"checkpoint de OTRA historia por debajo del pruning point (peer hostil)",
				[]dagconfig.Checkpoint{{BlueScore: 1, Hash: otraHistoria}}, true},
			{"checkpoint por encima del pruning point (todavia no llega; lo revisa la regla por encabezado)",
				[]dagconfig.Checkpoint{{BlueScore: ppBlue + 1000, Hash: otraHistoria}}, false},
		}
		for i, caso := range casos {
			cfg := *consensusConfig
			cfg.Checkpoints = caso.checkpoints
			cfg.CheckpointsExpireDAAScore = 0
			err := sincronizaHastaImportarPruningPoint(t, factory, &cfg, peer, pruningPoint, i)
			switch {
			case caso.quiereError && err == nil:
				t.Fatalf("[%s] el nodo nuevo IMPORTO un pruning point cuya lista no contiene al checkpoint", caso.nombre)
			case caso.quiereError && !errors.Is(err, ruleerrors.ErrCheckpointMismatch):
				t.Fatalf("[%s] rechazado, pero no por el checkpoint: %+v", caso.nombre, err)
			case !caso.quiereError && err != nil:
				t.Fatalf("[%s] el nodo nuevo debia importar el pruning point: %+v", caso.nombre, err)
			}
			if err != nil {
				t.Logf("[%s] rechazado como debe ser: %v", caso.nombre, err)
			} else {
				t.Logf("[%s] importado", caso.nombre)
			}
		}
	})
}

// sincronizaHastaImportarPruningPoint repite los pasos del test heredado hasta
// ValidateAndInsertImportedPruningPoint y devuelve su resultado. Todo lo anterior
// tiene que salir bien (t.Fatalf si no): lo que se prueba es solo la importacion.
func sincronizaHastaImportarPruningPoint(t *testing.T, factory consensus.Factory, cfg *consensus.Config,
	peer testapi.TestConsensus, pruningPoint *externalapi.DomainHash, n int) error {

	nombre := func(s string) string { return "TestNodoNuevoContraPeerHostil" + s + string(rune('A'+n)) }

	nuevo, teardownNuevo, err := factory.NewTestConsensus(cfg, nombre("Nuevo"))
	if err != nil {
		t.Fatalf("nuevo: %+v", err)
	}
	defer teardownNuevo(false)

	proof, err := peer.BuildPruningPointProof()
	if err != nil {
		t.Fatalf("BuildPruningPointProof: %+v", err)
	}
	if err := nuevo.ValidatePruningPointProof(proof); err != nil {
		t.Fatalf("ValidatePruningPointProof: %+v", err)
	}

	stagingConfig := *cfg
	stagingConfig.SkipAddingGenesis = true
	staging, teardownStaging, err := factory.NewTestConsensus(&stagingConfig, nombre("Staging"))
	if err != nil {
		t.Fatalf("staging: %+v", err)
	}
	defer teardownStaging(false)

	if err := staging.ApplyPruningPointProof(proof); err != nil {
		t.Fatalf("ApplyPruningPointProof: %+v", err)
	}
	pruningPointHeaders, err := peer.PruningPointHeaders()
	if err != nil {
		t.Fatalf("PruningPointHeaders: %+v", err)
	}
	violaFinalidad, err := nuevo.ArePruningPointsViolatingFinality(pruningPointHeaders)
	if err != nil {
		t.Fatalf("ArePruningPointsViolatingFinality: %+v", err)
	}
	if violaFinalidad {
		t.Fatalf("violacion de finalidad inesperada")
	}
	if err := staging.ImportPruningPoints(pruningPointHeaders); err != nil {
		t.Fatalf("ImportPruningPoints: %+v", err)
	}

	anticono, err := peer.PruningPointAndItsAnticone()
	if err != nil {
		t.Fatalf("PruningPointAndItsAnticone: %+v", err)
	}
	for _, h := range anticono {
		block, _, err := peer.GetBlock(h)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}
		daaWindow, err := peer.BlockDAAWindowHashes(h)
		if err != nil {
			t.Fatalf("BlockDAAWindowHashes: %+v", err)
		}
		ghostdagHashes, err := peer.TrustedBlockAssociatedGHOSTDAGDataBlockHashes(h)
		if err != nil {
			t.Fatalf("TrustedBlockAssociatedGHOSTDAGDataBlockHashes: %+v", err)
		}
		conDatos := &externalapi.BlockWithTrustedData{
			Block:        block,
			DAAWindow:    make([]*externalapi.TrustedDataDataDAAHeader, 0, len(daaWindow)),
			GHOSTDAGData: make([]*externalapi.BlockGHOSTDAGDataHashPair, 0, len(ghostdagHashes)),
		}
		for i, dh := range daaWindow {
			hdr, err := peer.TrustedDataDataDAAHeader(h, dh, uint64(i))
			if err != nil {
				t.Fatalf("TrustedDataDataDAAHeader: %+v", err)
			}
			conDatos.DAAWindow = append(conDatos.DAAWindow, hdr)
		}
		for _, gh := range ghostdagHashes {
			data, err := peer.TrustedGHOSTDAGData(gh)
			if err != nil {
				t.Fatalf("TrustedGHOSTDAGData: %+v", err)
			}
			conDatos.GHOSTDAGData = append(conDatos.GHOSTDAGData, &externalapi.BlockGHOSTDAGDataHashPair{Hash: gh, GHOSTDAGData: data})
		}
		if err := staging.ValidateAndInsertBlockWithTrustedData(conDatos, false); err != nil {
			t.Fatalf("ValidateAndInsertBlockWithTrustedData: %+v", err)
		}
	}

	// Encabezados desde el pruning point hasta la punta del peer.
	peerTip, err := peer.GetVirtualSelectedParent()
	if err != nil {
		t.Fatalf("GetVirtualSelectedParent: %+v", err)
	}
	faltantes, _, err := peer.GetHashesBetween(pruningPoint, peerTip, math.MaxUint64)
	if err != nil {
		t.Fatalf("GetHashesBetween: %+v", err)
	}
	for i, h := range faltantes {
		info, err := staging.GetBlockInfo(h)
		if err != nil {
			t.Fatalf("GetBlockInfo: %+v", err)
		}
		if info.Exists {
			continue
		}
		header, err := peer.GetBlockHeader(h)
		if err != nil {
			t.Fatalf("GetBlockHeader: %+v", err)
		}
		if err := staging.ValidateAndInsertBlock(&externalapi.DomainBlock{Header: header}, false); err != nil {
			t.Fatalf("ValidateAndInsertBlock (encabezado %d): %+v", i, err)
		}
	}

	// UTXO set del pruning point.
	var desde *externalapi.DomainOutpoint
	var utxos []*externalapi.OutpointAndUTXOEntryPair
	const paso = 100_000
	for {
		lote, err := peer.GetPruningPointUTXOs(pruningPoint, desde, paso)
		if err != nil {
			t.Fatalf("GetPruningPointUTXOs: %+v", err)
		}
		desde = lote[len(lote)-1].Outpoint
		utxos = append(utxos, lote...)
		if len(lote) < paso {
			break
		}
	}
	if err := staging.AppendImportedPruningPointUTXOs(utxos); err != nil {
		t.Fatalf("AppendImportedPruningPointUTXOs: %+v", err)
	}

	// Lo que se prueba.
	return staging.ValidateAndInsertImportedPruningPoint(pruningPoint)
}
