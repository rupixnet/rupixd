package consensusstatemanager

import (
"github.com/pkg/errors"

"github.com/rupixnet/rupixd/domain/consensus/model"
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
)

// calculateGemsHistory es el espejo de calculateKingsCount para los topes
// HISTORICOS de Diamante, Platino y Rodio. A diferencia de Kings (que cuenta
// gemas VIVAS y por eso resta al gastarse), los topes historicos SOLO SUBEN:
//
//historia(bloque) = historia(selectedParent) + gemas NACIDAS en el bloque
//
// donde "nacidas" = max(0, outputs(nivel) - inputs(nivel)) por transaccion.
// Asi las transferencias no inflan (out-in = 0), y las quemas para ascender
// NO devuelven cupo al nivel inferior. Una vez alcanzado el tope de un nivel,
// jamas nace otra gema de ese nivel en ninguna cadena valida — el bloque que
// lo intente es invalido por entero.
func (csm *consensusStateManager) calculateGemsHistory(stagingArea *model.StagingArea,
blockHash *externalapi.DomainHash,
acceptanceData externalapi.AcceptanceData,
blockGHOSTDAGData *externalapi.BlockGHOSTDAGData) (*externalapi.GemsHistory, error) {

if blockHash.Equal(csm.genesisHash) {
return &externalapi.GemsHistory{}, nil // genesis: cero gemas de todo nivel (sin premine)
}

parent, err := csm.gemsHistoryStore.Get(csm.databaseContext, stagingArea, blockGHOSTDAGData.SelectedParent())
if err != nil {
return nil, err
}
// Copia mutable a partir del padre
history := parent.Clone()

for _, blockAcceptanceData := range acceptanceData {
for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
if !transactionAcceptanceData.IsAccepted {
continue
}
tx := transactionAcceptanceData.Transaction

// Contar inputs y outputs por nivel en esta tx
inD, inP, inR := 0, 0, 0
for _, input := range tx.Inputs {
switch input.UTXOEntry.ScriptPublicKey().Version {
case constants.LevelDiamante:
inD++
case constants.LevelPlatino:
inP++
case constants.LevelRodio:
inR++
}
}
outD, outP, outR := 0, 0, 0
for _, output := range tx.Outputs {
switch output.ScriptPublicKey.Version {
case constants.LevelDiamante:
outD++
case constants.LevelPlatino:
outP++
case constants.LevelRodio:
outR++
}
}

// Nacimientos netos = max(0, out - in). Solo suma, nunca resta.
history.Diamante += nacidosNetos(outD, inD)
history.Platino += nacidosNetos(outP, inP)
history.Rodio += nacidosNetos(outR, inR)
}
}

// La muralla: si algun nivel supera su tope historico, el bloque es invalido.
if history.Diamante > constants.MaxDiamante {
return nil, errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
"el bloque %s llevaria los Diamantes historicos a %d, tope %d",
blockHash, history.Diamante, constants.MaxDiamante)
}
if history.Platino > constants.MaxPlatino {
return nil, errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
"el bloque %s llevaria los Platinos historicos a %d, tope %d",
blockHash, history.Platino, constants.MaxPlatino)
}
if history.Rodio > constants.MaxRodio {
return nil, errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
"el bloque %s llevaria los Rodios historicos a %d, tope %d",
blockHash, history.Rodio, constants.MaxRodio)
}

return history, nil
}

// nacidosNetos devuelve max(0, out-in): cuantas gemas nuevas de un nivel
// aparecen en una tx que no provienen de un input del mismo nivel.
func nacidosNetos(out, in int) uint64 {
if out <= in {
return 0
}
return uint64(out - in)
}
