package transactionvalidator

import (
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
"github.com/rupixnet/rupixd/domain/consensus/utils/txscript"
"github.com/pkg/errors"
)

// checkLevelRules valida las reglas de la escalera de niveles de Rupix.
// Se llama desde ValidateTransactionInContextAndPopulateFee, cuando los
// UTXOs de entrada ya estan poblados.
//
// Reglas:
//  1. Una tx que crea un output de gema (Version N >= 1) es un ASCENSO y debe:
//     a. Consumir exactamente constants.BurnRatio (10) inputs de Version N-1
//     b. Crear exactamente 1 output de Version N, con monto = constants.GemAmount
//     c. Ocurrir cuando povDaaScore >= LevelUnlockDaaScore(N) (nivel desbloqueado)
//     d. No devolver cambio del nivel quemado: cero outputs de Version N-1
//  2. Las gemas de input solo pueden: transferirse (mismo Version en output,
//     monto 1) o quemarse en un ascenso. Jamas fraccionarse.
//  3. Los inputs Gold (Version 0) adicionales pagan fees; su logica de montos
//     ya la cubre la validacion estandar.
func (v *transactionValidator) checkLevelRules(tx *externalapi.DomainTransaction, povDaaScore uint64) error {
// Conteo de inputs por nivel
inputsPorNivel := make(map[uint16]int)
for _, input := range tx.Inputs {
ver := input.UTXOEntry.ScriptPublicKey().Version
inputsPorNivel[ver]++
}

// Conteo de outputs por nivel + validacion de montos de gema
outputsPorNivel := make(map[uint16]int)
for _, output := range tx.Outputs {
ver := output.ScriptPublicKey.Version
outputsPorNivel[ver]++
if ver >= constants.LevelDiamante && output.Value != constants.GemAmount {
return errors.Wrapf(ruleerrors.ErrBadTxOutValue,
"output de gema (nivel %d) con monto %d: las gemas son piezas enteras, monto debe ser %d",
ver, output.Value, constants.GemAmount)
}
}

// Total de Gold quemado: outputs Version 0 con script imposible de gastar
// (OpReturn). Visible en la cadena para siempre, intocable para todos.
burnedGold := uint64(0)
for _, output := range tx.Outputs {
if output.ScriptPublicKey.Version == constants.LevelGold &&
txscript.IsUnspendable(output.ScriptPublicKey.Script) {
burnedGold += output.Value
}
}

// Revisar cada nivel de gema: ascensos y transferencias
for nivel := constants.LevelDiamante; nivel <= constants.LevelKings; nivel++ {
in := inputsPorNivel[nivel]
outMismoNivel := outputsPorNivel[nivel]
outSuperior := 0
if nivel < constants.LevelKings {
outSuperior = outputsPorNivel[nivel+1]
}

// ¿Esta tx crea gemas de este nivel sin transferir las existentes?
creadas := outMismoNivel - in // >0 significa que nacen gemas nuevas de este nivel
if creadas > 0 {
// ASCENSO: deben nacer de quemar el nivel inferior
inferior := inputsPorNivel[nivel-1]
transferidasInferior := outputsPorNivel[nivel-1]
quemadas := inferior - transferidasInferior

if nivel == constants.LevelDiamante {
// La puerta de entrada a la escalera: el Diamante nace quemando Gold.
// El Gold es divisible, asi que la regla es por MONTO, no por piezas:
// crear N Diamantes exige un output de quema (OpReturn) con
// exactamente N * 10 RUPIX. El cambio en Gold esta permitido
// (el Gold es dinero); la quema queda visible y intocable.
requerido := uint64(creadas) * constants.BurnRatio * constants.RupiaPerRupix
if burnedGold != requerido {
return errors.Wrapf(ruleerrors.ErrBadTxOutValue,
"ascenso a Diamante: crea %d pero quema %d rupias en OpReturn (se requieren exactamente %d)",
creadas, burnedGold, requerido)
}
unlock := constants.LevelUnlockDaaScore(nivel, v.blocksPerHalving)
if povDaaScore < unlock {
return errors.Wrapf(ruleerrors.ErrBadTxOutValue,
"nivel %d bloqueado: se desbloquea en DAA score %d (actual: %d)",
nivel, unlock, povDaaScore)
}
continue
}

if quemadas != constants.BurnRatio*creadas {
return errors.Wrapf(ruleerrors.ErrBadTxOutValue,
"ascenso a nivel %d: crea %d gemas pero quema %d del nivel %d (se requieren %d)",
nivel, creadas, quemadas, nivel-1, constants.BurnRatio*creadas)
}
// ¿El nivel ya esta desbloqueado?
unlock := constants.LevelUnlockDaaScore(nivel, v.blocksPerHalving)
if povDaaScore < unlock {
return errors.Wrapf(ruleerrors.ErrBadTxOutValue,
"nivel %d bloqueado: se desbloquea en DAA score %d (actual: %d)",
nivel, unlock, povDaaScore)
}
_ = outSuperior
} else if creadas < 0 {
// Se consumen mas gemas de las que se transfieren: deben ser quema de ascenso
quemadas := -creadas
nacenSuperior := 0
if nivel < constants.LevelKings {
nacenSuperior = outputsPorNivel[nivel+1] - inputsPorNivel[nivel+1]
}
if nacenSuperior <= 0 || quemadas != constants.BurnRatio*nacenSuperior {
return errors.Wrapf(ruleerrors.ErrBadTxOutValue,
"gemas de nivel %d desaparecen (%d) sin ascenso valido al nivel %d",
nivel, quemadas, nivel+1)
}
}
}
return nil
}
