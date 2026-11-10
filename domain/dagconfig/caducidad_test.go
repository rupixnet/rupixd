package dagconfig

import "testing"

// TestEstadoCheckpoints (Rupix, 10-nov-2026): aviso con menos de 450,000 DAA de margen,
// "caducada" al llegar a la caducidad, y silencio en redes sin checkpoints.
func TestEstadoCheckpoints(t *testing.T) {
	p := Params{Checkpoints: []Checkpoint{{BlueScore: 1}}, CheckpointsExpireDAAScore: 3_900_000}
	casos := []struct {
		daa              uint64
		caducada, pronto bool
		margen           uint64
	}{
		{3_000_000, false, false, 900_000},
		{3_450_000, false, false, 450_000}, // justo en el margen: todavia no
		{3_450_001, false, true, 449_999},
		{3_895_742, false, true, 4_258}, // el DAA real de la noche del 9-nov
		{3_900_000, true, false, 0},
		{4_500_000, true, false, 0},
	}
	for _, c := range casos {
		caducada, pronto, margen := p.EstadoCheckpoints(c.daa)
		if caducada != c.caducada || pronto != c.pronto || margen != c.margen {
			t.Errorf("DAA %d: (%v,%v,%d) quiero (%v,%v,%d)", c.daa, caducada, pronto, margen, c.caducada, c.pronto, c.margen)
		}
	}
	sin := Params{}
	if a, b, m := sin.EstadoCheckpoints(9_999_999); a || b || m != 0 {
		t.Errorf("sin checkpoints no hay aviso: %v %v %d", a, b, m)
	}
	// La testnet real: hoy caducada.
	if a, _, _ := TestnetParams.EstadoCheckpoints(3_900_000); !a {
		t.Errorf("TestnetParams debe reportar la lista caducada en su CheckpointsExpireDAAScore")
	}
}
