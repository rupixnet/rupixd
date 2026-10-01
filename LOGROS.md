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

**La emisión, verificada con enteros exactos** a DAA 256,208: la regla da 99,999 × 0.5 + 100,000 × 0.25 + 56,208 × 0.125 = **82,025.5 RUPIX**. En el nodo existían 81,915.87. La diferencia (109.63) es lo quemado: los 110 Gold de los 11 Diamantes de JC más la quema de cada transacción. *(Corrección del 29-sep: había 0.5 RUPIX más que la regla por calendario, no menos; se explica en MEMORIA, "El medio RUPIX del halving 2". La frase "nunca hay Gold de más" era demasiado fuerte: hay un exceso acotado en cada frontera de halving.)* La herramienta que lo calcula (`tools/verificar-emision.py`) queda en el repo para que cualquiera la corra contra su propio nodo.

**El primer Platino de v0.6.0** se forjó desde el seed: 10 Diamantes quemados, 1 Platino nacido (tx `c382e0a751e10003cc9692bfa2678723c12e965110b0793da6df24e3c942a81e`, DAA 256,577). Verificado en el nodo: la dirección quedó con una salida de nivel 2 y ninguna de nivel 1, y lo quemado subió exactamente 100 Gold. La noche anterior la red había rechazado un Platino en el DAA 185,738; pasado el halving, lo aceptó. La regla funciona en los dos sentidos.

## 29 de septiembre de 2026 — Halving 3, envíos de 1,000 y 10,000 RUPIX, y el primer Platino de la comunidad

**Halving 3.** La testnet #4 cruzó el DAA 300,000 en la mañana: la recompensa bajó de 0.125 a 0.0625 RUPIX por bloque y se abrió el Rodio. Los dos nodos (seed y JC) reportaron el mismo `pruningPointHash` (`7e2ece39…`) antes y después: no hubo bifurcación.

**Los envíos grandes, probados de verdad.** Con la wallet de la rama `wallet-envios-grandes` (daemon de prueba, copia de llaves) se le mandaron a JC desde la wallet del minero del seed, la de ~78,000 pedazos de 0.25 y 0.5:
- **1,000 RUPIX:** 24 transacciones, 8.6 segundos.
- **10,000 RUPIX:** 232 transacciones (231 de consolidación + 1 de pago), 46.1 segundos.
Antes del arreglo, 1,000 se cortaba a los 2 minutos sin salir; 10,000 no era posible. Verificado desde el seed: la dirección de JC pasó a 13,305.75 RUPIX.

**El primer Platino de la comunidad.** JC, desde su PC con Windows y su propio nodo, forjó un Platino quemando 10 Diamantes que él mismo había forjado con Gold que él minó: tx `bc1bb97590ef7d81fb7c7b3c79cdce62f219b7e11ef7d4376ba3d41b650f563a`, nacido en el bloque DAA **348,162**. Es la misma operación que la red le rechazó el 28-sep en el DAA 185,738 (`nivel 2 bloqueado`); pasado el halving 2, la aceptó.

**Dos nodos independientes, la misma respuesta.** Cada uno consultó su propio nodo con `rupixctl GetUtxosByAddresses` sobre la dirección de JC, sin copiarse:

```
seed                                            JC (Windows)
42ccd70e52bf716f idx 0 nivel 1 DAA 181403       42ccd70e52bf716f idx 0 nivel 1 DAA 181403
926bf4a0711f639d idx 0 nivel 1 DAA 185382       926bf4a0711f639d idx 0 nivel 1 DAA 185382
bc1bb97590ef7d81 idx 0 nivel 2 DAA 348162       bc1bb97590ef7d81 idx 0 nivel 2 DAA 348162
```

Mismas tres gemas, mismos IDs, mismos bloques de nacimiento, mismo punto de poda. Los 10 Diamantes quemados no aparecen en ningún nodo. Nadie tuvo que confiar en nadie.

## 29 de septiembre de 2026 (noche) — Visto bueno del auditor, v0.6.1 y el primer checkpoint

El auditor externo revisó las cuatro ramas contra `1e38bace` y dio el visto bueno para mergearlas: checkpoints para DAG, prueba automática de los niveles en el mempool, selección lineal de pedazos en la wallet y claves de la wallet (contraseña en pantalla, frase semilla en papel). Confirmó además, con el código en la mano, la explicación del medio RUPIX del halving 2 (ver MEMORIA) y recomendó no tocar el consenso por eso.

Merge sin conflictos, suite en verde, y **v0.6.1** publicada. Con ella, el **primer checkpoint de Rupix**: testnet #4, blue score 86,400 (DAA 86,399), hash `7e2ece393c7d991c86e7ba915276cd85b5fc19e8647d5d197fa26bf116604fad`, caduca en DAA 2,000,000. Es el punto de poda que el seed y el nodo de JC ya tenían en común antes de que nadie lo declarara: el checkpoint no impone una historia, confirma la que los nodos ya compartían. `TestCheckpointsPublicados` obliga a que el código y `CHECKPOINTS.md` digan lo mismo.

