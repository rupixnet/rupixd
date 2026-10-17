#!/bin/sh
# instalar-nodo.sh (Rupix, 16-oct-2026): un nodo de la testnet de Rupix en un comando.
#
#   curl -sSL https://rupix.network/instalar.sh | sh
#
# Que hace, en orden, y se detiene al primer fallo:
#   1. Detecta el sistema (Linux x86_64 o macOS) y la ultima release publicada.
#   2. Baja el zip de la release y la lista SHA256SUMS-<os>.txt que publica el CI.
#   3. VERIFICA el sha256 de cada binario contra esa lista; si uno no coincide, no instala nada.
#   4. Instala rupixd, rupixctl, rupixwallet y rupixminer (en /usr/local/bin si puede; si no, en ~/.local/bin).
#   5. Pregunta como quieres que se llame tu nodo en el mapa (https://explorer.rupix.network/mapa.html).
#   6. En Linux con systemd y permisos: crea el servicio `rupixd-testnet` (arranca solo, sobrevive reinicios) y lo enciende.
#      En macOS o sin systemd: deja los programas instalados y te dice el comando para arrancar el nodo.
#   7. Corre `rupixwallet --testnet verificar` para que veas tu pantalla de OK.
#
# Opciones (tambien por variables de entorno):
#   --nombre=TuNombre        nombre en el mapa (RUPIX_NOMBRE). Sin espacios. Vacio = "nodo".
#   --version=v0.6.6         una release concreta (RUPIX_VERSION). Por defecto, la ultima.
#   --seed=IP.PUBLICA        modo seed: puerto 17211 abierto a internet, sin indice de UTXOs (menos RAM).
#   --sin-servicio           solo instalar los binarios; no crear ni tocar servicios.
#   --prefix=/ruta           donde instalar (RUPIX_PREFIX). Por defecto /usr/local/bin o ~/.local/bin.
#   --sin-verificar          no correr `verificar` al final.
#
# Es texto plano: leelo antes de correrlo. Su sha256 esta publicado junto a cada release y en la guia.
# No pide contrasenas, no toca wallets, no manda nada a ningun lado que no sea GitHub (descarga) y tu propio nodo.
# Repo y codigo: https://github.com/rupixnet/rupixd  ·  No confies, verifica.

set -eu

ULTIMA_CONOCIDA="v0.6.6"   # se actualiza en cada release; es el ultimo recurso si GitHub no contesta
REPO="rupixnet/rupixd"
SEED1="178.104.69.148:17211"
PUERTO_P2P=17211

NOMBRE="${RUPIX_NOMBRE:-}"
VERSION="${RUPIX_VERSION:-}"
PREFIX="${RUPIX_PREFIX:-}"
SEED_IP=""
SERVICIO=1
VERIFICAR=1
for arg in "$@"; do
	case "$arg" in
		--nombre=*) NOMBRE="${arg#*=}" ;;
		--version=*) VERSION="${arg#*=}" ;;
		--prefix=*) PREFIX="${arg#*=}" ;;
		--seed=*) SEED_IP="${arg#*=}" ;;
		--sin-servicio) SERVICIO=0 ;;
		--sin-verificar) VERIFICAR=0 ;;
		-h|--help) sed -n '2,30p' "$0" 2>/dev/null || true; exit 0 ;;
		*) echo "opcion desconocida: $arg"; exit 2 ;;
	esac
done

decir() { printf '\n\033[1;33m==> %s\033[0m\n' "$*"; }
fallo() { printf '\n\033[1;31mFALLO: %s\033[0m\n' "$*" >&2; exit 1; }
tiene() { command -v "$1" >/dev/null 2>&1; }

# 1) Sistema
SO=$(uname -s); ARQ=$(uname -m)
case "$SO" in
	Linux) ETIQUETA=linux; SHA="sha256sum" ;;
	Darwin) ETIQUETA=osx; SHA="shasum -a 256" ;;
	*) fallo "sistema no soportado por este instalador: $SO (en Windows sigue la guia: https://rupix.network/guia.html)" ;;
esac
case "$ARQ" in
	x86_64|amd64) ;;
	arm64|aarch64) [ "$SO" = Darwin ] || fallo "Linux en $ARQ: las releases son x86_64; compila desde el codigo (go build) o usa una maquina x86_64" ;;
	*) fallo "arquitectura no soportada: $ARQ" ;;
esac
tiene curl || fallo "falta curl"
tiene unzip || fallo "falta unzip (Linux: apt install unzip)"

# 2) Version
if [ -z "$VERSION" ]; then
	VERSION=$(curl -sS -o /dev/null -w '%{redirect_url}' "https://github.com/$REPO/releases/latest" 2>/dev/null | sed -n 's#.*/tag/\(v[0-9][0-9.]*\)$#\1#p' || true)
	[ -n "$VERSION" ] || VERSION="$ULTIMA_CONOCIDA"
fi
case "$VERSION" in v[0-9]*) ;; *) fallo "version rara: $VERSION" ;; esac
decir "Rupix $VERSION para $ETIQUETA ($ARQ)"

