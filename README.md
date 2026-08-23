# Rupix

**Rupix existe para ser dinero de todos: sin dueño, sin premine (sin monedas preguardadas por su creador), sin permiso para entrar. Con un techo de 42 millones que nadie puede cambiar — y una cantidad que baja cada vez que se usa. Mientras el dinero normal se imprime, Rupix se hace menos. Y no tienes que confiar en nadie: verifícalo.**

[rupix.network](https://rupix.network) | [@RupixNetwork](https://x.com/RupixNetwork) | [Changelog](./CHANGELOG.md)

---

## Estado actual (agosto 2026)

La testnet está **viva**: un nodo corre 24/7 como servicio, minando sobre un génesis propio con cero premine, y la economía completa vive en el consenso.

- ✅ **Génesis propio minado el 3 de agosto de 2026** — subsidy en cero, byte por byte, regenerable con `cmd/genesisgen`
- ✅ **Testnet operativa 24/7** — 200,000+ bloques minados, dificultad ajustándose sola
- ✅ **Primera transacción de la historia (19 de agosto)** — 4 txs, cada una quemando por ley; el nodo rechazó primero la que no quemaba (`ErrInsufficientBurn`) y aceptó las que cumplen
- ✅ **Economía completa en consenso**: escalera de 5 niveles, quema 10:1, burn por transacción, muralla de 2,100 Kings — 18 escenarios de ataque cubiertos por tests
- ✅ **Identidad completa**: direcciones `rupix:`/`rupixtest:`, llaves extendidas `rpub`/`rtub`, RPC en rupias

## Qué es Rupix

Rupix es una blockchain Layer 1 con consenso Proof of Work sobre un BlockDAG (no una cadena lineal). Está construida sobre la arquitectura GHOSTDAG, el protocolo de consenso desarrollado por el equipo de investigación de DAGLabs y publicado en código abierto bajo licencia ISC. Reconocemos y agradecemos ese trabajo: sin esa base, Rupix no existiría.

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
go build -o rupixminer ./cmd/kaspaminer
go build -o rupixwallet ./cmd/kaspawallet
go build -o rupixctl ./cmd/kaspactl
```

Conectarte al testnet:

```
./rupixd --testnet --utxoindex
```

*(La testnet pública con nodos semilla abiertos está en camino — hoy la red corre en modo laboratorio.)*

## Camino a mainnet

- Explorador público (en construcción — con contador de ceniza, supplies por nivel y countdown de halving)
- Testnet pública sincronizable por cualquiera
- Auditoría externa del código de consenso
- Infraestructura redundante (múltiples nodos semilla)
- Hashrate inicial comprometido

**Fecha de mainnet: la anunciaremos cuando el código esté listo, no antes.** Preferimos lanzar tarde y bien que pronto y comprometidos.

## Filosofía

Rupix no es un fork por novedad ni por hype. Es una arquitectura económica nueva sobre un motor de consenso probado. Tomamos lo que ya estaba bien hecho (el motor GHOSTDAG) y construimos encima una propuesta económica original que apuesta por la escasez verificable y la honestidad radical.

Quien creó Rupix mina desde el bloque 1, como cualquiera. No hay direcciones privilegiadas, no hay sales, no hay rondas. La única ventaja del que llega temprano es haber estado despierto cuando arrancó la red.

## Licencia

ISC — Rupix developers, 2026.

## Reconocimientos

A los investigadores y desarrolladores que crearon y publicaron GHOSTDAG bajo licencia abierta. Su trabajo permite que proyectos como Rupix existan.

**No confíes, verifica.** ER
