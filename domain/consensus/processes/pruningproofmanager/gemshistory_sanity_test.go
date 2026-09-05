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
valido := &externalapi.GemsHistory{Diamante: 1000, Platino: 100, Rodio: 10, Kings: 1}
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

// Caso 7: CERO MENTIROSO - Kings sin Diamantes (incoherente) -> RECHAZADO
mentiroso := &externalapi.GemsHistory{Diamante: 0, Platino: 0, Rodio: 0, Kings: 5}
if err := validateGemsHistorySanity(mentiroso); err == nil {
t.Fatal("cero mentiroso (Kings sin Diamantes) NO fue rechazado")
}
t.Log("OK: cero mentiroso (Kings sin niveles inferiores) rechazado")

// Caso 8: Platino sin Diamante (incoherente) -> RECHAZADO
mentiroso2 := &externalapi.GemsHistory{Diamante: 0, Platino: 3, Rodio: 0, Kings: 0}
if err := validateGemsHistorySanity(mentiroso2); err == nil {
t.Fatal("Platino sin Diamante NO fue rechazado")
}
t.Log("OK: Platino sin Diamante rechazado")

// Caso 9: escalera coherente (todos los niveles presentes) -> ACEPTADO
coherente := &externalapi.GemsHistory{Diamante: 10000, Platino: 1000, Rodio: 100, Kings: 10}
if err := validateGemsHistorySanity(coherente); err != nil {
t.Fatalf("conteo coherente rechazado: %v", err)
}
t.Log("OK: escalera coherente aceptada")

// Caso 10: EL CASO DEL AUDITOR - ratios violados (Kings:5 pero pocos inferiores)
// 5 Kings exigen >=50 Rodios, >=500 Platinos, >=5000 Diamantes.
auditorCaso := &externalapi.GemsHistory{Diamante: 1, Platino: 1, Rodio: 1, Kings: 5}
if err := validateGemsHistorySanity(auditorCaso); err == nil {
t.Fatal("(Kings:5, Rodio:1, Platino:1, Diamante:1) NO fue rechazado - viola ratios 10:1")
}
t.Log("OK: conteo que viola ratios 10:1 rechazado (caso del auditor)")

// Caso 11 (HONESTIDAD): un conteo COHERENTE pero falsamente bajo AUN PASA.
// Esto NO lo cierra la cordura - solo el commitment en header. Documentado.
coherenteBajo := &externalapi.GemsHistory{Diamante: 10000, Platino: 1000, Rodio: 100, Kings: 10}
if err := validateGemsHistorySanity(coherenteBajo); err != nil {
t.Fatalf("conteo coherente bajo rechazado (no deberia - la cordura no lo detecta): %v", err)
}
t.Log("HONESTO: conteo coherente bajo PASA (la cordura NO cierra esto; solo commitment en header)")
}
