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
// Rupix: COHERENCIA DE ESCALERA POR RATIOS (mata conteos incoherentes).
// Por la mecanica 10:1, cada gema de un nivel exigio quemar 10 del
// inferior. Entonces el conteo HISTORICO de cada nivel debe ser al
// menos 10x el del nivel superior. Un conteo que viole esto es imposible.
// NOTA: esto valida COHERENCIA, NO es verificable total. Un conteo
// coherente pero falsamente bajo aun pasaria. Solo el commitment en
// header (pendiente) cierra el ataque realista por completo.
const ratio = 10
if gemsHistory.Rodio < ratio*gemsHistory.Kings {
	return errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
		"conteo incoherente: Rodio (%d) < 10x Kings (%d): imposible por la escalera 10:1",
		gemsHistory.Rodio, gemsHistory.Kings)
}
if gemsHistory.Platino < ratio*gemsHistory.Rodio {
	return errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
		"conteo incoherente: Platino (%d) < 10x Rodio (%d): imposible por la escalera 10:1",
		gemsHistory.Platino, gemsHistory.Rodio)
}
if gemsHistory.Diamante < ratio*gemsHistory.Platino {
	return errors.Wrapf(ruleerrors.ErrGemsCapExceeded,
		"conteo incoherente: Diamante (%d) < 10x Platino (%d): imposible por la escalera 10:1",
		gemsHistory.Diamante, gemsHistory.Platino)
}
return nil
}