# 3) Descargar y verificar
TMP=$(mktemp -d 2>/dev/null || mktemp -d -t rupix)
trap 'rm -rf "$TMP"' EXIT
BASE="https://github.com/$REPO/releases/download/$VERSION"
ZIP="rupix-$VERSION-$ETIQUETA.zip"
LISTA="SHA256SUMS-$ETIQUETA.txt"
decir "Bajando $ZIP y $LISTA"
curl -sSL -o "$TMP/$ZIP" "$BASE/$ZIP" || fallo "no pude bajar $BASE/$ZIP"
curl -sSL -o "$TMP/$LISTA" "$BASE/$LISTA" || fallo "no pude bajar $BASE/$LISTA (releases anteriores a v0.6.5 no la tienen; usa --version=v0.6.6 o posterior)"
mkdir -p "$TMP/bin" && unzip -q -o "$TMP/$ZIP" -d "$TMP/bin"
decir "Verificando sha256 de cada binario contra la lista publicada"
( cd "$TMP/bin" && $SHA -c "../$LISTA" ) || fallo "un binario NO coincide con la lista publicada. No se instala nada. Reporta esto en github.com/$REPO/issues"
for b in rupixd rupixctl rupixwallet rupixminer; do [ -f "$TMP/bin/$b" ] || fallo "falta $b en el zip"; done

# 4) Instalar
SUDO=""
if [ -z "$PREFIX" ]; then
	if [ "$(id -u)" = 0 ]; then PREFIX=/usr/local/bin
	elif tiene sudo && sudo -n true 2>/dev/null; then PREFIX=/usr/local/bin; SUDO=sudo
	else PREFIX="$HOME/.local/bin"; fi
fi
mkdir -p "$PREFIX" 2>/dev/null || $SUDO mkdir -p "$PREFIX"
PUEDE_ROOT=0; { [ "$(id -u)" = 0 ] || [ -n "$SUDO" ]; } && PUEDE_ROOT=1
CON_SERVICIO=0; [ "$SERVICIO" = 1 ] && [ "$SO" = Linux ] && [ "$PUEDE_ROOT" = 1 ] && tiene systemctl && CON_SERVICIO=1
if [ "$CON_SERVICIO" = 1 ]; then
	$SUDO systemctl stop rupixd-testnet 2>/dev/null || true   # actualizar sin "Text file busy"
fi
for b in rupixd rupixctl rupixwallet rupixminer; do
	chmod +x "$TMP/bin/$b"
	$SUDO cp "$TMP/bin/$b" "$PREFIX/$b.nuevo" && $SUDO mv -f "$PREFIX/$b.nuevo" "$PREFIX/$b"
done
decir "Instalado en $PREFIX: $("$PREFIX/rupixd" --version 2>/dev/null || echo rupixd)"
case ":$PATH:" in *":$PREFIX:"*) ;; *) echo "   (agrega $PREFIX a tu PATH, o llama a los programas con la ruta completa)";; esac

# 5) Nombre en el mapa
if [ -z "$NOMBRE" ] && [ -r /dev/tty ]; then
	printf '\n¿Como quieres que se llame tu nodo en el mapa? (sin espacios; vacio = "nodo"): '
	read -r NOMBRE < /dev/tty || NOMBRE=""
fi
NOMBRE=$(printf '%s' "$NOMBRE" | tr -d ' <>&"/' | cut -c1-32)
UA=""; [ -n "$NOMBRE" ] && UA="--uacomment=$NOMBRE"

# 6) Servicio (Linux + systemd + permisos) o instrucciones
if [ -n "$SEED_IP" ]; then
	FLAGS="--testnet --listen=0.0.0.0:$PUERTO_P2P --externalip=$SEED_IP --rpclisten=127.0.0.1:17210 --addpeer=$SEED1 $UA"
else
	FLAGS="--testnet --utxoindex --rpclisten=127.0.0.1:17210 --addpeer=$SEED1 $UA"
fi
if [ "$CON_SERVICIO" = 1 ]; then
	if ! id rupix >/dev/null 2>&1; then
		$SUDO useradd --system --home-dir /var/lib/rupix --create-home --shell /usr/sbin/nologin rupix 2>/dev/null || \
		$SUDO useradd -r -d /var/lib/rupix -m -s /sbin/nologin rupix
	fi
	$SUDO mkdir -p /var/lib/rupix && $SUDO chown rupix:rupix /var/lib/rupix
	UNIT=/etc/systemd/system/rupixd-testnet.service
	$SUDO sh -c "cat > $UNIT" <<EOF
[Unit]
Description=Rupix testnet node ($VERSION)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=rupix
ExecStart=$PREFIX/rupixd --appdir=/var/lib/rupix/testnet $FLAGS
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
	$SUDO systemctl daemon-reload
	$SUDO systemctl enable --now rupixd-testnet
	decir "Servicio rupixd-testnet encendido (arranca solo al reiniciar). Log: journalctl -u rupixd-testnet -f"
	if [ -n "$SEED_IP" ] && tiene ufw; then
		echo "   Modo seed: recuerda abrir el puerto: ufw allow $PUERTO_P2P/tcp"
	fi
	if [ "$VERIFICAR" = 1 ]; then
		decir "Esperando 25 s a que el nodo arranque y verificando"
		sleep 25
		"$PREFIX/rupixwallet" --testnet verificar || true
		echo
		echo "Mientras sincroniza es normal ver AVISO en 'sincronizado'. Tu luz: https://explorer.rupix.network/mapa.html"
	fi
else
	decir "Binarios instalados. Para arrancar tu nodo (deja la ventana abierta):"
	echo
	echo "   $PREFIX/rupixd $FLAGS"
	echo
	echo "En otra ventana, para comprobar: $PREFIX/rupixwallet --testnet verificar"
	echo "Tu luz en el mapa: https://explorer.rupix.network/mapa.html"
fi

decir "Listo. Guia completa: https://rupix.network/guia.html · No confies, verifica."
