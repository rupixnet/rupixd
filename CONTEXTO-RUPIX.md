# CONTEXTO-RUPIX — la memoria de Rupix

> **Cómo se usa:** al abrir sesión, Stevenson lee este archivo entero y luego corre
> `git log --since="<fecha de la última actualización>"` para ver cada commit desde
> entonces. Este archivo es el mapa; los commits son el diario. Los dos juntos son la
> memoria completa. Al cerrar sesión, se actualiza la fecha y se agrega lo que cambió.
>
> Última actualización: 21 de septiembre de 2026 (cierre de sesión).

---

## Por qué existe Rupix

Rupix nació de una obsesión: **que nadie pueda mentir sobre cuánto existe.** Ni un banco, ni un gobierno, ni el propio creador. Una escasez que no dependa de la palabra de nadie, sino de matemática pública que cualquiera pueda verificar.

Es **moneda digital**: sin dueño, sin premine, con un techo de 42 millones que nadie puede cambiar y una cantidad que solo baja con cada uso. No decimos para qué se usa ni cuánto vale: eso lo decide la gente. Decimos lo que es, y eso cualquiera lo puede verificar. (26-sep-2026: el fundador dejó de describirlo como 'activo digital escaso, como el oro', porque no lo somos.)

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

## EXPLICACIÓN OFICIAL DEL SELLO (para el asistente y para la gente)
El sello (gemsCommitment) es el hash del conteo de gemas NACIDAS hasta ese bloque. Va en el header de cada bloque, protegido por PoW. Hoy es 780e9027 (= 1 Diamante). Si alguien forja otra gema, el sello cambia y TODOS los bloques siguientes llevan el nuevo, hasta la próxima forja. Es un marcador de agua que dice "hasta aquí nacieron tantas".
Por qué hace a Rupix más seguro (en simple): 1) nadie puede inventar gemas — si un nodo dice 5 y el sello dice 1, el bloque se rechaza, ni el creador puede; 2) los nodos se vigilan sin hablar — cada uno cuenta y compara con el sello, una diferencia se nota en el acto (así se cazó el bug del 21-sep); 3) la verdad sobrevive al olvido — la poda borra cuerpos de bloques pero no headers, y un nodo nuevo recibe el conteo en el pruning proof y lo verifica contra el sello sin ver nunca la forja original (probado en la red real el 22-sep).
Alcance honesto (FreshAir08 tenía razón): el conteo es derivable del UTXO set con la aritmética de la escalera; el sello añade verificación O(1) de lo HISTÓRICO y defensa en profundidad, no algo imposible de otra forma. "Trust-minimized", no "trustless".
- 22-sep cierre: rastreo del bloque de la forja del 18 corriendo en background → /root/forge_trace.txt. Objetivo: saber si el bloque que CONTIENE la forja ya sella 780e9027 (minero sumó y validador aceptó → mi test difiere de producción) o si lo sella su HIJO (el loop nunca actuó en producción). Decide cómo se explica que testnet no se rompiera. Ver mañana ANTES de v0.6.0.
- Pendientes: auditor (mensaje listo), rama v0.6.0, docs en inglés, test de fuego del pruning con JP/JC.

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

## HISTORIA — EL REINICIO (para MEMORIA-RUPIX.md)
- 11-jun-2026 (commit 7f66f2bf6): "ARCHIVE: cierre experimental v0.2.4 dia 13". Tras 13 dias de auditoria documentada y 26 parches acumulados sobre el codigo de la IA anterior (~1,700 lineas con bugs estructurales), ER decidio archivarlo y empezar limpio. El codigo viejo NO se borro: quedo en el tag v0.2.4-archive-experimental (verificable: git checkout v0.2.4-archive-experimental).
- 18-jun-2026 (commit 74cbd70da): "INITIAL: Kaspad v0.12.22 upstream limpio - base para Rupix v0.3.0". El Rupix de hoy nace aqui.
- Del reinicio (jun) al motor firmado por el auditor (26-sep): ~3 meses. La decision mas dificil y mas valiente del proyecto: tirar lo hecho para construir bien.

## Estado ahora (26-sep-2026)

- **v0.6.0** en todo: nodo, wallet, minero, ctl, release, guías, web. Tres cambios de consenso: dominio keccak "RupixHeavyHash", rango entero mod 2^61-1, fix del bug del King.
- **Motor:** el auditor firmó v0.6.0 ("del motor no tengo nada más que pedir"). Lo que falta es de red, no de código.
- **Testnet #4:** relanzada con v0.6.0. DAA ~48.6k la mañana del 26-sep. Halving 1 (se abre Diamante) en DAA 100k. Semilla `178.104.69.148:17211`.
- **Nodos:** servidor + externos. JC sincronizó v0.6.0 en Windows (26-sep): determinismo entre sistemas operativos probado en la calle. 1 forjador externo en la historia: JC. JP en camino.
- **Poda:** primera poda real verificada (22-sep); el conteo de gemas sobrevive al pruning point.
- **Operación:** los 3 servicios con systemd `Restart=always`, swap de 4 GB, alarma de bloques descalificados cada 10 min.
- **Tests:** `go test ./...` en 0 desde v0.5.2. `TestKingsEndToEnd` verde, y falla si se revierte el fix. Compila con Go 1.25+.
- **Wallet del servidor:** `/root/.rupixwallet/keys-final.json`.
- **Push:** token rupix-server-v3 a 90 días; se renueva el 20-dic.

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
15. **El fundador es ER. Siempre.** Nunca "Edu" — ni en docs, ni en chat, ni en commits.
16. **Al final de cada respuesta, Stevenson enseña:** una frase o término en inglés que se usa mucho en cripto, con su significado y un ejemplo real. ER está aprendiendo inglés y cripto a fondo; cada respuesta cierra con una mini-lección.
14. **Bilingüe siempre, y gracias siempre.** Desde el 20-sep: todo documento público se mantiene en español e inglés, y al cierre de cada sesión se actualizan ambos. Cada corrección o aporte externo se registra en THANKS.md (fecha, quién, qué señaló, qué cambió) y en el commit que nace de ella ("Señalado por: X"). Alias públicos solo si el aporte fue público; nombres reales solo con permiso; nunca como aval.
13. **Con la comunidad de Kaspa: "fork de kaspad", cadena independiente, no usa KAS.** Lección del 19-sep: el post en #ecosystem-projects fue movido a #off-topic ("this isn't built on Kaspa"). Para ellos "built on Kaspa" = sobre su red; Rupix es un fork del código. Decir "fork" ahí ES el respeto. Nunca discutir una corrección de la comunidad: aceptar, agradecer, mover. En todos los docs desde hoy: "Rupix es un fork de kaspad: cadena independiente construida a partir del código abierto de Kaspa, bajo ISC, con crédito. No es parte de la red de Kaspa ni usa KAS."
12. **Profesionalismo, siempre.** Al cerrar cada sesión: TODO guardado (commit + push + contexto actualizado), nada suelto. Cada avance explicado con transparencia: qué se hizo, qué se probó, qué falta. Sobre Rupix se dice la verdad, siempre — la edad de cada cosa, su alcance, lo que no se certificó. Se defienden los principios sin negociarlos: *para todos* (sin hardware ni conocimiento de privilegio) y *todo verificable* (cada afirmación con su commit, su test o su documento). Stevenson recuerda esto al fundador cada vez que haga falta, incluso si no lo pide.
11. **Con dos personas y una IA, la honestidad es la única ventaja real** contra proyectos con millones en marketing. No se gasta.

