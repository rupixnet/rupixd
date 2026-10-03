# retirados/

Piezas heredadas de kaspad que ya no describen a Rupix pero que no se borran (en Rupix nada viejo se borra: se aparta y se dice por qué).

- `docker-kaspa/`: los Dockerfiles heredados (compilaban `kaspad` con `golang:1.23-alpine` y rutas de kaspanet) y `build_and_test.sh` (usaba `go get -d`, retirado del CI el 30-sep-2026). Si algún día Rupix publica imágenes Docker, se escriben nuevas desde `deploy.yaml` (Go 1.26.6, flags reproducibles), no desde estas.
- `readmes-kaspa/`: los README de `cmd/rupixctl` y `cmd/rupixminer`, que eran el texto de kaspactl/kaspaminer. Las guías vigentes de Rupix son `GUIA-TESTNET.md` / `TESTNET-GUIDE.md` y `GUIA-FORJA.md` / `FORGE-GUIDE.md` en la raíz.

Apartados el 2 de octubre de 2026 (revisión del repo).
