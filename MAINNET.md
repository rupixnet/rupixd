# Luz verde para mainnet: qué tiene que ser verdad, y cómo lo verifica cualquiera

**BORRADOR (30-sep-2026, v0.6.2, testnet #4 en su día 30). No está en vigor.** Lo escribió Stevenson para discutirlo con ER; los números, la postura sobre checkpoints en mainnet y la dificultad del día 1 están por decidir. Entra en vigor cuando se quite esta línea.

Rupix no tiene fecha de mainnet. Tiene **criterios**. Cuando todos se cumplan, se declara la luz verde en público (este archivo, la web, GitHub, Reddit, X), y **desde ese día corren dos meses** de aviso antes de que exista el primer bloque de mainnet. La fecha sale de los criterios; nunca al revés. Si durante los dos meses algo deja de cumplirse, el reloj se detiene y se dice por qué.

Esto es lo contrario de lo habitual, y a propósito: nadie se entera de Rupix mainnet por un rumor, nadie llega tarde porque no supo, y nadie, incluido el fundador, tiene ventaja de minuto cero.

---

## 1. Los criterios (todos, no la mayoría)

Cada criterio dice qué tiene que ser verdad, con número, y **cómo lo comprueba un desconocido** sin confiar en nosotros.

### Código y verificación

| # | Qué tiene que ser verdad | Cómo se verifica | Estado al 30-sep |
|---|---|---|---|
| C1 | **Especificación completa**: cada regla de consenso en `ESPECIFICACION.md` con el ataque que detiene y el test que lo prueba, **sin huecos abiertos**. | Correr cada test citado: `go test -v -run <Nombre> ./...` debe dar `--- PASS`. La lista "Los huecos" vacía. | 7 de 8 cerrados (#0–#7). Falta #6 (fuzzing). |
| C2 | **Fuzzing** de `topes.Cabe`, `CalculateGemsCommitment`, RupixHeavyHash/`computeRank` y el parseo de `MsgPruningPoints`, con corpus en el repo y **≥ 24 horas acumuladas por objetivo sin fallo**. | `go test -fuzz=<Objetivo>` con el corpus de `testdata/fuzz/`; el registro de horas en `ESPECIFICACION.md`. | Pendiente. |
| C3 | **CI en verde** (`Tests` y `race` nocturno) sobre `main` durante las **4 semanas** previas a la luz verde, sin excepciones silenciadas. | Pestaña Actions del repo; `.github/workflows/`. | Verde desde el 30-sep (día 1 de 28). |
| C4 | **Binarios reproducibles** en cada release etiquetada: 4 de 4 SHA256 iguales. | `tools/verificar-binarios.sh vX.Y.Z` con Go 1.26.6. | v0.6.1 y v0.6.2: 4/4. |
| C5 | **Sin cambios de consenso** en las **6 semanas** previas a la luz verde. Si hay uno, el reloj de C5 vuelve a cero (y C3 también). | `CHANGELOG.md`: ninguna entrada marcada *cambio de consenso* en ese periodo. | Último: v0.6.2 (30-sep). Reloj corre desde ahí. |

### Revisión externa

| # | Qué tiene que ser verdad | Cómo se verifica | Estado al 30-sep |
|---|---|---|---|
| R1 | **Revisor con nombre**: al menos una persona con identidad pública y reputación verificable que haya leído `ESPECIFICACION.md`, compilado, corrido la suite y publicado su revisión **con su nombre**, sin hallazgos bloqueantes abiertos. | Enlace a su revisión pública desde el README; los hallazgos y su cierre en `CHANGELOG.md`. | Pendiente. (El auditor actual es anónimo: su trabajo cuenta, su firma no basta.) |
| R2 | **Reporte privado de vulnerabilidades** activo en GitHub y `SECURITY.md` con tiempos de respuesta públicos. | Pestaña Security del repo. | Pendiente (lo activa ER). |
| R3 | **Periodo de "rómpelo"**: ≥ 4 semanas con la invitación pública abierta (Reddit, X) a atacar la testnet, con los reportes recibidos y su respuesta publicados. | Hilo público + `LOGROS.md`. | Pendiente. |

### Red y operación

| # | Qué tiene que ser verdad | Cómo se verifica | Estado al 30-sep |
|---|---|---|---|
| N1 | **Al menos 3 seeds** independientes (distinto operador, distinto proveedor), en la lista de `DNSSeeds`/seeds del código. | `domain/dagconfig/params.go`; `rupixctl GetConnectedPeerInfo` desde un nodo nuevo. | 1 seed. |
| N2 | **≥ 10 nodos externos** sincronizados y en línea de forma sostenida (≥ 4 semanas), de **≥ 5 operadores** distintos. | Conteo de peers desde los seeds, publicado semanalmente en `LOGROS.md`. | 2 externos (JC, JP), intermitentes. |
| N3 | **Hashrate externo sostenido**: ≥ 50 % de los bloques minados por nodos que no son del fundador durante ≥ 4 semanas. | `tools/` (script de reparto de coinbase por dirección, pendiente de escribir) sobre el nodo de cualquiera. | JC llegó a ~40 % un día; no sostenido. |
| N4 | **Sincronización desde cero** verificada por ≥ 3 personas distintas en v-final, con el mismo `pruningPointHash` que los seeds. | Sus salidas en `LOGROS.md`. | 1 (JP, en v0.5.0; falta repetir en la versión final). |
| N5 | **Operación documentada** para que un tercero levante un seed sin preguntarnos: `OPERACIONES.md` probado por alguien externo. | Testimonio en `LOGROS.md`. | Documento existe; no probado por externo. |

### Economía y lanzamiento justo

| # | Qué tiene que ser verdad | Cómo se verifica | Estado al 30-sep |
|---|---|---|---|
| E1 | **Génesis de mainnet publicado con ≥ 2 semanas de anticipación** a la fecha: hash, subsidio **0**, sin premine, sin direcciones precargadas. | El bloque génesis en el código, byte por byte; `TestGenesis*`. | Pendiente (se hace dentro de los dos meses). |
| E2 | **Parámetros de mainnet congelados** y explicados en `matematica.html`/`ESPECIFICACION.md`: 42,000,000 de techo, 0.5 por bloque, halving cada 42,000,000 de bloques, escalera por halving, topes 2,100,000/210,000/21,000/2,100. | `domain/dagconfig/params.go` (`MainnetParams`) y `TestTotalSupply`. | Definidos; falta congelarlos con el génesis. |
| E3 | **Checkpoints de mainnet**: política publicada antes del lanzamiento (si habrá, cuántos, cuándo caducan, quién puede proponerlos). Hoy la postura es: **sin checkpoints en el génesis**; solo si la red los necesita y siempre con caducidad. | `CHECKPOINTS.md`. | Postura escrita aquí; falta la política formal. |
| E4 | **Dos meses de aviso**, día por día, desde la luz verde hasta el bloque 1, con la fecha y hora UTC fijas publicadas en todos los canales el mismo día. | Este archivo, con la fecha. | — |

---

## 1b. Lo que Bitcoin hizo en 2009–2014, y qué copiamos

Bitcoin no se protegió en sus inicios con criptografía: se protegió con cuatro cosas. (1) **No valía nada y nadie miraba**: un 51 % era trivial y nadie lo hizo porque no había nada que ganar; el ataque llega con el valor, y para entonces ya había gente. (2) **Checkpoints en el código**: el cliente trajo bloques fijos (11,111; 33,333; 74,000…) desde 2010 hasta 2014, puestos por el propio Satoshi, y los fue quitando cuando el hashrate ya no los necesitaba; es nuestra política de checkpoints temporales con caducidad, con la diferencia de que nosotros escribimos la regla de retiro (`CHECKPOINTS.md`). (3) **Un fundador con la mayoría del hashrate que no abusó**: Satoshi minó cerca de un millón de BTC el primer año y nunca lo usó contra la red; la lección para Rupix es que un fundador con la mayoría es un riesgo aunque sea honesto, porque la gente no verifica intenciones, verifica hashrate; de ahí N3. (4) **Una comunidad chica que coordinaba a mano**: el bug del "value overflow" de 2010 (184 mil millones de BTC de la nada) se arregló en horas con una versión nueva y una reorganización a propósito; nuestro equivalente es `OPERACIONES.md`, el aviso a los nodos y la regla de que todo cambio de consenso sale etiquetado y anunciado.

Lo que Bitcoin no tuvo y Rupix sí: una especificación con cada regla y su test, builds reproducibles y una regla escrita de cuándo se quitan los candados. No sustituye a la gente; hace que la gente buena, cuando llegue, pueda confiar sin confiar.

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

## 4. Qué pasa el día 1

- Bloque 1 a la hora anunciada. Nadie mina antes: el génesis fija el tiempo y la dificultad inicial se calibra para que las primeras horas no regalen la emisión a quien tenga más máquinas listas (se documenta en E2).
- El fundador mina con lo mismo que cualquiera y lo dice. Cero premine se puede verificar en el génesis; cero ventaja de arranque se verifica mirando quién minó los primeros bloques.
- La testnet #4 sigue viva como red de pruebas mientras alguien la use.

---

*Somos todos Rupix. No confíes, verifica.*
