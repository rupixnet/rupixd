# MEMORIA RUPIX

> La historia de Rupix contada en orden: qué pasó, por qué se decidió cada cosa y qué creímos y después cambiamos.
> Cada hecho lleva su fuente (commit, tag o archivo) para que cualquiera lo verifique.
> Es la base de lo que responderá Stevenson Rux, el asistente de Rupix.
>
> **Estado:** en construcción. Los capítulos marcados *(por escribir)* se escriben con calma, uno por uno, con revisión del fundador.
>
> **Sobre las fechas:** van en hora de la Ciudad de México. Los commits guardan la hora en UTC, así que el trabajo de la noche aparece en git con fecha del día siguiente.

## Índice

1. Por qué existe Rupix *(por escribir)*
2. El primer intento y el reinicio del 11 de junio *(por escribir; material: sección HISTORIA del CONTEXTO, commit `dc6e22c`, tag `v0.2.4-archive-experimental`)*
3. El arranque limpio del 18 de junio *(por escribir)*
4. La economía sellada *(por escribir)*
5. El génesis y la primera transacción *(por escribir)*
6. Los primeros de fuera *(por escribir)*
7. La costura: el sello de gemas y los bugs de septiembre *(por escribir)*
8. El auditor y la honestidad radical *(por escribir)*
9. v0.5: checkpoints y tests en verde *(por escribir)*
10. El test del King y la v0.6.0 *(por escribir)*
11. La comunidad de Kaspa *(por escribir)*
12. El 26 de septiembre *(por escribir)*
13. Hacia dónde va *(por escribir)*
14. Lo que creímos y después cambiamos *(empezado, abajo)*
- Apéndice: línea de tiempo *(por escribir)*

---

## 14. Lo que creímos y después cambiamos

Esta sección no se borra ni se maquilla. Cada entrada dice qué se creía, qué resultó cierto y dónde está la prueba.

### El algoritmo: RandomX y Autolykos → RupixHeavyHash
En notas de otra etapa se planeó migrar a RandomX (minado por CPU) y, después, a Autolykos v2. No se hizo. El algoritmo de Rupix es **RupixHeavyHash**: el motor de Kaspa (kHeavyHash) con generador y semilla propios desde v0.5.0 (16-sep), y dominio keccak propio desde v0.6.0. Las notas viejas quedaron marcadas como de otra etapa el 26-sep (commit `65a6fde`).

### Cómo se describe Rupix: "activo digital escaso" → "moneda digital que no controla nadie"
Hasta el 26-sep, el README, la web y el CONTEXTO decían "activo digital escaso, como el oro" y "valor para guardar". Ese día el fundador lo cambió: "no lo somos". En testnet no tiene valor de mercado, y prometer valor sería vender humo. La regla desde entonces: se dice lo que Rupix es (moneda digital, sin dueño, sin premine, 42 millones que nadie puede cambiar, cada vez más escasa) y no para qué se usa ni cuánto vale; el rumbo lo pone la gente. Commits `c9a3719` y `27f864a` (repo) y `a33b753` (web).

### La alarma de descalificados que no podía sonar
El 23-sep el auditor pidió vigilar los bloques descalificados. La alarma se hizo la noche del 25-sep (commit `0a76230`) y caminaba hacia atrás por la cadena seleccionada. Por construcción nunca podía ver un bloque descalificado, porque esos quedan fuera de esa cadena: dijo "OK" 134 veces sin poder sonar. El mensaje de su commit afirmaba que "habría cazado el bug del King el día 18"; era falso. El 26-sep el nodo empezó a registrar cada descalificación y la alarma pasó a leer ese registro; si no ve actividad, dice ERROR en vez de OK (commit `195940d`). Probada con un log falso. Falta la prueba de punta a punta en devnet.

### "El mempool no valida la escalera"
Una nota de trabajo (`NOTAS-H1-MEMPOOL.md`) decía que el mempool solo revisaba anti-spam y formato. Al leer el código el 26-sep resultó falso: al insertar una transacción, el mempool llama a la validación del consenso, que ejecuta `checkLevelRules`. Quedan pendientes un test que lo demuestre y el Nivel B (topes históricos). Commit `8167c10b`.

### El README que se quedó atrás
El 26-sep el README todavía decía que el fix del bug del King "saldría en v0.6.0", cuando ya había salido, y fechaba al tercer nodo externo (16-sep) en v0.6.0, cuando era v0.5.0. Además presentaba el primer Diamante real, el primer forjador externo y el primer Platino (14-sep) sin aclarar que fueron en testnets anteriores, no en la #4. Corregido en el commit `8167c10b`.

