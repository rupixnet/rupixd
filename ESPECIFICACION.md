# Especificación de Rupix: cada regla, el ataque que detiene y el test que lo prueba

Primera versión: 30 de septiembre de 2026 (v0.6.1). Formato sugerido por el auditor externo: tres columnas, nada más. **Donde la tercera columna dice "— (hueco)", ahí está el trabajo pendiente.** Esta tabla se llena desde el código y los tests reales del repositorio, no de memoria; cada test citado existe y se puede correr con `go test -run <Nombre> ./...`.

Convenciones: "rupia" = 1/100,000,000 RUPIX. "DAA" = DAA score (el número de bloque efectivo). "Nivel" = campo `Version` del `ScriptPublicKey` de una salida: 0 Gold, 1 Diamante, 2 Platino, 3 Rodio, 4 Kings.

## 1. Emisión

| Regla | Qué ataque detiene | Test que la viola y confirma el rechazo |
|---|---|---|
| El génesis no emite nada (`SubsidyGenesisReward = 0`) en todas las redes. | Premine escondido en el bloque 0. | `TestNoPremine` (`coinbasemanager_test.go`): falla si alguna red pone subsidio en el génesis. |
| La recompensa por bloque es `50,000,000 >> (DAA / BlocksPerHalving)` rupias: 0.5 RUPIX, mitad en cada halving, cero a partir de la era 26. | Cambiar el calendario de emisión; inflación por decreto. | `TestTotalSupply`: suma el calendario y exige exactamente 41,999,994.96 RUPIX; documenta la cota del exceso por fronteras (0.0031 %). |
| Cada bloque cobra según **su propio** DAA, no el del bloque que lo mergea. Por eso hay un exceso pequeño y acotado en cada frontera de halving (medido: 0.5 RUPIX en el halving 2 de la testnet). | (No es un ataque: es la letra chica, publicada.) | — (hueco: falta un test que construya bloques a ambos lados de una frontera y afirme el exceso exacto). |
| `MaxRupia` (42,000,000 RUPIX) es un tope **por transacción**, no de emisión. | Salidas absurdas que desbordan aritmética. | Heredado de Kaspa (`transactionvalidator`, `checkTransactionAmountRanges`). |
| PoW no se puede apagar en ninguna red pública (`SkipProofOfWork = false`). | Un nodo que acepte bloques sin trabajo. | `TestSkipProofOfWork` (`params_test.go`). |
| El algoritmo es RupixHeavyHash: matriz de rango 64 generada por un PRNG propio, distinto del de Kaspa; entero, no flotante. | Que un ASIC de Kaspa mine Rupix; que dos implementaciones discrepen por redondeo. | `TestRupixPRNGDistintoDeXoshiro`, `TestRupixPRNGDeterminista`, `TestRupixMatrizRango64`, `TestRupixHeavyHashCambia`, `TestRankIntCoincideConFloat` (`utils/pow`), `TestPOW` (`blockvalidator`). |

## 2. Quema por transacción

| Regla | Qué ataque detiene | Test que la viola y confirma el rechazo |
|---|---|---|
| Toda transacción que no sea coinbase quema al menos `BurnBase + BurnPerByte × bytes` (1,000 rupias + 10 por byte) en una salida `OpReturn` de Gold. | Mover Gold gratis; que la escasez no crezca con el uso. | `TestLevelRules` → "ataque: tx sin quema muere", "ataque: quema por debajo del minimo". |
| La coinbase está exenta. | (Regla, no ataque.) | `TestLevelRules` → "coinbase exenta del burn". |
| Las salidas de quema **no entran** al UTXO set. | Contar como circulante lo que se quemó; "resucitar" una quema. | Código: `mutable_utxo_diff.go` (`isBurnScript`). — (hueco: falta un test que intente gastar una salida OpReturn y confirme que no existe en el UTXO set). |

## 3. La escalera (gemas)

| Regla | Qué ataque detiene | Test que la viola y confirma el rechazo |
|---|---|---|
| Un Diamante nace quemando exactamente 10 Gold (10 RUPIX) en OpReturn, además de la quema por transacción. | Diamantes baratos; Diamante sin quema (el Gold iría al minero). | `TestLevelRules` → "ataque: quema insuficiente para diamante (9 gold)", "ataque: diamante sin OpReturn"; casos válidos "10 gold -> 1 diamante", "30 gold -> 3 diamantes". |
| Una gema de nivel N+1 nace consumiendo exactamente 10 gemas de nivel N (`BurnRatio = 10`). | Ascensos con 9; dos Kings con 10 Rodios; doble ascenso; cambio excesivo que esconde una quema corta. | `TestLevelRules` → "ataque: quemar solo 9"; `TestGemAscensionAttacksRound2` → "2 kings con solo 10 rodios", "cambio excesivo esconde quema insuficiente", "doble ascenso desde rodios insuficientes", "transferir y ascender con las mismas piezas"; control "king legitimo con cambio". |
| Cada nivel se abre en su halving: nivel N desde DAA `N × BlocksPerHalving`. Antes, la red rechaza la forja (`nivel N bloqueado`). | Forjar antes de tiempo (visto en vivo el 28-sep: Platino en DAA 185,738, rechazado). | `TestLevelRules` → "ataque: nivel bloqueado", "ataque: diamante antes del halving 1"; `TestGemAscensionAttacksRound2` → "king en unlock-1 (399) debe fallar", "king en unlock exacto (400) debe pasar"; **y en el mempool**: `TestMempoolRechazaNivelCerrado` (la misma tx rechazada antes y aceptada después). |
| Una gema es una salida de 1 rupia con `Version = nivel`; no se fracciona ni se funde. | Partir una gema; "gema" con nivel inexistente. | `TestLevelRules` → "ataque: gema fraccionada", "ataque: nivel fantasma (Version 99)". |
| Una gema solo puede: transferirse (misma cantidad, mismo nivel) o consumirse en un ascenso. No puede desaparecer ni destruirse "sin querer". | Destrucción disfrazada de ascenso; gemas que se esfuman. | `TestLevelRules` → "ataque: gemas desaparecen sin ascenso", "ataque: gema a OpReturn"; caso válido "transferencia de King pasa". |
| La coinbase no puede crear gemas. | Un minero que se regala un King. | `TestLevelRules` → "ataque: la coinbase se regala un King". |
| Corpus: 10,000 transacciones generadas (válidas e inválidas) con el veredicto esperado. | Combinaciones que a mano no se nos ocurren. | `TestLadderCorpus` (`testdata/rupix_ladder_corpus.json`). |
| Topes históricos: 2,100,000 Diamantes, 210,000 Platinos, 21,000 Rodios, 2,100 Kings. Cuentan las gemas que **nacieron** en toda la historia, no las que viven: alcanzado el tope, no nace otra aunque se quemen. Un bloque que los exceda se rechaza (`ErrGemsCapExceeded`, `ErrKingsCapExceeded`). | Más gemas de las que la ley permite. | `TestKingsCountArithmetic` ("el King 2,100 nace, el 2,101 muere"), `TestGemsHistoryArithmetic`, `gemshistory_sanity_test.go`. — (hueco parcial: los topes de Diamante/Platino/Rodio se prueban en la aritmética, no en un bloque real de consenso). |

