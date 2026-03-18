# RUPIX

**Red blockchain L1 independiente con sistema de escasez progresiva en 5 niveles.**

Supply máximo: 42,000,000 RUPIX | Zero pre-mine | Proof of Work | Deflationary burn

rupix.network

## Qué es Rupix

Rupix es una blockchain L1 construida sobre DAGKnight/GHOSTDAG (el mismo consenso de Kaspa), con un modelo económico propio:

- **42,000,000 RUPIX** máximo — sellado en el genesis, nadie puede crear más
- **Sistema de 5 niveles**: Gold → Diamond → Platinum → Rhodium → Kings Rupix
- **Burn deflacionario**: cada transacción destruye tokens permanentemente
- **Zero pre-mine**: el primer RUPIX fue minado después del genesis, como Bitcoin

## Niveles de escasez

| Nivel | Nombre | Supply máximo | Cómo se obtiene |
|-------|--------|---------------|-----------------|
| L1 | Gold | 42,000,000 | Minando |
| L2 | Diamond | 2,100,000 | Quemar 10 L1 |
| L3 | Platinum | 210,000 | Quemar 10 L2 |
| L4 | Rhodium | 21,000 | Quemar 10 L3 |
| L5 | Kings Rupix | 2,100 | Quemar 10 L4 |

## Correr un nodo

**Requisitos:** Go 1.21+, Git, 4GB RAM, 50GB disco
```bash
git clone https://github.com/rupixnet/rupixd.git
cd rupixd
go build -o rupixd .
go build -o rupixminer ./cmd/rupixminer
go build -o rupixwallet ./cmd/rupixwallet
go build -o rupixctl ./cmd/rupixctl
```

**Testnet:**
```bash
./rupixd --testnet --utxoindex
```

**Verificar supply (siempre 42M):**
```bash
./rupixctl --testnet GetCoinSupply
```

## Verificabilidad

Todo es verificable ejecutando un nodo:
```bash
# Verificar que PoW no puede desactivarse en producción
go test ./domain/dagconfig/... -run TestSkipProofOfWork

# Verificar supply máximo on-chain
./rupixctl --testnet GetCoinSupply
# maxRupia: 4200000000000000 = 42,000,000 RUPIX exactos
```

## Diferencias técnicas vs Kaspa

- Supply 42M vs 28.7B de Kaspa — 680x más escaso
- Sistema de 5 niveles de escasez — único en crypto
- Burn antispam destruye tokens (Kaspa paga fees al minero)
- TxTypeBurn — tipo de transacción nuevo a nivel de protocolo
- Corrección del bug DAA score off-by-one del codebase original
- Protección BlueWork nil mejorada

## Licencia

MIT — Rupix developers 2026

*No confíes, verifica. — rupix.network*