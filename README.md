# Rupix

**Una blockchain L1 con escasez progresiva en 5 niveles.**

Supply máximo: 42,000,000 RUPIX. Sin pre-mine. Sin reserva de fundador. Sin atajos.

[rupix.network](https://rupix.network) · [Código](https://github.com/rupixnet/rupixd) · [Changelog](./CHANGELOG.md)

---

## Qué es Rupix

Rupix es una blockchain Layer 1 con consenso Proof of Work sobre un BlockDAG (no una cadena lineal). Está construida sobre la arquitectura GHOSTDAG, el protocolo de consenso desarrollado por el equipo de investigación de DAGLabs y publicado en código abierto bajo licencia ISC. Reconocemos y agradecemos ese trabajo: sin esa base, Rupix no existiría.

Lo que Rupix añade encima:

- **Modelo económico propio de 5 niveles** con quema permanente para acceder a cada nivel superior
- **Supply absoluto de 42,000,000 RUPIX**, sellado en el protocolo
- **Tokenomics deflacionaria**: cada transacción destruye tokens permanentemente
- **Genesis sin pre-mine**: el primer RUPIX se mineó después del bloque 0, como Bitcoin
- **Identidad propia**: parámetros de red, prefijos de dirección, denominaciones y reglas económicas únicas

## Por qué creemos en el trilema

El llamado *trilema de blockchain* (Vitalik Buterin) plantea que cualquier red distribuida tiene que elegir entre tres propiedades y solo puede tener dos al mismo tiempo: **descentralización, seguridad y escalabilidad**.

Rupix se construye sobre la premisa de que un BlockDAG con Proof of Work permite empujar las tres al mismo tiempo más lejos que las arquitecturas tradicionales. No decimos que el trilema esté "resuelto" — decimos que está siendo empujado en una dirección que respeta los tres principios:

- **Descentralización**: PoW sin pre-mine, código abierto, sin gobernanza centralizada, anyone-can-mine
- **Seguridad**: validación criptográfica completa, sin atajos, sin "trusted parties"
- **Escalabilidad**: BlockDAG permite múltiples bloques paralelos sin perder consistencia

## Los 5 niveles de escasez

| Nivel | Nombre | Supply máximo | Cómo se obtiene |
|-------|--------|---------------|-----------------|
| L1 | Rupix Gold | 42,000,000 | Minando |
| L2 | Rupix Diamante | 2,100,000 | Quemar 10 Gold |
| L3 | Rupix Platino | 210,000 | Quemar 10 Diamante |
| L4 | Rupix Rodio | 21,000 | Quemar 10 Platino |
| L5 | Kings Rupix | 2,100 | Quemar 10 Rodio |

Cada nivel es exactamente 10 veces más escaso que el anterior. Para crear **1 Kings Rupix** hay que haber destruido **10,000 Gold** a lo largo de toda la cadena. La quema es irreversible y queda registrada en la blockchain para siempre. Nadie puede revertirla — ni los fundadores, ni los mineros, ni ningún acuerdo social futuro.

## Deflación permanente

Cada transacción de Rupix destruye una pequeña cantidad de tokens:
burn = 1,000 rupias + (bytes_de_la_tx × 10 rupias)

(1 RUPIX = 100,000,000 rupias)

Estos tokens no van a un fondo, no van al minero, no van a nadie. Desaparecen. **El supply total solo puede bajar.**

## Estado actual del proyecto

Rupix está en desarrollo activo. Trabajamos en público y documentamos cada avance verificable en [CHANGELOG.md](./CHANGELOG.md).

**Lo que funciona hoy:**
- Red de testnet operativa con nodo semilla en Hetzner
- Minado funcional con `rupixminer`
- Sincronización inicial (IBD) sin nil pointer crashes — corregido en `v0.2.1`
- Sistema de 5 niveles implementado a nivel de protocolo
- Tipo de transacción `TxTypeBurn` para transiciones de nivel

**Lo que estamos arreglando:**
- Parámetros de pruning del testnet (FIX-002 en curso)
- Revisión de parches inseguros heredados durante el desarrollo inicial
- Calibración del sistema de burn entre niveles

**Lo que falta antes de mainnet:**
- Testnet pública estable y sincronizable por cualquier usuario
- Auditoría externa del código de consenso
- Infraestructura redundante (múltiples nodos semilla)
- Block explorer público
- Hashrate inicial comprometido

**Fecha de mainnet:** la anunciaremos cuando el código esté listo, no antes. Preferimos lanzar tarde y bien que pronto y comprometidos.

## Cómo correr un nodo

**Requisitos:** Go 1.21+, 4 GB RAM, 50 GB de disco.
git clone https://github.com/rupixnet/rupixd.git
cd rupixd
go build -o rupixd .
go build -o rupixminer ./cmd/rupixminer
go build -o rupixctl ./cmd/rupixctl

**Conectarte al testnet:**
./rupixd --testnet --utxoindex

**Verificar el supply (que coincida con lo prometido):**
./rupixctl --testnet GetCoinSupply

Resultado esperado: `maxRupia: 4200000000000000` — son exactamente 42,000,000 RUPIX.

## Verificabilidad

Todo en Rupix se puede verificar leyendo el código. No confíes en nosotros, verifica:

- **Que no hay pre-mine**: revisa `domain/dagconfig/genesis.go`
- **Que el supply está sellado en 42M**: revisa `domain/consensus/utils/constants/constants.go`
- **Que el sistema de niveles existe**: revisa `domain/consensus/processes/burnmanager/`
- **Que PoW no se puede desactivar en ninguna red**: corre `go test ./domain/dagconfig/... -run TestSkipProofOfWork`

## Filosofía

Rupix no es un fork por novedad ni por hype. Es una arquitectura económica nueva sobre un motor de consenso probado. Tomamos lo que ya estaba bien hecho — el motor GHOSTDAG — y construimos encima una propuesta económica original que apuesta por la escasez verificable y la honestidad radical.

Quien fundó Rupix mina desde el bloque 0, como cualquiera. No hay direcciones privilegiadas, no hay sales, no hay rondas. La única ventaja del que llega temprano es haber estado despierto cuando arrancó la red.

## Licencia

ISC — Rupix developers, 2026.

## Reconocimientos

A los investigadores y desarrolladores que crearon y publicaron GHOSTDAG bajo licencia abierta. Su trabajo permite que proyectos como Rupix existan.

---

*No confíes, verifica.*