### El CHANGELOG sin su versión más importante
La v0.6.0 salió y el CHANGELOG pasó 6 días sin su entrada. Se agregó el 26-sep (commit `90995c9`).

### "Dos forjadores externos"
El CONTEXTO decía que había dos forjadores externos en la historia. Era uno: JC. Corregido el 26-sep (commit `b15cbf1`).

### "22 vulnerabilidades"
Un plan viejo hablaba de 22 vulnerabilidades en las dependencias; el README dice 0 según govulncheck. El 26-sep se agregó al CONTEXTO un pendiente basado en ese plan viejo, sin verificar, y se quitó el mismo día (commit `be765fc`). Queda correr govulncheck otra vez para confirmar el 0.

### La devnet olvidada
La devnet de pruebas de v0.6.0 quedó prendida dos días en el seed, con 1.8 GB de RAM y su puerto P2P abierto a internet. Se apagó el 26-sep y quedó escrita la regla: las devnets solo corren mientras se usan (commit `8d6884f`).

### La falsa alarma del explorer
El 26-sep se creyó que el explorer mostraba nombres viejos de la escalera ("Zafiro", "Esmeralda"). Venía de una copia vieja en el espacio de trabajo del asistente: el explorer real ya decía Diamante, Platino, Rodio y Kings. La lección: verificar contra la fuente viva antes de corregir.

### "Un halving cada décadas"
El comentario del parámetro más visible de la economía, en `params.go`, decía que el halving de mainnet ocurría cada "~décadas". Son **~16 meses**: 42 millones de bloques a un bloque por segundo son unos 486 días. Lo encontró el auditor en su barrido del 27-sep. La emisión siempre estuvo bien (el test `TestTotalSupply` suma exactamente 42M); lo que mentía era el comentario.

### La constante de 150 que "definía la emisión"
Al revisar ese comentario apareció `constants.BlocksPerHalving = 150`, con un comentario que decía que definía la emisión y el calendario de niveles. Nadie la usaba: el consenso toma el valor de cada red (100,000 en testnet, 42 millones en mainnet). Un revisor la habría leído como si el halving fuera cada 150 bloques. Se eliminó el 27-sep.

### La página del explorer que se quedó en 10,000
El explorer sirve su página desde una carpeta del servidor, y el repo tenía otra copia. El 27-sep resultó que la del repo estaba vieja: decía que el halving de la testnet era cada 10,000 bloques. La viva decía 100,000, que es lo correcto. Si alguien hubiera desplegado la del repo, la escalera se habría mostrado mal. Se sincronizaron y quedó escrito cuál es la fuente.

### El explorer que nadie cuidaba y el nodo v0.3.0 dormido
El 27-sep, al versionar la operación del seed, apareció que el explorer llevaba desde el 9-sep lanzado a mano: su servicio estaba muerto y tras un reinicio no habría vuelto. Y el servicio del nodo v0.3.0 seguía habilitado, apuntando a la misma carpeta y puerto que el nodo actual: al reiniciar se habrían peleado. Nadie lo vio porque todo funcionaba. Se arregló antes del reinicio del kernel y el servicio viejo se guardó, no se borró.

### El checkpoint que habría partido la red
El 18-sep los checkpoints pasaron su prueba en devnet. El 27-sep, al preparar el primero real sobre el Diamante de JC, se releyó el código: rechaza cualquier bloque con el DAA del checkpoint y otro hash. En la devnet había un solo minero y nunca hubo dos bloques con el mismo DAA; en la red real, con dos mineros, sí puede haberlos. Un checkpoint en el lugar equivocado habría dejado a los nodos nuevos sin poder sincronizar. Se encontró antes de publicar nada.

### Los 1,000 que no salían
Mandar 100 RUPIX desde el seed tardó 9 segundos; mandar 1,000 se cortaba a los 2 minutos exactos. En lugar de subir el tiempo de espera y seguir, se midió: armar el envío tardó 2.2, 6.9, 16.0, 28.6, 44.2 y 64.1 segundos para 100 a 600 RUPIX. Duplicar el monto cuadruplicaba el tiempo. La causa estaba en la wallet heredada: recalcula la comisión rearmando toda la transacción por cada pedazo que agrega, y el minero tenía 78,329 pedazos de 0.25 y 0.5. No era la red. Se explicó con números antes de arreglarlo. La noche siguiente se arregló: la comisión se estima en tiempo lineal durante la selección y la exacta se calcula una sola vez al final. Misma wallet, mismo método: los 1,000 pasaron de cortarse a los 120 segundos a salir en 0.82.

