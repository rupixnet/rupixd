package consensus_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/consensus/utils/merkle"
	"github.com/rupixnet/rupixd/domain/consensus/utils/testutils"
	"github.com/rupixnet/rupixd/domain/consensus/utils/txscript"
	"github.com/rupixnet/rupixd/domain/dagconfig"
)

// TestTopeDeDiamantesEnBloqueReal (hueco #3 de ESPECIFICACION.md): el tope historico de
// Diamantes (MaxDiamante) lo aplica el consenso sobre bloques reales, no solo una
// funcion aritmetica. Forjar 2,100,000 Diamantes en un test es inviable, asi que se
// SIEMBRA el conteo historico del padre en MaxDiamante-1 (los stores que la validacion
// lee) y se construyen bloques reales encima:
//   - un bloque que forja 1 Diamante llega EXACTAMENTE al tope: entra (UTXOValid) y el
//     conteo guardado es MaxDiamante;
//   - un bloque que forja 1 mas (el 2,100,001) muere: o el builder honesto se niega, o
//     el validador lo descalifica (ErrGemsCapExceeded); en ningun caso queda UTXOValid.
func TestTopeDeDiamantesEnBloqueReal(t *testing.T) {
	params := dagconfig.DevnetParams
	params.BlocksPerHalving = 50 // Diamante abierto desde DAA 50
	cfg := consensus.Config{Params: params}
	cfg.SkipProofOfWork = true
	cfg.BlockCoinbaseMaturity = 0
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&cfg, "TestTopeDeDiamantesEnBloqueReal")
	if err != nil {
		t.Fatalf("NewTestConsensus: %+v", err)
	}
	defer teardown(false)

	opTrueSPK, redeem := testutils.OpTrueScript()
	sigScript, err := txscript.PayToScriptHashSignatureScript(redeem, nil)
	if err != nil {
		t.Fatalf("sigScript: %+v", err)
	}
	forjaDiamante := func(cb *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		diez := uint64(constants.BurnRatio) * uint64(constants.RupiaPerRupix)
		v := cb.Outputs[0].Value
		if v < diez+2*uint64(constants.RupiaPerRupix) {
			t.Fatalf("coinbase de %d rupias: no alcanza para un Diamante", v)
		}
		return &externalapi.DomainTransaction{Version: constants.MaxTransactionVersion, Payload: []byte{},
			Inputs: []*externalapi.DomainTransactionInput{{
				PreviousOutpoint: externalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(cb), Index: 0},
				SignatureScript:  sigScript, Sequence: constants.MaxTxInSequenceNum}},
			Outputs: []*externalapi.DomainTransactionOutput{
				{Value: constants.GemAmount, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: opTrueSPK.Script, Version: constants.LevelDiamante}},
				{Value: diez, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{txscript.OpReturn}, Version: constants.LevelGold}},
				{Value: v - diez - uint64(constants.RupiaPerRupix), ScriptPublicKey: &externalapi.ScriptPublicKey{Script: opTrueSPK.Script, Version: constants.LevelGold}},
			}}
	}
	estado := func(h *externalapi.DomainHash) externalapi.BlockStatus {
		st, err := tc.BlockStatusStore().Get(tc.DatabaseContext(), model.NewStagingArea(), h)
		if err != nil {
			t.Fatalf("BlockStatusStore: %+v", err)
		}
		return st
	}
	historia := func(h *externalapi.DomainHash) *externalapi.GemsHistory {
		gh, err := tc.GemsHistoryStore().Get(tc.DatabaseContext(), model.NewStagingArea(), h)
		if err != nil {
			t.Fatalf("GemsHistoryStore.Get: %+v", err)
		}
		return gh
	}

	// Minar hasta abrir el Diamante, guardando coinbases.
	tip := cfg.GenesisHash
	var coinbases []*externalapi.DomainTransaction
	for i := 0; i < 60; i++ {
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

	// Sembrar el conteo historico: el padre (tip) y el virtual "ya vieron" MaxDiamante-1.
	// Es lo que la validacion lee para el siguiente bloque; todo lo demas es real.
	sembrado := &externalapi.GemsHistory{Diamante: constants.MaxDiamante - 1}
	dbTx, err := tc.DatabaseContext().Begin()
	if err != nil {
		t.Fatalf("Begin: %+v", err)
	}
	sa := model.NewStagingArea()
	tc.GemsHistoryStore().Stage(sa, tip, sembrado.Clone())
	tc.GemsHistoryStore().Stage(sa, model.VirtualBlockHash, sembrado.Clone())
	if err := sa.Commit(dbTx); err != nil {
		t.Fatalf("Commit staging: %+v", err)
	}
	if err := dbTx.Commit(); err != nil {
		t.Fatalf("Commit db: %+v", err)
	}
	if historia(tip).Diamante != constants.MaxDiamante-1 {
		t.Fatalf("la siembra no quedo guardada")
	}

	// 1) El Diamante 2,100,000 (exactamente el tope) NACE. La forja va en bTope y la
	// ACEPTA el bloque siguiente (asi funciona la aceptacion en el DAG): el conteo se
	// refleja en bAcepta.
	bTope, _, err := tc.AddBlock([]*externalapi.DomainHash{tip}, nil, []*externalapi.DomainTransaction{forjaDiamante(coinbases[0])})
	if err != nil {
		t.Fatalf("el bloque con la forja del Diamante %d debe entrar: %+v", constants.MaxDiamante, err)
	}
	bAcepta, _, err := tc.AddBlock([]*externalapi.DomainHash{bTope}, nil, nil)
	if err != nil {
		t.Fatalf("el bloque que acepta la forja del tope debe entrar: %+v", err)
	}
	if st := estado(bAcepta); st != externalapi.StatusUTXOValid {
		t.Fatalf("bloque que acepta el Diamante %d: esperado UTXOValid, obtenido %s", constants.MaxDiamante, st)
	}
	if got := historia(bAcepta).Diamante; got != constants.MaxDiamante {
		t.Fatalf("conteo historico tras aceptar el tope: esperado %d, obtenido %d", constants.MaxDiamante, got)
	}

	// 2) El Diamante 2,100,001 va en bExceso. La pregunta que importa: que le pasa al
	// bloque HONESTO que lo mergea. Lo sano: la forja queda NO aceptada y la cadena sigue.
	// Lo peligroso: el bloque honesto muere por aceptar una tx ajena (veneno cerca del tope).
	bExceso, _, errExceso := tc.AddBlock([]*externalapi.DomainHash{bAcepta}, nil, []*externalapi.DomainTransaction{forjaDiamante(coinbases[1])})
	if errExceso != nil {
		if !errors.Is(errExceso, ruleerrors.ErrGemsCapExceeded) {
			t.Fatalf("insertar el bloque con la forja %d fallo por otra cosa: %+v", constants.MaxDiamante+1, errExceso)
		}
		t.Logf("insertar el bloque con la forja %d devolvio ErrGemsCapExceeded: %v", constants.MaxDiamante+1, errExceso)
		if bExceso != nil {
			if st, err := tc.BlockStatusStore().Get(tc.DatabaseContext(), model.NewStagingArea(), bExceso); err == nil {
				t.Logf("  ...y aun asi el bloque quedo guardado con estado %s", st)
			} else {
				t.Logf("  ...y el bloque NO quedo guardado (%v)", err)
			}
		}
	} else {
		t.Logf("bloque con la forja %d: estado %s", constants.MaxDiamante+1, estado(bExceso))
		hijo, _, err := tc.AddBlock([]*externalapi.DomainHash{bExceso}, nil, nil)
		switch {
		case err != nil:
			if !errors.Is(err, ruleerrors.ErrGemsCapExceeded) {
				t.Fatalf("construir sobre el bloque con la forja %d fallo por otra cosa: %+v", constants.MaxDiamante+1, err)
			}
			t.Logf("VENENO: el builder honesto no puede construir sobre el bloque que lleva la forja %d: %v", constants.MaxDiamante+1, err)
		default:
			st := estado(hijo)
			hist := historia(hijo)
			t.Logf("hijo del bloque con la forja %d: estado %s, conteo %d", constants.MaxDiamante+1, st, hist.Diamante)
			if hist.Diamante > constants.MaxDiamante {
				t.Fatalf("el conteo historico supero el tope: %d", hist.Diamante)
			}
			if st != externalapi.StatusUTXOValid {
				t.Logf("VENENO: el bloque honesto que mergea la forja %d queda %s", constants.MaxDiamante+1, st)
			}
		}
	}

	// 2b) EL ATACANTE no usa el builder honesto: fabrica el bloque a mano. Se construye un
	// bloque vacio valido, se le mete la forja y se recalcula la raiz de Merkle (nada mas
	// del encabezado depende de las propias transacciones). Es un bloque valido en si
	// mismo: su tx solo se acepta cuando alguien lo mergea. La pregunta: que pasa con el
	// virtual (que lo mergea al insertarlo) y con la red.
	veneno, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{bAcepta}, nil, nil)
	if err != nil {
		t.Fatalf("BuildBlockWithParents (vacio): %+v", err)
	}
	veneno.Transactions = append(veneno.Transactions, forjaDiamante(coinbases[1]))
	hdr := veneno.Header.ToMutable()
	hdr.SetHashMerkleRoot(merkle.CalculateHashMerkleRoot(veneno.Transactions))
	veneno.Header = hdr.ToImmutable()
	hashVeneno := consensushashing.BlockHash(veneno)
	errVeneno := tc.ValidateAndInsertBlock(veneno, true)
	// Lo observado el 30-sep-2026 y lo que se exige desde entonces: la actualizacion del
	// virtual (que mergea al bloque nuevo) detecta el tope, la insercion entera se
	// deshace y el bloque NO queda guardado. Ni veneno para bloques honestos ni perdida
	// de liveness: para un nodo honesto ese bloque no existe.
	if !errors.Is(errVeneno, ruleerrors.ErrGemsCapExceeded) {
		t.Fatalf("el bloque fabricado con la forja %d debe rechazarse con ErrGemsCapExceeded; devolvio: %v", constants.MaxDiamante+1, errVeneno)
	}
	if _, err := tc.BlockStatusStore().Get(tc.DatabaseContext(), model.NewStagingArea(), hashVeneno); err == nil {
		t.Fatalf("el bloque fabricado con la forja %d quedo guardado; no debe existir", constants.MaxDiamante+1)
	}
	hv, err := tc.GemsHistoryStore().Get(tc.DatabaseContext(), model.NewStagingArea(), model.VirtualBlockHash)
	if err != nil {
		t.Fatalf("GemsHistoryStore virtual: %+v", err)
	}
	if hv.Diamante != constants.MaxDiamante {
		t.Fatalf("conteo del virtual tras el veneno: esperado %d, obtenido %d", constants.MaxDiamante, hv.Diamante)
	}
	t.Logf("bloque fabricado con la forja %d: rechazado al insertar, no guardado; virtual en %d", constants.MaxDiamante+1, hv.Diamante)

	// Un hermano honesto que NO mergea al bloque envenenado: debe entrar sin problema.
	hermano, _, err := tc.AddBlock([]*externalapi.DomainHash{bAcepta}, nil, nil)
	if err != nil {
		t.Fatalf("LIVENESS: un bloque honesto hermano (sin mergear la forja %d) no pudo entrar: %+v", constants.MaxDiamante+1, err)
	}
	t.Logf("hermano honesto sobre bAcepta: estado %s", estado(hermano))

	// 3) Y el minero de produccion (construye sobre el virtual, sin elegir padres): sigue vivo?
	if _, err := tc.BuildBlock(&externalapi.DomainCoinbaseData{ScriptPublicKey: opTrueSPK, ExtraData: nil}, nil); err != nil {
		t.Fatalf("LIVENESS: el minero de produccion no puede construir tras la forja %d: %+v", constants.MaxDiamante+1, err)
	}
	sp, err := tc.GetVirtualSelectedParent()
	if err != nil {
		t.Fatalf("GetVirtualSelectedParent: %+v", err)
	}
	t.Logf("selected parent del virtual: %s (veneno=%s, hermano=%s)", sp, hashVeneno, hermano)
	if got := historia(sp).Diamante; got > constants.MaxDiamante {
		t.Fatalf("el virtual lleva un conteo por encima del tope: %d", got)
	}
}
