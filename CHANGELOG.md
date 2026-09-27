# Changelog — Rupix

Historial honesto de cambios. Todo verificable contra los commits del repositorio.

Formato: [versión] - fecha - descripción técnica

---

## [Sin publicar]
- `go.mod`: minimo de Go de 1.26.6 a 1.25.0 (la minima real del arbol de dependencias). Un revisor con Go 1.25 ya puede compilar y correr la suite. Los binarios de release se siguen compilando con 1.26.6.
- Residuos del fork: el daemon de rupixwallet escribe su log en `~/.rupixwallet` (antes `~/.kaspawallet`); el minero dice "Rupixd is not synced" (antes "Kaspad").
- Alarma de bloques descalificados **corregida**. La v1 (25-sep, pedida por el auditor el 23-sep) caminaba hacia atras por la cadena seleccionada y, por construccion, nunca podia ver un bloque descalificado: dijo "OK" 134 veces sin poder sonar. La v2 lee la decision del propio nodo en su log (solo lo nuevo desde la revision anterior) y dice ERROR si el log no crece. Probada con un log falso: suena. Falta probarla provocando un bloque descalificado en devnet.
- Nodo: `resolve_block_status.go` registra cada bloque descalificado en nivel Warn (`BLOQUE DESCALIFICADO <hash>: <motivo>`); antes era Debug y no se veia. Cambio de log, no de consenso: no requiere relanzamiento.
- README (es/en): seccion "Como minar Rupix".
- `changelog.txt` renombrado a `CHANGELOG-kaspad-upstream.txt` para no confundirlo con este archivo.

## [v0.6.0] - 2026-09-26 — Tres cambios de consenso + testnet #4
- Dominio keccak propio: `heavyHashDomain` pasa de "HeavyHash" a "RupixHeavyHash". El hash de Rupix es distinto al de Kaspa desde el primer bit.
- `computeRank` con aritmetica entera (mod 2^61-1) en lugar de float64 con tolerancia: el resultado es identico en cualquier CPU y sistema operativo. Era el riesgo silencioso mas grave segun el auditor; el lo verifico por su cuenta.
- Fix del bug del King: `newBlockGemsCommitment` sumaba las forjas del propio bloque ademas de las del mergeset, asi que el minero sellaba un conteo distinto al que calcula el validador. El primer King real habria sido rechazado. Lo cazo `TestKingsEndToEnd`, que falla si se revierte el fix.
- Coinbase con `Version: 0` en lugar de `MaxScriptPublicKeyVersion`.
- Testnet relanzada (#4) con los tres cambios. Vectores de `pruning_test` regenerados.

## [v0.5.2] - 2026-09-20 — Suite de tests en verde
- `go test ./...` de 18 paquetes rojos a 0, sin tocar consenso. Causa raíz (predicha por el auditor como H-5): tests heredados creaban outputs con `Version = MaxScriptPublicKeyVersion` (= 4 = Kings en Rupix) y la escalera los rechazaba. Corregido en el framework de test, vectores bip32/txscript/ventana/poda regenerados desde el código. `DisasmString` aceptaba solo la versión máxima; ahora 0..4. `go vet` limpio en producción.
- README en inglés + aclaración: fork de kaspad, cadena independiente, no usa KAS.

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