### El gitignore que escondía la wallet
El `.gitignore` tenía la línea `rupixwallet` para ignorar el binario compilado. Sin la barra inicial, esa línea también tapaba la carpeta `cmd/rupixwallet`: `git add` ahí fallaba en silencio. Se descubrió el 28-sep porque un commit del arreglo de la wallet nunca ocurrió y la medición siguiente compiló el código viejo, dando los mismos tiempos de siempre. La moraleja quedó en el método: después de cada commit se verifica en GitHub que de verdad exista, y la medición imprime de qué commit salió el binario.

### El medio RUPIX del halving 2
El 28-sep escribimos "el Gold emitido cuadró con la regla al entero exacto" y "nunca hay Gold de más". Al día siguiente, revisando los mismos números para la web: la regla daba 82,025.5, el nodo tenía 81,915.87, y con 110 quemados más una recompensa sin cobrar el nodo debía tener como máximo 81,915.38. Había 0.5 RUPIX de más. El auditor lo explicó con el código: cada bloque cobra según su propio DAA score, no según el del bloque que lo mergea; cuatro bloques minados en 199,99x y cobrados pasado el 200,000 conservan los 0.25. La emisión por era no es exactamente bloques × recompensa; es eso más un exceso acotado en cada frontera. Unos pocos RUPIX en toda la vida de la cadena, como Bitcoin tampoco emite exactamente 21M. La lección no es el medio RUPIX: es que "al entero exacto" se escribió sin haber restado todo lo que había que restar. Se corrigió en README, LOGROS y la web, y `TestTotalSupply` ahora dice "por calendario" y documenta la cota.

### El CI que decía rojo por cosas de otro
El 30-sep, verificando la release v0.6.1, el workflow `Tests` de GitHub salió en rojo. Al mirar el historial: llevaba rojo en cada push desde hacía semanas y nadie lo había leído, mientras `go test ./...` pasaba en el seed. Reproducido paso por paso: `gofmt` limpio, `go vet` limpio, la suite completa limpia; lo que fallaba eran piezas heredadas de Kaspa (sus "stability tests", un `go get -d` viejo) y, escondidos entre ellas, dos avisos reales de `staticcheck` que sí eran nuestros. Uno de ellos era un bug pequeño en la wallet: el output de quema se agregaba al mock de la transacción *después* de armarlo, así que nunca contaba en la comisión estimada; un colchón de 100 rupias lo tapaba desde hacía semanas. Un CI que siempre está rojo no avisa de nada, y lo que avisaba de verdad quedó enterrado. El workflow heredado quedó guardado en `.github/workflows-retirados/`; el nuevo corre solo lo que es de Rupix, para que el verde signifique algo.

### La verificación que no podía fallar
Al publicar la especificación, un bucle corría cada test citado y mostraba `ok`. Los diecinueve dieron `ok`. Pero cada línea decía `[no tests to run]`: el `head -1` tomaba la primera línea que devuelve `go test ./domain/...`, la del paquete raíz, donde ningún test se llama así. Una verificación que no puede fallar no verifica nada. Se repitió con `-v` buscando `--- PASS: <nombre>` y entonces sí apareció lo que había que ver: `TestPOW` no existía como test que corre, estaba en `t.Skip`. La lección es doble: la comprobación tiene que poder decir que no, y un `Skip` sale verde igual que un `PASS`.

### El test que no era idempotente
Un bloque de comandos se pegó dos veces y el script de parche insertó `DefaultAppDir` y `languageSubCmd` por segunda vez: el código dejó de compilar. La comprobación "¿ya está el texto viejo?" no sirve cuando el texto nuevo contiene al viejo. Desde entonces cada inserción comprueba la marca nueva antes de tocar nada, y una doble corrida no rompe.

### La forja se acepta en el bloque siguiente
El primer intento del test de topes decía que el Diamante 2,100,000 "no se contó": el bloque entró, pero el conteo guardado no subió. No era un bug: en un DAG las transacciones de un bloque las acepta el bloque que lo mergea, no él mismo. Ya estaba escrito en el test del mempool ("el siguiente bloque acepta las forjas") y se había olvidado. Lo que salió de reescribirlo valió más que el test: el builder honesto simula el virtual y se niega a incluir una forja que rompa el tope, y un bloque fabricado a mano con ella se rechaza entero al insertarlo.