## 4. El sello de gemas en el encabezado

| Regla | Qué ataque detiene | Test que la viola y confirma el rechazo |
|---|---|---|
| Cada bloque lleva en su encabezado `GemsCommitment` = hash de los conteos históricos (Diamante, Platino, Rodio, Kings) después de aplicar el bloque. Está bajo el PoW. | Mentir sobre cuántas gemas existen; un nodo que "olvida" gemas. | Minero y validador calculan lo mismo: `TestKingsCommitmentMineroIgualValidador`; una transferencia no infla: `TestKingsTransferenciaNoInfla`; de punta a punta con el minero de producción: `TestKingsEndToEnd`. |
| Un bloque cuyo sello no coincide con lo calculado se rechaza (`ErrBadUTXOCommitment` en `verify_and_build_utxo.go`). | Un bloque con conteo falso. | — **(hueco: no hay test automático que fabrique un bloque con sello falso y confirme el rechazo; se probó en vivo en la testnet v0.4.2 pero no está en la suite).** |
| El conteo se persiste por bloque (`GemsHistoryStore`) y se valida contra la pruning proof. | Que un nodo nuevo herede un conteo inventado. | `gemshistory_sanity_test.go` (`pruningproofmanager`). |

## 5. Checkpoints temporales

| Regla | Qué ataque detiene | Test que la viola y confirma el rechazo |
|---|---|---|
| Lista vacía = sin efecto. | (Regla.) | `TestSinCheckpointsPasaTodo`. |
| Con checkpoint `{X, H}`: todo bloque con blue score ≥ X + MergeDepth debe tener a H en el pasado de alguno de sus padres. Un hermano de H no se ve afectado. | Reescribir la historia anterior a H con más hashrate (51 %). | `TestCheckpointDAG` (honesto entra, atacante rechazado justo en el umbral, hermano tardío aceptado; verificado que falla si la regla se apaga), `TestCheckpointAplicaDesdeElMargen`. |
| Pasado `CheckpointsExpireDAAScore`, se ignoran. | (Caducidad verificable.) | `TestCheckpointCaducadoSeIgnora`. |
| El código y `CHECKPOINTS.md` dicen lo mismo. | Un checkpoint cambiado a escondidas. | `TestCheckpointsPublicados`. |
| Un nodo que sincroniza desde cero valida la lista de puntos de poda contra los checkpoints. | Un peer hostil que sirve otro punto de poda con su propia prueba. | — **(hueco: código pendiente para v0.6.2; hasta entonces el checkpoint protege a nodos ya sincronizados).** |

## 6. Red y sincronización (heredado de Kaspa, no reescrito)

GHOSTDAG, ventana DAA, dificultad, poda y pruning proof, merge depth, madurez de coinbase, firma de transacciones: se heredan de kaspad con sus tests (`go test ./...` en verde en Linux y macOS desde el 30-sep-2026). Rupix no los modificó salvo lo listado arriba. Los cambios de Rupix se pueden ver con `git log --oneline` y `CHANGELOG.md`.

## Los huecos, en orden de importancia (el trabajo de la semana)

1. **Sello falso rechazado, automatizado.** Fabricar un bloque válido en todo salvo el `GemsCommitment` y afirmar `ErrBadUTXOCommitment`. Es la afirmación más fuerte del README y no tiene test en la suite.
2. **Nodo nuevo contra peer hostil** (v0.6.2): validar `MsgPruningPoints` después de `ArePruningPointsInValidChain`; test en devnet con un peer que sirve otro punto de poda.
3. **Topes de Diamante/Platino/Rodio en un bloque de consenso**, no solo en la aritmética.
4. **Exceso por frontera de halving**: bloques a ambos lados de un halving, afirmar el exceso exacto y que no puede crecer más que la cota.
5. **Gastar una quema**: intentar gastar una salida OpReturn y confirmar que no existe.
6. **Fuzzing** de `checkLevelRules` y de la validación de transacciones con entradas aleatorias (Go tiene `go test -fuzz`).

Cuando un hueco se cierra, se mueve de esta lista a su fila de la tabla, con el nombre del test.

*No confíes, verifica.*
