# Lo que Rupix logra hoy / What Rupix does today

🇲🇽 Cada punto de esta lista está probado y es verificable contra el código y la cadena. No es una promesa; es lo que funciona hoy. Actualizado: 23 de septiembre de 2026.

🇬🇧 Every item here is tested and verifiable against the code and the chain. Not a promise; what works today. Updated: 23 September 2026.

## Base / Foundation
- 🇲🇽 Fork de kaspad (Go): BlockDAG con GHOSTDAG, Proof of Work, pruning y red P2P heredados de Kaspa, con crédito. Cadena independiente, moneda propia, no usa KAS. / 🇬🇧 Fork of kaspad (Go): BlockDAG with GHOSTDAG, PoW, pruning and P2P from Kaspa, credited. Independent chain, own coin, does not use KAS.
- 🇲🇽 Cero premine, verificable byte a byte en el génesis (subsidio 0x00). / 🇬🇧 Zero premine, verifiable byte by byte in genesis.
- 🇲🇽 Techo de 42,000,000 RUPIX sellado en el protocolo. / 🇬🇧 42,000,000 RUPIX cap sealed in the protocol.

## La escalera de gemas / The gem ladder
- 🇲🇽 5 niveles: Gold (minado) → Diamante → Platino → Rodio → Kings. Cada nivel se forja quemando 10 del anterior. Diamante quema 10 RUPIX en OpReturn; los niveles altos consumen 10 piezas sin devolverlas. Techos sobre lo nacido: 2.1M / 210k / 21k / 2,100. Cada nivel se desbloquea en un halving. / 🇬🇧 5 tiers, each forged by burning 10 of the tier below. Caps on born: 2.1M / 210k / 21k / 2,100. Each unlocks at a halving.
- 🇲🇽 PROBADO DE PUNTA A PUNTA: 1000 Diamantes → 100 Platinos → 10 Rodios → 1 King forjados con el minero real (block_builder.go) y validados por el validador real (verify_and_build_utxo.go). TestKingsEndToEnd. / 🇬🇧 PROVEN END-TO-END: full ladder up to a King, forged by the real miner and validated by the real validator.
- 🇲🇽 Quema por transacción: cada tx destruye rupias para siempre (OpReturn). El suministro solo baja. / 🇬🇧 Per-transaction burn: supply only goes down.

## El sello del conteo / The count commitment
- 🇲🇽 El conteo de gemas nacidas va sellado (hash) en el header de cada bloque, protegido por PoW. Un conteo falso hace que el bloque se rechace: ni el creador puede inventar gemas. / 🇬🇧 The count of born gems is committed (hashed) into every block header under PoW. A false count gets the block rejected.
- 🇲🇽 Defensa en profundidad: todos los nodos deben coincidir en el conteo o el bloque es inválido, así una divergencia entre implementaciones se detecta en el acto. / 🇬🇧 Defense in depth: nodes must agree on the count or the block is invalid.

## Pruning verificable / Verifiable pruning
- 🇲🇽 PROBADO EN LA RED REAL (22-sep): la testnet podó por primera vez (punto en DAA 345,694). Un Diamante forjado 172,000 bloques antes, con su bloque ya podado, sigue contado: el conteo histórico sobrevive a la poda y viaja en el pruning proof. / 🇬🇧 PROVEN ON THE REAL NETWORK: first pruning; a gem count survives pruning of the block that created it, carried in the pruning proof.

## Algoritmo de minado / Mining algorithm
- 🇲🇽 RupixHeavyHash: motor kHeavyHash de Kaspa con generador de matriz propio (rupixPRNG) y dominio keccak propio ("RupixHeavyHash"). Los ASIC de Kaspa quedan fuera. Arranque justo con GPU/CPU. Alcance honesto: ventaja de meses, no independencia permanente. / 🇬🇧 RupixHeavyHash: Kaspa's engine, own matrix PRNG and keccak domain. Kaspa ASICs locked out. Months of head start, not permanent independence.
- 🇲🇽 Rango de la matriz con aritmética entera (mod 2^61-1): determinista en x86 y ARM, sin riesgo de que dos CPUs difieran y partan la cadena. / 🇬🇧 Matrix rank via integer arithmetic: deterministic across CPUs.

## Defensa y operación / Defense and operations
- 🇲🇽 Checkpoints temporales con caducidad dentro del consenso, probados en devnet. Defensa declarada contra el 51% mientras el hashrate crece. / 🇬🇧 Temporary checkpoints with in-consensus expiry, tested.
- 🇲🇽 Nodo semilla como servicios systemd con reinicio automático (self-healing). / 🇬🇧 Seed node as self-healing systemd services.
- 🇲🇽 go test ./... en verde, go vet limpio, binarios verificables con SHA256, 0 vulnerabilidades. / 🇬🇧 Full test suite green, go vet clean, SHA256-verifiable binaries, 0 vulnerabilities.

## Honestidad / Honesty
- 🇲🇽 Cinco rondas de revisión externa, todos los hallazgos publicados. Bugs documentados en público el día que se encuentran. Una versión marcada como defectuosa cuando lo fue. Ver THANKS.md, CHANGELOG.md, rupix_traspaso_auditoria.md. / 🇬🇧 Five rounds of external review, all findings published. Bugs documented publicly the day they are found. See THANKS.md.

*No confíes, verifica. / Don't trust, verify.* — ER

## 26 de septiembre de 2026 — Primer nodo externo v0.6.0 sincronizado (JC)

JC ("Coroking") corrió `rupix-v0.6.0-win64` en Windows y sincronizó al 100%.
Verificado por RPC: `serverVersion: 0.6.0`, `isSynced: true`, `blockCount: 48,645`,
`pruningPointHash: d2561df4…` presente.

**Qué prueba:** un segundo nodo independiente, en otro sistema operativo, validó
toda la cadena v0.6.0 sin rechazar un bloque → determinismo de validación cross-OS
confirmado (la razón exacta del cambio a rango entero mod 2^61-1). Su nodo carga
pruning point → cruzó la poda durante la sincronización.

**Pendiente de cierre (cuando JC vuelva):** reconciliación de tip en vivo
(GetBlockDagInfo simultáneo seed↔JC), y como HITO aparte, que JC mine un bloque
(determinismo de producción, no solo de validación).

## 26 de septiembre de 2026 (noche) — Primer minero externo en v0.6.0: JC

JC creó su wallet a las 10:21 y minó todo el día desde su PC con Windows (~50 KH/s).
A las 19:45, **14,187 bloques suyos habían recibido recompensa**: 7,093.5 RUPIX.
Verificado desde el seed, no solo desde su wallet: `GetBalanceByAddress` da 7,111 RUPIX en su dirección (siguió minando).

**Qué prueba:**
- Su minero produce bloques que el consenso acepta. El determinismo va en las dos direcciones: validar y producir.
- Cerca del 40% de los bloques desde la mañana fueron suyos. La dificultad pasó de 37,724 a 67,204: ese salto fue su hashrate.
- Reconciliación en vivo cerrada: su tip (`ed6d61e2…`) está en el seed como chain block, y los dos nodos tienen el mismo pruning point (`d2561df4…`).
- Su nodo acepta conexiones entrantes: puede servir la cadena a otros.

La red ya no es solo del servidor.