*(30-sep, tarde: esa última frase era el hueco. Ver "El veneno del tope".)*

### El veneno del tope
"Un bloque que exceda el tope se rechaza entero" sonaba a la regla más dura posible. El auditor vio lo que eso permite: si el bloque envenenado llega a ser punta, cada bloque honesto que lo mergee o construya encima se rechaza con él. Una forja de 10 Gold para tirar a los mineros honestos. La regla correcta es la que ya usábamos para toda transacción inválida por UTXO: el bloque entra y la transacción, sola, no se acepta. Sin efectos colaterales: el Gold sigue vivo, el sello no cambia, nadie más pierde. La lección: la severidad de un rechazo no es una virtud; en un DAG, rechazar un bloque es castigar a quien lo mergea, y eso lo puede provocar cualquiera.

### Archivos completos, no parches
El primer arreglo del tope se aplicó con un heredoc y varios `sed` encadenados en el seed. Uno de ellos no encontró su ancla, no dijo nada, y el código quedó a medias: `mergeSetHashes[0]` en un mergeset vacío, pánico en el primer bloque después del génesis. ER lo dijo claro: "no me gusta parchar en términos breves". Desde entonces, cuando se toca consenso, viaja el archivo completo, con su hash comprobado al llegar, y el test se corre con `-v` esperando `--- PASS: <nombre>`. Un parche que puede fallar en silencio no es más rápido: es más lento, porque el fallo aparece después y lejos.

### "Todavía nadie preguntó"
Al pasar `TestTopeDeDiamantesEnBloqueReal` a la regla nueva, dos rondas se fueron en afirmar que un bloque era UTXOValid cuando el consenso decía `UTXOPendingVerification`. No era un veredicto: en un DAG, el UTXO de un bloque solo se resuelve si está en la cadena seleccionada; un hermano que empata en peso, o un hijo construido solo sobre él, se quedan en "todavía nadie preguntó". La prueba correcta hace lo que haría un minero honesto: construir sobre la punta actual **y** el bloque dudoso a la vez; ese bloque es el más pesado, el virtual lo sigue, y ahí sí hay respuesta. El auditor lo dijo mejor: *el estado de un bloque fuera de la cadena seleccionada no es veredicto; es la trampa clásica al escribir tests sobre un DAG, y dos rondas es barato por aprenderla.*

### El exit que decía dos cosas
`go test ./domain/... | grep -v '^ok'` con `pipefail` devuelve 1 cuando el `grep` no encuentra nada (todo ok) y también cuando `go test` falla. La misma cifra para el mejor y el peor caso. Y en ese mismo bloque un tecleo (`&<`) mandó la cadena a segundo plano y corrió la suite en paralelo con la copia del archivo. Se repitió con `go test` a un archivo y su código de salida solo, y el filtro aparte. Es la tercera vez que la lección es la misma: una verificación tiene que poder decir que no, y tiene que decir solo una cosa.

### Los trabajadores huérfanos del fuzzer
La primera noche de fuzzing el seed llegó a carga 22 sobre 4 núcleos. `pkill -f fuzz-noche.sh` mataba el script, pero los procesos que de verdad gastan CPU se llaman `.test -test.fuzzworker`, y cada reinicio dejaba los anteriores vivos. Tres arranques, nueve huérfanos. La lección es de siempre con otra cara: antes de reiniciar algo, comprobar que lo anterior murió de verdad (`pgrep` del nombre real, no del que uno cree), y correr lo pesado con `nice` y en su propia copia de trabajo (`git worktree`) para no pisar la rama en la que se sigue trabajando.

### Repetir una verdad no la hace más verdad
El 2-oct, ER lo dijo claro: "ya me lo dijiste muchas veces". Tenía razón. Una verdad incómoda (el revisor con nombre) dicha tres veces en una noche deja de ser información y se vuelve ruido; peor, hace que lo demás suene negativo cuando no lo es. Lo que hay que decir se dice una vez, bien dicho, se escribe donde corresponde (MAINNET.md), y no se vuelve a traer hasta que haya novedad. La honestidad incluye saber callarse.

---

*No confíes, verifica.*
