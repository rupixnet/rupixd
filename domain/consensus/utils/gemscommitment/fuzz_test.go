package gemscommitment

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
)

// FuzzCalculateGemsCommitment (hueco #6 de ESPECIFICACION.md): el sello de gemas que
// va en cada encabezado. Invariantes:
//  1. Misma historia -> mismo sello (determinista, dos llamadas y dos structs iguales).
//  2. Cambiar UN solo conteo en 1 cambia el sello (ningun conteo se ignora).
//  3. El parametro kingsCount se ignora cuando la historia trae sus propios Kings
//     (fuente unica), y manda cuando la historia es nil.
//  4. Nunca devuelve nil ni el hash cero.
//
// Correr: go test -fuzz=FuzzCalculateGemsCommitment -fuzztime=30m ./domain/consensus/utils/gemscommitment/
func FuzzCalculateGemsCommitment(f *testing.F) {
	f.Add(uint64(0), uint64(0), uint64(0), uint64(0), uint64(0))
	f.Add(uint64(2_100_000), uint64(210_000), uint64(21_000), uint64(2_100), uint64(7))
	f.Add(uint64(1), uint64(0), uint64(0), uint64(0), uint64(0))
	f.Add(^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0))

	f.Fuzz(func(t *testing.T, d, p, r, k, kc uint64) {
		h := &externalapi.GemsHistory{Diamante: d, Platino: p, Rodio: r, Kings: k}
		s1 := CalculateGemsCommitment(h, kc)
		s2 := CalculateGemsCommitment(&externalapi.GemsHistory{Diamante: d, Platino: p, Rodio: r, Kings: k}, kc+1)
		if s1 == nil || s2 == nil {
			t.Fatalf("sello nil")
		}
		if !s1.Equal(s2) {
			t.Fatalf("no determinista, o kingsCount influyo teniendo historia: %s vs %s", s1, s2)
		}
		var cero externalapi.DomainHash
		if s1.Equal(&cero) {
			t.Fatalf("sello cero")
		}
		// 2) cada conteo cuenta.
		variantes := []*externalapi.GemsHistory{
			{Diamante: d + 1, Platino: p, Rodio: r, Kings: k},
			{Diamante: d, Platino: p + 1, Rodio: r, Kings: k},
			{Diamante: d, Platino: p, Rodio: r + 1, Kings: k},
			{Diamante: d, Platino: p, Rodio: r, Kings: k + 1},
		}
		for i, v := range variantes {
			if CalculateGemsCommitment(v, kc).Equal(s1) {
				t.Fatalf("cambiar el conteo %d en 1 no cambio el sello (%+v)", i, *h)
			}
		}
		// 3) sin historia manda kingsCount.
		n1 := CalculateGemsCommitment(nil, kc)
		n2 := CalculateGemsCommitment(nil, kc+1)
		if n1.Equal(n2) {
			t.Fatalf("sin historia, kingsCount no influye")
		}
		if !n1.Equal(CalculateGemsCommitment(&externalapi.GemsHistory{Kings: kc}, 0)) {
			t.Fatalf("nil+kingsCount debe equivaler a historia {0,0,0,kingsCount}")
		}
	})
}
