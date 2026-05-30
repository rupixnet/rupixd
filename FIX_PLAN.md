# Plan de limpieza profunda — Rupix

**Iniciado:** 28 de mayo de 2026
**Estado:** auditoria activa
**Branch:** fix-deep-cleanup
**Safety tag:** v0.2.2-safe (estado pre-cleanup)

## Contexto

Durante el desarrollo inicial de Rupix (antes de FIX-001), se aplicaron
multiples parches al codigo de consenso heredado de Kaspa. Estos parches
fueron metidos individualmente para suprimir crashes especificos durante
IBD y operacion temprana, pero NO fueron documentados ni discutidos.

Auditoria del dia 3 (28 mayo 2026) revela que estos parches:

1. Suprimen validaciones criticas de Kaspa
2. Causan que el flujo de coinbase rewards no llegue al UTXO index
3. Producen el bug observable: balance del minero = 0

## Filosofia del cleanup

- **No prisa con mainnet.** Trabajamos hasta que quede al 100%.
- **No parchar mas.** Cada parche se revierte y se reemplaza por
  el comportamiento original de Kaspa o por un fix bien documentado.
- **Tests primero.** Antes de aplicar cualquier reversion, escribimos
  el test que la valida.
- **Verificable.** Cada cambio en este branch tiene su commit, su
  entrada en FIX_PLAN.md, y referencia al codigo Kaspa upstream.

## Inventario de parches identificados

### CRITICOS (rompen tokenomics)

#### P-001: validateAcceptedIDMerkleRoot deshabilitado
- Archivo: domain/consensus/processes/consensusstatemanager/verify_and_build_utxo.go
- Lineas: 27-30
- Estado: comentado, no se ejecuta
- Riesgo: bloques no validan que su acceptance data coincida con el merkle root
- Sintoma: AcceptedIDMerkleRoot identico en todos los bloques
- Plan: revertir comentario, agregar test que valide consistencia

#### P-002: validateUTXOCommitment deshabilitado
- Archivo: domain/consensus/processes/consensusstatemanager/verify_and_build_utxo.go
- Lineas: ~25 (comentado el primero)
- Estado: comentado
- Riesgo: no se verifica que el UTXO multiset coincida con el commitment del header
- Plan: revertir y validar

#### P-003: ghostdagDataStore retorna nil en notfound
- Archivo: domain/consensus/datastructures/ghostdagdatastore/ghostdag_data_store.go
- Linea: 63-65
- Codigo actual: \`return externalapi.NewBlockGHOSTDAGData(0, new(big.Int), nil, nil, nil, nil), nil\`
- Estado: parche activo
- Riesgo: caller recibe selectedParent=nil silenciosamente, dispara branches no probados
- Plan: propagar error, ajustar callers para manejarlo correctamente

### MEDIOS (afectan robustez)

#### P-004: checkCoinbaseBlueScore deshabilitado
- Archivo: domain/consensus/processes/blockvalidator/block_body_in_isolation.go
- Lineas: 51-55
- Plan: revertir, probar con cadena fresca

#### P-005: daablocksstore retorna 0 en notfound
- Archivo: domain/consensus/datastructures/daablocksstore/daa_blocks_store.go
- Lineas: 64, 91
- Plan: propagar error, identificar caller que necesitaba el parche

#### P-006: reachabilityDataStore retorna vacio en notfound
- Archivo: domain/consensus/datastructures/reachabilitydatastore/reachability_data_store.go
- Lineas: 94, 112
- Plan: propagar error

#### P-007: RUPIX ignore durante IBD (multiples)
- Archivos varios en domain/consensus/processes/blockvalidator/
- Plan: revisar caso por caso. Algunos pueden ser legitimos para IBD,
  otros pueden estar enmascarando bugs.

## Tests necesarios antes de validar fix

- T-001: minero local mina 110+ bloques → balance > 0
- T-002: GetCoinSupply muestra circulatingRupia > 0 despues de 110 bloques
- T-003: AcceptedIDMerkleRoot DIFERENTE entre bloques consecutivos
- T-004: IBD desde Windows con BD limpia sigue funcionando
- T-005: Nodo Windows con BD limpia ve mismo balance que el seed

## Bitacora

### 2026-05-28 - Inicio
- Tag v0.2.2-safe creado
- Branch fix-deep-cleanup creada
- FIX_PLAN.md inicial committeado
- Auditoria a fondo pendiente

### 2026-05-30 - CORRECCION HONESTA

FIX-004 + FIX-005 aplicados con verificación incompleta.

- v0.2.3 fue tagueado prematuramente (BORRADO).
- Verificación a 120s mostró cadena estable y balance creciendo (28 RUPIX).
- Pero verificación a 180s revelo: cadena se atasca en bloque 67 igual.
- 50 rechazos + 150 panics en los 60s adicionales.

Conclusion: FIX-004 + FIX-005 son CORRECTOS pero INSUFICIENTES.
Hay otro parche oculto que se manifiesta despues de N bloques (N=67?).

Hipotesis: probable culprit en pick_virtual_parents.go RUPIX-017
(deduplicacion final) o consensusstatemanager (selectVirtualSelectedParent
con StatusUTXOPendingVerification aceptado como valido).

No taguear v0.2.3 hasta que cadena supere 200+ bloques estable.

Lesson: verificar empirically al doble del umbral conocido antes de
declarar victoria. 120s no fue suficiente cuando ayer la cadena
duraba ~120s antes de atascarse.
