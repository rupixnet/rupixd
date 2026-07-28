# Changelog — Rupix

Historial honesto de cambios. Todo verificable contra los commits del repositorio.

Formato: [versión] - fecha - descripción técnica

---

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

