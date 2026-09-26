package pow

import (
"crypto/rand"
"testing"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
)

// computeRankInt: rango de la matriz con aritmetica ENTERA modulo un primo grande.
// Sin float64, sin eps: determinista en cualquier arquitectura (x86, ARM, ...).
// Candidato a reemplazar computeRank (float64) tras verificar que coinciden.
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
rf, ri := mat.computeRankFloat(), mat.computeRank()
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
