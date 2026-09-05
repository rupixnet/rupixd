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
return nil
}
