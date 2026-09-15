package pow

import (
"testing"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
)

// TestRupixPRNGDistintoDeXoshiro: la semilla propia DEBE producir una
// secuencia distinta a xoshiro256++ desde el mismo hash. Si coincidiera,
// los ASICs de Kaspa seguirian sirviendo.
func TestRupixPRNGDistintoDeXoshiro(t *testing.T) {
hash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
})
r := newRupixPRNG(hash)
x := newxoShiRo256PlusPlus(hash)
iguales := 0
for i := 0; i < 64; i++ {
if r.Uint64() == x.Uint64() {
iguales++
}
}
if iguales > 0 {
t.Fatalf("rupixPRNG coincide con xoshiro en %d de 64 salidas: no es un PRNG distinto", iguales)
}
}

// TestRupixPRNGDeterminista: mismo hash -> misma secuencia, siempre.
// Sin esto, el PoW no seria verificable.
func TestRupixPRNGDeterminista(t *testing.T) {
hash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{42})
a := newRupixPRNG(hash)
b := newRupixPRNG(hash)
for i := 0; i < 100; i++ {
if a.Uint64() != b.Uint64() {
t.Fatalf("rupixPRNG no es determinista en la iteracion %d", i)
}
}
}

// TestRupixMatrizRango64: la matriz generada con rupixPRNG debe alcanzar
// rango 64 (generateMatrix reintenta hasta lograrlo). Si nunca lo lograra,
// el minado se colgaria. Probamos varios hashes.
func TestRupixMatrizRango64(t *testing.T) {
for seed := byte(0); seed < 8; seed++ {
hash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{seed, seed + 1, seed + 2})
mat := generateMatrix(hash)
if mat.computeRank() != 64 {
t.Fatalf("matriz con hash semilla %d no alcanzo rango 64", seed)
}
// ningun valor pasa de 15 (nibble): sin desbordamiento uint16
for i := 0; i < 64; i++ {
for j := 0; j < 64; j++ {
if mat[i][j] > 15 {
t.Fatalf("valor de matriz fuera de rango nibble: %d", mat[i][j])
}
}
}
}
}

// TestRupixHeavyHashCambia: el HeavyHash con rupixPRNG debe ser distinto
// al que daria xoshiro para el mismo header. Es la prueba de que los
// bloques viejos no validan y el hardware heredado no sirve.
func TestRupixHeavyHashCambia(t *testing.T) {
	hash := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{0x8f, 0x2a, 0xc4, 0x91, 0x17, 0xe3, 0x5b, 0xd8, 0x6c, 0x39, 0xa2, 0x7e, 0x50, 0xbd, 0x04, 0xf6, 0x1d, 0x9a, 0x63, 0xc7, 0x2b, 0xe8, 0x45, 0x0f, 0xb1, 0x76, 0xd2, 0x38, 0x9c, 0x51, 0xea, 0x07})
matRupix := generateMatrix(hash)
// matriz "como xoshiro" para comparar
var matXo matrix
gen := newxoShiRo256PlusPlus(hash)
for i := range matXo {
for j := 0; j < 64; j += 16 {
val := gen.Uint64()
for shift := 0; shift < 16; shift++ {
matXo[i][j+shift] = uint16(val >> (4 * shift) & 0x0F)
}
}
}
if matRupix.HeavyHash(hash).Equal(matXo.HeavyHash(hash)) {
t.Fatal("HeavyHash con rupixPRNG dio igual que con xoshiro: el cambio no tuvo efecto")
}
}
