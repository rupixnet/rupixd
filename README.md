# RUPIX

**Red blockchain L1 independiente con sistema de escasez progresiva en 5 niveles.**

Supply máximo: 42,000,000 RUPIX | Zero pre-mine | Proof of Work | Deflationary burn

rupix.network

## Qué es Rupix

Rupix es una blockchain L1 construida sobre DAGKnight/GHOSTDAG con un modelo económico propio:

- **42,000,000 RUPIX** máximo — sellado en el genesis, nadie puede crear más
- **Sistema de 5 niveles**: Gold → Diamond → Platinum → Rhodium → Kings Rupix
- **Burn deflacionario**: cada transacción destruye tokens permanentemente
- **Zero pre-mine**: el primer RUPIX fue minado después del genesis, como Bitcoin

## Niveles de escasez

El oro es valioso porque es escaso y porque cuesta trabajo extraerlo. Rupix lleva esa idea más lejos: para acceder a los niveles superiores hay que quemar tokens del nivel anterior — permanentemente. Esos tokens no van a ningún minero ni a ninguna tesorería. Desaparecen para siempre, grabados en la blockchain como prueba irreversible.

| Nivel | Nombre | Supply máximo | Cómo se obtiene |
|-------|--------|---------------|-----------------|
| L1 | Gold | 42,000,000 | Minando |
| L2 | Diamond | 2,100,000 | Quemar 10 Gold |
| L3 | Platinum | 210,000 | Quemar 10 Diamond |
| L4 | Rhodium | 21,000 | Quemar 10 Platinum |
| L5 | Kings Rupix | 2,100 | Quemar 10 Rhodium |

Cada nivel es 10 veces más escaso que el anterior. Para crear 1 Kings Rupix hay que haber destruido 10,000 Gold a lo largo de toda la cadena. Solo pueden existir 2,100 Kings Rupix en el mundo — para siempre. El burn es irreversible. Nadie puede revertirlo, ni los fundadores.

## Deflación permanente

Cada transacción destruye tokens:

    burn = 1,000 rupias + (tamaño_tx × 10 rupias)

El supply total solo puede bajar — nunca subir.

## Correr un nodo

**Requisitos:** Go 1.21+, 4GB RAM, 50GB disco

    git clone https://github.com/rupixnet/rupixd.git
    cd rupixd
    go build -o rupixd .
    go build -o rupixminer ./cmd/rupixminer
    go build -o rupixwallet ./cmd/rupixwallet
    go build -o rupixctl ./cmd/rupixctl

**Testnet:**

    ./rupixd --testnet --utxoindex

**Verificar supply:**

    ./rupixctl --testnet GetCoinSupply
    # maxRupia: 4200000000000000 = exactamente 42,000,000 RUPIX

## Verificabilidad

No tienes que confiar en nosotros. Puedes verificarlo tú mismo:

    go test ./domain/dagconfig/... -run TestSkipProofOfWork
    ./rupixctl --testnet GetCoinSupply

El código es público. El genesis no tiene pre-mine. Las reglas no las controla nadie — las controla el protocolo.

## Mejoras técnicas

- Supply máximo sellado: 42,000,000 RUPIX — criptográficamente imposible de modificar
- Sistema de 5 niveles de escasez a nivel de protocolo — único en blockchain
- Burn deflacionario: cada transacción destruye tokens permanentemente
- TxTypeBurn — tipo de transacción nativo para transiciones de nivel
- Corrección del DAA score off-by-one en construcción de bloques
- Protección BlueWork nil en sincronización de nodos nuevos
- Zero pre-mine — verificable en el bloque genesis

## Estado del proyecto

- ✅ Simnet funcionando — primer burn L1→L2 exitoso
- ✅ Sistema de 5 niveles funcional a nivel de protocolo
- 🔄 Testnet pública — próximamente
- 🔄 Mainnet — Mayo/Junio 2026

## Licencia

ISC — Rupix developers 2026

*No confíes, verifica. — rupix.network*