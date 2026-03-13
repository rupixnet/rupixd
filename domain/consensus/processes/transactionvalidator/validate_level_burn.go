package transactionvalidator

import (
    "github.com/rupixnet/rupixd/domain/consensus/model"
    "github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
    "github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
    "github.com/rupixnet/rupixd/domain/consensus/utils/constants"
    "github.com/pkg/errors"
)

func (v *transactionValidator) validateLevelBurn(tx *externalapi.DomainTransaction, povBlockHash *externalapi.DomainHash, stagingArea *model.StagingArea) error {
    if len(tx.Inputs) == 0 || len(tx.Outputs) == 0 {
        return nil
    }
    if tx.Inputs[0].UTXOEntry == nil {
        return nil
    }
    firstLevel := tx.Inputs[0].UTXOEntry.ScriptPublicKey().Version
    if firstLevel == constants.LevelOro {
        return nil
    }
    for _, input := range tx.Inputs {
        if input.UTXOEntry == nil {
            return nil
        }
        if input.UTXOEntry.ScriptPublicKey().Version != firstLevel {
            return errors.Wrapf(ruleerrors.ErrInvalidLevelBurn, "inputs de niveles mixtos")
        }
    }
    if uint64(len(tx.Inputs)) != constants.BurnRatio {
        return errors.Wrapf(ruleerrors.ErrInvalidLevelBurn, "se requieren exactamente %d inputs, tiene %d", constants.BurnRatio, len(tx.Inputs))
    }
    if len(tx.Outputs) != 1 {
        return errors.Wrapf(ruleerrors.ErrInvalidLevelBurn, "debe tener exactamente 1 output")
    }
    targetLevel := firstLevel + 1
    if tx.Outputs[0].ScriptPublicKey.Version != targetLevel {
        return errors.Wrapf(ruleerrors.ErrInvalidLevelBurn, "output debe ser nivel %d", targetLevel)
    }
    povDAAScore, err := v.daaBlocksStore.DAAScore(v.databaseContext, stagingArea, povBlockHash)
    if err != nil {
        return err
    }
    switch targetLevel {
    case constants.LevelDiamante:
        if povDAAScore < constants.LevelDiamanteUnlockScore {
            return errors.Wrapf(ruleerrors.ErrInvalidLevelBurn, "nivel Diamante no desbloqueado aun")
        }
    case constants.LevelPlatino:
        if povDAAScore < constants.LevelPlatinoUnlockScore {
            return errors.Wrapf(ruleerrors.ErrInvalidLevelBurn, "nivel Platino no desbloqueado aun")
        }
    case constants.LevelRodio:
        if povDAAScore < constants.LevelRodioUnlockScore {
            return errors.Wrapf(ruleerrors.ErrInvalidLevelBurn, "nivel Rodio no desbloqueado aun")
        }
    case constants.LevelKings:
        if povDAAScore < constants.LevelKingsUnlockScore {
            return errors.Wrapf(ruleerrors.ErrInvalidLevelBurn, "Kings no desbloqueado aun")
        }
    default:
        return errors.Wrapf(ruleerrors.ErrInvalidLevelBurn, "nivel destino invalido: %d", targetLevel)
    }
    return nil
}