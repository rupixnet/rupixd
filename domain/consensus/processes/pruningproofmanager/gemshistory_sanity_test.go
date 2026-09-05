package pruningproofmanager

import (
"testing"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
"github.com/pkg/errors"
)

// TestValidateGemsHistorySanity (Rupix) prueba la validacion de cordura del
// pruning verificable: acepta conteos validos, rechaza conteos imposibles.
func TestValidateGemsHistorySanity(t *testing.T) {
// Caso 1: conteo valido -> sin error
valido := &externalapi.GemsHistory{Diamante: 5, Platino: 2, Rodio: 1}
if err := validateGemsHistorySanity(valido); err != nil {
t.Fatalf("conteo valido rechazado: %v", err)
}
t.Log("OK: conteo valido aceptado")

// Caso 2: nil -> sin error (genesis)
if err := validateGemsHistorySanity(nil); err != nil {
t.Fatalf("nil rechazado: %v", err)
}
t.Log("OK: nil aceptado (genesis)")

// Caso 3: Diamantes exceden el tope -> RECHAZADO
excesivo := &externalapi.GemsHistory{Diamante: 9_000_000, Platino: 0, Rodio: 0}
err := validateGemsHistorySanity(excesivo)
if err == nil {
t.Fatal("conteo excesivo de Diamantes NO fue rechazado (deberia serlo)")
}
if !errors.Is(err, ruleerrors.ErrGemsCapExceeded) {
t.Fatalf("error incorrecto: %v", err)
}
t.Log("OK: exceso de Diamantes rechazado correctamente")

// Caso 4: Platinos exceden el tope -> RECHAZADO
excesivoP := &externalapi.GemsHistory{Diamante: 0, Platino: 500_000, Rodio: 0}
if err := validateGemsHistorySanity(excesivoP); err == nil {
t.Fatal("exceso de Platinos NO fue rechazado")
}
t.Log("OK: exceso de Platinos rechazado")

// Caso 5: en el tope exacto -> ACEPTADO (no lo excede)
enTope := &externalapi.GemsHistory{Diamante: 2_100_000, Platino: 210_000, Rodio: 21_000}
if err := validateGemsHistorySanity(enTope); err != nil {
t.Fatalf("conteo en el tope exacto rechazado (no deberia): %v", err)
}
t.Log("OK: conteo en el tope exacto aceptado (sin off-by-one)")

// Caso 6: Kings exceden el tope -> RECHAZADO
excesivoK := &externalapi.GemsHistory{Diamante: 0, Platino: 0, Rodio: 0, Kings: 3000}
if err := validateGemsHistorySanity(excesivoK); err == nil {
t.Fatal("exceso de Kings NO fue rechazado")
}
t.Log("OK: exceso de Kings rechazado")
}
