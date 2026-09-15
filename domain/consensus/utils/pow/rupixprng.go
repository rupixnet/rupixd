package pow

import (
"encoding/binary"
"math/bits"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
)

// rupixSeal es el sello de Rupix mezclado en la semilla: los bytes de
// "RUPIX" seguidos de una constante impar. Hace que el estado inicial del
// generador no coincida con ningun xoshiro sembrado desde el mismo hash.
const rupixSeal uint64 = 0x5255504958A5C3E1

// rupixMul es una constante impar grande para la mezcla multiplicativa.
// La multiplicacion (no-lineal) es lo que estructuralmente separa este PRNG
// de xoshiro256++ (que solo usa sumas, XOR y rotaciones).
const rupixMul uint64 = 0x9E3779B97F4A7C15

// rupixPRNG es el generador propio de Rupix para la matriz de HeavyHash.
// Mismo estado de 256 bits sembrado desde el hash del bloque, pero con una
// formula estructuralmente distinta a xoshiro256++: suma cruzada s1+s2,
// mezcla multiplicativa, rotaciones 29/11/37 y shift 19. Los ASICs de Kaspa
// (xoshiro en silicio) no pueden producir esta secuencia.
type rupixPRNG struct {
s0 uint64
s1 uint64
s2 uint64
s3 uint64
}

func newRupixPRNG(hash *externalapi.DomainHash) *rupixPRNG {
hashArray := hash.ByteArray()
return &rupixPRNG{
s0: binary.LittleEndian.Uint64(hashArray[:8]) ^ rupixSeal,
s1: binary.LittleEndian.Uint64(hashArray[8:16]) ^ bits.RotateLeft64(rupixSeal, 17),
s2: binary.LittleEndian.Uint64(hashArray[16:24]) ^ bits.RotateLeft64(rupixSeal, 34),
s3: binary.LittleEndian.Uint64(hashArray[24:32]) ^ bits.RotateLeft64(rupixSeal, 51),
}
}

// Uint64 avanza el generador y devuelve 64 bits. Sale con rango completo
// (0..2^64-1); generateMatrix toma nibbles (& 0x0F), asi que ningun valor
// de la matriz pasa de 15 y no hay desbordamiento en uint16.
func (r *rupixPRNG) Uint64() uint64 {
// salida: suma cruzada + mezcla multiplicativa + rotacion
res := bits.RotateLeft64((r.s1+r.s2)*rupixMul, 29) + r.s0

t := r.s1 << 19
r.s2 ^= r.s0
r.s3 ^= r.s1
r.s1 ^= r.s2
r.s0 ^= r.s3

r.s2 ^= t
r.s3 = bits.RotateLeft64(r.s3, 37)
r.s0 = bits.RotateLeft64(r.s0, 11) ^ r.s2
return res
}
