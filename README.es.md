# Rupix

🇬🇧 [English version](./README.md)

🇬🇧 [English version](./README.md)

**Rupix es un activo digital escaso para todos: sin dueño, sin premine (sin monedas preguardadas por su creador), sin permiso para entrar. Con un techo de 42 millones que nadie puede cambiar, y una cantidad que solo baja. Mientras el dinero normal se imprime, Rupix se hace más escaso. Y no tienes que confiar en nadie: verifícalo.**

[rupix.network](https://rupix.network) | [@RupixNetwork](https://x.com/RupixNetwork) | [Changelog](./CHANGELOG.md) | [Gracias](./THANKS.md)

---

## Estado actual (Rupix v0.5.2)

- ✅ **Algoritmo de minado propio — RupixHeavyHash**: variante de kHeavyHash (el algoritmo de Kaspa; Rupix es un fork de kaspad bajo licencia ISC, con agradecimiento). Rupix conserva el motor probado de Kaspa (matriz 64x64, HeavyHash) y reemplaza el generador que llena la matriz (xoshiro256++) por uno propio, con una fórmula estructuralmente distinta (multiplicación no-lineal que xoshiro no tiene) y un sello "RUPIX" en la semilla. Efecto: los ASIC fabricados para Kaspa no pueden minar Rupix — su hardware produce una matriz incorrecta y la red lo rechaza. Arranque justo: minable con GPU/CPU, sin ventaja de hardware heredado. Probado en devnet (22,000+ bloques, 0 rechazos, commitment y economía intactos) y en la testnet pública. El minero (rupixminer, incluido en cada release) usa la misma función interna que el nodo, por lo que mina con RupixHeavyHash: no hay dos algoritmos, minero y validador comparten una sola fuente. Motor de Kaspa, semilla de Rupix.

**Alcance honesto:** el generador propio fue analizado de forma independiente — es lineal sobre GF(2), biyectivo (matriz de transición de rango 256, sin estados transitorios) y sin ciclos cortos en las pruebas. Se descartó el colapso catastrófico; **no** se certificó el período máximo teórico (2^256−1), lo cual exigiría verificar la primitividad del polinomio característico. Y sobre los ASIC: RupixHeavyHash impide el hardware fijo fabricado para Kaspa, lo que da una **ventaja de meses, no independencia permanente** — un FPGA puede reprogramarse en semanas. Es un arranque justo para que todos empiecen parejos, no una barrera eterna.

La **testnet pública está viva**: acepta nodos externos, mina sobre un génesis propio con cero premine, y la economía completa vive en el consenso. Cualquiera puede conectar su nodo — ver [GUIA-TESTNET.md](./GUIA-TESTNET.md).

- ✅ **Testnet pública operativa 24/7** — génesis propio, subsidy en cero, dificultad ajustándose sola, escuchando conexiones (seed público)
- ✅ **Binarios descargables** para Windows, macOS y Linux — ver [Releases](https://github.com/rupixnet/rupixd/releases/latest) — usa siempre la última versión
- ✅ **Explorador público en vivo** — [explorer.rupix.network](https://explorer.rupix.network)
- ✅ **Economía completa en consenso**: escalera de 5 niveles, quema 10:1, burn por transacción, murallas históricas (2.1M/210k/21k/2,100) — 25+ escenarios de ataque cubiertos por tests
- ✅ **Identidad completa**: direcciones `rupix:`/`rupixtest:`, llaves extendidas `rpub`/`rtub`, RPC en rupias
- ✅ **Verificación total del conteo de gemas (commitment en header)**: el conteo (Diamante/Platino/Rodio/Kings) se sella en el hash de cada bloque, protegido por el minado (PoW). La red recalcula y valida el sello al recibir cada bloque, y persiste en disco. Un conteo falso NO pasa: el sello no cuadra y se rechaza. Verificable desde el génesis, sin confiar en nadie.
- ✅ **Forja funcionando de punta a punta**: minar Gold → quemarlo → forjar una gema (Diamante) → el commitment refleja el conteo real. Probado en vivo: la forja se mina, el sello del bloque cuadra con la validación, el estado persiste. La escalera vive.
- ✅ **Diamante real sellado en la cadena** (14-sep-2026): con la testnet pasando los 100,000 bloques (halving 1, Diamante desbloqueado), se forjó un Diamante real. El commitment de la cadena refleja el conteo (sello `780e9027…`), sin discrepancia entre minero y validador. Código revisado en dos rondas de auditoría externa (huecos de verificación cerrados) y con test de regresión en el CI.
- ✅ **Primer forjador externo — comunidad real** (14-sep-2026): un segundo usuario, desde su propia computadora y su propio nodo, forjó 3 Diamantes reales. Minó Gold, quemó para forjar, y la red selló el conteo en el commitment. La verificación total funciona entre varias personas, no solo el creador.
- ✅ **Primer Platino de la red** (14-sep-2026): 10 Diamantes quemados para siempre, 1 Platino nacido. El segundo escalón de la escalera, probado en la cadena real.
- ✅ **Tercer nodo externo** (16-sep-2026): un tercer participante sincronizó su nodo desde cero con la testnet de RupixHeavyHash (v0.5.2) y recibió RUPIX. La red pública corre en 3 nodos: el servidor semilla y dos externos.
- ✅ **Binarios verificables con SHA256**: cada release publica la huella de cada binario, generada por el CI — descargas, comparas, y confirmas que nadie lo alteró
- ✅ **`go test ./...` en verde y `go vet` limpio** (20-sep-2026): la suite completa pasa. Los últimos 18 paquetes rojos no eran bugs de consenso: eran tests heredados de Kaspa con expectativas de Kaspa (prefijos `kaspa:`, recompensa de 500, orden por hash) y outputs creados con `Version = MaxScriptPublicKeyVersion`, que en Rupix es 4 = Kings — la escalera los rechazaba como Kings falsos. Regenerados desde el código, verificables.
- ✅ **0 vulnerabilidades** (govulncheck), compilado con Go 1.26.6
- ✅ **Red de más de un nodo**: primer nodo externo conectado y sincronizado, primera transacción entre dos personas registrada en la cadena

## Qué es Rupix

Rupix es una blockchain Layer 1 con consenso Proof of Work sobre un BlockDAG (no una cadena lineal). **Rupix es un fork de kaspad**: una cadena independiente, con su propia moneda, construida a partir del código abierto de Kaspa (GHOSTDAG, kHeavyHash) bajo licencia ISC. **Rupix no es parte de la red de Kaspa ni usa KAS.** Reconocemos y agradecemos ese trabajo: sin el código que el equipo de Kaspa publicó, Rupix no existiría.

Lo que Rupix añade encima:

- **Modelo económico propio de 5 niveles** con quema permanente para forjar cada nivel superior
- **Supply absoluto de 42,000,000 RUPIX**, sellado en el protocolo
- **Burn por transacción**: cada transferencia destruye rupias para siempre
- **Génesis sin premine**: el primer RUPIX se minó después del bloque 0, como Bitcoin
- **Desbloqueo por halvings**: cada nivel de la escalera se abre con un halving — la escasez tiene calendario

## La escalera — los 5 niveles

| Nivel | Nombre | Supply máximo | Cómo se forja | Se desbloquea |
|-------|--------|---------------|---------------|----------------|
| L1 | Gold | 42,000,000 | Minando | Desde el génesis |
| L2 | Diamante | 2,100,000 | Quemar 10 Gold | Halving 1 |
| L3 | Platino | 210,000 | Quemar 10 Diamante | Halving 2 |
| L4 | Rodio | 21,000 | Quemar 10 Platino | Halving 3 |
| L5 | Kings | 2,100 | Quemar 10 Rodio | Halving 4 |

Cada nivel se forja quemando 10 unidades del anterior. Crear 1 Kings implica haber destruido 10,000 Gold a lo largo de la cadena; llenar los 2,100 Kings destruiría 21 millones de Gold — la mitad de todo el que existirá jamás. La quema es irreversible y queda en la cadena para siempre. Nadie puede revertirla: ni el creador, ni los mineros, ni ningún acuerdo futuro.

## El fuego — deflación en cada uso

Cada transacción de Rupix destruye una pequeña cantidad, exigida por consenso:

```
burn = 1,000 rupias + (bytes_de_la_tx × 10 rupias)
```

Donde 1 RUPIX = 100,000,000 rupias. Estas rupias no van a un fondo ni al minero: **desaparecen**, en un output OpReturn visible para siempre e imposible de gastar. Esto ya funciona: la primera transacción de la red lo pagó.

## Por qué creemos en el trilema

El trilema de blockchain plantea que cualquier red distribuida tiene que elegir entre descentralización, seguridad y escalabilidad, y solo puede tener dos al mismo tiempo.

Rupix se construye sobre la premisa de que un BlockDAG con Proof of Work permite empujar las tres al mismo tiempo más lejos que las arquitecturas tradicionales. No decimos que el trilema esté resuelto: decimos que estamos empujándolo en una dirección que respeta los tres principios.

- **Descentralización**: PoW sin premine, código abierto, sin gobernanza centralizada, anyone-can-mine
- **Seguridad**: validación criptográfica completa, sin atajos, sin trusted parties
- **Escalabilidad**: BlockDAG permite múltiples bloques paralelos sin perder consistencia

## Verifícalo tú mismo

No confíes en nosotros. Compruébalo:

- **Que no hay premine**: `go run ./cmd/genesisgen` regenera el génesis y muestra el subsidy en cero, byte por byte
- **Que el techo es 42M**: revisa `domain/consensus/utils/constants/constants.go` (MaxRupia)
- **Que la economía tiene tests**: `go test ./domain/...` — la escalera, el burn y los Kings con sus escenarios de ataque
- **Que PoW no se puede desactivar**: `go test ./domain/dagconfig/...`
- **El supply en vivo** (con un nodo corriendo): `rupixctl GetCoinSupply` → `maxRupias: 4200000000000000` — exactamente 42,000,000 RUPIX

## Cómo correr un nodo

Requisitos: Go 1.21+, 4 GB RAM, 50 GB de disco.

```
git clone https://github.com/rupixnet/rupixd.git
cd rupixd
go build -o rupixd .
go build -o rupixminer ./cmd/rupixminer
go build -o rupixwallet ./cmd/rupixwallet
go build -o rupixctl ./cmd/rupixctl
```

Conectarte al testnet:

```
./rupixd --testnet --utxoindex
```

*(La testnet pública está viva 24/7 con seed abierto: `--addpeer=178.104.69.148:17211`. Cualquiera puede conectarse — ver [GUIA-TESTNET.md](./GUIA-TESTNET.md).)*

## Camino a mainnet

- **Minado accesible para todos — HECHO (v0.5.2)** — Rupix migró del algoritmo heredado de Kaspa a RupixHeavyHash, su propio algoritmo. Los ASIC de Kaspa ya no pueden minar Rupix; se mina desde una computadora normal (GPU/CPU). Rupix es para todos.
- ✅ **Verificación total con commitment en header** — el conteo de gemas se sella en el hash de cada bloque (protegido por PoW), se valida al recibir, y persiste en disco. Un conteo falso se rechaza: el sello no cuadra. Esto CIERRA el verificable total — probado en vivo (primera transferencia entre nodos, testnet v0.4.2).
- **Go vs Rust — riesgo declarado.** Rupix corre en kaspad-go, la implementación legacy; el desarrollo activo de Kaspa está en rusty-kaspa. kaspad-go hace todo lo que Rupix necesita hoy (GHOSTDAG, pruning, kHeavyHash) pero no recibe mejoras ni correcciones upstream. La migración a rusty-kaspa es un objetivo del segundo año, condicionado a tener contribuidores que la sostengan. No es una promesa; es una dirección declarada.
- **Auditoría externa del código de consenso**
- **Infraestructura redundante** (múltiples nodos semilla) y **hashrate comprometido**
- ✅ **Checkpoints temporales — HECHO (v0.5.2)** — con caducidad dentro del consenso, probados en devnet, política pública en [CHECKPOINTS.md](./CHECKPOINTS.md). Falta publicar el primero real.

**Fecha de mainnet: la anunciaremos cuando el código esté listo, no antes.** Preferimos lanzar tarde y bien que pronto y comprometidos.

## Filosofía

Rupix no es un fork por novedad ni por hype. Es una arquitectura económica nueva sobre un motor de consenso probado. Tomamos lo que ya estaba bien hecho (el motor GHOSTDAG) y construimos encima una propuesta económica original que apuesta por la escasez verificable y la honestidad radical.

Quien creó Rupix mina desde el bloque 1, como cualquiera. No hay direcciones privilegiadas, no hay sales, no hay rondas. La única ventaja del que llega temprano es haber estado despierto cuando arrancó la red.

## Licencia

ISC — Rupix developers, 2026.

## Reconocimientos

A los investigadores y desarrolladores que crearon y publicaron GHOSTDAG bajo licencia abierta. Su trabajo permite que proyectos como Rupix existan.

**No confíes, verifica.** ER
