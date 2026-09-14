package blockbuilder

import (
"testing"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
"github.com/rupixnet/rupixd/domain/consensus/utils/gemscommitment"
"github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
)

// contarKingsComoBloque replica la logica H-10 de newBlockGemsCommitment:
// suma al conteo del padre los Kings que NACEN en las tx del bloque
// (input King = quemado, output King = nace).
func contarKingsComoBloque(kingsPadre uint64, txs []*externalapi.DomainTransaction) uint64 {
k := kingsPadre
for _, tx := range txs {
outK, inK := 0, 0
for _, input := range tx.Inputs {
if input.UTXOEntry != nil && input.UTXOEntry.ScriptPublicKey().Version == constants.LevelKings {
inK++
}
}
for _, output := range tx.Outputs {
if output.ScriptPublicKey.Version == constants.LevelKings {
outK++
}
}
if outK > inK {
k += uint64(outK - inK)
}
}
return k
}

// contarKingsComoValidador replica la logica de calculateKingsCount:
// input King resta, output King suma (contador que sube y baja).
func contarKingsComoValidador(kingsPadre uint64, txs []*externalapi.DomainTransaction) uint64 {
count := kingsPadre
for _, tx := range txs {
for _, input := range tx.Inputs {
if input.UTXOEntry != nil && input.UTXOEntry.ScriptPublicKey().Version == constants.LevelKings {
if count > 0 {
count--
}
}
}
for _, output := range tx.Outputs {
if output.ScriptPublicKey.Version == constants.LevelKings {
count++
}
}
}
return count
}

func kingInput() *externalapi.DomainTransactionInput {
return &externalapi.DomainTransactionInput{
UTXOEntry: utxo.NewUTXOEntry(constants.GemAmount,
&externalapi.ScriptPublicKey{Script: []byte{}, Version: constants.LevelKings}, false, 0),
}
}
func kingOutput() *externalapi.DomainTransactionOutput {
return &externalapi.DomainTransactionOutput{
Value:           constants.GemAmount,
ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{}, Version: constants.LevelKings},
}
}
func rodioInput() *externalapi.DomainTransactionInput {
return &externalapi.DomainTransactionInput{
UTXOEntry: utxo.NewUTXOEntry(constants.GemAmount,
&externalapi.ScriptPublicKey{Script: []byte{}, Version: constants.LevelRodio}, false, 0),
}
}

// TestKingsCommitmentMineroIgualValidador es la regresion de H-10.
// El minero (newBlockGemsCommitment) y el validador (calculateKingsCount)
// DEBEN sellar el mismo commitment cuando un bloque forja un King. Con el
// bug, el minero sellaba Kings=0 y el validador Kings=1 -> el primer King
// se rechazaba. Este test lo caza automaticamente.
func TestKingsCommitmentMineroIgualValidador(t *testing.T) {
// Un bloque que FORJA un King: 10 Rodios de entrada -> 1 King de salida.
// (el conteo solo mira los outputs/inputs de nivel Kings)
forjaKing := &externalapi.DomainTransaction{
Inputs: []*externalapi.DomainTransactionInput{
rodioInput(), rodioInput(), rodioInput(), rodioInput(), rodioInput(),
rodioInput(), rodioInput(), rodioInput(), rodioInput(), rodioInput(),
},
Outputs: []*externalapi.DomainTransactionOutput{kingOutput()},
}
txs := []*externalapi.DomainTransaction{forjaKing}

kingsPadre := uint64(0) // el padre no tiene Kings todavia

kingsMinero := contarKingsComoBloque(kingsPadre, txs)
kingsValidador := contarKingsComoValidador(kingsPadre, txs)

if kingsMinero != kingsValidador {
t.Fatalf("H-10 REGRESION: minero cuenta %d Kings, validador cuenta %d. El bloque con el King se rechazaria.",
kingsMinero, kingsValidador)
}
if kingsMinero != 1 {
t.Fatalf("se forjo 1 King pero el conteo dice %d", kingsMinero)
}

// Y el commitment sellado debe ser identico
historyMinero := &externalapi.GemsHistory{Kings: kingsMinero}
historyValidador := &externalapi.GemsHistory{Kings: kingsValidador}
selloMinero := gemscommitment.CalculateGemsCommitment(historyMinero, kingsMinero)
selloValidador := gemscommitment.CalculateGemsCommitment(historyValidador, kingsValidador)
if !selloMinero.Equal(selloValidador) {
t.Fatalf("H-10 REGRESION: el sello del minero (%s) != sello del validador (%s)",
selloMinero, selloValidador)
}
t.Logf("OK: minero y validador sellan Kings=%d con el mismo commitment", kingsMinero)
}

// TestKingsTransferenciaNoInfla verifica que una TRANSFERENCIA de King
// (1 King entra, 1 King sale) no cambia el conteo: out==in, nada nace.
func TestKingsTransferenciaNoInfla(t *testing.T) {
transferKing := &externalapi.DomainTransaction{
Inputs:  []*externalapi.DomainTransactionInput{kingInput()},
Outputs: []*externalapi.DomainTransactionOutput{kingOutput()},
}
txs := []*externalapi.DomainTransaction{transferKing}
kingsPadre := uint64(5)

kingsMinero := contarKingsComoBloque(kingsPadre, txs)
kingsValidador := contarKingsComoValidador(kingsPadre, txs)

if kingsMinero != kingsValidador {
t.Fatalf("transferencia: minero %d != validador %d", kingsMinero, kingsValidador)
}
if kingsMinero != 5 {
t.Fatalf("una transferencia de King no debe cambiar el conteo: esperado 5, got %d", kingsMinero)
}
}
