package consensus_test

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/model/testapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/consensus/utils/merkle"
	"github.com/rupixnet/rupixd/domain/consensus/utils/testutils"
	"github.com/rupixnet/rupixd/domain/consensus/utils/txscript"
	"github.com/rupixnet/rupixd/domain/dagconfig"
)

// Hueco #7 de ESPECIFICACION.md (hallazgo del auditor, 30-sep-2026): una forja que
// rompe un tope historico debe quedar NO ACEPTADA, no matar al bloque honesto que la
// mergea. Estos tests se escribieron ANTES del arreglo y fallaban; el commit que los
// pone en verde es el arreglo (utils/topes + conteo corriente en la aceptacion).

type labTopes struct {
	t         *testing.T
	tc        testapi.TestConsensus
	cfg       *consensus.Config
	sigScript []byte
	spk       *externalapi.ScriptPublicKey
	coinbases []*externalapi.DomainTransaction
	tip       *externalapi.DomainHash
}

func nuevoLabTopes(t *testing.T, nombre string) (*labTopes, func(bool)) {
	params := dagconfig.DevnetParams
	params.BlocksPerHalving = 50
	cfg := consensus.Config{Params: params}
	cfg.SkipProofOfWork = true
	cfg.BlockCoinbaseMaturity = 0
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&cfg, nombre)
	if err != nil {
		t.Fatalf("NewTestConsensus: %+v", err)
	}
	spk, redeem := testutils.OpTrueScript()
	sig, err := txscript.PayToScriptHashSignatureScript(redeem, nil)
	if err != nil {
		t.Fatalf("sigScript: %+v", err)
	}
	return &labTopes{t: t, tc: tc, cfg: &cfg, sigScript: sig, spk: spk, tip: cfg.GenesisHash}, teardown
}

// minar agrega n bloques vacios a la punta y guarda sus coinbases.
func (l *labTopes) minar(n int) {
	for i := 0; i < n; i++ {
		h, _, err := l.tc.AddBlock([]*externalapi.DomainHash{l.tip}, nil, nil)
		if err != nil {
			l.t.Fatalf("minar: %+v", err)
		}
		blk, _, err := l.tc.GetBlock(h)
		if err != nil {
			l.t.Fatalf("GetBlock: %+v", err)
		}
		if len(blk.Transactions[0].Outputs) > 0 {
			l.coinbases = append(l.coinbases, blk.Transactions[0])
		}
		l.tip = h
	}
}

// forja arma la tx que quema 10 Gold de una coinbase y crea 1 Diamante (salida 0).
func (l *labTopes) forja(cb *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	diez := uint64(constants.BurnRatio) * uint64(constants.RupiaPerRupix)
	v := cb.Outputs[0].Value
	if v < diez+2*uint64(constants.RupiaPerRupix) {
		l.t.Fatalf("coinbase de %d rupias: no alcanza", v)
	}
	return &externalapi.DomainTransaction{Version: constants.MaxTransactionVersion, Payload: []byte{},
		Inputs: []*externalapi.DomainTransactionInput{{
			PreviousOutpoint: externalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(cb), Index: 0},
			SignatureScript:  l.sigScript, Sequence: constants.MaxTxInSequenceNum}},
		Outputs: []*externalapi.DomainTransactionOutput{
			{Value: constants.GemAmount, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: l.spk.Script, Version: constants.LevelDiamante}},
			{Value: diez, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{txscript.OpReturn}, Version: constants.LevelGold}},
			{Value: v - diez - uint64(constants.RupiaPerRupix), ScriptPublicKey: &externalapi.ScriptPublicKey{Script: l.spk.Script, Version: constants.LevelGold}},
		}}
}

// mueveGema arma la tx que gasta el Diamante recien nacido (salida 0 de la forja) y
// lo pasa a otra salida del mismo nivel, pagando la quema con una coinbase.
func (l *labTopes) mueveGema(forja, cb *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	v := cb.Outputs[0].Value
	return &externalapi.DomainTransaction{Version: constants.MaxTransactionVersion, Payload: []byte{},
		Inputs: []*externalapi.DomainTransactionInput{
			{PreviousOutpoint: externalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(forja), Index: 0}, SignatureScript: l.sigScript, Sequence: constants.MaxTxInSequenceNum},
			{PreviousOutpoint: externalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(cb), Index: 0}, SignatureScript: l.sigScript, Sequence: constants.MaxTxInSequenceNum},
		},
		Outputs: []*externalapi.DomainTransactionOutput{
			{Value: constants.GemAmount, ScriptPublicKey: &externalapi.ScriptPublicKey{Script: l.spk.Script, Version: constants.LevelDiamante}},
			{Value: uint64(constants.RupiaPerRupix), ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{txscript.OpReturn}, Version: constants.LevelGold}},
			{Value: v - 2*uint64(constants.RupiaPerRupix), ScriptPublicKey: &externalapi.ScriptPublicKey{Script: l.spk.Script, Version: constants.LevelGold}},
		}}
}

