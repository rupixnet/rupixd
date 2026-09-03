# H-1: TOPES EN MEMPOOL — mapa completo (para atacar fresco)

## EL PROBLEMA
El mempool (validate_transaction.go: validateTransactionInContext) solo valida
anti-spam y estándar. NO valida la escalera. Una tx tramposa de gemas entra al
mempool y solo se rechaza al ARMAR el bloque (tarde) → griefing al minero.

## CÓMO FUNCIONA HOY (dos validaciones de escalera, en lugares distintos)
1. checkLevelRules (transactionvalidator/level_ascension.go:26)
   → mecánica de la tx: ratio 10:1, ascensos, niveles desbloqueados.
   → NO necesita conteo acumulado. Solo mira la tx.
   → Se llama desde transaction_in_context.go:93
2. calculateGemsHistory (consensusstatemanager/gemshistory.go:78)
   → topes históricos: rechaza si excede MaxDiamante/Platino/Rodio.
   → SÍ necesita el gemshistory acumulado (del padre/virtual).

## PLAN (dos niveles de alcance)
### NIVEL A (más fácil, gran valor) — hacer primero
Que el mempool llame a checkLevelRules antes de aceptar la tx.
Rechaza txs de gemas MAL FORMADAS antes del minero.
NO necesita conteo acumulado. Requiere exponer/llamar checkLevelRules desde el mempool.

### NIVEL B (completo) — evaluar después
Que además valide el tope histórico:
- El mempool tiene consensusReference → pedir el gemshistory del virtual/tip
- Validar con validateGemsHistorySanity (ya existe) o similar
Más invasivo pero posible (el acceso ya existe).

## PIEZAS A FAVOR (ya existen)
- Mempool tiene consensusReference (mempool.go:21)
- checkLevelRules ya existe
- gemshistory se calcula y expone (hecho en el pruning)
- validateGemsHistorySanity ya existe (pruningproofmanager/gemshistory_sanity.go)

## DIFICULTAD: MEDIA. Es CONECTAR piezas existentes, no crear.
## MÉTODO: rama de prueba, tests, empezar por Nivel A.
