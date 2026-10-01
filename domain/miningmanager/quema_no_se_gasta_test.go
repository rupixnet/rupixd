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

// TestQuemaNoSeGasta (hueco #5 de ESPECIFICACION.md): lo que se quema no existe.
// Una salida OpReturn (quema) no entra al UTXO set, asi que nadie puede gastarla:
// ni por el mempool ni metiendola en un bloque. La salida de cambio de la MISMA
// transaccion si existe y si se gasta (control).
func TestQuemaNoSeGasta(t *testing.T) {
	params := dagconfig.DevnetParams
	cfg := consensus.Config{Params: params}
	cfg.SkipProofOfWork = true
	cfg.BlockCoinbaseMaturity = 0
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&cfg, "TestQuemaNoSeGasta")
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
	burn := func(v uint64) *externalapi.DomainTransactionOutput {
		return &externalapi.DomainTransactionOutput{Value: v, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{txscript.OpReturn}, Version: constants.LevelGold}}
	}
	mkTx := func(ins []*externalapi.DomainTransactionInput, outs []*externalapi.DomainTransactionOutput) *externalapi.DomainTransaction {
		return &externalapi.DomainTransaction{Version: constants.MaxTransactionVersion, Inputs: ins, Outputs: outs, Payload: []byte{}}
	}

	// Unos bloques para tener una coinbase gastable.
	tip := cfg.GenesisHash
	var cb *externalapi.DomainTransaction
	for i := 0; i < 5; i++ {
		h, _, err := tc.AddBlock([]*externalapi.DomainHash{tip}, nil, nil)
		if err != nil {
			t.Fatalf("mine: %+v", err)
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
	if cb == nil {
		t.Fatalf("no hubo coinbase con salidas")
	}

	// tx1: gasta la coinbase; salida 0 = cambio en Gold, salida 1 = QUEMA de 1 RUPIX.
	const unRupix = uint64(constants.RupiaPerRupix)
	v := cb.Outputs[0].Value
	if v < 2*unRupix {
		t.Fatalf("coinbase de %d rupias, no alcanza", v)
	}
	tx1 := mkTx([]*externalapi.DomainTransactionInput{spend(cb, 0)},
		[]*externalapi.DomainTransactionOutput{gold(v - unRupix - unRupix/10), burn(unRupix)})
	h, _, err := tc.AddBlock([]*externalapi.DomainHash{tip}, nil, []*externalapi.DomainTransaction{tx1})
	if err != nil {
		t.Fatalf("bloque con la quema: %+v", err)
	}
	tip = h
	if _, _, err := tc.AddBlock([]*externalapi.DomainHash{tip}, nil, nil); err != nil {
		t.Fatalf("mine: %+v", err)
	}
	id1 := *consensushashing.TransactionID(tx1)

	// 1) En el UTXO set del virtual: el cambio EXISTE, la quema NO.
	padresVirtual, err := tc.GetVirtualInfo()
	if err != nil {
		t.Fatalf("GetVirtualInfo: %+v", err)
	}
	utxos, err := tc.GetVirtualUTXOs(padresVirtual.ParentHashes, nil, 100000)
	if err != nil {
		t.Fatalf("GetVirtualUTXOs: %+v", err)
	}
	hayCambio, hayQuema := false, false
	for _, u := range utxos {
		if u.Outpoint.TransactionID.Equal(&id1) {
			switch u.Outpoint.Index {
			case 0:
				hayCambio = true
			case 1:
				hayQuema = true
			}
		}
	}
	if !hayCambio {
		t.Fatalf("la salida de cambio de tx1 deberia estar en el UTXO set")
	}
	if hayQuema {
		t.Fatalf("la QUEMA de tx1 esta en el UTXO set: lo quemado podria gastarse")
	}

	// 2) Mempool: gastar la quema se rechaza (el outpoint no existe); gastar el cambio entra.
	gastaQuema := mkTx([]*externalapi.DomainTransactionInput{spend(tx1, 1)}, []*externalapi.DomainTransactionOutput{gold(unRupix / 2)})
	if _, err := mm.ValidateAndInsertTransaction(gastaQuema, false, false); err == nil {
		t.Fatalf("el mempool ACEPTO gastar una quema")
	} else if !strings.Contains(strings.ToLower(err.Error()), "orphan") {
		t.Fatalf("el mempool rechazo gastar la quema, pero no por outpoint inexistente: %v", err)
	}
	gastaCambio := mkTx([]*externalapi.DomainTransactionInput{spend(tx1, 0)}, []*externalapi.DomainTransactionOutput{gold(v - 2*unRupix)})
	if _, err := mm.ValidateAndInsertTransaction(gastaCambio, false, false); err != nil {
		t.Fatalf("control: gastar el cambio de tx1 debe entrar al mempool: %v", err)
	}

	// 3) Por bloque: un bloque que gasta la quema no queda UTXOValid.
	h, _, err = tc.AddBlock([]*externalapi.DomainHash{tip}, nil, []*externalapi.DomainTransaction{gastaQuema})
	if err == nil {
		st, err := tc.BlockStatusStore().Get(tc.DatabaseContext(), model.NewStagingArea(), h)
		if err != nil {
			t.Fatalf("BlockStatusStore: %+v", err)
		}
		if st == externalapi.StatusUTXOValid {
			t.Fatalf("un bloque que gasta una quema quedo UTXOValid")
		}
		t.Logf("bloque que gasta la quema: %s", st)
	} else {
		t.Logf("bloque que gasta la quema rechazado al construir/insertar: %v", err)
	}
}
