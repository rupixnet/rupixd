# Luz verde para mainnet: qué tiene que ser verdad, y cómo lo verifica cualquiera

**Aprobado por ER el 15-nov-2026 (v0.7.0, testnet #4 en su día 51). Entra en vigor el Día 0, con la testnet #5.** Lo escribió Stevenson el 30-sep como borrador; el 15-nov ER fijó los números (20 nodos / 8 personas, consenso binario, checkpoints de mainnet con retiro escrito, la única excepción de los dos meses) y la sección 0 es el texto público. Las tablas de abajo son el detalle verificable de cada regla.

Rupix no tiene fecha de mainnet. Tiene **criterios**. Cuando todos se cumplan, se declara la luz verde en público (este archivo, la web, GitHub, Reddit, X), y **desde ese día corren dos meses** de aviso antes de que exista el primer bloque de mainnet. La fecha sale de los criterios; nunca al revés. Si durante los dos meses algo deja de cumplirse, el reloj se detiene y se dice por qué.

Esto es lo contrario de lo habitual, y a propósito: nadie se entera de Rupix mainnet por un rumor, nadie llega tarde porque no supo, y nadie, incluido el fundador, tiene ventaja de minuto cero.

---

## 0. Las 17 reglas, en llano (aprobadas por ER el 15-nov-2026, regla 12 bajada el 23-nov; en vigor desde el Día 0, con la testnet #5)

**Testnet #5 es la fase final. Rupix saldrá a mainnet cuando se cumpla TODO esto, y no antes.** No hay fecha: la pone la red cuando cumpla las reglas. Las 17 van en una sola lista: las primeras siete dependen de nosotros, de la 8 a la 13 de la red, y de la 14 a la 17 son promesas para el día que todo se cumpla. Donde dice "medido en 4 semanas" no es un plazo, es la vara: para que "sostenido" no se cumpla con una foto de una tarde.

1. **Cada regla, escrita y probada.** Cada regla de consenso está en `ESPECIFICACION.md` con el ataque que detiene y el test que lo demuestra. Sin huecos abiertos.
2. **La versión candidata no toca el consenso.** Si se cambia una sola regla de consenso, la Prueba pública vuelve a empezar desde cero. Es binario: cero cambios o se reinicia.
3. **Pruebas automáticas en verde**, incluidas las nocturnas, durante toda la Prueba pública, sin excepciones escondidas.
4. **Binarios que cualquiera puede reproducir**: compilas el código en el sistema declarado y obtienes, byte por byte, lo que publicamos. 4 de 4, siempre.
5. **Búsqueda automática de fallos (fuzzing)** sobre las piezas críticas, con más de 24 horas acumuladas en cada una y los resultados publicados.
6. **Operación documentada**: cualquiera levanta un seed siguiendo `OPERACIONES.md`, sin preguntarnos; probado por alguien externo.
7. **Canal privado para reportar vulnerabilidades**, con tiempos de respuesta públicos (`SECURITY.md`).

8. **Al menos 3 seeds** de operadores distintos, en proveedores distintos.
9. **Al menos 20 nodos ajenos, de al menos 8 personas distintas**, en línea de forma sostenida (medido en 4 semanas seguidas, visible en el mapa). Si vamos en 12, se dice "vamos en 12" y se sigue; la cifra no se baja.
10. **Más de la mitad de los bloques minados por gente que no somos nosotros**, sostenido (medido en 4 semanas seguidas). Mientras el fundador tenga la mayoría del hashrate, no hay mainnet.
11. **Tres personas distintas sincronizan desde cero** la versión candidata y obtienen el mismo punto de poda que los seeds.
12. **Revisión abierta.** La especificación está publicada para que cualquiera la lea, compile y corra las pruebas durante toda la Prueba pública; cada hallazgo recibido, de quien sea, se publica con su respuesta. Si una persona con nombre y reputación pública publica su revisión, se enlaza desde el README. No se paga ni se regala nada por ello; el único reconocimiento es público y verificable. (El 23-nov-2026 ER bajó esta regla de «revisor con nombre obligatorio» a «revisión abierta»: un requisito que no depende de nadie no puede bloquear mainnet para siempre; la puerta sigue abierta.)
13. **Prueba pública**: invitación abierta a romper la testnet, con cada reporte recibido y su respuesta publicados (medido: al menos 4 semanas con la invitación abierta).

14. **Dos meses de aviso**, con fecha y hora UTC fijas, publicadas el mismo día en todos los canales, y **no se detiene**. Única excepción, escrita aquí de antemano: una vulnerabilidad que ponga fondos en riesgo; si aparece, se dice en público, se arregla, y los dos meses empiezan de nuevo. Pocos nodos, mercado o miedo no detienen nada.
15. **Génesis publicado al menos dos semanas antes**: sin premine, sin direcciones precargadas, subsidio cero en el bloque 0. Cualquiera lo verifica.
16. **Parámetros congelados y explicados**: 42,000,000 de techo, 0.5 por bloque, halving cada 42,000,000 de bloques, la escalera de gemas.
17. **Checkpoints de mainnet, con fin escrito antes de nacer.** Mainnet arranca sin ningún checkpoint en el código. Durante el primer año puede haber checkpoints temporales: uno cada ~30 días de bloques (2,592,000), propuesto en público por cualquiera, y solo entra si lo confirman al menos 2 nodos ajenos; cada lista caduca a los 2 meses. Se sueltan del todo cuando pase lo primero de: el bloque 31,536,000 (un año) o tres meses seguidos con más del 80 % del hashrate fuera del fundador. Soltar es una versión sin ese código, anunciada como cualquier release. Después, nunca más, pase lo que pase.

**Lo que no va a pasar:** no hay fecha hasta que se cumplan las 17. No se regala nada por participar. No se pide dinero ni seeds a nadie. Nadie puede congelar ni revertir fondos, ni nosotros. Si algo de esto cambia, se escribe con fecha en este archivo antes que en cualquier otro lado.

*No confíes, verifica. ER y Claude.*

---

## 1. Los criterios (todos, no la mayoría)

Cada criterio dice qué tiene que ser verdad, con número, y **cómo lo comprueba un desconocido** sin confiar en nosotros.

### Código y verificación

| # | Qué tiene que ser verdad | Cómo se verifica | Estado al 30-sep |
|---|---|---|---|
| C1 | **Especificación completa**: cada regla de consenso en `ESPECIFICACION.md` con el ataque que detiene y el test que lo prueba, **sin huecos abiertos**. | Correr cada test citado: `go test -v -run <Nombre> ./...` debe dar `--- PASS`. La lista "Los huecos" vacía. | 8 de 8 huecos cerrados (#0–#7); el #6 (fuzzing) tiene sus objetivos y corpus en el repo y sus horas en `ESPECIFICACION.md` §6 (28-oct-2026). |
| C2 | **Fuzzing** de `topes.Cabe`, `CalculateGemsCommitment`, RupixHeavyHash/`computeRank` y el parseo de `MsgPruningPoints`, con corpus en el repo y **≥ 24 horas acumuladas por objetivo sin fallo**. | `go test -fuzz=<Objetivo>` con el corpus de `testdata/fuzz/`; el registro de horas en `ESPECIFICACION.md`. | Casi: 28 h en `Cabe`, `CalculateGemsCommitment`, `GenerateMatrix` y `ComputeRank`; 26 en `MsgPruningPoints`; 24 en `CheckLevelRules`; **22 en `KaspadMessage`** (noche #9 terminada el 25-oct, 0 fallos). Falta una corrida de ≥ 2 h de `KaspadMessage`. |
| C3 | **CI en verde** (`Tests` y `race` nocturno) sobre `main` durante las **4 semanas** previas a la luz verde, sin excepciones silenciadas. | Pestaña Actions del repo; `.github/workflows/`. | Verde desde el 30-sep; el reloj de 4 semanas se cuenta desde el último cambio de consenso (v0.6.2, 30-sep) y se reinicia si hay otro. |
| C4 | **Binarios reproducibles** en cada release etiquetada: 4 de 4 SHA256 iguales, contra un **sistema declarado y fijo** (Linux: Ubuntu 24.04 + Go 1.26.6 + gcc 13.x; `go-secp256k1` y `go-muhash` son C, así que el enlazador del sistema entra al binario). Pendiente (v0.7.x): contenedor con imagen fijada por digest. | `tools/verificar-binarios.sh vX.Y.Z` en Ubuntu 24.04 con Go 1.26.6 (exige ambos). | v0.6.1–v0.6.6: 4/4 (con `ubuntu-latest`, que entonces era 24.04); **v0.6.7: 4/4 DIFIERE** (runner pasó a GCC 15.2); v0.6.8 no compiló; v0.6.9 fija el sistema; **v0.7.0: 4/4 IGUAL** (14-nov-2026). |
| C5 | **Cero cambios de consenso en la versión candidata.** Si se cambia una sola regla de consenso, la Prueba pública (R3) vuelve a empezar desde cero. Binario, sin ventanas (ER, 15-nov-2026). | `CHANGELOG.md`: ninguna entrada marcada *cambio de consenso* en ese periodo. | Último: v0.6.2 (30-sep). Reloj corre desde ahí. |

### Revisión externa

| # | Qué tiene que ser verdad | Cómo se verifica | Estado al 30-sep |
|---|---|---|---|
| R1 | **Revisión abierta** (bajada el 23-nov-2026 de «revisor con nombre obligatorio»): la especificación publicada para que cualquiera la lea, compile y corra la suite durante toda la Prueba pública; cada hallazgo, de quien sea, publicado con su respuesta. Si una persona con nombre y reputación pública publica su revisión, se enlaza desde el README. Nada se paga ni se regala. | Discussions/issues públicos + `CHANGELOG.md` con cada hallazgo y su cierre; enlace a revisiones con nombre desde el README. | Bajada por ER el 23-nov-2026: un requisito que no depende de nadie no puede bloquear mainnet para siempre; la puerta sigue abierta. |
| R2 | **Reporte privado de vulnerabilidades** activo en GitHub y `SECURITY.md` con tiempos de respuesta públicos. | Pestaña Security del repo. | **Cumplido el 7-oct-2026**: Private vulnerability reporting activado por ER (la pestaña Security muestra "Report a vulnerability"); `SECURITY.md` ya existía. |
| R3 | **Periodo de "rómpelo"**: ≥ 4 semanas con la invitación pública abierta (Reddit, X) a atacar la testnet, con los reportes recibidos y su respuesta publicados. | Hilo público + `LOGROS.md`. | Pendiente. Nombre público acordado el 1-nov: **«Prueba pública»** (4 semanas), dentro del **Día 0 del camino a mainnet**; se abre cuando la v0.7.0 (mejoras para el usuario) esté publicada. Reportes: pestaña Security (privado) o Issues (público); cada reporte y su respuesta se publicarán en `LOGROS.md`. |

### Red y operación

| # | Qué tiene que ser verdad | Cómo se verifica | Estado al 30-sep |
|---|---|---|---|
| N1 | **Al menos 3 seeds** independientes (distinto operador, distinto proveedor), en la lista de `DNSSeeds`/seeds del código. | `domain/dagconfig/params.go`; `rupixctl GetConnectedPeerInfo` desde un nodo nuevo. | 1 seed (Núremberg, Hetzner). El 2.º se pospuso el 28-oct por costo; **compromiso de ER (1-nov): se paga en EE. UU. al publicar v0.7.0** (Contabo, otro proveedor, ~$6.5/mes). El 3.º debe ser de otro operador. |
| N2 | **≥ 20 nodos externos**, de **≥ 8 operadores** distintos, sincronizados y en línea de forma sostenida (medido: 4 semanas seguidas, visible en el mapa). La cifra no se baja nunca (ER, 15-nov-2026: antes decía 10/5). | Conteo de peers desde los seeds, publicado semanalmente en `LOGROS.md`. | 3 externos en el libro de nodos (JC #2 intermitente desde Windows; nodo #3 anónimo, 12 min el 24-oct; JP antes de v0.6.0). Ninguno sostenido 4 semanas. |
| N3 | **Hashrate externo sostenido**: ≥ 50 % de los bloques minados por nodos que no son del fundador durante ≥ 4 semanas. | `tools/` (script de reparto de coinbase por dirección, pendiente de escribir) sobre el nodo de cualquiera. | JC llegó a ~40 % un día; no sostenido. |
| N4 | **Sincronización desde cero** verificada por ≥ 3 personas distintas en v-final, con el mismo `pruningPointHash` que los seeds. | Sus salidas en `LOGROS.md`. | 2 (JP en v0.5.0; JC en v0.6.6 el 19-oct-2026, mismo `pruningPointHash` que el seed, confirmó el checkpoint #3 antes de publicarse). La sincronización de ER desde Windows (14-nov-2026, nodo #5) no cuenta: misma persona. Falta repetir en la versión final. |
| N5 | **Operación documentada** para que un tercero levante un seed sin preguntarnos: `OPERACIONES.md` probado por alguien externo. | Testimonio en `LOGROS.md`. | Documento existe; no probado por externo. |

### Economía y lanzamiento justo

| # | Qué tiene que ser verdad | Cómo se verifica | Estado al 30-sep |
|---|---|---|---|
| E1 | **Génesis de mainnet publicado con ≥ 2 semanas de anticipación** a la fecha: hash, subsidio **0**, sin premine, sin direcciones precargadas. | El bloque génesis en el código, byte por byte; `TestGenesis*`. | Pendiente (se hace dentro de los dos meses). |
| E2 | **Parámetros de mainnet congelados** y explicados en `matematica.html`/`ESPECIFICACION.md`: 42,000,000 de techo, 0.5 por bloque, halving cada 42,000,000 de bloques, escalera por halving, topes 2,100,000/210,000/21,000/2,100. | `domain/dagconfig/params.go` (`MainnetParams`) y `TestTotalSupply`. | Definidos; falta congelarlos con el génesis. |
| E3 | **Checkpoints de mainnet, con fin escrito antes de nacer**: mainnet arranca sin ninguno en el código; durante el primer año puede haber temporales, uno cada ~30 días de bloques (2,592,000), propuesto en público por cualquiera y solo si lo confirman ≥ 2 nodos ajenos, cada lista caduca a los 2 meses; se sueltan del todo al bloque 31,536,000 (un año) o tras tres meses seguidos con > 80 % del hashrate fuera del fundador, lo primero; después, nunca más. | `CHECKPOINTS.md` (política) y el código de cada release. | Política fijada por ER el 15-nov-2026 (sección 0, regla 17). |
| E4 | **Dos meses de aviso**, día por día, desde la luz verde hasta el bloque 1, con la fecha y hora UTC fijas publicadas en todos los canales el mismo día, y **sin detenerse**: única excepción, una vulnerabilidad que ponga fondos en riesgo (se dice en público, se arregla y los dos meses empiezan de nuevo). | Este archivo, con la fecha. | Excepción única fijada por ER el 15-nov-2026. |

---

## 1b. Gracias a Bitcoin: lo que aprendimos de sus primeros años

Rupix existe porque Bitcoin demostró, entre 2009 y 2014, que una red de dinero puede nacer sin permiso de nadie y sostenerse con gente que verifica. De esos años tomamos, con agradecimiento, cuatro cosas: los **checkpoints temporales** que el cliente trajo desde 2010 y se retiraron cuando la red ya no los necesitaba (de ahí nuestra política con caducidad escrita en `CHECKPOINTS.md`); la lección de que **ningún participante debe concentrar el hashrate**, ni siquiera uno honesto, porque la gente verifica hashrate y no intenciones (de ahí N3); la **coordinación abierta y rápida** ante un error de consenso, con una versión nueva anunciada a todos (de ahí `OPERACIONES.md` y la regla de que todo cambio de consenso sale etiquetado y avisado); y la idea de que la seguridad llega con la gente, no antes que ella, por lo que estos criterios piden personas y no fechas.

Todo lo que Rupix escribe (la especificación con cada regla y su test, los builds reproducibles, la regla de retiro de los candados) es trabajo hecho sobre ese camino que Bitcoin abrió. No lo sustituye ni lo mejora: lo agradece y lo continúa.

## 2. Qué NO es criterio

- **Precio, exchanges, "listados".** Rupix no promete ninguno, no paga por ninguno y no los necesita para arrancar.
- **Cantidad de seguidores.** Diez nodos honestos valen más que diez mil cuentas.
- **La wallet gráfica (v0.7) ni el pool (v0.8).** Son comodidad; mainnet arranca con la wallet de terminal si hace falta. Comodidad nunca va antes que seguridad.
- **Que todo esté perfecto.** Que todo lo que afirmamos sea **verificable**, sí.

## 3. Qué pasa el día de la luz verde

1. Este archivo cambia de "criterios" a "cumplidos", con la evidencia de cada fila y la **fecha y hora UTC del bloque 1**, exactamente dos meses después.
2. El mismo texto se publica en la web, en GitHub (release notes de la versión candidata), en Reddit y en X, el mismo día.
3. Durante los dos meses: génesis publicado (E1), versión final etiquetada y reproducible (C4), guías de minado y de nodo para mainnet, al menos un ensayo público de "día 1" en una devnet con la misma versión.
4. Si algo deja de cumplirse (un hallazgo bloqueante, un cambio de consenso, un seed caído sin reemplazo), el reloj se detiene, se explica en este archivo y se reanuda cuando vuelva a cumplirse. Preferimos un retraso explicado a un lanzamiento con asterisco.

## 3b. Cómo cambia Rupix (la regla para cambiar las reglas; escrita el 23-nov-2026)

Las 17 reglas dicen cómo se abre mainnet. Esta sección dice cómo puede cambiar Rupix **después**, para que la respuesta nunca sea «quien tenga las llaves de GitHub».

Un cambio de **consenso** (reglas de validez de bloques y transacciones, emisión, halvings, techo, escalera de gemas, quema por transacción, checkpoints) solo entra así, y en este orden:

1. **Propuesta pública** (issue en el repo) con el problema o ataque que resuelve, escrito para que lo entienda alguien que no programa.
2. **Test** que demuestra el problema antes del cambio y su ausencia después, en el repo, en la misma propuesta.
3. **Dos meses de aviso** desde que la propuesta se declara aceptada hasta que entra en vigor, publicados el mismo día en todos los canales, con la altura de bloque o el DAA de activación.
4. **Confirmación de al menos 2 nodos ajenos** en testnet con la versión que lo trae, antes de etiquetarla para mainnet.
5. **Versión etiquetada y reproducible** (4 de 4), con el cambio descrito en `CHANGELOG.md`, `ESPECIFICACION.md` y en las notas del release.

**Ninguna persona es excepción**: ni el fundador, ni quien herede las llaves, ni quien tenga más hashrate o más monedas. Tener las llaves de GitHub da el poder de publicar, no el de saltarse esto. Una versión que cambie consenso sin este camino no es «una actualización de Rupix»: es otra moneda, y cualquiera puede señalarlo contra este documento; los nodos que verifiquen no la seguirán.

Lo que **no** es consenso (explorador, wallet, web, instaladores, mapa) cambia con el procedimiento normal: test, `CHANGELOG.md`, release. Y lo que nunca cambia, pase lo que pase y cambien las manos que cambien: nada se regala ni se paga por participar; nunca se pide dinero ni seeds; nunca se publica ni se guarda la IP de un usuario; nadie puede congelar ni revertir fondos; todo cambio se escribe con fecha antes de hacerse.

---

## 4. Qué pasa el día 1

- Bloque 1 a la hora anunciada. Nadie mina antes: el génesis fija el tiempo y la dificultad inicial se calibra para que las primeras horas no regalen la emisión a quien tenga más máquinas listas (se documenta en E2).
- El fundador mina con lo mismo que cualquiera y lo dice. Cero premine se puede verificar en el génesis; cero ventaja de arranque se verifica mirando quién minó los primeros bloques.
- La testnet #4 sigue viva como red de pruebas mientras alguien la use.

---

*Somos todos Rupix. No confíes, verifica.*
