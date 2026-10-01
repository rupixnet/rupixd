package consensus_test

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/blockheader"
	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/dagconfig"
)

// TestSelloFalsoRechazado (hueco #1 de ESPECIFICACION.md): un bloque valido en todo
// salvo el sello de gemas del encabezado (GemsCommitment) NO puede ser parte de la
// cadena. Es la afirmacion mas fuerte del README ("un conteo falso se rechaza") y
// hasta hoy solo se habia probado en vivo, no en la suite.
//
// Como funciona el rechazo en un DAG: el bloque con sello falso no se descarta como
// encabezado (su PoW y su estructura son validos), pero al verificar su UTXO el sello
// no cuadra (ErrBadUTXOCommitment en verify_and_build_utxo.go) y queda
// StatusDisqualifiedFromChain: nunca sera el selected parent del virtual, y todo
// bloque que lo tome como selected parent hereda la descalificacion. El mismo bloque
// con el sello correcto queda StatusUTXOValid.
func TestSelloFalsoRechazado(t *testing.T) {
	params := dagconfig.DevnetParams
	cfg := consensus.Config{Params: params}
	cfg.SkipProofOfWork = true
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&cfg, "TestSelloFalsoRechazado")
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

	// Unos bloques honestos para tener historia.
	tip := cfg.GenesisHash
	for i := 0; i < 5; i++ {
		tip, _, err = tc.AddBlock([]*externalapi.DomainHash{tip}, nil, nil)
		if err != nil {
			t.Fatalf("bloque honesto %d: %+v", i, err)
		}
	}

	// El bloque honesto siguiente, construido por el block builder real.
	honesto, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{tip}, nil, nil)
	if err != nil {
		t.Fatalf("BuildBlockWithParents: %+v", err)
	}
	h := honesto.Header

	// El mismo bloque con UN cambio: el sello de gemas apunta a otro conteo.
	selloFalso := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{0xde, 0xad, 0xbe, 0xef})
	if selloFalso.Equal(h.GemsCommitment()) {
		t.Fatalf("el sello falso coincide con el real; elige otro")
	}
	falso := &externalapi.DomainBlock{
		Header: blockheader.NewImmutableBlockHeader(
			h.Version(), h.Parents(), h.HashMerkleRoot(), h.AcceptedIDMerkleRoot(),
			h.UTXOCommitment(), selloFalso,
			h.TimeInMilliseconds(), h.Bits(), h.Nonce(), h.DAAScore(), h.BlueScore(), h.BlueWork(), h.PruningPoint()),
		Transactions: honesto.Transactions,
	}
	hashFalso := consensushashing.BlockHash(falso)
	hashHonesto := consensushashing.BlockHash(honesto)
	if hashFalso.Equal(hashHonesto) {
		t.Fatalf("cambiar el sello debe cambiar el hash del bloque")
	}

	// 1) El bloque con sello falso entra como encabezado pero queda descalificado.
	if err := tc.ValidateAndInsertBlock(falso, true); err != nil {
		t.Fatalf("el bloque con sello falso debe insertarse (descalificado), no fallar: %+v", err)
	}
	if st := estado(hashFalso); st != externalapi.StatusDisqualifiedFromChain {
		t.Fatalf("bloque con sello falso: esperado StatusDisqualifiedFromChain, obtenido %s", st)
	}
	sp, err := tc.GetVirtualSelectedParent()
	if err != nil {
		t.Fatalf("GetVirtualSelectedParent: %+v", err)
	}
	if sp.Equal(hashFalso) {
		t.Fatalf("el virtual NO debe seguir a un bloque con sello falso")
	}

	// 2) El mismo bloque con el sello correcto es UTXOValid y el virtual lo sigue.
	if err := tc.ValidateAndInsertBlock(honesto, true); err != nil {
		t.Fatalf("el bloque honesto debe entrar: %+v", err)
	}
	if st := estado(hashHonesto); st != externalapi.StatusUTXOValid {
		t.Fatalf("bloque honesto: esperado StatusUTXOValid, obtenido %s", st)
	}
	sp, err = tc.GetVirtualSelectedParent()
	if err != nil {
		t.Fatalf("GetVirtualSelectedParent: %+v", err)
	}
	if !sp.Equal(hashHonesto) {
		t.Fatalf("el virtual debe seguir al bloque honesto, sigue a %s", sp)
	}

	// 3) Un minero honesto NI SIQUIERA puede construir sobre el falso: el block builder
	// se niega ("selected parent ... DisqualifiedFromChain").
	if _, _, err := tc.AddBlock([]*externalapi.DomainHash{hashFalso}, nil, nil); err == nil {
		t.Fatalf("el block builder no debe construir sobre un bloque descalificado")
	}

	// (La herencia de la descalificacion por un hijo fabricado a mano la aplica
	// resolve_block_status.go: si el selected parent esta descalificado, el hijo tambien.
	// Fabricar un hijo valido en todo salvo su padre requiere mas maquinaria; queda
	// como refinamiento en ESPECIFICACION.md.)
	t.Logf("sello falso descalificado (%s), honesto UTXOValid (%s); el builder honesto no construye sobre el falso", hashFalso, hashHonesto)
}
