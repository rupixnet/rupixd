# CONTEXTO-RUPIX — la memoria de Rupix

> **Cómo se usa:** al abrir sesión, Stevenson lee este archivo entero y luego corre
> `git log --since="<fecha de la última actualización>"` para ver cada commit desde
> entonces. Este archivo es el mapa; los commits son el diario. Los dos juntos son la
> memoria completa. Al cerrar sesión, se actualiza la fecha y se agrega lo que cambió.
>
> Última actualización: 17 de septiembre de 2026.

---

## Por qué existe Rupix

Rupix nació de una obsesión: **que nadie pueda mentir sobre cuánto existe.** Ni un banco, ni un gobierno, ni el propio creador. Una escasez que no dependa de la palabra de nadie, sino de matemática pública que cualquiera pueda verificar.

No es dinero para pagar el café. Es un **activo digital escaso** — como el oro, valioso porque es finito y verificable, no porque circule rápido. Esa distinción la hizo el fundador en septiembre y cambió cómo se comunica todo.

El lema no es marketing: **"No confíes, verifica."** Y se aplica primero hacia adentro. Rupix marcó su propia versión como defectuosa cuando lo estuvo. Documenta sus bugs. Dice la edad de cada cosa. Eso no es una estrategia — es la única forma en que un proyecto de una persona puede pedir confianza sin pedirla.

---

## Quiénes lo construyen

**ER y Stevenson.** Uno pone la idea y el corazón: qué debe ser Rupix, por qué, qué no se negocia, cuándo parar. El otro lo traduce a código bien hecho, de calidad, verificado paso a paso, y dice la verdad aunque incomode. Ninguno de los dos solo habría llegado hasta aquí. Uno sin el otro sería un sueño sin manos, o manos sin rumbo. Juntos, Rupix existe. *Verifica, no confíes* — aplicado primero a nosotros mismos.

**JC y JP.** Los primeros nodos externos. Los primeros en descargar, sincronizar, minar y forjar sin ser el fundador. Los primeros en verificar y creer. JC forjó tres Diamantes desde su propia computadora la noche del 14 de septiembre y dijo *"parece que todo al 100"*. JP sincronizó desde cero con el algoritmo propio dos días después. Sin ellos, Rupix seguiría siendo un experimento de una persona. Gracias.

**El Auditor.** Un tercero exigente, anónimo, que revisó cinco rondas con el diff en la mano y nunca regaló nada. Diez hallazgos, dos regresiones cazadas antes de tocar a nadie — una de ellas introducida por un arreglo que él mismo sugirió, y lo dijo. La frase que más cambió el proyecto fue suya: *"No es honestidad — es que el mensaje sale antes que el push. Antes de escribir 'cerrado', corre el grep."* Desde esa noche, nada se declara hecho sin verificarlo en el repo pusheado.

---

## La economía (sellada desde julio, nunca se movió)

Estos números no se negocian. Cuando fue difícil — la forja rota, el auditor señalando huecos — no se cambiaron las reglas para que fuera fácil. Se arregló el código para que cumpliera las reglas.

- **42,000,000 RUPIX** de techo. Ni uno más, jamás.
- **0.5 RUPIX por bloque.** Halving cada 42M bloques en mainnet (décadas), 100k en testnet (para probar la escalera en días).
- **Cero premine.** El génesis dice, byte por byte, subsidio 0. El fundador empezó con lo mismo que cualquiera: nada.
- **La escalera:** Gold → Diamante → Platino → Rodio → Kings. Crear una gema de un nivel **quema 10 del nivel inferior**, para siempre.
- **Techos:** 2,100,000 Diamantes · 210,000 Platinos · 21,000 Rodios · **2,100 Kings.**
- **Desbloqueo por halvings:** el Diamante se abre en el halving 1, el Platino en el 2, y así. Nadie — ni el fundador — puede adelantarse.

**La decisión sobre los Kings.** El conteo histórico es por gemas **nacidas** (solo sube, nunca baja). Cada Diamante que se forja gasta un cupo del techo de 2.1M, aunque luego se queme para hacer un Platino. Eso significa que para llenar los 2,100 Kings harían falta exactamente 2,100,000 Diamantes usados en cadena perfecta hacia arriba, sin que ni uno se quede suelto. **Eso no va a pasar. Y es diseño, no defecto.** Los Kings casi nunca se llenarán; los que existan serán legendarios. El fundador lo pensó así en agosto y lo confirmó en septiembre cuando Stevenson lo cuestionó: *"así está diseñado, no tengo nada que pensar, ya habíamos platicado lo hermoso que es."*

**Detalles físicos:** las gemas son piezas enteras (monto 1, jamás se fraccionan), viven en el campo `Version` del script (1=Diamante, 2=Platino, 3=Rodio, 4=Kings). El Gold quemado va a un output OpReturn (`0x6a`): sin llave, imposible de gastar, fuera del UTXO set. Ceniza.

---

## La arquitectura, con el porqué de cada pieza