## 30 de septiembre de 2026 (madrugada) — Primer CI en verde y compilación reproducible

**El CI dice la verdad por primera vez.** El workflow `Tests` heredado de Kaspa llevaba semanas en rojo en cada push, por piezas que no eran nuestras, y debajo escondía dos avisos reales de `staticcheck` (uno, un bug chico en la estimación de comisión de la wallet). Se retiró a `.github/workflows-retirados/` y se reemplazó por lo que sí es de Rupix: gofmt, go vet, staticcheck, build y `go test ./...` en Linux y macOS. Primer verde: commit `bb1db440`, los dos sistemas. Desde hoy, rojo significa que algo nuestro se rompió y verde significa que no.

**Los binarios publicados son exactamente el código.** Se bajó `rupix-v0.6.1-linux.zip` de la release (SHA256 `b534d843f2295b0f396bf81d151430749f6e890abd8996a3db74bbbf44217f67`), se compiló el tag `v0.6.1` (`a12c2a05`) en el seed con los mismos flags que el CI y Go 1.26.6, y se comparó binario por binario:

| Binario | SHA256 (idéntico en CI y en el seed) |
|---|---|
| rupixd | `ea973ce0d720678a431093e1794fe3401c6186647f747153b49b0f41259fd82a` |
| rupixctl | `033f60a457d5e0bd827a5202a2f557d91acb369adac1614c1111c36a16383c3d` |
| rupixwallet | `8f1f013816911f80e7de5d5febdb41c457325c852c3a9e1c611dbdb3e47e7b34` |
| rupixminer | `d3a40158d20e0511e572774ff052a2fdddb00bf3d6518fb310965038bad11ad8` |

Cuatro de cuatro. Cualquiera con Go 1.26.6 puede repetirlo con `tools/verificar-binarios.sh v0.6.1`. Eso cierra un pendiente viejo: ya no hay que confiar en GitHub ni en nosotros para saber que lo que se descarga es lo que dice el código.

## 30 de septiembre de 2026 — Halving 4: la escalera completa abierta, y el checkpoint sobrevive a la poda

**La testnet #4 recorrió toda la escalera en cuatro días**, como se diseñó (26-sep → 30-sep): el DAA 400,000 se cruzó a mediodía; recompensa 0.03125 RUPIX por bloque, era 5, y los cinco niveles abiertos: Gold, Diamante, Platino, Rodio y **Kings**. Un King cuesta 10 Rodios = 100 Platinos = 1,000 Diamantes = 10,000 Gold quemados; desde hoy cualquiera que los tenga puede intentarlo.

**Emisión verificada a DAA 407,293** (`tools/verificar-emision.py`): regla por calendario 93,977.40625; en el nodo 93,757.42; diferencia 219.99 = los 220 Gold quemados en 22 Diamantes (12 de JC, 10 del seed) más la quema por transacción, menos el exceso acotado de las fronteras 3 y 4 (centavos, como se documentó ayer). El nodo nunca tiene más de lo que permite la regla.

