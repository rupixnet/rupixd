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

## 27 de septiembre de 2026 (~00:18) — Halving 1 de la testnet #4: el Diamante se abre

La testnet #4 cruzó el DAA 100,000. El explorer pasó el Diamante a "✅ desbloqueado" y la recompensa bajó de 0.5 a 0.25 RUPIX por bloque.

**Se puede verificar a mano:** a DAA 100,209 el Gold emitido era **50,052.5 RUPIX** = 100,000 bloques × 0.5 + 210 bloques × 0.25. La emisión siguió exacto el calendario escrito desde el génesis. Nadie lo decidió esa noche.

## 27 de septiembre de 2026 (~22:58, hora de México) — Primer Diamante de v0.6.0 en la red pública, forjado por JC

JC, desde su PC con Windows, quemó 10 Gold y forjó el primer Diamante de la testnet #4. No lo forjó el fundador: lo forjó un minero de la comunidad, con Gold que él mismo minó.

**Se puede verificar a mano:** la gema es la salida de la tx `8653650fc729d4cef85c3fe11ee4c67b0dfcc7e038e973c7d93d62944f8b3a4b`, que entró en el bloque con DAA **180,710**. Desde el seed (no desde la wallet de JC) se pidieron al nodo todas las salidas de su dirección: 33,201 de Gold y 1 de nivel 1, el Diamante.

## 28 de septiembre de 2026 (madrugada) — Pruebas en vivo con un nodo de la comunidad

Con JC conectado desde su PC, se probó la red de punta a punta. Todo verificado desde el seed, no desde las wallets:

- **Misma historia en dos nodos independientes:** el punto de poda (`pruningPointHash`) del nodo de JC y el del seed son idénticos (`d2561df4…3cf9`).
- **Una gema viajó entre dos wallets distintas y volvió:** JC mandó un Diamante al seed y el seed se lo devolvió (tx `e721d29b…`). En la cadena, el Diamante que salió de JC ya no existe y el que regresó sí.
- **La red rechazó un Platino antes de tiempo:** JC intentó forjar un Platino en el DAA 185,738, antes de que se abra en el 200,000. Su propio nodo lo rechazó al recibirlo: `nivel 2 bloqueado: se desbloquea en DAA score 200000 (actual: 185738)` (tx `a51bdc64…`). La wallet no revisa esa regla; la aplica el consenso. Ni el que tiene los Diamantes puede adelantarse.
- **JC tiene 11 Diamantes en la cadena**, todos forjados con Gold que minó él.

## 28 de septiembre de 2026 (noche) — Halving 2 verificado y primer Platino de v0.6.0

La testnet #4 cruzó el DAA 200,000 en la madrugada: la recompensa bajó de 0.25 a 0.125 RUPIX por bloque y se abrió el Platino.

**La emisión, verificada con enteros exactos** a DAA 256,208: la regla da 99,999 × 0.5 + 100,000 × 0.25 + 56,208 × 0.125 = **82,025.5 RUPIX**. En el nodo existían 81,915.87. La diferencia (109.63) es lo quemado: los 110 Gold de los 11 Diamantes de JC más la quema de cada transacción. Nunca hay Gold de más. La herramienta que lo calcula (`tools/verificar-emision.py`) queda en el repo para que cualquiera la corra contra su propio nodo.

**El primer Platino de v0.6.0** se forjó desde el seed: 10 Diamantes quemados, 1 Platino nacido (tx `c382e0a751e10003cc9692bfa2678723c12e965110b0793da6df24e3c942a81e`, DAA 256,577). Verificado en el nodo: la dirección quedó con una salida de nivel 2 y ninguna de nivel 1, y lo quemado subió exactamente 100 Gold. La noche anterior la red había rechazado un Platino en el DAA 185,738; pasado el halving, lo aceptó. La regla funciona en los dos sentidos.
