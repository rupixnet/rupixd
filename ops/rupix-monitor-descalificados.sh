#!/bin/bash
# Rupix: alarma de bloques descalificados (v2.1, 27-sep-2026).
# Lee la decision del propio nodo: resolve_block_status.go escribe
# "BLOQUE DESCALIFICADO <hash>: <motivo>" (Warn) cada vez que descalifica un bloque.
# Lee solo lo que el log agrego desde la revision anterior. Guarda inode + posicion:
# si el log roto (rupixd.log pasa a otro nombre con el mismo inode), termina de leer
# el archivo viejo desde donde iba y luego lee el nuevo desde el principio.
# Si no puede leer todo, o no hay actividad nueva, dice ERROR: nunca un OK sin datos.
# La v1 (25-sep, pedida por el auditor el 23-sep) caminaba por la cadena seleccionada
# y por construccion nunca podia ver un bloque descalificado.
LOG=${RUPIX_ALARMA_LOG:-/root/rupix-descalificados.log}
STATE=${RUPIX_ALARMA_STATE:-/root/.rupix-alarma-offset}
F=${1:-$(ls /root/rupix-testnet/*/logs/rupixd.log 2>/dev/null | head -1)}
ts=$(date -u +%FT%TZ)
if [ -z "$F" ] || [ ! -f "$F" ]; then echo "$ts ERROR: no encuentro el log del nodo" >> $LOG; exit 1; fi
ino=$(stat -c %i "$F"); size=$(stat -c %s "$F")
[ -f "$STATE" ] && read pino pos < "$STATE"
if [ -z "$pino" ] || [ -z "$pos" ]; then
  pino=$ino; pos=$(( size > 2000000 ? size - 2000000 : 0 ))
fi
previo=""
if [ "$pino" != "$ino" ]; then
  viejo=""
  for g in "$F".*; do
    [ -f "$g" ] && [ "$(stat -c %i "$g")" = "$pino" ] && viejo="$g" && break
  done
  if [ -z "$viejo" ]; then
    echo "$ts ERROR: el log roto y no encuentro el archivo anterior (inode $pino). Pudo perderse un tramo." >> $LOG
    echo "$ino 0" > "$STATE"; exit 1
  fi
  previo=$(tail -c +$((pos + 1)) "$viejo")
  pos=0
fi
[ "$size" -lt "$pos" ] && pos=0
nuevo="$previo
$(tail -c +$((pos + 1)) "$F")"
echo "$ino $size" > "$STATE"
vistos=$(printf "%s\n" "$nuevo" | grep -c "Processed")
if [ "$vistos" -eq 0 ]; then
  echo "$ts ERROR: el log del nodo no muestra actividad desde la ultima revision. La alarma no puede afirmar nada." >> $LOG
  exit 1
fi
hits=$(printf "%s\n" "$nuevo" | grep "BLOQUE DESCALIFICADO")
n=$(printf "%s" "$hits" | grep -c "BLOQUE DESCALIFICADO")
if [ "$n" -gt 0 ]; then
  echo "$ts ⚠️  ALARMA: el nodo descalifico $n bloque(s) desde la ultima revision:" >> $LOG
  printf "%s\n" "$hits" | tail -5 >> $LOG
else
  echo "$ts OK: el nodo no descalifico ningun bloque ($vistos lineas de actividad leidas)" >> $LOG
fi