**El checkpoint sobrevivió al avance de la poda.** El punto de poda pasó de H (`7e2ece39…`, blue score 86,400, el checkpoint #1) a `b80cafc8…` (blue score 172,800). El seed **sigue teniendo a H** como bloque de cadena y sigue aplicando la regla: los puntos de poda pasados se conservan aunque la poda avance. Es la razón de fondo para publicar checkpoints solo sobre puntos de poda: nunca desaparecen del nodo. Verificado en vivo, sin ningún rechazo en el journal.

## 30 de septiembre de 2026 (tarde) — La wallet para principiantes, la especificación y cinco huecos cerrados

**Wallet para principiantes** (rama `wallet-principiantes`, 4 commits, para v0.6.2): `forge` revisa antes de mandar (gemas y Gold contra lo que pide la escalera, y si el nivel está abierto preguntando al nodo) y no manda nada si falta algo; `forge`, `send` y `transfer-gem` resumen lo que va a pasar (dirección y monto completos) y piden confirmación; `--gem-address` opcional; rechazos del nodo explicados en palabras sin esconder el original; **español o inglés** (`language es|en`, `--lang`, `RUPIX_LANG`, idioma del sistema); `balance` claro; al crear la wallet se piden dos palabras al azar de la frase semilla y ninguna bandera lo salta; si el daemon no está corriendo, cualquier comando dice cómo arrancarlo. Todo probado en vivo contra el daemon del seed sin mover nada. Guías de forja ES/EN con la tabla "Qué cambió y por qué". Criterio desde hoy para cada cambio: **comodidad, seguridad, transparencia**; una mejora vale si suma en una sin restar en las otras dos.

**`ESPECIFICACION.md`**: cada regla de consenso junto al ataque que detiene y al test que la viola y confirma el rechazo (formato del auditor). Llenada desde el código y comprobada test por test. Encontró seis huecos el mismo día:

- **#0 (el más grave): `TestPOW` estaba en `t.Skip` desde antes de v0.6.0.** Se cambió la función de hash, la matriz y el rango, y en todo ese tiempo ningún test verificaba que el validador rechace un nonce inválido: todo funcionaba porque el minero honesto siempre manda PoW válido. Ahora corre en simnet y devnet en 0.04 s (PoW inválido, target alto, target negativo, bloque bien minado). Palabras del auditor: "la rama más valiosa de la semana, y no por el test nuevo sino por el hallazgo".
- **#1**: `TestSelloFalsoRechazado`: un bloque con el sello de gemas falso queda descalificado, el virtual no lo sigue, el builder honesto no construye encima; el mismo bloque con sello correcto es `UTXOValid`. Falla si se apaga la regla.
- **#3**: `TestTopeDeDiamantesEnBloqueReal`: sembrando el conteo histórico en 2,099,999, la forja 2,100,000 entra y un bloque **fabricado a mano** con la 2,100,001 se rechaza al insertarlo (no queda guardado, el virtual sigue en el tope, la red sigue viva). El builder honesto ni siquiera lo construye.
- **#4**: `TestExcesoEnFronteraDeHalving`: dos hermanos en el último DAA de una era cobran la recompensa vieja; el exceso es exactamente la caída (25,000,000,000 rupias en devnet). El medio RUPIX del halving 2, ahora con test.
- **#5**: `TestQuemaNoSeGasta`: la quema no está en el UTXO set, el mempool y el consenso rechazan gastarla, el cambio de la misma transacción sí se gasta.

Quedan el #2 (validar la lista de puntos de poda para nodos nuevos, código de v0.6.2) y el #6 (fuzzing). Rama `test-sello-falso`, 4 commits.

**El auditor, segunda y tercera ronda:** sin hallazgos en consenso ni en la infraestructura de verificación; tres precisiones aplicadas (el verificador de binarios exige la versión de Go del CI; renovación de checkpoints sin hueco; `--yes` no existe en nada que muestre la frase semilla). "Cuando llegue el revisor con nombre, `ESPECIFICACION.md` es lo primero que debe leer."

## 30 de septiembre de 2026 (noche) — v0.6.2: el tope ya no mata bloques, y el nodo nuevo valida sus puntos de poda

**El hallazgo del auditor, y la respuesta el mismo día.** Revisando `test-sello-falso` encontró que nuestra regla más dura, "un bloque que exceda el tope de gemas se rechaza entero", era en un DAG un veneno: si ese bloque llegaba a ser punta, cada bloque honesto que lo mergeara o construyera encima caía con él. Una forja de 10 Gold para tirar a los mineros honestos. Lo bloqueante era para mainnet, no para hoy (en la testnet ningún tope es alcanzable), pero se cerró hoy, con su método: **primero tres tests en rojo, luego el patrón, luego los tres en verde.** La regla pasó a ser la que ya usábamos para toda transacción inválida: el bloque entra y la forja que excede el tope, sola, no se acepta. El Gold sigue vivo, el sello no cambia, nadie más pierde. Dos consensos independientes coinciden en cuál de dos hermanos con la última gema se la lleva. Visto bueno explícito del auditor a `826be325`, con dos ajustes aplicados en `fb123dfe` (las invariantes gritan en `Error`; las dos fuentes de Kings se comparan en el test de punta a punta).

**Hueco #2, lo que le prometimos a JC y JP en v0.6.1.** El checkpoint protegía a nodos ya sincronizados; un nodo que sincroniza desde cero recibe de un peer la lista de puntos de poda y, contra un peer hostil que sirviera otra historia, no tenía defensa. Desde v0.6.2 la tiene: antes de importar el punto de poda, cada checkpoint activo por debajo tiene que estar en esa lista o ser un ancestro conocido; si no, el punto de poda no se importa. `TestNodoNuevoContraPeerHostil`, rojo antes y verde después, con cinco configuraciones sobre la sincronización real en las cuatro redes.

**Seis de siete huecos de la especificación cerrados y en `main`** (queda el fuzzing). Y una pregunta que desde hoy se le hace a toda regla nueva de Rupix: *¿y si no es punta?*

**v0.6.2** = las cuatro ramas (`wallet-principiantes`, `test-sello-falso`, `tope-no-acepta`, `ibd-checkpoint`), el test del tope en bloque real afirmando la regla nueva, y la suite completa en verde. Primer cambio de consenso desde que hay checkpoints. Binarios reproducibles como en v0.6.1.

**Publicada a las 19:24:** `main` = tag `v0.6.2` = `80f857d7`; binarios reproducibles **4 de 4** (`tools/verificar-binarios.sh v0.6.2`); seed actualizado y minando. Auditor, última ronda: sin hallazgos en `80f857d7`. *"El código deja de ser el cuello de botella. Lo que sigue es el revisor con nombre, y ahora sí tiene con qué recibirlo: ESPECIFICACION, los tests que violan cada regla, builds reproducibles y CI con race."*

