package consensus_test

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/dagconfig"
)

// TestExcesoEnFronteraDeHalving (hueco #4 de ESPECIFICACION.md): en un DAG, cada bloque
// cobra segun SU PROPIO DAA score, aunque lo mergee un bloque que ya cruzo el halving.
// Dos bloques paralelos minados en el ultimo DAA de una era (49, con halving en 50) se
// cobran los dos con la recompensa vieja; el calendario por DAA solo habria pagado
// una vez la vieja (DAA 49) y una vez la nueva (DAA 50). El exceso es exactamente la
// caida de recompensa en esa frontera. Es lo que se midio en la testnet (0.5 RUPIX en
// el halving 2) y lo que documenta ESPECIFICACION.md; aqui queda afirmado con numeros.
func TestExcesoEnFronteraDeHalving(t *testing.T) {
	params := dagconfig.DevnetParams
	params.BlocksPerHalving = 50
	cfg := consensus.Config{Params: params}
	cfg.SkipProofOfWork = true
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&cfg, "TestExcesoEnFronteraDeHalving")
	if err != nil {
		t.Fatalf("NewTestConsensus: %+v", err)
	}
	defer teardown(false)

	subsidio := func(daa uint64) uint64 {
		return params.DeflationaryPhaseBaseSubsidy >> (daa / params.BlocksPerHalving)
	}
	daaDe := func(h *externalapi.DomainHash) uint64 {
		d, err := tc.DAABlocksStore().DAAScore(tc.DatabaseContext(), model.NewStagingArea(), h)
		if err != nil {
			t.Fatalf("DAAScore: %+v", err)
		}
		return d
	}
	pagaEnCoinbase := func(h *externalapi.DomainHash) uint64 {
		blk, _, err := tc.GetBlock(h)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}
		var total uint64
		for _, o := range blk.Transactions[0].Outputs {
			total += o.Value
		}
		return total
	}

	// Cadena lineal hasta DAA 48.
	tip := cfg.GenesisHash
	for daaDe(tip) < 48 {
		tip, _, err = tc.AddBlock([]*externalapi.DomainHash{tip}, nil, nil)
		if err != nil {
			t.Fatalf("mine: %+v", err)
		}
	}

	// Dos hermanos en DAA 49 (el ultimo de la era 0).
	s1, _, err := tc.AddBlock([]*externalapi.DomainHash{tip}, nil, nil)
	if err != nil {
		t.Fatalf("s1: %+v", err)
	}
	s2, _, err := tc.AddBlock([]*externalapi.DomainHash{tip}, nil, nil)
	if err != nil {
		t.Fatalf("s2: %+v", err)
	}
	if daaDe(s1) != 49 || daaDe(s2) != 49 {
		t.Fatalf("los hermanos deben tener DAA 49, tienen %d y %d", daaDe(s1), daaDe(s2))
	}

	// El bloque que los mergea ya esta en la era 1 (DAA 51): paga a los dos hermanos.
	c, _, err := tc.AddBlock([]*externalapi.DomainHash{s1, s2}, nil, nil)
	if err != nil {
		t.Fatalf("c: %+v", err)
	}
	if daaDe(c) != 51 {
		t.Fatalf("el bloque que mergea debe tener DAA 51, tiene %d", daaDe(c))
	}
	pagado := pagaEnCoinbase(c)
	real := 2 * subsidio(49)                  // cada hermano cobra por SU DAA (era 0)
	calendario := subsidio(49) + subsidio(50) // el calendario solo tiene un DAA 49 y un DAA 50
	if pagado != real {
		t.Fatalf("la coinbase que mergea pago %d rupias; esperado %d (dos recompensas de la era 0)", pagado, real)
	}
	exceso := pagado - calendario
	if exceso != subsidio(49)-subsidio(50) {
		t.Fatalf("exceso %d rupias; esperado exactamente la caida de recompensa %d", exceso, subsidio(49)-subsidio(50))
	}

	// Y el bloque siguiente cobra a C con la recompensa nueva: el exceso no se repite.
	d, _, err := tc.AddBlock([]*externalapi.DomainHash{c}, nil, nil)
	if err != nil {
		t.Fatalf("d: %+v", err)
	}
	if pagaEnCoinbase(d) != subsidio(51) {
		t.Fatalf("el bloque D debe pagar a C con la recompensa de la era 1 (%d), pago %d", subsidio(51), pagaEnCoinbase(d))
	}
	t.Logf("frontera en DAA 50: pagado %d, calendario %d, exceso %d rupias = exactamente la caida de recompensa", pagado, calendario, exceso)
}