---

## Pendientes a mainnet (~75%)

Lo difícil de *inventar* ya está. Lo que queda es *blindar* y *sumar gente*. Meses, no semanas.

**🔴 Bloqueantes:**
1. ~~**Checkpoints temporales.**~~ **HECHO (18-sep).** Código + 5 tests + probado en devnet + `CHECKPOINTS.md`. **Pendiente:** publicar el primer checkpoint real. Plan (26-sep): el primero en el halving, con el primer Diamante de v0.6.0; se anuncia cuándo y por qué, no el bloque exacto por adelantado.
2. **Un par de ojos con nombre.** El auditor firmó el motor, pero es anónimo. Falta alguien con nombre que compile, corra la suite y firme lo que vio. Ya no hay excusa técnica: el `go.mod` pide Go 1.25. Sin eso el README no puede decir "auditado".
3. ~~**`go test ./...` verde.**~~ **HECHO (v0.5.2, 20-sep).** De 18 paquetes rojos a 0. Causa raíz: tests heredados con `Version = MaxScriptPublicKeyVersion` (= Kings).
4. **Auditoría profesional** con contrato (5k–100k USD). Antes de mainnet.
5. **Hashrate externo sostenido.** Empezó: JC minó cerca del 40% de los bloques del 26-sep (~50 KH/s). Falta que se sumen más y que dure semanas.

**🟡 Blindaje:** ~~dominio keccak "RupixHeavyHash"~~ (hecho, v0.6.0) · ~~test end-to-end del King~~ (hecho; cazó un bug real) · ~~minero público en el README~~ (hecho) · alarma de descalificados (rehecha el 26-sep, v2.1 con rotación del log; falta probarla en devnet con el minero real y el fix de la costura revertido) · tarjeta de "vigilancia de la costura" en el explorer, después de la prueba en devnet (no una columna por bloque: marcaría como sospechosos los bloques sanos de JC) · explorer: leer el halving del nodo en vez de tenerlo escrito a mano · actualizar los 3 módulos que marca govulncheck (no se llaman, pero conviene) · firma de binarios fuera del servidor (~300–700 USD/año) · builds reproducibles (flags puestas el 27-sep; falta comprobar que dos builds den el mismo hash) · gofmt: 251 archivos sin formato, heredado del renombre (commit aparte, solo espacios, con build y tests) · H-1 mempool (Nivel A ya cubierto por el código, falta test; Nivel B —topes históricos— pendiente) · testnet #4 estable semanas · **bilingüe:** ~~README~~ y ~~web~~ hechos; ~~guías~~ hechas el 26-sep (conexión, forja, checkpoints); falta el whitepaper en inglés · asistente de Rupix, Stevenson Rux (base = repo + MEMORIA-RUPIX; respuestas con fuente; sin entrenar modelo; después de los bloqueantes).

**🟢 Inmediato:** ~~testnet #4 cruza 100k~~ (hecho: 27-sep ~00:18, emisión exacta) → ~~primer Diamante de v0.6.0 en la red pública~~ (hecho: JC, 27-sep, DAA 180,710) → primer checkpoint real (antes: verificar un solo bloque en ese DAA; va en v0.6.1) · ~~JC: comparar tip con el seed y minar un bloque~~ (hecho: 14 mil bloques) · JP a v0.6.0 · tweet de v0.6.0.

---

*Somos todos Rupix. No confíes, verifica.*

