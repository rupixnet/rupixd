# Changelog — Rupix

Historial honesto de cambios. Todo verificable contra los commits del repositorio.

Formato: [versión] - fecha - descripción técnica

---

## [v0.5.1] - 2026-09-18 — Checkpoints temporales
- Defensa contra el 51% mientras el hashrate es bajo: bloque en el DAA score de un checkpoint debe tener el hash canónico o `ErrCheckpointMismatch`. Caducidad dentro del consenso. Probado en devnet (correcto acepta, falso rechaza). Lista vacía: compatible con v0.5.0, sin relanzamiento. Política en CHECKPOINTS.md.
- Primer Diamante con RupixHeavyHash en la red pública (commitment `780e9027…`). README bilingüe.

## [v0.5.0] - 2026-09-16 — RupixHeavyHash (algoritmo propio)
- Generador de la matriz de kHeavyHash reemplazado (xoshiro256++ → rupixPRNG con multiplicación y sello RUPIX). Los ASIC de Kaspa quedan fuera. Honesto: ventaja de meses, no independencia. Génesis re-minado (misma fecha y mensaje). Testnet relanzada (#3).
- Versión del binario corregida (reportaba 0.4.0). Dependencias: grpc 1.83.2, 0 vulnerabilidades reales. Guard del bucle de generateMatrix.

## [v0.4.5] - 2026-09-14 — Diamante real, primer forjador externo
- Bug de la forja resuelto (el virtual no calculaba su gemsHistory). Primer Diamante real sellado (`780e9027…`). Segundo usuario forjó 3 Diamantes desde su nodo. Primer Platino. Auditoría externa: H-8, H-9, H-10 cerrados.
- **v0.4.4 marcada como defectuosa — no usar** (regresión de verificación introducida por un fix).

## [v0.4.2] - 2026-09-12 — Commitment en header (verificación total)
- El conteo de gemas se sella en el hash de cada bloque, protegido por PoW, validado al recibir, persistido en disco. Halving de testnet a 100,000 bloques. Testnet relanzada (#2). Pruning verificable (gemsHistory viaja en el proof).

## [v0.4.0] - 2026-09-03 — Testnet pública
- Primera testnet pública con binarios verificables (SHA256, compilados por CI). Génesis sin premine (04/03/2026 — RUPIX IS ALIVE). Escalera completa en consenso (5 niveles, 10:1, techos históricos). Burn por transacción. Explorador público. Primer nodo externo.

## [Reinicio] - 2026-06-13 — Desde Kaspa limpio
- El código anterior (v0.2.x, con 1,700 líneas de una IA previa) se descartó por bugs estructurales en cascada. Rupix se reconstruyó desde kaspad limpio, con la economía como ley. Todo lo anterior a esta fecha es historia, no código vigente.

## [v0.2.2] - 2026-05-25 — Primera testnet sincronizable

### Milestone histórico

Por primera vez en la historia del proyecto, un nodo Windows con base de
datos limpia logra sincronizar completamente desde el nodo semilla de
Hetzner y registra en log:

    [INF] PROT: IBD with peer ... finished successfully

Hora exacta: 25 mayo 2026, 23:46 hora México (05:46 UTC).

### Fixed

- **FIX-002 (FinalityDuration testnet)**: cambiado de `1 * time.Minute` a
  `12 * time.Hour`. El valor anterior violaba la invariante de Kaspa
  `MergeSetSizeLimit << FinalityDuration/TargetTimePerBlock`, causando
  poda agresiva e `ErrPrunedBlock` durante IBD. El valor `12 * time.Hour`
  es el estándar de Kaspa testnet.

- **FIX-003 (MaxBlockParents testnet)**: cambiado de `1` a `10`. Con
  `MaxBlockParents=1` el DAG no convergía: una red de 695 bloques
  generaba 196 tips paralelos (28% sin merge), causando
  `ErrMissingParents` durante header sync porque los headers llegaban
  fuera de orden topológico. Valor `10` es el estándar de Kaspa y es
  coherente con `defaultGHOSTDAGK=18`.

### Scope

Solo se modificaron parámetros de **testnet**. Mainnet, Simnet y Devnet
sin cambios.

### Conocido (pendiente)

- `ErrMissingParents` aparece marcado como "non-critical" durante relay
  post-IBD. El nodo lo maneja correctamente, pero se debe investigar si
  el comportamiento es óptimo o si hay margen de mejora.
- Pendiente validación: cliente Windows debe alcanzar `blockCount` igual
  al seed y exponer balance correcto vía `GetBalanceByAddress`.

### Verificable

- Commit del fix: `a01809d3`
- Commit del merge: `b83e9c41`
- Tag: `v0.2.2`
- Log de prueba: cliente Windows con BD limpia (`C:\\RUPIX\\ibd-test-fix003.log`)

---

## [v0.2.1] - 2026-05-21 — IBD-001 cerrado

### Fixed
- **IBD-001 (crítico)**: nodo Windows con base de datos limpia ya no crashea
  con `nil pointer dereference` al sincronizar desde cero.
- Causa raíz: la función `ValidateAndInsertBlockAsTrusted` saltaba las
  validaciones que construyen y almacenan GHOSTDAG data. Sin esa data,
  llamadas subsiguientes al store retornaban not-found, lo cual era
  enmascarado por parches que devolvían datos vacíos con `selectedParent=nil`,
  causando el nil pointer en el pruning manager.
- Fix aplicado: alineamos el flujo IBD con el comportamiento de kaspad
  upstream (referencia: kaspad master, `ibd.go` líneas 519 y 691).

### Removed
- Atajo inseguro "ignore not found durante IBD" en `processHeader` 
  (silenciaba errores legítimos).
- Import huérfano `reachabilitydata` en `reachability_data_store.go`.

### Known issues exposed
- **ErrPrunedBlock**: con el nil pointer eliminado, sale a la luz que
  el testnet poda demasiado agresivo por `FinalityDuration=1min`.
  Pendiente para FIX-002.

### Verificable
- Commit del fix: `00d2f969`
- Commit de merge: `e8a006c7`
- Tag: `v0.2.1`

---

## Cómo verificar el código

```bash
git clone https://github.com/rupixnet/rupixd.git
cd rupixd
git checkout v0.2.1
go build -o rupixd .
```

No confíes, verifica.

