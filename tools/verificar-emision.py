#!/usr/bin/env python3
# verificar-emision.py — compara el Gold que existe en tu nodo con lo que dice la regla.
# No confia en nadie: le pregunta a TU nodo (rupixctl) y calcula la regla aqui mismo.
# Uso: python3 verificar-emision.py [--mainnet]
import json, subprocess, sys

MAIN = "--mainnet" in sys.argv
RED = [] if MAIN else ["--testnet"]
HALVING = 42_000_000 if MAIN else 100_000
BASE = 50_000_000            # 0.5 RUPIX en rupias
R = 100_000_000              # rupias por RUPIX

def rpc(cmd):
    out = subprocess.run(["rupixctl", *RED, cmd], capture_output=True, text=True, timeout=120).stdout
    return json.loads(out)

def buscar(d, trozo):
    for k, v in d.items():
        if isinstance(v, dict):
            r = buscar(v, trozo)
            if r is not None:
                return r
        elif trozo in k.lower():
            return int(v)
    return None

def emitido(bloques):
    # Suma exacta de la recompensa de los bloques con DAA 1..bloques (el genesis paga 0).
    total, era, falta, d = 0, 0, bloques, 1
    while falta > 0:
        recompensa = BASE >> era if era < 64 else 0
        if recompensa == 0:
            break
        fin_era = (era + 1) * HALVING          # primer DAA de la era siguiente
        n = min(falta, fin_era - d)
        total += n * recompensa
        falta -= n; d += n; era += 1
    return total

daa = buscar(rpc("GetBlockDagInfo"), "virtualdaascore")
circ = buscar(rpc("GetCoinSupply"), "circulating")
era = daa // HALVING
recompensa = BASE >> era
esperado = emitido(daa - 1)
diferencia = esperado - circ

print("Red:                 ", "mainnet" if MAIN else "testnet", "· halving cada", f"{HALVING:,}", "bloques")
print("DAA actual:          ", f"{daa:,}", "· era", era + 1)
print("Recompensa ahora:    ", recompensa / R, "RUPIX por bloque")
print("Niveles abiertos:    ", ", ".join(["Gold"] + [n for i, n in enumerate(["Diamante", "Platino", "Rodio", "Kings"], 1) if daa >= i * HALVING]))
print("Segun la regla:      ", f"{esperado / R:,.8f}", "RUPIX emitidos")
print("En tu nodo existen:  ", f"{circ / R:,.8f}", "RUPIX")
print("Diferencia:          ", f"{diferencia / R:,.8f}", "RUPIX")
# La diferencia es lo quemado (forjas, quema por transaccion) mas las recompensas de los
# ultimos bloques que el siguiente bloque todavia no ha pagado. Nunca puede ser negativa.
if diferencia < -2 * recompensa:
    print("ALERTA: tu nodo tiene MAS Gold del que permite la regla.")
    sys.exit(1)
print("OK: nunca hay mas Gold del que permite la regla.")