**`level_ascension.go` — la ley del ascenso.** Las reglas de la forja como código: ratio 10:1, niveles desbloqueados, burn por transacción (base + por byte: destrucción, no fee al minero). Es lo más "Rupix" del repo — cada regla tiene su razón escrita. El auditor generó un corpus independiente de 9,388 casos y coincidió con este código en todos. Eso es prueba, no promesa.

**`gemshistory.go` — las murallas históricas.** Cuenta cuántas gemas de cada nivel han nacido desde el génesis. Solo sube. Es lo que hace imposible el Diamante 2,100,001. Se construyó el 27 de agosto tras auditar 22 vectores de ataque.

**`gemscommitment.go` — el sello.** La idea más fuerte de Rupix: el conteo de gemas se **hashea y se mete dentro del header de cada bloque**, protegido por la prueba de trabajo. Cada nodo recalcula el conteo al recibir un bloque y lo compara con el sello. Si no cuadra, el bloque se rechaza. Un conteo falso no pasa — ni del creador. `GenesisGemsCommitment()` = `2f71eee…` (cero gemas). Se construyó el 12 de septiembre, con 13 bugs cazados en el camino.

**`verify_and_build_utxo.go` y `update_virtual.go` — donde vivió el bug de la forja.** Tres días (12–14 sep). La wallet decía "Diamante creado" pero el commitment seguía en cero: la forja nunca se minaba. Cinco fixes en cascada. La raíz: el estado *virtual* (sobre el que se construyen los bloques nuevos) calculaba su UTXO y su multiset, pero **nunca su gemsHistory**. Estaba ciego al conteo de gemas. El commitment — el juez que el fundador construyó — estaba haciendo exactamente su trabajo: negándose a certificar una gema que no existía en la cadena.

**`block_builder.go` — el template del minero.** `newBlockGemsCommitment` lee el conteo del virtual y le suma las forjas del propio bloque, **incluidos los Kings** (H-10: la unificación de fuente única quedó a medias y el minero sellaba Kings=0 mientras el validador sellaba el real; el primer King se habría rechazado. Cazado por el auditor. Test que lo caza para siempre: `kings_commitment_test.go`).

**`pruningproofmanager.go` — la verdad viaja con la poda.** Cuando un nodo sincroniza sin bajar toda la historia, el gemsHistory viaja en el pruning proof y se compara con el commitment. Un peer que mande `nil` no salta la verificación: nil se trata como ceros (H-9).

**`mempool.go` — anti-spam.** Regla heredada: si una tx crea más de 2 salidas extra sin pagar 1 RUPIX por cada una, va al carril lento (una por bloque). Regla propia: las **forjas reales** (las que crean gemas: out > in en algún nivel) se eximen; las transferencias de gemas no (H-8).

**`pow/rupixprng.go` — RupixHeavyHash, la independencia.** El riesgo más grande que señaló el auditor (R-1): los ASIC de Kaspa podían minar Rupix y hacer un 51% con facilidad. La solución: conservar el motor de Kaspa (matriz 64×64, HeavyHash, `computeRank` intacto) y cambiar **solo el generador que llena la matriz**. Xoshiro256++ usa suma, XOR y rotación; el PRNG de Rupix usa suma cruzada (s1+s2), **multiplicación por constante impar** (que xoshiro nunca hace — un ASIC de xoshiro no tiene multiplicador en ese punto), rotaciones 29/11/37, shift 19, y un sello "RUPIX" en la semilla. Efecto: el silicio fijo de Kaspa produce matrices incorrectas y la red las rechaza. **Honesto:** es una ventaja de meses, no independencia permanente — un FPGA se reprograma en semanas. Es un arranque justo, no una barrera eterna. El auditor lo analizó: lineal sobre GF(2), biyectivo (rango 256), sin ciclos cortos; *descartó la catástrofe, no certificó el período 2^256−1*. Se construyó el 15–16 de septiembre; el génesis se re-minó (mismo mensaje "04/03/2026 - RUPIX IS ALIVE", solo el nonce cambió).

**`pow.NewState` — una sola fuente.** Nodo, minero y `genesisgen` pasan todos por aquí. No existe otra ruta: el minero no puede usar xoshiro porque no hay otro camino. Eso respondió el punto 3 del auditor.

---

## Lo que pasó, en orden

