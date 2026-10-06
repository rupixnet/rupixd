#!/bin/sh
# checkpoint-propuesto.sh — propone el siguiente checkpoint de la testnet a partir del
# punto de poda ACTUAL del nodo local, y lo coteja con el punto de poda de un nodo
# externo si se le pasa. No cambia nada: imprime las tres lineas que hay que pegar
# (checkpoints.go, checkpoints_test.go, CHECKPOINTS.md) y las condiciones que se
# cumplen o no. La regla de CHECKPOINTS.md: solo sobre un punto de poda, nunca sobre
# bloques recientes, y publicado antes de que caduque el anterior.
#
# Uso: sh tools/checkpoint-propuesto.sh                      (solo el nodo local)
#      sh tools/checkpoint-propuesto.sh <pruningPointHash de otro nodo>
#      RUPIXCTL="rupixctl --testnet" sh tools/checkpoint-propuesto.sh
set -u
RUPIXCTL="${RUPIXCTL:-rupixctl --testnet}"
EXTERNO="${1:-}"

campo() { # nombre json -> valor sin comillas (primera aparicion)
  grep -oE "\"$1\":[[:space:]]*\"?[0-9a-f]+\"?" | head -1 | sed 's/.*:[[:space:]]*"\{0,1\}\([0-9a-f]*\)"\{0,1\}/\1/'
}

info=$($RUPIXCTL GetBlockDagInfo 2>/dev/null) || { echo "no se pudo hablar con el nodo ($RUPIXCTL)"; exit 2; }
pp=$(printf '%s' "$info" | campo pruningPointHash)
daaVirtual=$(printf '%s' "$info" | campo virtualDaaScore)
[ -n "$pp" ] || { echo "GetBlockDagInfo no trajo pruningPointHash"; exit 2; }

bloque=$($RUPIXCTL GetBlock "$pp" false 2>/dev/null) || { echo "GetBlock $pp fallo"; exit 2; }
blue=$(printf '%s' "$bloque" | campo blueScore)
daa=$(printf '%s' "$bloque" | campo daaScore)
[ -n "$blue" ] && [ -n "$daa" ] || { echo "GetBlock no trajo blueScore/daaScore"; exit 2; }

# Caducidad propuesta: ~23 dias de testnet despues del checkpoint (como el #1:
# 86,400 -> 2,000,000), redondeada a 100,000. El siguiente se publica antes de
# llegar a caducidad - 300,000 (regla de renovacion sin hueco).
caduca=$(( (daa + 2000000) / 100000 * 100000 ))
limite=$(( caduca - 300000 ))

echo "== checkpoint propuesto · $(date -u +%FT%TZ) · nodo: $RUPIXCTL"
echo "punto de poda local : $pp"
echo "blue score          : $blue   (DAA $daa)   · virtual DAA $daaVirtual · profundidad $((daaVirtual - daa)) bloques"
if [ -n "$EXTERNO" ]; then
  if [ "$EXTERNO" = "$pp" ]; then echo "nodo externo        : MISMO punto de poda (se puede publicar como compartido)";
  else echo "nodo externo        : DISTINTO ($EXTERNO). No publicar hasta entender por que."; fi
else
  echo "nodo externo        : no se paso ninguno (si se publica asi, CHECKPOINTS.md debe decir que nadie externo lo confirmo)"
fi
[ "$((daaVirtual - daa))" -ge 100000 ] && echo "profundidad >= 100k : si" || echo "profundidad >= 100k : NO (espera)"
echo
echo "--- checkpoints.go (dentro de testnetCheckpoints):"
echo "	// #N — $(date -u +%d-%b-%Y), testnet #4, vX.Y.Z. Punto de poda del seed; externo: <hash o 'ninguno'>."
echo "	{BlueScore: $blue, Hash: mustHash(\"$pp\")},"
echo "--- params.go: CheckpointsExpireDAAScore: $caduca  (publicar el siguiente antes del DAA $limite)"
echo "--- CHECKPOINTS.md (tabla):"
echo "| testnet #4 | blue score $blue (DAA $daa) | \`$pp\` | $(date -u +%d-%b-%Y) | vX.Y.Z · caduca en DAA $caduca |"
