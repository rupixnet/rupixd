package consensus_test

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/consensus/utils/gemscommitment"
	"github.com/rupixnet/rupixd/domain/consensus/utils/testutils"
	"github.com/rupixnet/rupixd/domain/consensus/utils/txscript"
	"github.com/rupixnet/rupixd/domain/dagconfig"
)

// TestKingsEndToEnd (Rupix) — el test que el auditor pidio ver.
// Forja un King de punta a punta con el codigo REAL: coinbases del block builder,
// forjas como txs reales validadas por level_ascension.go, bloques construidos por
// block_builder.go y validados por verify_and_build_utxo.go. Al final,
// gemsHistory.Kings == 1 leido del store y el header REAL lo sella. Si se revierte
// H-10 (el minero deja de contar Kings), el bloque del King se descalifica y falla.
//
// Laboratorio: devnet con BlocksPerHalving=20 -> Kings se desbloquea en DAA 80.
// Devnet paga 500 RUPIX por bloque; una coinbase da para 49 Diamantes (10 RUPIX c/u).
func TestKingsEndToEnd(t *testing.T) {
	params := dagconfig.DevnetParams
	params.BlocksPerHalving = 20
	cfg := consensus.Config{Params: params}
	cfg.SkipProofOfWork = true
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&cfg, "TestKingsEndToEnd")
	if err != nil {
		t.Fatalf("NewTestConsensus: %+v", err)
	}
	defer teardown(false)

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
	addValid := func(tip *externalapi.DomainHash, txs []*externalapi.DomainTransaction, what string) *externalapi.DomainHash {
		h, _, err := tc.AddBlock([]*externalapi.DomainHash{tip}, nil, txs)
		if err != nil {
			t.Fatalf("AddBlock %s: %+v", what, err)
		}
		st, err := tc.BlockStatusStore().Get(tc.DatabaseContext(), model.NewStagingArea(), h)
		if err != nil {
			t.Fatalf("status %s: %+v", what, err)
		}
		if st != externalapi.StatusUTXOValid {
			t.Fatalf("bloque con %s no es UTXOValid: %s", what, st)
		}
		return h
	}
	gemsAt := func(h *externalapi.DomainHash) *externalapi.GemsHistory {
		gh, err := tc.GemsHistoryStore().Get(tc.DatabaseContext(), model.NewStagingArea(), h)
		if err != nil {
			t.Fatalf("gemsHistory: %+v", err)
		}
		return gh
	}

	// En un BlockDAG las txs de un bloque las ACEPTA (y cuenta) el bloque siguiente de la
	// cadena seleccionada, igual que el UTXO set. Para leer el conteo que incluye la
	// ultima forja, se mina un bloque vacio encima.
	seal := func(tip *externalapi.DomainHash) *externalapi.DomainHash {
		return addValid(tip, nil, "sello")
	}
	// 1) Minar: 40 bloques para desbloquear Kings + madurez de coinbases (100) + margen.
	tip := cfg.GenesisHash
	var coinbases []*externalapi.DomainTransaction
	for i := 0; i < 220; i++ {
		h, _, err := tc.AddBlock([]*externalapi.DomainHash{tip}, nil, nil)
		if err != nil {
			t.Fatalf("mine %d: %+v", i, err)
		}
		blk, _, _ := tc.GetBlock(h)
		if len(blk.Transactions[0].Outputs) > 0 {
			coinbases = append(coinbases, blk.Transactions[0])
		}
		tip = h
	}
	t.Logf("%d coinbases utiles; Kings se desbloquea en DAA %d", len(coinbases), constants.LevelUnlockDaaScore(constants.LevelKings, params.BlocksPerHalving))

	const diez = 10 * constants.RupiaPerRupix
	//  2. 1000 Diamantes, en cadena: cada forja gasta el cambio (output 2) de la anterior.
	//     Cuando el cambio no alcanza, se toma la siguiente coinbase.
	t.Log("1000 Diamantes...")
	var diamantes []*externalapi.DomainTransaction
	ci := 0
	src, srcIdx, disp := coinbases[ci], uint32(0), coinbases[ci].Outputs[0].Value
	// Kaspa no permite gastar en el mismo bloque un output creado en ese bloque
	// (ErrChainedTransactions): una forja por bloque, cada una gasta el cambio de la anterior.
	for len(diamantes) < 1000 {
		for disp < 2*diez+2 { // la fuente no alcanza: siguiente coinbase con Gold suficiente
			ci++
			if ci >= len(coinbases) {
				t.Fatalf("se acabaron las coinbases con %d Diamantes forjados: minar mas bloques", len(diamantes))
			}
			src, srcIdx, disp = coinbases[ci], 0, coinbases[ci].Outputs[0].Value
		}
		tx := mkTx([]*externalapi.DomainTransactionInput{spend(src, srcIdx)},
			[]*externalapi.DomainTransactionOutput{gem(constants.LevelDiamante), burn(diez), gold(disp - diez - 1)})
		tip = addValid(tip, []*externalapi.DomainTransaction{tx}, "Diamante")
		diamantes = append(diamantes, tx)
		src, srcIdx, disp = tx, 2, tx.Outputs[2].Value
	}
	tip = seal(tip)
	if got := gemsAt(tip).Diamante; got != 1000 {
		// diagnostico: en que bloque de la cadena se perdio uno?
		prev := uint64(0)
		for i, d := range diamantes {
			_ = d
			if i < 3 || i > 996 {
				t.Logf("  tras Diamante #%d: (ver abajo)", i+1)
			}
		}
		t.Logf("gemsHistory en el tip: %+v (prev=%d)", gemsAt(tip), prev)
		t.Fatalf("Diamantes nacidos: %d, esperaba 1000", got)
	}

	//  3. 100 Platinos: 10 Diamantes (output 0) -> 1 Platino. Gold extra para el burn por tx
	//     no hace falta en devnet (BurnBase=0); si la regla exige OpReturn, el error lo dira.
	forgeUp := func(level uint16, fuel []*externalapi.DomainTransaction) *externalapi.DomainTransaction {
		var ins []*externalapi.DomainTransactionInput
		for _, f := range fuel {
			ins = append(ins, spend(f, 0))
		}
		return mkTx(ins, []*externalapi.DomainTransactionOutput{gem(level)})
	}
	t.Log("100 Platinos...")
	var platinos []*externalapi.DomainTransaction
	for len(platinos) < 100 {
		var txs []*externalapi.DomainTransaction
		for k := 0; k < 25 && len(platinos)+len(txs) < 100; k++ {
			i := (len(platinos) + len(txs)) * 10
			txs = append(txs, forgeUp(constants.LevelPlatino, diamantes[i:i+10]))
		}
		tip = addValid(tip, txs, "Platinos")
		platinos = append(platinos, txs...)
	}
	tip = seal(tip)
	if got := gemsAt(tip).Platino; got != 100 {
		t.Fatalf("Platinos nacidos: %d, esperaba 100", got)
	}

	// 4) 10 Rodios
	t.Log("10 Rodios...")
	var rodios []*externalapi.DomainTransaction
	for k := 0; k < 10; k++ {
		rodios = append(rodios, forgeUp(constants.LevelRodio, platinos[k*10:k*10+10]))
	}
	tip = addValid(tip, rodios, "Rodios")
	tip = seal(tip)
	if got := gemsAt(tip).Rodio; got != 10 {
		t.Fatalf("Rodios nacidos: %d, esperaba 10", got)
	}

	//  5. EL KING — construido por el MINERO DE PRODUCCION (tc.BuildBlock = block_builder.go,
	//     el mismo camino que rupixminer), no por el test builder. Asi el header lo sella
	//     newBlockGemsCommitment (H-10) y lo valida verify_and_build_utxo. Si H-10 se
	//     revierte (Kings=0 en el minero), el validador rechaza el bloque y esto falla.
	// Un Diamante minado por el block_builder de PRODUCCION (cobertura: el minero real forja bien)
	{
		for disp < 2*diez+2 {
			ci++
			src, srcIdx, disp = coinbases[ci], 0, coinbases[ci].Outputs[0].Value
		}
		dtx := mkTx([]*externalapi.DomainTransactionInput{spend(src, srcIdx)},
			[]*externalapi.DomainTransactionOutput{gem(constants.LevelDiamante), burn(diez), gold(disp - diez - 1)})
		dblk, err := tc.BuildBlock(&externalapi.DomainCoinbaseData{ScriptPublicKey: opTrueSPK}, []*externalapi.DomainTransaction{dtx})
		if err != nil {
			t.Fatalf("BuildBlock Diamante (minero real): %+v", err)
		}
		if err := tc.ValidateAndInsertBlock(dblk, true); err != nil {
			t.Fatalf("insert Diamante (minero real): %+v", err)
		}
		dh := consensushashing.BlockHash(dblk)
		st, _ := tc.BlockStatusStore().Get(tc.DatabaseContext(), model.NewStagingArea(), dh)
		if st != externalapi.StatusUTXOValid {
			t.Fatalf("Diamante minado por el block_builder real no es UTXOValid: %s", st)
		}
		tip = dh
		src, srcIdx, disp = dtx, 2, dtx.Outputs[2].Value
	}
	t.Log("EL KING (minado con el block_builder de produccion)...")
	king := forgeUp(constants.LevelKings, rodios)
	// el minero de produccion construye sobre el virtual: el tip debe ser el selected parent
	kingBlk, err := tc.BuildBlock(&externalapi.DomainCoinbaseData{ScriptPublicKey: opTrueSPK, ExtraData: nil}, []*externalapi.DomainTransaction{king})
	if err != nil {
		t.Fatalf("BuildBlock (minero real) del King: %+v", err)
	}
	if err := tc.ValidateAndInsertBlock(kingBlk, true); err != nil {
		t.Fatalf("el validador RECHAZO el bloque del King minado por el block_builder real: %+v", err)
	}
	kingHash := consensushashing.BlockHash(kingBlk)
	st, err := tc.BlockStatusStore().Get(tc.DatabaseContext(), model.NewStagingArea(), kingHash)
	if err != nil {
		t.Fatalf("status King: %+v", err)
	}
	if st != externalapi.StatusUTXOValid {
		t.Fatalf("bloque del King (minero real) no es UTXOValid: %s — el minero sello un conteo que el validador no acepta (H-10?)", st)
	}
	// El bloque que ACEPTA al King (minado tambien por produccion) tiene el conteo con Kings=1
	sealBlk, err := tc.BuildBlock(&externalapi.DomainCoinbaseData{ScriptPublicKey: opTrueSPK}, nil)
	if err != nil {
		t.Fatalf("BuildBlock sello: %+v", err)
	}
	if err := tc.ValidateAndInsertBlock(sealBlk, true); err != nil {
		t.Fatalf("sello del King rechazado: %+v", err)
	}
	sealHash := consensushashing.BlockHash(sealBlk)
	gh := gemsAt(sealHash)
	if gh.Kings != 1 {
		t.Fatalf("Kings en el store tras aceptar al King: %d, esperaba 1", gh.Kings)
	}
	// El header del bloque sello (minado por produccion) DEBE sellar Kings=1, no Kings=0.
	esperado := gemscommitment.CalculateGemsCommitment(gh, gh.Kings)
	if !sealBlk.Header.GemsCommitment().Equal(esperado) {
		t.Fatalf("el minero de produccion sello %s pero el conteo real es %s (Kings=%d)", sealBlk.Header.GemsCommitment(), esperado, gh.Kings)
	}
	sinKing := gh.Clone()
	sinKing.Kings = 0
	if sealBlk.Header.GemsCommitment().Equal(gemscommitment.CalculateGemsCommitment(sinKing, 0)) {
		t.Fatalf("el minero de produccion sella Kings=0: H-10 revertido")
	}
	t.Logf("King minado por block_builder.go, validado por verify_and_build_utxo.go: %s | %+v", sealBlk.Header.GemsCommitment(), gh)
}
