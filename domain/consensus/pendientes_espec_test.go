package consensus_test

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/blockheader"
	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/consensus/utils/pow"
	"github.com/rupixnet/rupixd/domain/dagconfig"
)

// Los dos "pendientes" que ESPECIFICACION.md declaraba con nuestra propia letra
// (secciones 3 y 4). Se cierran aqui, el 7-oct-2026.

// TestHijoDeSelloFalsoHeredaDescalificacion (seccion 4): un atacante no puede "limpiar"
// un bloque con sello falso construyendo encima. El builder honesto se niega (eso ya lo
// probaba TestSelloFalsoRechazado); aqui el atacante fabrica A MANO un hijo valido en
// todo salvo su padre (mismo cuerpo que el hijo honesto, padres cambiados al bloque
// falso). Ese hijo entra como encabezado, pero queda StatusDisqualifiedFromChain por
// herencia (resolve_block_status.go), el virtual no lo sigue, y la cadena honesta sigue.
func TestHijoDeSelloFalsoHeredaDescalificacion(t *testing.T) {
	params := dagconfig.DevnetParams
	cfg := consensus.Config{Params: params}
	cfg.SkipProofOfWork = true
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&cfg, "TestHijoDeSelloFalsoHeredaDescalificacion")
	if err != nil {
		t.Fatalf("NewTestConsensus: %+v", err)
	}
	defer teardown(false)
	estado := func(h *externalapi.DomainHash) externalapi.BlockStatus {
		st, err := tc.BlockStatusStore().Get(tc.DatabaseContext(), model.NewStagingArea(), h)
		if err != nil {
			t.Fatalf("BlockStatusStore.Get: %+v", err)
		}
		return st
	}
	conSello := func(b *externalapi.DomainBlock, sello *externalapi.DomainHash) *externalapi.DomainBlock {
		h := b.Header
		return &externalapi.DomainBlock{Header: blockheader.NewImmutableBlockHeader(
			h.Version(), h.Parents(), h.HashMerkleRoot(), h.AcceptedIDMerkleRoot(), h.UTXOCommitment(), sello,
			h.TimeInMilliseconds(), h.Bits(), h.Nonce(), h.DAAScore(), h.BlueScore(), h.BlueWork(), h.PruningPoint()),
			Transactions: b.Transactions}
	}
	conPadres := func(b *externalapi.DomainBlock, padres []externalapi.BlockLevelParents) *externalapi.DomainBlock {
		h := b.Header
		return &externalapi.DomainBlock{Header: blockheader.NewImmutableBlockHeader(
			h.Version(), padres, h.HashMerkleRoot(), h.AcceptedIDMerkleRoot(), h.UTXOCommitment(), h.GemsCommitment(),
			h.TimeInMilliseconds(), h.Bits(), h.Nonce(), h.DAAScore(), h.BlueScore(), h.BlueWork(), h.PruningPoint()),
			Transactions: b.Transactions}
	}

	tip := cfg.GenesisHash
	for i := 0; i < 5; i++ {
		if tip, _, err = tc.AddBlock([]*externalapi.DomainHash{tip}, nil, nil); err != nil {
			t.Fatalf("bloque honesto %d: %+v", i, err)
		}
	}
	honesto, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{tip}, nil, nil)
	if err != nil {
		t.Fatalf("BuildBlockWithParents: %+v", err)
	}
	// Sello falso con el MISMO nivel de bloque que el honesto, para que el hijo fabricado
	// tenga exactamente los mismos padres indirectos y lo unico distinto sea el padre.
	var falso *externalapi.DomainBlock
	for i := byte(1); i < 200; i++ {
		cand := conSello(honesto, externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{0xde, 0xad, i}))
		if pow.BlockLevel(cand.Header, params.MaxBlockLevel) == pow.BlockLevel(honesto.Header, params.MaxBlockLevel) {
			falso = cand
			break
		}
	}
	if falso == nil {
		t.Fatalf("no se encontro un sello falso con el mismo nivel de bloque")
	}
	hashFalso, hashHonesto := consensushashing.BlockHash(falso), consensushashing.BlockHash(honesto)
	if err := tc.ValidateAndInsertBlock(falso, true); err != nil {
		t.Fatalf("el bloque con sello falso debe insertarse (descalificado): %+v", err)
	}
	if err := tc.ValidateAndInsertBlock(honesto, true); err != nil {
		t.Fatalf("el honesto debe entrar: %+v", err)
	}
	if st := estado(hashFalso); st != externalapi.StatusDisqualifiedFromChain {
		t.Fatalf("precondicion: el falso debe estar descalificado, esta %s", st)
	}

	// El hijo honesto, construido por el builder sobre el bloque honesto...
	hijoHonesto, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{hashHonesto}, nil, nil)
	if err != nil {
		t.Fatalf("BuildBlockWithParents(hijo): %+v", err)
	}
	// ...y el mismo hijo con los padres cambiados al falso: eso es lo unico que cambia.
	padres := make([]externalapi.BlockLevelParents, len(hijoHonesto.Header.Parents()))
	for i, nivel := range hijoHonesto.Header.Parents() {
		padres[i] = make(externalapi.BlockLevelParents, len(nivel))
		for j, p := range nivel {
			if p.Equal(hashHonesto) {
				padres[i][j] = hashFalso
			} else {
				padres[i][j] = p
			}
		}
	}
	hijoFalso := conPadres(hijoHonesto, padres)
	hashHijoFalso := consensushashing.BlockHash(hijoFalso)

	// 1) Entra como encabezado (es valido en si mismo)...
	if err := tc.ValidateAndInsertBlock(hijoFalso, true); err != nil {
		t.Fatalf("el hijo fabricado sobre el falso es valido en si mismo y debe guardarse, no fallar: %+v", err)
	}
	// 2) ...pero hereda la descalificacion: nunca sera parte de la cadena.
	if st := estado(hashHijoFalso); st != externalapi.StatusDisqualifiedFromChain {
		t.Fatalf("el hijo de un bloque con sello falso debe heredar StatusDisqualifiedFromChain; quedo %s", st)
	}
	sp, err := tc.GetVirtualSelectedParent()
	if err != nil {
		t.Fatalf("GetVirtualSelectedParent: %+v", err)
	}
	if sp.Equal(hashHijoFalso) || sp.Equal(hashFalso) {
		t.Fatalf("el virtual sigue a la rama del sello falso (%s)", sp)
	}
	// 3) El hijo honesto entra UTXOValid y el virtual lo sigue: la mentira no gano nada.
	if err := tc.ValidateAndInsertBlock(hijoHonesto, true); err != nil {
		t.Fatalf("el hijo honesto debe entrar: %+v", err)
	}
	hashHijoHonesto := consensushashing.BlockHash(hijoHonesto)
	if st := estado(hashHijoHonesto); st != externalapi.StatusUTXOValid {
		t.Fatalf("hijo honesto: esperado UTXOValid, obtenido %s", st)
	}
	if sp, _ = tc.GetVirtualSelectedParent(); !sp.Equal(hashHijoHonesto) {
		t.Fatalf("el virtual debe seguir al hijo honesto, sigue a %s", sp)
	}
	t.Logf("hijo fabricado sobre el sello falso: descalificado por herencia (%s); hijo honesto UTXOValid (%s)", hashHijoFalso, hashHijoHonesto)
}

