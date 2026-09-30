package miningmanager_test

import (
	"strings"
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/consensus/utils/testutils"
	"github.com/rupixnet/rupixd/domain/consensus/utils/txscript"
	"github.com/rupixnet/rupixd/domain/consensusreference"
	"github.com/rupixnet/rupixd/domain/dagconfig"
	"github.com/rupixnet/rupixd/domain/miningmanager"
	"github.com/rupixnet/rupixd/domain/miningmanager/mempool"
)

// TestMempoolRechazaNivelCerrado (H-1, nivel A): el mempool aplica las reglas de la
// escalera ANTES de aceptar una transaccion. Es la prueba automatica de lo que se vio
// en vivo el 28-sep-2026, cuando el nodo de JC rechazo un Platino en el DAA 185,738.
//
// Laboratorio: devnet con BlocksPerHalving=50 -> Diamante en DAA 50, Platino en DAA 100.
//  1. Con el Platino cerrado, el mempool rechaza la forja con "nivel 2 bloqueado".
//  2. La misma transaccion, pasado el DAA 100, el mempool la acepta.
//
// El par rechazo/aceptacion demuestra que el rechazo viene de la regla del nivel y no
// de otra cosa (firma, comision, formato): es la misma transaccion en los dos casos.
func TestMempoolRechazaNivelCerrado(t *testing.T) {
	params := dagconfig.DevnetParams
	params.BlocksPerHalving = 50
	cfg := consensus.Config{Params: params}
	cfg.SkipProofOfWork = true
	cfg.BlockCoinbaseMaturity = 0
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&cfg, "TestMempoolRechazaNivelCerrado")
	if err != nil {
		t.Fatalf("NewTestConsensus: %+v", err)
	}
	defer teardown(false)

	tcAsConsensus := tc.(externalapi.Consensus)
	tcAsConsensusPointer := &tcAsConsensus
	ref := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
	mm := miningmanager.NewFactory().NewMiningManager(ref, &cfg.Params, mempool.DefaultConfig(&cfg.Params))

	opTrueSPK, redeem := testutils.OpTrueScript()
	sigScript, err := txscript.PayToScriptHashSignatureScript(redeem, nil)
	if err != nil {
		t.Fatalf("sigScript: %+v", err)
	}
	spend := func(tx *externalapi.DomainTransaction, idx uint32) *externalapi.DomainTransactionInput {
		return &externalapi.DomainTransactionInput{
			PreviousOutpoint: externalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(tx), Index: idx},
			SignatureScript:  sigScript,
			Sequence:         constants.MaxTxInSequenceNum,
		}
	}
	gold := func(v uint64) *externalapi.DomainTransactionOutput {
		return &externalapi.DomainTransactionOutput{Value: v, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: opTrueSPK.Script, Version: constants.LevelGold}}
	}
	gem := func(level uint16) *externalapi.DomainTransactionOutput {
		return &externalapi.DomainTransactionOutput{Value: 1, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: opTrueSPK.Script, Version: level}}
	}
	burn := func(v uint64) *externalapi.DomainTransactionOutput {
		return &externalapi.DomainTransactionOutput{Value: v, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{txscript.OpReturn}, Version: constants.LevelGold}}
	}
	mkTx := func(ins []*externalapi.DomainTransactionInput, outs []*externalapi.DomainTransactionOutput) *externalapi.DomainTransaction {
		return &externalapi.DomainTransaction{Version: constants.MaxTransactionVersion, Inputs: ins, Outputs: outs, Payload: []byte{}}
	}
	daa := func() uint64 {
		d, err := tc.GetVirtualDAAScore()
		if err != nil {
			t.Fatalf("GetVirtualDAAScore: %+v", err)
		}
		return d
	}

	// 1) Minar hasta abrir el Diamante (DAA 50), guardando las coinbases con Gold.
	tip := cfg.GenesisHash
	var coinbases []*externalapi.DomainTransaction
	mine := func(n int) {
		for i := 0; i < n; i++ {
			h, _, err := tc.AddBlock([]*externalapi.DomainHash{tip}, nil, nil)
			if err != nil {
				t.Fatalf("mine: %+v", err)
			}
			blk, _, err := tc.GetBlock(h)
			if err != nil {
				t.Fatalf("GetBlock: %+v", err)
			}
			if len(blk.Transactions[0].Outputs) > 0 {
				coinbases = append(coinbases, blk.Transactions[0])
			}
			tip = h
		}
	}
	mine(60)
	if d := daa(); d < 50 || d >= 100 {
		t.Fatalf("el laboratorio necesita 50 <= DAA < 100, esta en %d", d)
	}

	// 2) 10 Diamantes en un bloque, cada uno de una coinbase distinta.
	const diez = 10 * constants.RupiaPerRupix
	var diamantes []*externalapi.DomainTransaction
	for i := 0; i < 10; i++ {
		cb := coinbases[i]
		v := cb.Outputs[0].Value
		if v < diez+2 {
			t.Fatalf("coinbase %d con %d rupias: no alcanza para un Diamante", i, v)
		}
		diamantes = append(diamantes, mkTx([]*externalapi.DomainTransactionInput{spend(cb, 0)},
			[]*externalapi.DomainTransactionOutput{gem(constants.LevelDiamante), burn(diez), gold(v - diez - 1)}))
	}
	h, _, err := tc.AddBlock([]*externalapi.DomainHash{tip}, nil, diamantes)
	if err != nil {
		t.Fatalf("bloque con 10 Diamantes: %+v", err)
	}
	tip = h
	mine(1) // el siguiente bloque de la cadena acepta las forjas

	// 3) La forja del Platino: 10 Diamantes + 1 coinbase de Gold para pagar la comision.
	const comision = constants.RupiaPerRupix / 100
	cbGold := coinbases[20]
	ins := []*externalapi.DomainTransactionInput{spend(cbGold, 0)}
	for _, d := range diamantes {
		ins = append(ins, spend(d, 0))
	}
	platino := mkTx(ins, []*externalapi.DomainTransactionOutput{gem(constants.LevelPlatino), gold(cbGold.Outputs[0].Value - comision)})

	if d := daa(); d >= 100 {
		t.Fatalf("el Platino ya esta abierto (DAA %d): la prueba no probaria el rechazo", d)
	}
	_, err = mm.ValidateAndInsertTransaction(platino, false, false)
	if err == nil {
		t.Fatalf("el mempool ACEPTO un Platino con el nivel cerrado (DAA %d)", daa())
	}
	if !strings.Contains(err.Error(), "nivel 2 bloqueado") {
		t.Fatalf("el mempool rechazo el Platino, pero no por el nivel cerrado: %v", err)
	}
	t.Logf("DAA %d: rechazado como se esperaba: %v", daa(), err)

	// 4) Pasado el DAA 100, la MISMA transaccion entra al mempool.
	for daa() < 101 {
		mine(1)
	}
	if _, err := mm.ValidateAndInsertTransaction(platino, false, false); err != nil {
		t.Fatalf("con el Platino abierto (DAA %d) la misma forja debe entrar al mempool, dio: %v", daa(), err)
	}
	t.Logf("DAA %d: la misma forja entro al mempool", daa())
}
