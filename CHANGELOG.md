# Changelog — Rupix

Historial honesto de cambios. Todo verificable contra los commits del repositorio.

Formato: [versión] - fecha - descripción técnica

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

