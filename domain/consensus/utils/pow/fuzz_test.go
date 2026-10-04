package pow

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
)

// FuzzGenerateMatrix (hueco #6 de ESPECIFICACION.md): RupixHeavyHash. Para cualquier
// hash de pre-PoW de 32 bytes:
//  1. generateMatrix no entra en panico (tiene un tope de 64 intentos: si algun hash lo
//     agotara, un bloque con ese pre-PoW tiraria al nodo; el fuzzer busca ese hash).
//  2. La matriz tiene rango 64 por el camino entero (computeRank) y el camino flotante
//     heredado (computeRankFloat) dice lo mismo: las dos implementaciones coinciden.
//  3. Todo es determinista: misma entrada, misma matriz, mismo HeavyHash.
//  4. El HeavyHash del mismo pre-PoW con el mismo hash de entrada no es el hash cero.
//
// Correr: go test -fuzz=FuzzGenerateMatrix -fuzztime=2h ./domain/consensus/utils/pow/
func FuzzGenerateMatrix(f *testing.F) {
	f.Add(make([]byte, 32))
	f.Add([]byte("rupix-heavyhash-semilla-de-prueba"))
	todoUnos := make([]byte, 32)
	for i := range todoUnos {
		todoUnos[i] = 0xFF
	}
	f.Add(todoUnos)

	f.Fuzz(func(t *testing.T, semilla []byte) {
		var arr [externalapi.DomainHashSize]byte
		copy(arr[:], semilla)
		hash := externalapi.NewDomainHashFromByteArray(&arr)

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("PANICO con hash %s: %v", hash, r)
			}
		}()
		m1 := generateMatrix(hash)
		if rk := m1.computeRank(); rk != 64 {
			t.Fatalf("rango entero %d != 64 para %s", rk, hash)
		}
		if rf := m1.computeRankFloat(); rf != 64 {
			t.Fatalf("rango flotante %d != 64 (el entero dio 64) para %s", rf, hash)
		}
		m2 := generateMatrix(hash)
		if *m1 != *m2 {
			t.Fatalf("generateMatrix no es determinista para %s", hash)
		}
		h1 := m1.HeavyHash(hash)
		h2 := m2.HeavyHash(hash)
		if !h1.Equal(h2) {
			t.Fatalf("HeavyHash no es determinista para %s", hash)
		}
		var cero externalapi.DomainHash
		if h1.Equal(&cero) {
			t.Fatalf("HeavyHash cero para %s", hash)
		}
	})
}

// FuzzComputeRank: matrices arbitrarias de nibbles (0..15, el dominio real del PRNG),
// no solo las que genera el PRNG. El rango entero (el de consenso desde v0.6.0) debe
// estar en 0..64, no entrar en panico, y coincidir con el camino flotante heredado,
// que sigue en el repo como segunda opinion. Se construyen tres familias: nibbles
// crudos, casi-identidad con ruido (rango alto), y filas repetidas (rango bajo).
//
// Correr: go test -fuzz=FuzzComputeRank -fuzztime=1h ./domain/consensus/utils/pow/
func FuzzComputeRank(f *testing.F) {
	f.Add(make([]byte, 64*64), uint8(0))
	f.Add(make([]byte, 64*64), uint8(1))
	f.Add(make([]byte, 64*64), uint8(2))
	f.Fuzz(func(t *testing.T, datos []byte, modo uint8) {
		var mat matrix
		for i := 0; i < 64; i++ {
			for j := 0; j < 64; j++ {
				idx := i*64 + j
				var v byte
				if idx < len(datos) {
					v = datos[idx]
				}
				switch modo % 3 {
				case 0:
					mat[i][j] = uint16(v & 0x0F)
				case 1:
					if i == j {
						mat[i][j] = 1 + uint16(v&0x0E)
					} else {
						mat[i][j] = uint16(v & 0x01)
					}
				default:
					// filas repetidas: la fila i copia la fila i%4 -> rango <= 4
					src := (i % 4) * 64
					var w byte
					if src+j < len(datos) {
						w = datos[src+j]
					}
					mat[i][j] = uint16(w & 0x0F)
				}
			}
		}
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("PANICO en computeRank: %v", r)
			}
		}()
		ri := mat.computeRank()
		rf := mat.computeRankFloat()
		if ri < 0 || ri > 64 {
			t.Fatalf("rango entero fuera de 0..64: %d", ri)
		}
		if modo%3 == 2 && ri > 4 {
			t.Fatalf("filas repetidas (4 distintas) y el rango entero dio %d", ri)
		}
		if ri != rf {
			t.Fatalf("rango entero %d != flotante %d (modo %d)", ri, rf, modo%3)
		}
	})
}