// TestVenenoQueLlegaSinSerPunta (seccion 3): la pregunta fija del auditor, "¿y si no es
// punta?", en el DAG real. Hay mas puntas que padres caben en un bloque
// (MaxBlockParents = 10), asi que el bloque con la forja 2,100,001 llega y el virtual
// NO lo toma como padre: su transaccion no se evalua al insertarlo. Horas despues
// (aqui: un bloque despues) un minero honesto lo mergea entre sus 10 padres. Ahi se
// evalua la forja por primera vez y debe simplemente no aceptarse: el bloque honesto
// UTXOValid, el conteo en el tope, el Gold del forjador vivo.
func TestVenenoQueLlegaSinSerPunta(t *testing.T) {
	l, teardown := nuevoLabTopes(t, "TestVenenoQueLlegaSinSerPunta")
	defer teardown(false)
	l.minar(60)
	l.sembrar(l.tip, constants.MaxDiamante)
	base := l.tip

	// Mas puntas que padres: 12 hermanos honestos sobre la misma base.
	var hermanos []*externalapi.DomainHash
	for i := 0; i < 12; i++ {
		h, _, err := l.tc.AddBlock([]*externalapi.DomainHash{base}, nil, nil)
		if err != nil {
			t.Fatalf("hermano %d: %+v", i, err)
		}
		hermanos = append(hermanos, h)
	}
	forja := l.forja(l.coinbases[0])
	veneno := l.aMano(base, forja)
	hashVeneno := consensushashing.BlockHash(veneno)
	if err := l.tc.ValidateAndInsertBlock(veneno, true); err != nil {
		t.Fatalf("el bloque con la forja %d es valido en si mismo y debe guardarse: %+v", constants.MaxDiamante+1, err)
	}
	esPadreDelVirtual := func() bool {
		vi, err := l.tc.GetVirtualInfo()
		if err != nil {
			t.Fatalf("GetVirtualInfo: %+v", err)
		}
		for _, p := range vi.ParentHashes {
			if p.Equal(hashVeneno) {
				return true
			}
		}
		return false
	}
	// Si por desempate de hash el virtual lo tomo de padre, se agregan hermanos hasta
	// que lo desplacen: el escenario exige que NO sea punta del virtual.
	for intentos := 0; esPadreDelVirtual(); intentos++ {
		if intentos >= 40 {
			t.Fatalf("no se logro dejar al veneno fuera de los padres del virtual tras 40 hermanos mas")
		}
		h, _, err := l.tc.AddBlock([]*externalapi.DomainHash{base}, nil, nil)
		if err != nil {
			t.Fatalf("hermano extra: %+v", err)
		}
		hermanos = append(hermanos, h)
	}
	if got := l.historia(model.VirtualBlockHash).Diamante; got != constants.MaxDiamante {
		t.Fatalf("el virtual lleva %d sin haber mergeado el veneno; debe estar en %d", got, constants.MaxDiamante)
	}
	if !l.utxoVivo(l.coinbases[0]) {
		t.Fatalf("el Gold de la forja se gasto sin que nadie la evaluara")
	}

	// El minero honesto que llega despues y lo mergea entre sus 10 padres.
	padres := append([]*externalapi.DomainHash{hashVeneno}, hermanos[:9]...)
	honesto, _, err := l.tc.AddBlock(padres, nil, nil)
	if err != nil {
		t.Fatalf("el bloque honesto que mergea el veneno no pudo construirse/insertarse: %+v", err)
	}
	if st := l.estado(honesto); st != externalapi.StatusUTXOValid {
		t.Fatalf("el bloque honesto que mergeo el veneno quedo %s; debe ser UTXOValid", st)
	}
	if l.aceptada(honesto, forja) {
		t.Fatalf("la forja %d fue aceptada al mergearla tarde", constants.MaxDiamante+1)
	}
	if got := l.historia(honesto).Diamante; got != constants.MaxDiamante {
		t.Fatalf("conteo tras mergear tarde: %d, debe seguir en %d", got, constants.MaxDiamante)
	}
	if !l.utxoVivo(l.coinbases[0]) {
		t.Fatalf("el Gold de una forja no aceptada debe seguir vivo")
	}
	sigue, _, err := l.tc.AddBlock([]*externalapi.DomainHash{honesto}, nil, nil)
	if err != nil {
		t.Fatalf("LIVENESS: %+v", err)
	}
	if st := l.estado(sigue); st != externalapi.StatusUTXOValid {
		t.Fatalf("el bloque siguiente quedo %s", st)
	}
	t.Logf("veneno fuera de las puntas del virtual (%d hermanos), mergeado tarde por un honesto: UTXOValid, forja no aceptada, conteo %d", len(hermanos), constants.MaxDiamante)
}
