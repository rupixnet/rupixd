package consensusstatemanager

import (
"testing"

"github.com/pkg/errors"
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
"github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
)

// gemsHistoryDelta reproduce la aritmetica de calculateGemsHistory: cuenta
// nacimientos netos por nivel (solo sube) y aplica los topes historicos.
func gemsHistoryDelta(parent *externalapi.GemsHistory, acceptanceData externalapi.AcceptanceData) (*externalapi.GemsHistory, error) {
h := parent.Clone()
for _, bad := range acceptanceData {
for _, tad := range bad.TransactionAcceptanceData {
if !tad.IsAccepted {
continue
}
tx := tad.Transaction
inD, inP, inR := 0, 0, 0
for _, in := range tx.Inputs {
switch in.UTXOEntry.ScriptPublicKey().Version {
case constants.LevelDiamante:
inD++
case constants.LevelPlatino:
inP++
case constants.LevelRodio:
inR++
}
}
outD, outP, outR := 0, 0, 0
for _, out := range tx.Outputs {
switch out.ScriptPublicKey.Version {
case constants.LevelDiamante:
outD++
case constants.LevelPlatino:
outP++
case constants.LevelRodio:
outR++
}
}
h.Diamante += nacidosNetos(outD, inD)
h.Platino += nacidosNetos(outP, inP)
h.Rodio += nacidosNetos(outR, inR)
}
}
if h.Diamante > constants.MaxDiamante {
return nil, errors.Wrapf(ruleerrors.ErrGemsCapExceeded, "diamante %d > %d", h.Diamante, constants.MaxDiamante)
}
if h.Platino > constants.MaxPlatino {
return nil, errors.Wrapf(ruleerrors.ErrGemsCapExceeded, "platino %d > %d", h.Platino, constants.MaxPlatino)
}
if h.Rodio > constants.MaxRodio {
return nil, errors.Wrapf(ruleerrors.ErrGemsCapExceeded, "rodio %d > %d", h.Rodio, constants.MaxRodio)
}
return h, nil
}

func gemIn(level uint16) *externalapi.DomainTransactionInput {
return &externalapi.DomainTransactionInput{
UTXOEntry: utxo.NewUTXOEntry(1, &externalapi.ScriptPublicKey{Version: level}, false, 0)}
}
func gemOut(level uint16) *externalapi.DomainTransactionOutput {
return &externalapi.DomainTransactionOutput{Value: 1, ScriptPublicKey: &externalapi.ScriptPublicKey{Version: level}}
}
func accept(txs ...*externalapi.DomainTransaction) externalapi.AcceptanceData {
tads := []*externalapi.TransactionAcceptanceData{}
for _, tx := range txs {
tads = append(tads, &externalapi.TransactionAcceptanceData{Transaction: tx, IsAccepted: true})
}
return externalapi.AcceptanceData{{TransactionAcceptanceData: tads}}
}

func TestGemsHistoryArithmetic(t *testing.T) {
D, P, R := constants.LevelDiamante, constants.LevelPlatino, constants.LevelRodio

// 1. Nacen 3 Diamantes sobre 0 (forja desde Gold: sin input de gema)
ad := accept(&externalapi.DomainTransaction{Outputs: []*externalapi.DomainTransactionOutput{gemOut(D), gemOut(D), gemOut(D)}})
h, err := gemsHistoryDelta(&externalapi.GemsHistory{}, ad)
if err != nil || h.Diamante != 3 {
t.Fatalf("3 diamantes nacen: esperado 3, got %d (%v)", h.Diamante, err)
}

// 2. TRANSFERENCIA de 1 Diamante (1 entra, 1 sale) -> NO infla el historico
ad = accept(&externalapi.DomainTransaction{
Inputs:  []*externalapi.DomainTransactionInput{gemIn(D)},
Outputs: []*externalapi.DomainTransactionOutput{gemOut(D)}})
h, _ = gemsHistoryDelta(&externalapi.GemsHistory{Diamante: 100}, ad)
if h.Diamante != 100 {
t.Fatalf("transferencia NO debe inflar: esperado 100, got %d", h.Diamante)
}

// 3. QUEMA: 10 Diamantes -> 1 Platino. El historico de Diamante NO baja (solo sube).
inputs := []*externalapi.DomainTransactionInput{}
for i := 0; i < 10; i++ {
inputs = append(inputs, gemIn(D))
}
ad = accept(&externalapi.DomainTransaction{
Inputs:  inputs,
Outputs: []*externalapi.DomainTransactionOutput{gemOut(P)}})
h, _ = gemsHistoryDelta(&externalapi.GemsHistory{Diamante: 1000, Platino: 5}, ad)
if h.Diamante != 1000 {
t.Fatalf("HISTORICO: quemar diamantes NO baja el conteo: esperado 1000, got %d", h.Diamante)
}
if h.Platino != 6 {
t.Fatalf("nace 1 platino: esperado 6, got %d", h.Platino)
}

// 4. El Diamante 2,100,000 (el ultimo permitido) NACE
ad = accept(&externalapi.DomainTransaction{Outputs: []*externalapi.DomainTransactionOutput{gemOut(D)}})
h, err = gemsHistoryDelta(&externalapi.GemsHistory{Diamante: constants.MaxDiamante - 1}, ad)
if err != nil || h.Diamante != constants.MaxDiamante {
t.Fatalf("el Diamante 2,100,000 debe nacer: got %d (%v)", h.Diamante, err)
}

// 5. EL DIAMANTE 2,100,001 MUERE
_, err = gemsHistoryDelta(&externalapi.GemsHistory{Diamante: constants.MaxDiamante}, ad)
if !errors.Is(err, ruleerrors.ErrGemsCapExceeded) {
t.Fatalf("el Diamante 2,100,001 debe rechazarse, got: %v", err)
}

// 6. Platino en su tope
adP := accept(&externalapi.DomainTransaction{Outputs: []*externalapi.DomainTransactionOutput{gemOut(P)}})
_, err = gemsHistoryDelta(&externalapi.GemsHistory{Platino: constants.MaxPlatino}, adP)
if !errors.Is(err, ruleerrors.ErrGemsCapExceeded) {
t.Fatalf("el Platino 210,001 debe rechazarse, got: %v", err)
}

// 7. Rodio en su tope
adR := accept(&externalapi.DomainTransaction{Outputs: []*externalapi.DomainTransactionOutput{gemOut(R)}})
_, err = gemsHistoryDelta(&externalapi.GemsHistory{Rodio: constants.MaxRodio}, adR)
if !errors.Is(err, ruleerrors.ErrGemsCapExceeded) {
t.Fatalf("el Rodio 21,001 debe rechazarse, got: %v", err)
}

t.Logf("Historico sellado: Diamante 2,100,000 nace, 2,100,001 muere. Quemar no baja el conteo.")
}
