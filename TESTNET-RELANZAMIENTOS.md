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
