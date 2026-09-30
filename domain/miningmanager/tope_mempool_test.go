package miningmanager_test

import (
	"strings"
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/consensus/model"
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

// TestForjaSobreTopeEnMempool: escenario 3 del auditor, el mas barato. Con el tope de
// Diamantes lleno, una forja mas llega al mempool. Debe rechazarse en la entrada
// (H-1 nivel B) y la plantilla de bloque debe seguir saliendo. Antes del arreglo el
// mempool la aceptaba y el builder fallaba la plantilla completa: un atacante sin
// hashrate dejaba a todos los mineros sin poder construir.
func TestForjaSobreTopeEnMempool(t *testing.T) {
	params := dagconfig.DevnetParams
	params.BlocksPerHalving = 50
	cfg := consensus.Config{Params: params}
	cfg.SkipProofOfWork = true
	cfg.BlockCoinbaseMaturity = 0
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&cfg, "TestForjaSobreTopeEnMempool")
	if err != nil {
		t.Fatalf("NewTestConsensus: %+v", err)
	}
	defer teardown(false)
	tcAsConsensus := tc.(externalapi.Consensus)
	tcAsConsensusPointer := &tcAsConsensus
	ref := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
	mm := miningmanager.NewFactory().NewMiningManager(ref, &cfg.Params, mempool.DefaultConfig(&cfg.Params))

	spk, redeem := testutils.OpTrueScript()
	sig, err := txscript.PayToScriptHashSignatureScript(redeem, nil)
	if err != nil {
		t.Fatalf("sigScript: %+v", err)
	}
	tip := cfg.GenesisHash
	var cb *externalapi.DomainTransaction
	for i := 0; i < 60; i++ {
		h, _, err := tc.AddBlock([]*externalapi.DomainHash{tip}, nil, nil)
		if err != nil {
			t.Fatalf("minar: %+v", err)
		}
		blk, _, err := tc.GetBlock(h)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}
		if cb == nil && len(blk.Transactions[0].Outputs) > 0 {
			cb = blk.Transactions[0]
		}
		tip = h
	}
	// Tope lleno.
	dbTx, err := tc.DatabaseContext().Begin()
	if err != nil {
		t.Fatalf("Begin: %+v", err)
	}
	sa := model.NewStagingArea()
	tc.GemsHistoryStore().Stage(sa, tip, &externalapi.GemsHistory{Diamante: constants.MaxDiamante})
	tc.GemsHistoryStore().Stage(sa, model.VirtualBlockHash, &externalapi.GemsHistory{Diamante: constants.MaxDiamante})
	if err := sa.Commit(dbTx); err != nil {
		t.Fatalf("Commit: %+v", err)
	}
	if err := dbTx.Commit(); err != nil {
		t.Fatalf("Commit db: %+v", err)
	}

	diez := uint64(constants.BurnRatio) * uint64(constants.RupiaPerRupix)
	v := cb.Outputs[0].Value
	forja := &externalapi.DomainTransaction{Version: constants.MaxTransactionVersion, Payload: []byte{},
		Inputs: []*externalapi.DomainTransactionInput{{
			PreviousOutpoint: externalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(cb), Index: 0},
			SignatureScript:  sig, Sequence: constants.MaxTxInSequenceNum}},
		Outputs: []*externalapi.DomainTransactionOutput{
			{Value: constants.GemAmount, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: spk.Script, Version: constants.LevelDiamante}},
			{Value: diez, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{txscript.OpReturn}, Version: constants.LevelGold}},
			{Value: v - diez - uint64(constants.RupiaPerRupix), ScriptPublicKey: &externalapi.ScriptPublicKey{Script: spk.Script, Version: constants.LevelGold}},
		}}

	_, err = mm.ValidateAndInsertTransaction(forja, false, false)
	if err == nil {
		t.Fatalf("el mempool ACEPTO la forja %d con el tope lleno", constants.MaxDiamante+1)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "tope") {
		t.Fatalf("el mempool rechazo la forja, pero no por el tope: %v", err)
	}
	// Y la plantilla sigue saliendo.
	if _, _, err := mm.GetBlockTemplate(&externalapi.DomainCoinbaseData{ScriptPublicKey: spk, ExtraData: nil}); err != nil {
		t.Fatalf("LIVENESS: la plantilla de bloque fallo: %+v", err)
	}
	t.Logf("forja %d rechazada en la puerta del mempool: %v; la plantilla sigue saliendo", constants.MaxDiamante+1, err)
}
