#!/bin/sh
# verificar-binarios.sh — comprueba que los binarios publicados en una release son
# exactamente los que produce el codigo: compila el tag con los mismos flags que el CI
# (Go 1.26.6, Linux) y compara SHA256 binario por binario. No confia en GitHub ni en nadie.
# Uso: sh tools/verificar-binarios.sh v0.6.1
set -e
V="${1:?uso: verificar-binarios.sh vX.Y.Z}"
D=$(mktemp -d); cd "$D"
echo "== Go: $(go version)  (el CI usa 1.26.6; con otra version el hash cambia)"
curl -sL -o linux.zip "https://github.com/rupixnet/rupixd/releases/download/$V/rupix-$V-linux.zip"
curl -sL -o linux.zip.sha256 "https://github.com/rupixnet/rupixd/releases/download/$V/rupix-$V-linux.zip.sha256"
sed "s/rupix-$V-linux.zip/linux.zip/" linux.zip.sha256 | sha256sum -c
python3 -m zipfile -e linux.zip ci
git clone -q --depth 1 --branch "$V" https://github.com/rupixnet/rupixd src
cd src
go build -trimpath -buildvcs=false -ldflags="-s -w -buildid= -extldflags=-static" -tags netgo,osusergo -o ../local/ .
go build -trimpath -buildvcs=false -ldflags="-s -w -buildid= -extldflags=-static" -tags netgo,osusergo -o ../local/ ./cmd/rupixctl ./cmd/rupixwallet ./cmd/rupixminer
cd ..
ok=1
for f in rupixd rupixctl rupixwallet rupixminer; do
  a=$(sha256sum "ci/$f" | cut -d' ' -f1); b=$(sha256sum "local/$f" | cut -d' ' -f1)
  if [ "$a" = "$b" ]; then echo "IGUAL    $f  $a"; else echo "DIFIERE  $f  ci=$a  local=$b"; ok=0; fi
done
[ "$ok" = 1 ] && echo "OK: los binarios de $V son exactamente lo que produce el codigo." || { echo "ALERTA: los binarios publicados NO salen del codigo."; exit 1; }