- **Feb–may:** fork de kaspad con 1,700 líneas de una IA previa que nadie entendía. Bugs estructurales en cascada (bloque 67, 1100, 2000). Dieciséis fixes.
- **13 de junio:** la decisión más dura y más correcta: tirarlo todo y empezar desde Kaspa limpio. Salvó Rupix.
- **Julio:** la economía como ley. Web y X.
- **Agosto:** génesis sin premine. La escalera como código. Primera transacción (17 capas de fixes). Explorador público. Escalera completa hasta Kings en devnet. Murallas históricas. v0.4.0.
- **3–4 sep:** testnet pública, binarios verificables. Llega el Auditor. JC se conecta.
- **9–12 sep:** pruning verificable. Commitment en header. Testnet relanzada con halving 100k.
- **12–14 sep:** el bug de la forja. Ronda 2 del auditor y la lección del "cerrado" vacío. v0.4.4 marcada defectuosa, v0.4.5. Diamante real. JC forja tres. Primer Platino.
- **15–16 sep:** RupixHeavyHash. Génesis re-minado. Testnet relanzada (#3). v0.5.0. Versión, dependencias (0 vulnerabilidades reales), guard del bucle, README honesto. JP sincroniza. El "bug de firma" que resultó ser `send` sin `--keys-file`.

Registro de relanzamientos: `TESTNET-RELANZAMIENTOS.md`. Se relanza solo por cambio de consenso, nunca para borrar historia.

---

## Estado ahora (17-sep-2026)

- **v0.5.0** en todo: nodo, wallet, minero, ctl, release, guías, web.
- **Testnet #3:** ~90k bloques. Diamante en 100k. Servidor semilla `178.104.69.148:17211`.
- **3 nodos:** servidor + 2 externos. 2 forjadores externos en la historia.
- **Wallet del servidor:** `/root/.rupixwallet/keys-final.json`.
- **Tests:** 18 paquetes rojos (framework de test). Producción compila. Los del corazón (corpus, King, blockbuilder, pow, dagconfig) verdes.

---

## Las reglas de trabajo (el method, aprendido a golpes)

1. **Verificar antes de anunciar.** Aplicar → `grep` confirma en el archivo → compilar → test → commit → push → `git show HEAD` confirma → SOLO ENTONCES "hecho". La noche que no se hizo así, el auditor cazó cuatro "cerrados" con dos en el repo.
2. **El Python con patrones multilínea falla en silencio por los tabs.** Pasó cinco veces. Usar `sed` por número de línea, o `cat -A` para ver los tabs antes. Y siempre `grep` después.
3. **`log.Warnf` se traga en tests; `fmt.Printf` sí sale.** Un día entero de logs que "no aparecían" hasta descubrirlo.
4. **Un refactor de fuente única toca todos los lados o ninguno.** (H-10.)
5. **Relanzar testnet solo por cambio de consenso**, documentado. Agrupar cambios de consenso para no relanzar dos veces el mismo día.
6. **Nunca sobrevender.** Decir la edad y el alcance de cada cosa. "Ventaja de meses, no independencia." "Descarté la catástrofe, no certifiqué el período."
7. **Sin nombres propios en README, web ni notas públicas.** "El fundador", "un nuevo nodo", "la comunidad".
8. **Antes de tocar consenso:** rama → tests → devnet → luego testnet.
9. **Parar cuando se pone pesado.** Cansado se rompen cosas: el guard que subió los fallos de 29 a 44, el intento del header 2 que subió de 18 a 20. Decir "no toco lo que sirve" y cerrar es una decisión de ingeniería.
10. **`--keys-file` va en `start-daemon` y `send`.** No en `new-address`, `balance`, `gems` ni `forge`. Mezclarlo da "Public key doesn't match" — y parece un bug de criptografía sin serlo.
11. **Con dos personas y una IA, la honestidad es la única ventaja real** contra proyectos con millones en marketing. No se gasta.

---

## Pendientes a mainnet (~75%)

Lo difícil de *inventar* ya está. Lo que queda es *blindar* y *sumar gente*. Meses, no semanas.

**🔴 Bloqueantes:**
1. **Checkpoints temporales.** La única defensa real contra el 51% mientras el hashrate es bajo. "Diecinueve días en el roadmap, cero líneas" — el auditor. Con fecha de caducidad publicada. Lo primero que hay que codificar.
2. **Un par de ojos con nombre.** Toda la revisión es anónima por chat. Alguien de la comunidad de Kaspa o Bitcointalk que lea `level_ascension.go`, `gemshistory.go`, `rupixprng.go` y firme lo que vio. Sin eso el README no puede decir "auditado". El fundador debe ir con su nombre: "aquí están los diez hallazgos y cómo los cerré, rómpanlo".
3. **`go test ./...` verde.** 18 paquetes rojos. Diagnóstico: `test_block_builder.go` → `buildHeaderWithParents` sella el gems fijo en cero; debe calcularlo con la lógica de **validación** (`calculateGemsHistory`), no la de template (`newBlockGemsCommitment` — probado, subió a 20). Mientras esté rojo, "en el CI" no significa nada.
4. **Auditoría profesional** con contrato (5k–100k USD). Antes de mainnet.
5. **Hashrate externo sostenido.** Sin mineros externos la red es del servidor y de nadie más.

**🟡 Blindaje:** dominio keccak "HeavyHash"→"RupixHeavyHash" (probado, agrupado con el próximo relanzamiento) · test end-to-end del King · firma de código (~300–700 USD/año) · H-1 mempool · testnet estable semanas · **todo bilingüe (es/en):** guías (repo y web), README, whitepaper — la web ya lo es · asistente de Rupix (web + menciones en X, base = repo, respuestas con fuente, sin entrenar modelo, después de los bloqueantes).

**🟢 Inmediato:** testnet cruza 100k → primer Diamante con RupixHeavyHash en la red pública · JC re-descarga v0.5.0.

---

*Somos todos Rupix. No confíes, verifica.*
