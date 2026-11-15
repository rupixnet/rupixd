#!/bin/sh
# Prueba de instalar-nodo.sh (Rupix, 15-nov-2026): nombre_previo lee el --uacomment de un
# servicio existente y no inventa nada si no lo hay. Correr: sh tools/instalar-nodo_test.sh
set -eu
T=$(mktemp -d); trap 'rm -rf "$T"' EXIT
# Extrae la funcion tal cual del instalador, para probar exactamente lo que se instala.
sed -n '/^nombre_previo() {/,/^}/p' "$(dirname "$0")/instalar-nodo.sh" > "$T/f.sh"
. "$T/f.sh"
printf 'ExecStart=/usr/local/bin/rupixd --testnet --addpeer=x --uacomment=Coroking --rpclisten=1\n' > "$T/u1"
printf 'ExecStart=/usr/local/bin/rupixd --testnet --addpeer=x\n' > "$T/u2"
[ "$(nombre_previo "$T/u1")" = "Coroking" ] || { echo "FALLO: debia leer Coroking, leyo '$(nombre_previo "$T/u1")'"; exit 1; }
[ -z "$(nombre_previo "$T/u2")" ] || { echo "FALLO: sin uacomment debe dar vacio"; exit 1; }
[ -z "$(nombre_previo "$T/no-existe")" ] || { echo "FALLO: sin archivo debe dar vacio"; exit 1; }
sh -n "$(dirname "$0")/instalar-nodo.sh" || { echo "FALLO: sintaxis del instalador"; exit 1; }
echo "OK instalar-nodo_test: 3 casos + sintaxis"