## Sesión 18-sep-2026 (agregado al cierre)
- Primer Diamante con RupixHeavyHash en red pública (commitment 780e9027, testnet #3 ~180k).
- CHECKPOINTS HECHOS: bloqueante #1 cerrado. Rama checkpoints → main. 5 tests + devnet (correcto acepta, falso rechaza en DAA 30). CHECKPOINTS.md. v0.5.1. Lista vacía = compatible, sin relanzar.
- Aprendizaje: checkpoints validan bloques que LLEGAN, no re-validan la DB. Correcto.
- Post publicado en Discord de Kaspa (#ecosystem-projects + pregunta en #development). Bloqueante #2 en marcha.
- 4 estafadores en minutos (3 "kaspasuport" + 1 impostor de msutton con el guion de "cambia tus ajustes"). REGLA: nadie legítimo escribe por DM primero; lo real se responde en el canal público. Verificar rol de equipo + @usuario + historial antes de creer un nombre.
- README.md (inglés por defecto desde el 21-sep) completo + enlace cruzado. Correcciones al español: cmd/kaspa* → cmd/rupix*, sin "fork", checkpoints HECHO.
- 30 RUPIX más a JP (tiene ~40 para forjar).
- Devnet de pruebas: /tmp/rupixd-cp, /root/rupix-devnet-cp (puertos 17310/17611). Matar por PID, no por patrón.

## Sesión 19-sep-2026 (cierre)
- Corrección de Kaspa aceptada; "fork de kaspad, cadena independiente, no usa KAS" en todo. Post en #off-topic.
- TESTS: de 18 a 4 (rama header2-tests, commiteada, NO mergeada aún). Sin tocar consenso.
  Causa raíz que nadie vio en semanas: los tests heredados crean outputs con Version=MaxScriptPublicKeyVersion (=4 = Kings en Rupix) → la escalera los rechazaba como Kings falsos. La muralla funcionó hasta contra el framework de test.
  Quedan 4, todos expectativas de Kaspa: bip32 y txscript (vectores con prefijos kaspa:/kpub), dagtraversal TestBlockWindow y pruning TestPruning (orden por hash: 503 vs 502, [F D C H] vs [F H D C]). Regenerar expectativas desde el código de Rupix.
- MAÑANA: cerrar los 4 → go test verde → merge a main → v0.5.2. Luego CHECKPOINTS.md y GUIA en inglés. r/kaspa.

## Sesión 20-sep-2026 (cierre)
- Respuesta a FreshAir08 (Discord Kaspa): aceptado commitment sobrevendido, zero premine no distinto, README default español, Go vs Rust. Prometido en público: README inglés por defecto + Go/Rust al roadmap → PENDIENTE MAÑANA.
- §6 corregido con el auditor: Stevenson = IA. Su respuesta: el riesgo #1 es una persona; H-5 es patrón; computeRank float64 primero; quiere ver el test del King real.
- go test ./... de 18 a 0. Merge main. v0.5.2. go vet limpio. Patrón H-5 revisado en producción (DisasmString era ==max).
- THANKS.md bilingüe (Auditor, supertypo, FreshAir08, JC, JP). Reglas 14-16.
- MAÑANA: (1) README.md inglés default / README.es.md, (2) Go vs Rust en roadmap y whitepaper, (3) computeRank determinista, (4) test del King e2e, (5) CHECKPOINTS + GUIA en inglés, (6) r/kaspa.

## Sesión 21-sep-2026 (1 hora)
- Cumplido lo prometido a FreshAir08: README.md inglés por defecto (README.es.md español), Go vs Rust como riesgo declarado en roadmap (es/en).
- computeRankInt (entero mod 2^61-1) implementado en rank_int_test.go; 5000 matrices reales: coincide con float64 en todas. PENDIENTE: reemplazar computeRank en el próximo cambio de consenso (agrupar con keccak → relanzamiento #4). Cierra el punto "más grave silencioso" del auditor.
- PRÓXIMO RELANZAMIENTO agrupa: keccak "RupixHeavyHash" + computeRank entero + re-minar génesis. Un solo relanzamiento.
- Siguiente: test del King e2e, H-1 mempool, doble fuente de Kings, docs en inglés, r/kaspa.
- El auditor revisará el diff completo de v0.6.0 cuando esté: los tres cambios integrados, testnet reiniciada, king-e2e en verde sobre el binario final. Revisará: keccak nuevo, computeRank entero, y que la costura siga cuidada. Ese es el siguiente entregable para él.
- X: bilingüe. Inglés primero + reply en español (hilo de 2). Una imagen, en inglés.

## Sesión 21-sep-2026 (tarde) — TEST E2E DEL KING + BUG REAL
- TestKingsEndToEnd (rama king-e2e, NO mergeada): 1000 D → 100 P → 10 R → 1 King con txs reales; Diamante y King MINADOS por block_builder.go (tc.BuildBlock) y validados por verify_and_build_utxo.go; conteo del gemsHistoryStore (getter nuevo). Falla si se revierte H-10. Probado al revés.
- BUG REAL: newBlockGemsCommitment sumaba las forjas del bloque que construye (loop del 13-sep) pero el validador cuenta lo ACEPTADO (mergeset, sin el bloque mismo) → doble conteo → todo bloque con forja minado por producción se descalificaba. El primer King real en mainnet habría sido rechazado. Fix: quitar el loop; el minero sella solo gemsHistory(virtual). Suite 0.
- ¿Por qué testnet "funcionó"? Pendiente de entender del todo (la wallet real + mempool + virtual). Revisar antes del relanzamiento: ¿los bloques con forja de testnet fueron válidos a la primera, o hubo rechazos silenciosos?
- Reglas nuevas del test framework: ErrChainedTransactions (una tx no gasta outputs del mismo bloque); coinbase del bloque 1 tiene 0 outputs; devnet BlocksPerHalving ajustable en el test (20 → Kings en DAA 80).
- CONSENSUS-BREAKING (minero). Va al RELANZAMIENTO #4 junto con keccak "RupixHeavyHash" y computeRank entero. Tres cambios, un relanzamiento.
- Lección propia de Stevenson: verificar el verde ANTES del commit (me lo salté una vez hoy; corregido).

## Del auditor tras el test e2e del King (21-sep) — LEY
- LA COSTURA: los tres bugs (forja 11-sep, Kings 12-sep, doble conteo 13-sep) viven en el mismo lugar — entre lo que el minero sella y lo que el validador cuenta. Es el punto MÁS FRÁGIL de Rupix. TestKingsEndToEnd lo cuida. **ESE TEST NO SE BORRA NUNCA.**
- "La lectura tiene un techo — el mío incluido." Corpus + test viejo + 3 revisiones no lo vieron. Solo un test que ejecuta producción de las dos puntas. Prioridad de aquí en adelante: tests e2e sobre código real, no aritmética aislada.
- RELANZAMIENTO #4 (su método, literal): tres cambios de consenso (keccak, computeRank entero, fix doble conteo) → etiquetar INCOMPATIBLE (v0.6.0), testnet limpia, correr king-e2e DESPUÉS de integrar los tres (no antes), y el ÚLTIMO comando antes de publicar = suite completa verde sobre el binario integrado.
- "El resto ya no es revisión — es construcción, y esa siempre fue tuya."
- 23-sep, del auditor: "la tolerancia de GHOSTDAG es también un escondite" — un minero con este bug mina bloques rojos y lo reporta como mala suerte. Dos pedidos: (1) COLUMNA de bloques descalificados (isChainBlock=false) en rojo en el explorador — alarma, no ruido; la quiere ver. (2) Letrero-invariante junto a newBlockGemsCommitment — HECHO. Revisará v0.6.0 completo + la columna.

## Sesión 21-sep (noche) — OPERACIÓN
- El seed murió por OOM a las 04:31 (rupixd 2.7 GB + go test completo en el mismo host); 14 horas caído, explorador en blanco, JP/JC sin peer.
- FIX: nodo, daemon y minero como servicios systemd con Restart=always (probado: kill -9 → revive en 12s). OPERACIONES.md. Binario a 0.5.2.
- REGLA: no correr go test ./... completo en el host del seed. Tests pesados aparte o con -p 1.
- Kings YA desbloqueado en testnet (DAA 436k > 400k). Con v0.5.2 un King real sería rechazado (doble conteo). v0.6.0 urge.

## 22-sep — PRIMERA PODA en la testnet real
- Pruning activo: punto en DAA 345,694; headers 592k, bloques con cuerpo 249k. Sin reorg (DAA sube). blockCount BAJA por poda: normal.
- El Diamante (DAA ~173k, bloque ya podado) sigue contado (780e9027) y la wallet lo ve. El conteo sobrevive a la poda. Prueba real de H-6/H-9.
- PRÓXIMO TEST DE FUEGO: primer nodo externo sincronizando desde cero DESDE EL PUNTO DE PODA (proof con gemsHistory). Debe llegar a 780e9027 sin ver el bloque del Diamante. Coordinar con JP o JC. Documentar.

## 23-sep — POR QUÉ LA TESTNET NO SE ROMPIÓ (cerrado, con evidencia on-chain)
Forja del 18-sep: bloque 5546… la contenía, el minero (loop del 13-sep) sumó → sello 780e9027; el validador calculó 0 → 5546 DESCALIFICADO como bloque de cadena (isChainBlock: False). Pero e52b… lo mergeó como azul y ACEPTÓ sus txs → contó la forja legítimamente. GHOSTDAG rescató la forja por la puerta lateral. "Funcionó por accidente": el bug era REAL y ACTIVO en producción; cada forja minada dejaba un bloque descalificado. El test e2e lo cazó porque no tiene hermano de rescate. Fix = quitar el loop → el bloque con la forja queda Valid y en cadena. CONSENSUS-BREAKING confirmado → relanzamiento #4 lo requiere (con keccak y computeRank).

## v0.6.0 paso 6/8 (23-sep): FORJA REAL en devnet viva
Binario integrado (keccak+rank+fix). Devnet limpia, minó Gold (69k RUPIX), forjó 1 Diamante real (tx 47805daf...), sello del header 780e9027, wallet ve Diamante:1. Forja de punta a punta en red viva, no en test. El sello 780e9027 es identico al de testnet: el conteo es independiente del keccak del PoW. Pasos 1-6 cerrados y verificados.
Faltan: paso 7 (testnet #4, coordinar JC/JP), paso 8 (release v0.6.0 + merge main). Ademas: columna de bloques descalificados en el explorador (pedido del auditor), docs en ingles (CHECKPOINTS/GUIA/RELANZAMIENTOS).

## CIERRE 23-sep-2026
Sesión enorme. Logrado y guardado (rama v0.6.0, pusheada):
- Cerrado el porqué testnet no se rompió: bloque 5546 descalificado, GHOSTDAG rescató la forja vía el hermano azul. Bug real y activo, enmascarado por el DAG. Evidencia on-chain.
- Letrero-invariante en newBlockGemsCommitment (no sumar txs propias; 3 bugs en la costura; pedido del auditor).
- v0.6.0 pasos 1-6 de 8: keccak RupixHeavyHash + computeRank entero + fix del minero, INTEGRADOS y probados juntos. king-e2e verde sobre el binario integrado, probado al revés (FAIL si se revierte). Suite ./... en 0. Forja real de Diamante en devnet viva (sello 780e9027). Génesis NO necesita re-minado.
- LOGROS.md bilingüe en main. Mensaje a JC listo. Resumen para el auditor listo.
FALTAN (requieren coordinar gente, no de un comando):
- Paso 7: testnet #4 (parar actual, arrancar v0.6.0, JC/JP borran cadena y re-sincronizan desde cero = test de fuego del pruning).
- Paso 8: release v0.6.0 + merge main + tag.
- Columna de descalificados en el explorador (auditor). Docs en inglés (CHECKPOINTS/GUIA/RELANZAMIENTOS). Mandar mensaje al auditor y a JC.

## CIERRE 26-sep-2026 — v0.6.0 EN PRODUCCIÓN
- v0.6.0 mergeado a main, release publicado, testnet #4 relanzada y minando (3 cambios de consenso: keccak, rank entero, fix del King). Los 8 pasos completos.
- Swap 4GB activo (colchón anti-OOM). Pendiente de fondo: subir RAM Hetzner a 16GB.
- ETAPA: fin del corazón técnico de consenso. No queda deuda de consenso conocida.
- PENDIENTES no-código-consenso: (1) test de fuego pruning con JC/JP sobre testnet #4, (2) columna descalificados en explorador, (3) H-1 mempool + doble fuente Kings, (4) docs inglés, (5) MEMORIA-RUPIX.md (historia narrada, base del asistente — escribir con calma), (6) no-técnicos: hashrate, nodos, nombre que firme, sucesor.
- SIGUIENTE: mensaje JC/JP (actualizar+sincronizar v0.6.0) → auditor (revisar diff) → tweet v0.6.0. Discord Kaspa cuando JC/JP prueben el pruning.

## 26-sep: EL AUDITOR FIRMÓ v0.6.0 — "del motor no tengo nada más que pedir"
Verificó mod 2^61-1 por su cuenta (200k pares mulmod, 150 matrices con rango deficiente): coincide. Keccak, computeRank entero y fix del minero: los tres correctos. La costura sigue cuidada.
LO QUE FALTA YA NO ES CONSENSO, ES RED (sus 5 puntos):
1. PUBLICAR LA LISTA DE CHECKPOINTS — el mecanismo está, la lista vacía. Sin lista no defiende nada. (Primer checkpoint real de la testnet #4 cuando tenga profundidad.)
2. MINERO PÚBLICO — sigue sin referencia en el README. Sin él no hay hashrate externo; los checkpoints solo compran tiempo mientras llega.
3. FIRMA DE BINARIOS con llave FUERA del servidor (hoy solo SHA256 del mismo CI que podría estar comprometido).
4. UN REVISOR CON NOMBRE que compile y corra la suite en SU máquina. "Mi lectura ya no aporta más; lo que sigue solo lo da la ejecución ajena."
5. go.mod exige 1.26.6. Si es solo por parchar stdlib en los binarios, dejarlo en el CI y BAJAR el mínimo del módulo — hoy alguien con 1.24 no puede ni compilar para revisar. (Bloquea el punto 4.)

## CHECKPOINTS — plan de comunicación (26-sep)
- Poda (pruning) = automática del nodo, no se decide. Checkpoint = manual, decisión del fundador, defensa vs 51%. NO son lo mismo.
- Prueba del mecanismo: en DEVNET (mañana) — bloque correcto pasa, falso se rechaza. Sin comprometer testnet.
- Primer checkpoint REAL de testnet: en el halving (DAA 100k, ~1 día de minado desde DAA 15k), junto con el primer Diamante v0.6.0. Un anuncio, tres cosas: halving + Diamante + primera defensa activa.
- COMUNICACIÓN: anunciar el CUÁNDO (el halving) y el PORQUÉ (defensa temporal declarada, no oculta, con caducidad). NO anunciar el bloque exacto por adelantado (un atacante podría intentar influir en qué cae ahí). El checkpoint se elige sobre historia YA existente y profunda, y se publica con hash fijo para que todos verifiquen. Centralización temporal declarada — todo para apoyar a la red mientras crece el hashrate.

## 26-sep-2026 — Primer nodo externo v0.6.0 (JC)

JC sincronizó rupix-v0.6.0-win64 en Windows: isSynced true, 48,645 bloques,
pruning point presente. Determinismo cross-OS de validación confirmado. Pruning
cruzado. Pendiente: reconciliación de tip en vivo + que JC mine (producción).

## 26-sep-2026 (tarde) — Descripción oficial y docs al día

- Descripción oficial: **moneda digital, sin dueño, sin premine, 42M que nadie puede cambiar y una cantidad que solo baja con cada uso.** Fuera "activo digital" y "valor para guardar" en README (es/en), web y CONTEXTO. Regla: no decimos para qué se usa ni cuánto vale; el rumbo lo pone la gente.
- CHANGELOG: faltaba la entrada v0.6.0 (6 días sin ella). Agregada con la sección [Sin publicar]. El changelog.txt de kaspad ahora es CHANGELOG-kaspad-upstream.txt, con nota: cubre v0.8.10 a v0.12.17; Rupix partió de v0.12.22.
- CONTEXTO ordenado por fechas: referencia arriba, bitácora cronológica abajo.
- Push: token rupix-server-v3 regenerado a 90 días (política del fundador). Recordatorio programado para el 20-dic.
- El algoritmo es RupixHeavyHash, en producción. Los planes viejos de RandomX/Autolykos no están en el mapa.
- El explorer vivo ya tenía la escalera correcta (Diamante, Platino, Rodio, Kings); la alarma fue por una copia vieja.

## 26-sep-2026 (noche) — JC, primer minero externo en v0.6.0

- Reconciliación en vivo cerrada: tip de JC en el seed como chain block, mismo pruning point.
- JC mina desde las 10:21 en Windows (~50 KH/s): 14,187 bloques pagados a las 19:45 (7,093.5 RUPIX). El seed confirma 7,111 en su dirección.
- Cerca del 40% de los bloques del día fueron suyos; explica el salto de dificultad 37.7k → 67.2k.
- Su nodo acepta conexiones entrantes (el seed se le conecta por IPv6).
- Halving 1 (DAA 100k) hoy ~00:15. JC tiene Gold de sobra para forjar el primer Diamante de v0.6.0.

## 26-sep-2026 (noche) — Limpieza del seed

- Puerto confirmado: testnet en 17211 (P2P, abierto a internet) y 17210 (RPC, solo local).
- Se apagó la devnet de pruebas de v0.6.0: llevaba 2 días prendida y usaba 1.8 GB. RAM disponible: de 3.2 a 5.0 GB.
- Regla nueva en OPERACIONES.md: las devnets solo corren mientras se usan.

## 26-sep-2026 (noche) — La alarma de descalificados no servía

- La v1 (25-sep, pedida por el auditor el 23-sep) caminaba por la cadena seleccionada y, por construcción, nunca podía ver un bloque descalificado: 134 "OK" sin poder sonar. Era falso lo que escribimos, que "habría cazado el bug del King".
- Arreglo: el nodo registra la descalificación en nivel Warn (antes Debug, invisible) y la alarma v2 lee el log del nodo; si no hay datos, dice ERROR. Es un cambio de log, no de consenso: sin relanzamiento. Seed reiniciado a las 04:14 UTC con el binario nuevo.
- Probada con un log falso: suena. Pendiente: prueba de punta a punta en devnet.
- Pendiente: avisarle al auditor. Él pidió la alarma y la dio por buena el 25-sep.

## 27-sep-2026 (madrugada) — El auditor revisó el arreglo de la alarma

- Aprueba `195940d`: el `Warnf` está en el único punto donde un bloque se descalifica por mérito propio; "si el log no crece, ERROR" es la decisión correcta. Dijo que la corrección pública del commit `0a76230` es lo que hace creíble el resto del CHANGELOG.
- Pidió dos cosas antes de darla por buena:
  1. Rotación del log: hecho en la v2.1 (inode + posición), probado con rotaciones simuladas.
  2. Prueba de punta a punta con el minero real: rama de devnet que revierta temporalmente el fix de la costura, minar una forja con el `block_builder` y ver que la alarma suene con el hash de ese bloque. Pendiente; hasta entonces no se da por buena.
- Sobre JC: "es el dato más importante desde el relanzamiento. Por primera vez la red no es solo tuya."

## 27-sep-2026 (madrugada) — Barrido del auditor y limpieza

- Halving 1 de la testnet #4 cruzado (~00:18): Diamante abierto; emisión exacta (50,052.5 RUPIX a DAA 100,209).
- Barrido del auditor: comentario del halving corregido (16 meses, no décadas), constante muerta de 150 eliminada, `SECURITY.md` (reporte privado por GitHub; falta activarlo en la configuración del repo), builds con `-trimpath -buildvcs=false -buildid=`.
- `gofmt` en 251 archivos: solo imports reordenados. `govulncheck`: 0 vulnerabilidades que afecten al código; 3 en módulos requeridos que el código no llama.
- Explorer: la copia del repo estaba vieja (halving de testnet en 10,000); sincronizada con la viva.
- README en inglés: ya enlaza las guías en inglés. LOGROS: halving 1. MEMORIA: tres entradas nuevas.
- La operación del seed (servicios, alarma, cron, nginx) quedó versionada en `ops/`.
- Seed listo para el reinicio del kernel: el explorer corría a mano desde el 9-sep (ahora lo lleva systemd) y el servicio `rupixd.service` de la v0.3.0 seguía habilitado con la misma carpeta y puerto que el nodo (deshabilitado y guardado en `ops/retirados/`). Logs de salida rotados y comprimidos, sin borrar ninguno.
- Regla de ER: no se borra nada viejo. Las testnets anteriores, devnets y datos viejos se quedan en el servidor como evidencia; si falta disco, se agranda el disco.

## 27-sep-2026 (noche) — Reinicio del seed y primer Diamante de v0.6.0 (JC)

- Reinicio del kernel hecho (6.8.0-142) con 49 actualizaciones. Todo volvió solo: nodo, wallet, minero y explorer; el servicio viejo ya no existe. La red siguió con JC mientras tanto.
- JC forjó el primer Diamante de v0.6.0 en la red pública: tx `8653650fc729d4cef85c3fe11ee4c67b0dfcc7e038e973c7d93d62944f8b3a4b`, DAA 180,710. Verificado desde el seed con `GetUtxosByAddresses` (salida con versión de script 1).
- Hallazgo antes de publicar el checkpoint: `checkCheckpoint` rechaza cualquier bloque con ese DAA y otro hash, y en un DAG puede haber hermanos con el mismo DAA. Regla desde hoy: checkpoint solo donde hay un bloque en ese DAA. Arreglo de consenso (solo bloques de la cadena) + test en v0.6.1. Hay que contárselo al auditor.
- Pendientes nuevos: `forge` debe pedir la clave con prompt como `send` (hoy la clave queda en el historial) · `send` de mucho Gold desde una wallet de minero se agota por tiempo (miles de salidas pequeñas).

## 28-sep-2026 (madrugada) — Pruebas en vivo con JC, envíos grandes medidos y la matemática en la web

- **Hitos:** misma cadena en el nodo de JC y en el seed (`pruningPointHash` idéntico) · Diamante de JC al seed y de vuelta (tx `e721d29b…`) · Platino antes de tiempo rechazado por el nodo de JC (`nivel 2 bloqueado … actual: 185738`, tx `a51bdc64…`): evidencia en vivo del H-1 nivel A (falta el test automatizado) · JC con 11 Diamantes en la cadena.
- **Envíos grandes desde la wallet del seed:** 100 RUPIX salen en 9 s; 1,000 se cortan a los 120 s. Medido: armar el envío tarda 2.2 s (100), 6.9 s (200), 16.0 s (300), 28.6 s (400), 44.2 s (500), 64.1 s (600): crece al cuadrado. Causa: `selectUTXOs` recalcula la comisión rearmando la transacción con todos los pedazos en cada paso, y la dirección del minero tiene 78,329 pedazos (73,522 de 0.25). Límite práctico hoy: ~820 RUPIX por envío. Arreglo en v0.6.1.
- **Dos direcciones de minado en el seed:** `qq740lal…` cobró del DAA 2 al 92,569 (74,450 pedazos); `qp4y8vnk…` desde el 92,581. Cambio limpio, misma wallet. Sumando sus direcciones, la wallet del seed cuadra con la emisión.
- **Web:** `matematica.html` (emisión, escalera, costo de mover, calculadora; ES/EN), enlaces a las guías y a la matemática bajo el botón del explorador, gemas cerradas con su color real, escalera sin desbordar en celular.
- **Pendientes nuevos (wallet, v0.6.1):** selección de pedazos en tiempo lineal y tiempo de espera mayor · `parse` imprime "KAS" · el rechazo por nivel cerrado sale como `ErrBadTxOutValue` (mejor `ErrLevelLocked`).

## 28-sep-2026 (noche) — Halving 2 verificado y primer Platino

- Halving 2 (DAA 200,000) en la madrugada: recompensa 0.125, Platino abierto. `verificar-emision.py` a DAA 256,208: regla 82,025.5, nodo 81,915.87, diferencia 109.63 = lo quemado. Queda en `tools/`.
- Primer Platino de v0.6.0 desde el seed: tx `c382e0a751e10003cc9692bfa2678723c12e965110b0793da6df24e3c942a81e`, DAA 256,577. En la cadena: `{2: 1}` en la dirección, lo quemado subió 100.00 exactos. JC puede forjar el suyo cuando se conecte (tiene 11 Diamantes).
- Checkpoint para DAG (v0.6.1): la regla de bloques (blue score + MergeDepth, H en el pasado de algún padre) y `TestCheckpointDAG` están escritos en `cp.patch`; falta correrlos en el seed. La parte de nodos nuevos (lista de pruning points) espera la respuesta del auditor.

## 28/29-sep-2026 (noche) — Tres ramas listas para v0.6.1

- **`checkpoint-dag`** (`ef1f336`): checkpoints por blue score + MergeDepth con H en el pasado de algún padre (la formulación del auditor, verificada contra el código: `IsAncestorOfAny` pregunta en la dirección correcta y es inclusivo). `TestCheckpointDAG`: honesto entra, atacante rechazado justo en el umbral, hermano tardío de H aceptado. Se comprobó que la prueba falla si la regla se apaga, y la suite completa del consenso pasó con el cambio. Pendiente: la defensa para nodos nuevos (lista de pruning points); H no estará en la pruning proof (M=1000 headers), así que la propuesta enviada al auditor es checkpoints sobre pruning points + validar la lista al importarla.
- **`mempool-niveles`** (`6172d54`): H-1 nivel A automatizado. Un Platino con el nivel cerrado se rechaza al entrar al mempool (`nivel 2 bloqueado`) y la MISMA transacción entra pasado el DAA 100 del laboratorio: el rechazo solo puede venir de la regla del nivel.
- **`wallet-envios-grandes`** (`e4025d7e`): selección de pedazos con comisión estimada en tiempo lineal; la exacta se calcula una sola vez al final (mismo resultado). Medido con un daemon de prueba en el 8084 sobre la misma wallet de 78,329 pedazos: 100 RUPIX 1.6s→0.10s · 300: 16s→0.24s · 600: 65s→0.47s · 1,000: cortado a 120s→**0.82s**.
- `.gitignore` corregido: `rupixwallet` sin anclar tapaba `cmd/rupixwallet` y hacía fallar `git add` en la carpeta de la wallet (así se descubrió: un commit que nunca ocurrió y una medición que compiló el código viejo).
- Las tres ramas quedan para revisión del auditor antes de mergear a main y publicar v0.6.1.

## 29-sep-2026 — Halving 3, envíos reales de 1,000 y 10,000, primer Platino de la comunidad

- **Halving 3** (DAA 300,000, ~09:00): recompensa 0.0625, Rodio abierto. Al principio se leyó como posible bifurcación (DAA distintos entre seed y JC); era solo el halving. `pruningPointHash` idéntico (`7e2ece39…`) en los dos nodos, seed `isSynced`, 1 peer.
- **Envíos reales con la wallet de `wallet-envios-grandes`** (`e4025d7e`): daemon de prueba en 8084 con copia de llaves (`keys-prueba-envios.json`), nunca la de producción. A JC: 1,000 RUPIX = 24 txs / 8.6 s; 10,000 RUPIX = 232 txs (231 consolidación + 1 pago) / 46.1 s. Verificado desde el seed con `GetUtxosByAddresses`: JC pasó a 13,305.75 RUPIX en su dirección. Daemon de prueba y copia de llaves borrados al terminar; balance del seed 65,853.9.
- **Primer Platino de la comunidad:** JC, `forge --level=2 --gem-address=…` desde Windows; tx `bc1bb97590ef7d81fb7c7b3c79cdce62f219b7e11ef7d4376ba3d41b650f563a`, DAA 348,162. Quedan 2 Diamantes (181,403 y 185,382); 10 quemados.
- **Comparación entre nodos:** misma consulta en el seed (python) y en la PC de JC (PowerShell) → mismas tres gemas, mismos IDs, mismos DAA; poda `7e2ece393c7d991c` en JC. Registrado en LOGROS.
- Detalles de operación aprendidos: `rupixctl GetUtxosByAddresses <dir>` toma la dirección como parámetro suelto (separadas por coma), no `--addresses`; al pegar PowerShell por chat se pierde el `_` de `$_`, usar `foreach ($u in $e)`.
- **Pendientes:** respuesta del auditor a las cuatro ramas → merge → v0.6.1 (binarios, guías sin `--password`, frase semilla al crear) → primer checkpoint real. Defensa de nodos nuevos (lista de pruning points). Web: hito del día.

## 29-sep-2026 (noche) — Visto bueno del auditor, merge, v0.6.1, checkpoint #1

- **Auditor:** visto bueno a las cuatro ramas contra `1e38bace`. Condiciones cubiertas: la estimación lineal solo guía la selección y la fee final se recalcula exacta (con relleno si falta); el mensaje de la frase semilla ya pide limpiar la pantalla. Precisiones para nodos nuevos (v0.6.2): validar la lista `MsgPruningPoints` **después** de `ArePruningPointsInValidChain`, no en `validateAndInsertPruningPoints`; no seguir los enlaces `header.PruningPoint()` uno por uno; el chequeo es "H está en la lista en su índice". Medio RUPIX: cada bloque cobra por su propio DAA; exceso acotado por ventana DAA × caída de recompensa; `MaxRupia` es tope por transacción. Recomendación: no tocar consenso; documentar "≈42M por calendario"; contador exacto como RPC fuera del sello; `TestTotalSupply` con cota.
- **Merge** de `checkpoint-dag`, `mempool-niveles`, `wallet-envios-grandes`, `wallet-claves` → `d857c54d`. Sin conflictos, build y tests en verde.
- **v0.6.1:** `appPatch=1`; checkpoint #1 de testnet en `domain/dagconfig/checkpoints.go` (blue score 86,400, hash `7e2ece393c7d991c…`, caduca DAA 2,000,000) + `TestCheckpointsPublicados`; `TestTotalSupply` por calendario con cota; CHANGELOG; CHECKPOINTS ES/EN con la regla nueva, el alcance (nodos sincronizados) y el registro; guías sin `--password` y con la frase semilla; README con la precisión de los 42M; MEMORIA "El medio RUPIX".
- **Despliegue en el seed:** binarios de `/root/bin-v061` a `/usr/local/bin`, reinicio de `rupixd-testnet`, `rupix-miner`, wallet daemon y explorer; verificar versión y que el DAA sigue.
- Pendientes que quedan: release en GitHub (tag `v0.6.1`, CI genera binarios + SHA256) → avisar a JC/JP para actualizar; validación de la lista de pruning points (v0.6.2); `rpc-gemas` + contador exacto; tarjetas de minado/quemado en explorador y web; segundo seed (JC/JP); reporte privado de vulnerabilidades; post "intenten romperlo".

- 23:21: `TestTotalSupply` salió rojo en `395ce512` por un umbral mal calculado en la cota (0.001% en vez de 0.01%; la cota real es 0.0031%) y el bloque no se detuvo porque `go test | tail` devuelve el estado de `tail`. Corregido en el commit siguiente; el tag `v0.6.1` se movió a ese commit antes de publicar la release. Regla nueva: `set -o pipefail` en todos los bloques.
- 00:30 (30-sep): release `v0.6.1` publicada desde el seed por API (`HTTP 201`); CI `Build and upload assets` en verde, seis archivos (`linux`/`win64`/`osx` `.zip` + `.sha256`). El workflow `Tests` salió rojo: llevaba semanas rojo por piezas heredadas; reproducido en el seed (gofmt/vet/suite limpios; dos SA4006 reales de staticcheck, uno un bug chico en `estimateFee`). **CI heredado retirado** a `workflows-retirados/`, workflow nuevo con lo nuestro; ver MEMORIA. Guías del repo y web ya nombran `rupix-v0.6.1-*.zip`. Pendiente: aviso a JC/JP en cuanto el nuevo `Tests` esté en verde.
- 00:50 (30-sep): **primer CI en verde** (`bb1db440`, Linux y macOS). **Compilación reproducible verificada:** el seed tiene Go 1.26.6 (igual que el CI); `v0.6.1` recompilado con los flags de `deploy.yaml` → 4/4 binarios con el mismo SHA256 que el `.zip` de la release (`rupixd` `ea973ce0d720678a…`). Queda `tools/verificar-binarios.sh`. `race.yaml` heredado (rama `master` inexistente) retirado y reemplazado por uno sobre `main`. Badge en README. Pendiente para el registro de mañana: hito en la web.

## 30-sep-2026 (día) — Halving 4 y escalera completa, checkpoint bajo la poda

- **Halving 4** cruzado ~11:50 (DAA 400,000): recompensa 0.03125, era 5, Kings abierto. `verificar-emision.py` a DAA 407,293: regla 93,977.40625, nodo 93,757.42, diferencia 219.99 (220 quemados en 22 Diamantes). La testnet #4 completó su diseño (escalera en ~4 días).
- **Poda avanzó** a `b80cafc8…` (blue 172,800). H (`7e2ece39…`, blue 86,400) sigue en el seed como chain block; journal sin rechazos. Documentado en CHECKPOINTS ES/EN.
- Estado: seed v0.6.1, 4 servicios activos, **0 peers** (JC y JP apagados; el aviso de v0.6.1 sale cuando vuelvan). CI `tests.yaml` verde en `31ad76f7`; el `race.yaml` nuevo corre por primera vez esta noche (03:17 UTC).
- Web: hitos del CI/reproducible y del halving 4 publicados hoy.
- Siguiente: aviso JC/JP + respuesta al auditor; wallet para principiantes (rama `wallet-principiantes`, devnet); Reddit; v0.6.2 (lista de pruning points, `rpc-gemas`).
- 14:20 (30-sep): **auditor, segunda ronda** (sin hallazgos en consenso ni en verificación): (1) `verificar-binarios.sh` ahora se niega a correr con otra versión de Go (antes solo avisaba: alguien con 1.25 habría "encontrado" binarios que no coinciden); (2) CHECKPOINTS ES/EN: política de renovación sin hueco (el #2 antes de que caduque el #1; para el #1, a más tardar en DAA 1,700,000); (3) el aviso a JC/JP dice que el candado protege a nodos ya sincronizados hasta v0.6.2. Su recomendación para lo que sigue: modelo de amenazas, tests adversarios por regla, fuzzing.
- **Rupix para máquinas** (idea del 30-sep, para la lista): `ESPECIFICACION.md` (cada regla de consenso como afirmación + el test que la cubre), `API.md` (gRPC de la wallet y RPC del nodo como contrato versionado), todo lo verificable por RPC (gemas, quema, checkpoint activo), y salida `--json` en los comandos. Lo que hace a Rupix legible para una inteligencia es lo mismo que la hace honesta para una persona.
- 14:45 (30-sep): **ESPECIFICACION.md, primera versión** (formato del auditor: regla · ataque que detiene · test que lo prueba). Llenada desde el código y los tests reales. Seis huecos encontrados, en orden: (1) no hay test automático de bloque con sello de gemas falso rechazado (la afirmación más fuerte del README solo se probó en vivo en v0.4.2); (2) nodo nuevo vs peer hostil (v0.6.2); (3) topes de Diamante/Platino/Rodio solo en aritmética, no en bloque de consenso; (4) exceso por frontera de halving sin test; (5) gastar una quema sin test; (6) fuzzing. El (1) es lo primero.
