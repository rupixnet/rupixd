#!/bin/sh
# escalera-hasta-king.sh (Rupix, 7-oct-2026): forja gema por gema hasta tener 1 King.
#
# Recorre la escalera entera con la wallet de esta maquina: 1 King = 10 Rodios =
# 100 Platinos = 1,000 Diamantes = 10,000 RUPIX quemados (mas la quema por tx).
# Es la primera prueba de la escalera completa en red viva, y de que explorador,
# `verificar` y `gems` se actualizan solos. `forge` hace una gema por llamada
# (pendiente v0.7: `forge --count`), asi que esto es un bucle: forja, espera a
# que la gema se vea en la wallet, sigue. Greedy: sube de nivel en cuanto puede.
#
# Uso (en el seed, con rupixwallet-testnet corriendo):
#   read -s CLAVE; export CLAVE          # la clave de la wallet; solo en memoria
#   nohup sh tools/escalera-hasta-king.sh /root/escalera-king.log >/dev/null 2>&1 &
#   unset CLAVE
#   tail -f /root/escalera-king.log
# Parar limpio: touch /root/escalera-king.stop
#
# La clave nunca se escribe: va por tuberia a un pty (`script`), con 2 s de espera
# para que la wallet ya haya apagado el eco. Al registro solo van las lineas de
# `tx:` y las de error; la salida completa de `forge` no se guarda.
set -u
LOG="${1:-/root/escalera-king.log}"
STOP="${STOP:-/root/escalera-king.stop}"
RED="${RED:---testnet}"
ESPERA_MAX="${ESPERA_MAX:-240}"   # segundos maximos para ver la gema nueva
[ -n "${CLAVE:-}" ] || { echo "falta CLAVE en el entorno (read -s CLAVE; export CLAVE)"; exit 2; }
command -v script >/dev/null || { echo "falta 'script' (util-linux)"; exit 2; }

log() { echo "$(date '+%F %T')  $*" >>"$LOG"; }

# cuenta NIVEL: cuantas gemas de ese nivel tiene la wallet (linea NIVEL+1 de `gems`).
cuenta() {
	rupixwallet $RED gems 2>/dev/null | awk -v n="$1" 'NR==n+1 {print $NF}'
}

# forjar NIVEL: una gema. Devuelve el exit de rupixwallet.
forjar() {
	( sleep 2; printf '%s\n' "$CLAVE" ) | \
		script -qec "rupixwallet $RED forge --level $1 --yes" /dev/null 2>&1 | \
		grep -E '^ *tx: |rror|FALLO|fall|no se pudo|rechaz' >>"$LOG"
	# el exit del grep no sirve; el de script si: se recupera por el pipefail si existe,
	# y si no, por la comprobacion de conteo que hace el bucle.
	return 0
}

esperar() { # esperar NIVEL ESPERADO: hasta que cuenta(NIVEL) >= ESPERADO
	t=0
	while [ "$(cuenta "$1")" -lt "$2" ]; do
		sleep 3; t=$((t+3))
		[ "$t" -ge "$ESPERA_MAX" ] && return 1
	done
	return 0
}

log "inicio: $(rupixwallet $RED balance 2>/dev/null | head -1)"
log "inicio: D=$(cuenta 1) P=$(cuenta 2) R=$(cuenta 3) K=$(cuenta 4)"
forjas=0
while :; do
	[ -e "$STOP" ] && { log "parado por $STOP"; break; }
	d=$(cuenta 1); p=$(cuenta 2); r=$(cuenta 3); k=$(cuenta 4)
	case "$d$p$r$k" in *[!0-9]*|"") log "no pude leer gems; paro"; exit 1;; esac
	[ "$k" -ge 1 ] && { log "KING: $k. Fin."; break; }
	if   [ "$r" -ge 10 ]; then n=4; antes=$k
	elif [ "$p" -ge 10 ]; then n=3; antes=$r
	elif [ "$d" -ge 10 ]; then n=2; antes=$p
	else                       n=1; antes=$d
	fi
	forjar "$n"
	forjas=$((forjas+1))
	if ! esperar "$n" $((antes+1)); then
		log "forja #$forjas nivel $n: la gema no aparecio en $ESPERA_MAX s (D=$(cuenta 1) P=$(cuenta 2) R=$(cuenta 3) K=$(cuenta 4)); paro"
		exit 1
	fi
	log "forja #$forjas nivel $n ok: D=$(cuenta 1) P=$(cuenta 2) R=$(cuenta 3) K=$(cuenta 4)"
done
log "fin: $forjas forjas. $(rupixwallet $RED balance 2>/dev/null | head -1)"
rupixwallet $RED gems >>"$LOG" 2>&1