// sembrar deja el conteo historico de Diamantes del padre y del virtual en n, como si
// la historia ya hubiera llegado ahi. Todo lo demas es real.
func (l *labTopes) sembrar(padre *externalapi.DomainHash, n uint64) {
	dbTx, err := l.tc.DatabaseContext().Begin()
	if err != nil {
		l.t.Fatalf("Begin: %+v", err)
	}
	sa := model.NewStagingArea()
	l.tc.GemsHistoryStore().Stage(sa, padre, &externalapi.GemsHistory{Diamante: n})
	l.tc.GemsHistoryStore().Stage(sa, model.VirtualBlockHash, &externalapi.GemsHistory{Diamante: n})
	if err := sa.Commit(dbTx); err != nil {
		l.t.Fatalf("Commit: %+v", err)
	}
	if err := dbTx.Commit(); err != nil {
		l.t.Fatalf("Commit db: %+v", err)
	}
}

// aMano fabrica, como haria un atacante, un bloque valido sobre padre con las txs
// dadas: bloque vacio del builder + txs + raiz de Merkle recalculada.
func (l *labTopes) aMano(padre *externalapi.DomainHash, txs ...*externalapi.DomainTransaction) *externalapi.DomainBlock {
	blk, _, err := l.tc.BuildBlockWithParents([]*externalapi.DomainHash{padre}, nil, nil)
	if err != nil {
		l.t.Fatalf("BuildBlockWithParents: %+v", err)
	}
	blk.Transactions = append(blk.Transactions, txs...)
	h := blk.Header.ToMutable()
	h.SetHashMerkleRoot(merkle.CalculateHashMerkleRoot(blk.Transactions))
	blk.Header = h.ToImmutable()
	return blk
}

func (l *labTopes) estado(h *externalapi.DomainHash) externalapi.BlockStatus {
	st, err := l.tc.BlockStatusStore().Get(l.tc.DatabaseContext(), model.NewStagingArea(), h)
	if err != nil {
		l.t.Fatalf("BlockStatusStore: %+v", err)
	}
	return st
}

func (l *labTopes) historia(h *externalapi.DomainHash) *externalapi.GemsHistory {
	gh, err := l.tc.GemsHistoryStore().Get(l.tc.DatabaseContext(), model.NewStagingArea(), h)
	if err != nil {
		l.t.Fatalf("GemsHistoryStore: %+v", err)
	}
	return gh
}

// aceptada dice si el bloque h acepto la transaccion tx (en su acceptance data).
func (l *labTopes) aceptada(h *externalapi.DomainHash, tx *externalapi.DomainTransaction) bool {
	ad, err := l.tc.AcceptanceDataStore().Get(l.tc.DatabaseContext(), model.NewStagingArea(), h)
	if err != nil {
		l.t.Fatalf("AcceptanceDataStore: %+v", err)
	}
	id := consensushashing.TransactionID(tx)
	for _, b := range ad {
		for _, ta := range b.TransactionAcceptanceData {
			if consensushashing.TransactionID(ta.Transaction).Equal(id) {
				return ta.IsAccepted
			}
		}
	}
	l.t.Fatalf("la transaccion %s no aparece en la acceptance data de %s", id, h)
	return false
}

// utxoVivo dice si el outpoint (cb, 0) sigue sin gastar en el virtual.
func (l *labTopes) utxoVivo(cb *externalapi.DomainTransaction) bool {
	vi, err := l.tc.GetVirtualInfo()
	if err != nil {
		l.t.Fatalf("GetVirtualInfo: %+v", err)
	}
	utxos, err := l.tc.GetVirtualUTXOs(vi.ParentHashes, nil, 1000000)
	if err != nil {
		l.t.Fatalf("GetVirtualUTXOs: %+v", err)
	}
	id := consensushashing.TransactionID(cb)
	for _, u := range utxos {
		if u.Outpoint.TransactionID.Equal(id) && u.Outpoint.Index == 0 {
			return true
		}
	}
	return false
}

