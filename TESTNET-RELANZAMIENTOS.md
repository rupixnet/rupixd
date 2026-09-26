# Registro de relanzamientos de la testnet de Rupix

Cada relanzamiento queda aquí: qué, cuándo y por qué. Verificable contra el
historial de git y los génesis en `domain/dagconfig/genesis.go`. Un relanzamiento
SOLO ocurre por un cambio de consenso (reglas fundamentales que hacen incompatibles
los bloques viejos). Nunca por conveniencia ni para borrar historia.

## Relanzamiento #3 — 15/16 de septiembre de 2026
Motivo: Algoritmo de minado propio (RupixHeavyHash).
Cambió: el generador de la matriz del HeavyHash pasó de xoshiro256++ (Kaspa) a
rupixPRNG (propio, sello RUPIX). Cambio de Proof-of-Work; génesis testnet
re-minado (nonce 0x179b8). Verificable: commits 2f1538560 y 436240dbd. Release v0.5.0.
Se retiró: la testnet anterior (~196k bloques, 4 Diamantes, 1 Platino, primer
forjador externo). Probado y documentado; se reinicia porque el consenso cambió.

## Relanzamiento #2 — 12 de septiembre de 2026
Motivo: Verificación total con commitment en header (el conteo de gemas sellado
en el hash de cada bloque). Cambio de consenso. Halving testnet a 100,000 bloques.
Release v0.4.1 / v0.4.2.

## Relanzamiento #1 — 3-4 de septiembre de 2026
Motivo: Testnet pública inicial (v0.4.0) tras detectar una testnet privada vieja
con código de la era 0.12.22.

---

## Hito: primera poda de la testnet #3 — 22 de septiembre de 2026

🇲🇽 La testnet cruzó ~592,000 bloques (DAA) y el nodo semilla activó el pruning por primera vez: punto de poda en DAA 345,694 (bloque `ace0d8b8…`). `headerCount` 592,089 / `blockCount` 249,048 — los ~343,000 bloques anteriores ya no tienen cuerpo en el nodo.

**Verificado:** el Diamante forjado en DAA ~173,000 (172,000 bloques antes del punto de poda, su bloque ya podado) sigue contado: commitment `780e9027…` intacto, la wallet ve la gema. El conteo histórico sobrevive a la poda del bloque que lo originó. Es la primera prueba en la red real de lo que H-6 y H-9 cerraron.

**Pendiente de máxima importancia:** el próximo nodo externo que sincronice desde cero lo hará por primera vez **desde el punto de poda**, recibiendo el pruning proof con el gemsHistory dentro. Debe llegar a `780e9027…` sin haber visto el bloque del Diamante. Ese será el test de fuego del pruning verificable. Se documentará quién y cuándo.

🇬🇧 The testnet crossed ~592,000 blocks (DAA) and the seed node pruned for the first time: pruning point at DAA 345,694. The Diamond forged at DAA ~173,000 — 172,000 blocks before the pruning point, its block already pruned — is still counted: commitment `780e9027…` intact. The historical count survives pruning of the block that created it. First real-network proof of what H-6/H-9 closed. **Next:** the first external node to sync from scratch will do so from the pruning point via the proof — it must reach `780e9027…` without ever seeing the Diamond's block. That is the real test of verifiable pruning.

---

## Relanzamiento #4 — 26 de septiembre de 2026 (v0.6.0)

🇲🇽 Tres cambios de consenso juntos: (1) dominio keccak propio RupixHeavyHash, (2) rango de matriz con aritmética entera mod 2^61-1 (determinista en toda CPU), (3) fix del doble conteo del minero (el bloque con forja ya no se descalifica). Versión INCOMPATIBLE con la testnet #3: génesis nuevo, cadena desde cero. La testnet #3 se detuvo en DAA ~853,000. Método del auditor seguido: los tres integrados, king-e2e verde sobre el binario final, suite completa en verde como último paso antes de publicar. Génesis no requirió re-minado (el nonce actual cumple el PoW nuevo). Los nodos externos (JC, JP) deben borrar su cadena y sincronizar desde cero.

🇬🇧 Three consensus changes together: own keccak domain (RupixHeavyHash), integer matrix rank (deterministic across CPUs), and the miner double-count fix. INCOMPATIBLE with testnet #3: new genesis, chain from zero. Auditor's method followed: all three integrated, king-e2e green on the final binary, full suite green as the last step. External nodes must wipe and resync.
