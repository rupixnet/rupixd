#!/usr/bin/env python3
# mineros-distintos.py (Rupix, 3-nov-2026): cuantos MINEROS DISTINTOS hicieron los ultimos N
# bloques, leyendo el payload de la coinbase (ahi va el script del minero aunque el premio
# sea 0 y el bloque no tenga salidas). No imprime direcciones ni IPs: solo un conteo y los
# primeros 12 caracteres de cada script, que no identifican a nadie. Es un dato de la red,
# no de una persona (regla: nunca investigar quien es un nodo).
#
# Uso (en una maquina con rupixctl y un nodo): python3 tools/mineros-distintos.py [N] [--testnet]
import collections
import json
import subprocess
import sys

n_max = 300
red = "--testnet"
for a in sys.argv[1:]:
    if a.isdigit():
        n_max = int(a)
    elif a.startswith("--"):
        red = a


def rpc(*args):
    out = subprocess.run(["rupixctl", red, *args], capture_output=True, text=True).stdout
    try:
        return json.loads(out)
    except ValueError:
        sys.exit("rupixctl no devolvio JSON para %s:\n%s" % (args, out[:400]))


info = rpc("GetBlockDagInfo").get("getBlockDagInfoResponse") or {}
if info.get("error") or not info.get("tipHashes"):
    sys.exit("GetBlockDagInfo: %s" % json.dumps(info)[:400])
h = info["tipHashes"][0]
mineros, programas, leidos = collections.Counter(), collections.Counter(), 0
for _ in range(n_max):
    b = rpc("GetBlock", h, "true").get("getBlockResponse") or {}
    if b.get("error") or not b.get("block"):
        print("fin en bloque %d (podado o error): %s" % (leidos, json.dumps(b)[:200]))
        break
    blk = b["block"]
    txs = blk.get("transactions") or []
    payload = txs[0].get("payload", "") if txs else ""
    # payload coinbase: blueScore(8) subsidy(8) version(2) largo(1) script extra...
    if len(payload) >= 38:
        largo = int(payload[36:38], 16)
        script = payload[38:38 + 2 * largo]
        extra = bytes.fromhex(payload[38 + 2 * largo:]).decode("utf-8", "replace")
        mineros[script[:12] + "…"] += 1
        programas[extra[:40]] += 1
        leidos += 1
    padres = blk["header"].get("parents") or []
    if not padres or not padres[0].get("parentHashes"):
        break
    h = padres[0]["parentHashes"][0]
print("bloques leidos:", leidos)
print("mineros distintos:", len(mineros))
for s, c in mineros.most_common():
    print("  %s  %d bloques" % (s, c))
print("programa minero (texto extra de la coinbase):")
for e, c in programas.most_common(5):
    print("  %r  %d" % (e, c))
