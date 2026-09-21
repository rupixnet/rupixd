package pow

import (
	"math/bits"
"crypto/rand"
"testing"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
)

// computeRankInt: rango de la matriz con aritmetica ENTERA modulo un primo grande.
// Sin float64, sin eps: determinista en cualquier arquitectura (x86, ARM, ...).
// Candidato a reemplazar computeRank (float64) tras verificar que coinciden.
const rankPrime uint64 = (1 << 61) - 1 // primo de Mersenne

func mulmod(a, b uint64) uint64 {
// a,b < 2^61 -> producto < 2^122: usar 128 bits via math/bits
hi, lo := bits.Mul64(a, b)
// reduccion modulo 2^61-1 usando hi,lo
// x = hi*2^64 + lo; 2^64 = 8 * 2^61 = 8 mod p (porque 2^61 = 1 mod p)
x := (lo & rankPrime) + (lo >> 61) + hi*8
x = (x & rankPrime) + (x >> 61)
if x >= rankPrime {
x -= rankPrime
}
return x
}

func powmod(a, e uint64) uint64 {
r := uint64(1)
a %= rankPrime
for e > 0 {
if e&1 == 1 {
r = mulmod(r, a)
}
a = mulmod(a, a)
e >>= 1
}
return r
}

func (mat *matrix) computeRankInt() int {
var B [64][64]uint64
for i := 0; i < 64; i++ {
for j := 0; j < 64; j++ {
B[i][j] = uint64(mat[i][j]) % rankPrime
}
}
rank := 0
var rowSelected [64]bool
for i := 0; i < 64; i++ {
var j int
for j = 0; j < 64; j++ {
if !rowSelected[j] && B[j][i] != 0 {
break
}
}
if j != 64 {
rank++
rowSelected[j] = true
inv := powmod(B[j][i], rankPrime-2) // inverso modular (Fermat)
for p := i; p < 64; p++ {
B[j][p] = mulmod(B[j][p], inv)
}
for k := 0; k < 64; k++ {
if k != j && B[k][i] != 0 {
f := B[k][i]
for p := i; p < 64; p++ {
// B[k][p] -= f * B[j][p]  (mod p)
sub := mulmod(f, B[j][p])
B[k][p] = (B[k][p] + rankPrime - sub) % rankPrime
}
}
}
}
}
return rank
}

// TestRankIntCoincideConFloat: miles de matrices reales de rupixPRNG. El rango
// entero (determinista) debe coincidir con el rango float64 (el de consenso hoy).
// Si coinciden en todas, el reemplazo no cambia consenso.
func TestRankIntCoincideConFloat(t *testing.T) {
const n = 5000
diff := 0
for i := 0; i < n; i++ {
var b [32]byte
rand.Read(b[:])
h := externalapi.NewDomainHashFromByteArray(&b)
// generar matriz cruda (sin el reintento de rango) para probar tambien las no-64
var mat matrix
g := newRupixPRNG(h)
for r := range mat {
for c := 0; c < 64; c += 16 {
v := g.Uint64()
for s := 0; s < 16; s++ {
mat[r][c+s] = uint16(v >> (4 * s) & 0x0F)
}
}
}
rf, ri := mat.computeRank(), mat.computeRankInt()
if rf != ri {
diff++
if diff <= 3 {
t.Logf("matriz %d: float=%d int=%d", i, rf, ri)
}
}
}
if diff != 0 {
t.Fatalf("%d de %d matrices dan rango distinto entre float64 e int", diff, n)
}
t.Logf("%d matrices: rango float64 == rango entero en todas", n)
}