// TestVenenoDeTopeNoMataAlHonesto: escenario 1 del auditor. El atacante mina A con la
// forja 2,100,001 (valida en aislamiento). A se guarda sin ser punta del virtual. El
// bloque honesto B que mergea a A debe quedar UTXOValid, con la forja NO aceptada, el
// Gold del forjador sin quemar, y una tx que dependia de la gema tampoco aceptada.
func TestVenenoDeTopeNoMataAlHonesto(t *testing.T) {
	l, teardown := nuevoLabTopes(t, "TestVenenoDeTopeNoMataAlHonesto")
	defer teardown(false)
	l.minar(60)
	l.sembrar(l.tip, constants.MaxDiamante) // el tope ya esta lleno

	forja := l.forja(l.coinbases[0])
	veneno := l.aMano(l.tip, forja)
	hashVeneno := consensushashing.BlockHash(veneno)
	// Se inserta como cualquier bloque (el virtual lo mergea). Antes del arreglo esto
	// fallaba aqui mismo (ErrGemsCapExceeded en la actualizacion del virtual); si el
	// bloque llegara sin ser punta, el rechazo se movia al bloque honesto que lo
	// mergeara: el mismo bucle de aceptacion, el mismo arreglo.
	if err := l.tc.ValidateAndInsertBlock(veneno, true); err != nil {
		t.Fatalf("VENENO: el bloque con la forja %d es valido en si mismo y debe guardarse: %+v", constants.MaxDiamante+1, err)
	}
	if got := l.historia(model.VirtualBlockHash).Diamante; got != constants.MaxDiamante {
		t.Fatalf("el virtual tras mergear el veneno lleva %d, debe seguir en %d", got, constants.MaxDiamante)
	}

	honesto, _, err := l.tc.AddBlock([]*externalapi.DomainHash{hashVeneno}, nil, nil)
	if err != nil {
		t.Fatalf("VENENO: el bloque honesto que mergea la forja %d no pudo construirse/insertarse: %+v", constants.MaxDiamante+1, err)
	}
	if st := l.estado(honesto); st != externalapi.StatusUTXOValid {
		t.Fatalf("VENENO: el bloque honesto quedo %s; debe ser UTXOValid", st)
	}
	if l.aceptada(honesto, forja) {
		t.Fatalf("la forja %d fue aceptada", constants.MaxDiamante+1)
	}
	if got := l.historia(honesto).Diamante; got != constants.MaxDiamante {
		t.Fatalf("conteo tras el bloque honesto: %d, debe seguir en %d", got, constants.MaxDiamante)
	}
	if !l.utxoVivo(l.coinbases[0]) {
		t.Fatalf("el Gold de una forja no aceptada NO debe gastarse: el forjador debe poder reintentar")
	}

	// La tx que dependia de la gema no nacida (gastarla) solo puede ir en un bloque
	// posterior (encadenar dentro del mismo bloque esta prohibido: ErrChainedTransactions).
	// El atacante la mete a mano: su bloque queda descalificado (gasta un UTXO que nunca
	// existio), el Gold ajeno que usaba para la quema sigue vivo, y un hermano honesto
	// sigue adelante.
	dependiente := l.mueveGema(forja, l.coinbases[1])
	bloqueDep := l.aMano(honesto, dependiente)
	hashDep := consensushashing.BlockHash(bloqueDep)
	if err := l.tc.ValidateAndInsertBlock(bloqueDep, true); err != nil {
		t.Logf("el bloque que gasta la gema no nacida se rechazo al insertar: %v", err)
	} else if st := l.estado(hashDep); st == externalapi.StatusUTXOValid {
		t.Fatalf("un bloque que gasta la gema no nacida quedo UTXOValid")
	} else {
		t.Logf("el bloque que gasta la gema no nacida quedo %s", st)
	}
	if !l.utxoVivo(l.coinbases[1]) {
		t.Fatalf("el Gold usado por la tx dependiente NO debe gastarse")
	}
	sigue, _, err := l.tc.AddBlock([]*externalapi.DomainHash{honesto}, nil, nil)
	if err != nil {
		t.Fatalf("LIVENESS: un bloque honesto tras el veneno no pudo construirse: %+v", err)
	}
	if st := l.estado(sigue); st != externalapi.StatusUTXOValid {
		t.Fatalf("el bloque honesto siguiente quedo %s", st)
	}
	t.Logf("veneno mergeado por un bloque honesto: honesto UTXOValid, forja no aceptada, Gold vivo, conteo %d; la tx dependiente solo dana al atacante", constants.MaxDiamante)
}

