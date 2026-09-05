package pruningproofmanager

import (
"github.com/pkg/errors"
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
)

// validateGemsHistorySanity (Rupix) verifica que el conteo de gemas de un
// pruning proof sea coherente: que ningun nivel exceda su tope historico.
// Es la validacion de cordura del pruning verificable: el nodo nuevo rechaza
// conteos imposibles sin tener que confiar ciegamente en el dato recibido.
// Un gemsHistory nil es valido (ej: pruning point en el genesis, cero gemas).
func validateGemsHistorySanity(gemsHistory *externalapi.GemsHistory) error {
if gemsHistory == nil {
return nil
}
if gemsHistory.Diamante > constants.MaxDiamante {
return errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
"conteo de Diamantes (%d) excede el tope historico (%d)",
gemsHistory.Diamante, constants.MaxDiamante)
}
if gemsHistory.Platino > constants.MaxPlatino {
return errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
"conteo de Platinos (%d) excede el tope historico (%d)",
gemsHistory.Platino, constants.MaxPlatino)
}
if gemsHistory.Rodio > constants.MaxRodio {
return errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
"conteo de Rodios (%d) excede el tope historico (%d)",
gemsHistory.Rodio, constants.MaxRodio)
}
if gemsHistory.Kings > constants.MaxKings {
	return errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
		"conteo de Kings (%d) excede el tope historico (%d)",
		gemsHistory.Kings, constants.MaxKings)
}
// Rupix: COHERENCIA DE ESCALERA (mata al 'cero mentiroso' incoherente).
// Por la mecanica de la escalera, un nivel superior solo existe si
// existieron los inferiores (para quemarlos). Un proof que reporte
// Kings sin Diamantes (o similar) es imposible: miente.
if gemsHistory.Kings > 0 && (gemsHistory.Rodio == 0 || gemsHistory.Platino == 0 || gemsHistory.Diamante == 0) {
	return errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
		"conteo incoherente: hay Kings (%d) pero falta un nivel inferior (Rodio=%d Platino=%d Diamante=%d)",
		gemsHistory.Kings, gemsHistory.Rodio, gemsHistory.Platino, gemsHistory.Diamante)
}
if gemsHistory.Rodio > 0 && (gemsHistory.Platino == 0 || gemsHistory.Diamante == 0) {
	return errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
		"conteo incoherente: hay Rodio (%d) pero falta Platino (%d) o Diamante (%d)",
		gemsHistory.Rodio, gemsHistory.Platino, gemsHistory.Diamante)
}
if gemsHistory.Platino > 0 && gemsHistory.Diamante == 0 {
	return errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
		"conteo incoherente: hay Platino (%d) pero Diamante en 0",
		gemsHistory.Platino)
}
return nil
}
