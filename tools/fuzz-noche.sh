#!/bin/sh
# fuzz-noche.sh — corre los objetivos de fuzzing de ESPECIFICACION.md (hueco #6), uno tras
# otro, el tiempo que se le diga a cada uno, y deja el registro en un archivo. Si un
# objetivo encuentra un fallo, go deja el caso en testdata/fuzz/<Objetivo>/ dentro del
# paquete: ese archivo se commitea tal cual (es la prueba) y el fallo se arregla.
# Uso: sh tools/fuzz-noche.sh 2h            (2 horas por objetivo; 6 objetivos = ~12 h)
#      sh tools/fuzz-noche.sh 2m            (humo: 2 minutos cada uno)
set -u
T="${1:?uso: fuzz-noche.sh <tiempo por objetivo, p.ej. 2h>}"
LOG="${2:-fuzz-$(date +%Y%m%d-%H%M).log}"
cd "$(dirname "$0")/.."
echo "== fuzz-noche $(date -u +%FT%TZ) · $T por objetivo · $(go version) · $(git rev-parse --short HEAD)" | tee "$LOG"
fallos=0
correr() { # paquete objetivo
  echo "--- $2 ($1) · inicio $(date -u +%T)" | tee -a "$LOG"
  antes=$(wc -l <"$LOG")
  if go test -run '^$' -fuzz="^$2\$" -fuzztime="$T" "$1" >>"$LOG" 2>&1; then
    ultimo=$(tail -n +"$antes" "$LOG" | grep 'execs:' | tail -1 | sed 's/.*execs: \([0-9]*\).*total: \([0-9]*\).*/\1 intentos, \2 caminos/')
    echo "    OK   $2 · $ultimo" | tee -a "$LOG"
  else
    fallos=$((fallos+1))
    echo "    FALLO $2 · revisa $LOG y $1/testdata/fuzz/$2/" | tee -a "$LOG"
  fi
}
correr ./domain/consensus/utils/topes/ FuzzCabe
correr ./domain/consensus/utils/gemscommitment/ FuzzCalculateGemsCommitment
correr ./domain/consensus/utils/pow/ FuzzGenerateMatrix
correr ./domain/consensus/utils/pow/ FuzzComputeRank
correr ./infrastructure/network/netadapter/server/grpcserver/protowire/ FuzzMsgPruningPoints
correr ./domain/consensus/processes/transactionvalidator/ FuzzCheckLevelRules
echo "== fin $(date -u +%FT%TZ) · fallos: $fallos · registro: $LOG" | tee -a "$LOG"
[ "$fallos" = 0 ]
