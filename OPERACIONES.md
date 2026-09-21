# Operación del nodo semilla / Seed node operations

🇲🇽 El seed de la testnet corre como tres servicios systemd con reinicio automático. Si el nodo muere (OOM, crash, reinicio del servidor), se levanta solo en segundos. Lección del 21-sep-2026: el nodo murió por OOM a las 04:31 y nadie lo supo hasta las 19:00 — 14 horas sin seed. Nunca más.

🇬🇧 The testnet seed runs as three systemd services with automatic restart. If the node dies (OOM, crash, server reboot), it comes back on its own within seconds. Lesson from 21-Sep-2026: the node was OOM-killed at 04:31 and nobody knew until 19:00 — 14 hours without a seed. Never again.

## Servicios / Services

| Servicio | Qué es | Depende de |
|---|---|---|
| `rupixd-testnet` | el nodo (RPC 17210, P2P 17211) | red |
| `rupixwallet-testnet` | daemon de la wallet del servidor (8082) | rupixd |
| `rupixminer-testnet` | minero | rupixd |

## Comandos / Commands

- Estado / status: `systemctl status rupixd-testnet rupixwallet-testnet rupixminer-testnet`
- Reiniciar nodo (daemon y minero lo siguen): `systemctl restart rupixd-testnet`
- Log del servicio: `journalctl -u rupixd-testnet -n 50`
- Log del nodo: `tail -f /root/rupix-testnet.log`
- Bloque actual: `rupixctl -s 127.0.0.1:17210 GetBlockDagInfo`

## Actualizar el binario / Update the binary

1. `go build -o /usr/local/bin/rupixd .` (y wallet, miner, ctl)
2. `systemctl restart rupixd-testnet` — la cadena en disco no se toca; daemon y minero se reinician solos.

## Memoria / Memory

El servidor tiene 7.7 GB. rupixd usa ~2.7 GB. **No correr `go test ./...` completo en este host** mientras el seed corre: fue la causa del OOM del 21-sep. Tests pesados en otra máquina o con `go test -p 1`.

## Reglas / Rules

- Si un proceso está `activating` en loop: `journalctl -u <servicio> -n 30` dice por qué.
- Nunca `kill -9` a mano: `systemctl stop/restart`.
- Tras un reinicio del servidor, los tres arrancan solos (`enabled`).