// TestDosHermanosUltimaGema: escenario 2 del auditor, sin atacante. Dos mineros
// honestos forjan el ultimo Diamante en bloques hermanos. El bloque que los junta
// debe entrar UTXOValid, aceptar exactamente UNA forja (la que dicta el orden GHOSTDAG)
// y dejar el Gold de la otra sin gastar. Y DOS consensos independientes deben llegar al
// mismo resultado: el bloque construido por uno lo valida el otro con el mismo sello.
func TestDosHermanosUltimaGema(t *testing.T) {
	l1, td1 := nuevoLabTopes(t, "TestDosHermanosUltimaGema1")
	defer td1(false)
	l2, td2 := nuevoLabTopes(t, "TestDosHermanosUltimaGema2")
	defer td2(false)

	l1.minar(60)
	// El nodo 2 recibe exactamente los mismos bloques.
	replicar := func(h *externalapi.DomainHash) {
		blk, _, err := l1.tc.GetBlock(h)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}
		if err := l2.tc.ValidateAndInsertBlock(blk, true); err != nil {
			t.Fatalf("nodo 2 no acepto un bloque del nodo 1: %+v", err)
		}
	}
	cadena, err := l1.tc.GetVirtualSelectedParentChainFromBlock(l1.cfg.GenesisHash)
	if err != nil {
		t.Fatalf("cadena: %+v", err)
	}
	for _, h := range cadena.Added {
		replicar(h)
	}
	l2.coinbases, l2.tip = l1.coinbases, l1.tip
	l1.sembrar(l1.tip, constants.MaxDiamante-1)
	l2.sembrar(l2.tip, constants.MaxDiamante-1)

	s1 := l1.aMano(l1.tip, l1.forja(l1.coinbases[0]))
	s2 := l1.aMano(l1.tip, l1.forja(l1.coinbases[1]))
	h1, h2 := consensushashing.BlockHash(s1), consensushashing.BlockHash(s2)
	for _, l := range []*labTopes{l1, l2} {
		if err := l.tc.ValidateAndInsertBlock(s1, false); err != nil {
			t.Fatalf("hermano 1: %+v", err)
		}
		if err := l.tc.ValidateAndInsertBlock(s2, false); err != nil {
			t.Fatalf("hermano 2: %+v", err)
		}
	}

	// El nodo 1 construye el bloque que junta a los dos.
	c, _, err := l1.tc.AddBlock([]*externalapi.DomainHash{h1, h2}, nil, nil)
	if err != nil {
		t.Fatalf("LIVENESS: el bloque que junta a los dos hermanos no pudo construirse: %+v", err)
	}
	if st := l1.estado(c); st != externalapi.StatusUTXOValid {
		t.Fatalf("el bloque que junta a los hermanos quedo %s", st)
	}
	f1, f2 := s1.Transactions[1], s2.Transactions[1]
	a1, a2 := l1.aceptada(c, f1), l1.aceptada(c, f2)
	if a1 == a2 {
		t.Fatalf("debe aceptarse exactamente una forja; aceptadas: %v, %v", a1, a2)
	}
	if got := l1.historia(c).Diamante; got != constants.MaxDiamante {
		t.Fatalf("conteo tras juntar: %d, esperado %d", got, constants.MaxDiamante)
	}
	perdedor := l1.coinbases[1]
	if a2 {
		perdedor = l1.coinbases[0]
	}
	if !l1.utxoVivo(perdedor) {
		t.Fatalf("el Gold del forjador que perdio la carrera NO debe gastarse")
	}

	// El nodo 2 valida el bloque del nodo 1: mismo sello, mismo veredicto.
	blkC, _, err := l1.tc.GetBlock(c)
	if err != nil {
		t.Fatalf("GetBlock c: %+v", err)
	}
	if err := l2.tc.ValidateAndInsertBlock(blkC, true); err != nil {
		t.Fatalf("el nodo 2 rechazo el bloque del nodo 1: %+v", err)
	}
	if st := l2.estado(c); st != externalapi.StatusUTXOValid {
		t.Fatalf("en el nodo 2 el bloque quedo %s: los dos nodos no aceptaron la misma forja (sello distinto)", st)
	}
	if l2.aceptada(c, f1) != a1 || l2.aceptada(c, f2) != a2 {
		t.Fatalf("los dos nodos aceptaron forjas distintas")
	}
	t.Logf("dos hermanos con la ultima gema: una aceptada (f1=%v f2=%v), nadie muere, dos nodos coinciden", a1, a2)
}
