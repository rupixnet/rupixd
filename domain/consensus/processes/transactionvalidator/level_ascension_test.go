package transactionvalidator

import (
"strings"
"testing"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
"github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
)

// helpers: construir inputs/outputs sinteticos por nivel
func gemInput(version uint16, amount uint64) *externalapi.DomainTransactionInput {
return &externalapi.DomainTransactionInput{
UTXOEntry: utxo.NewUTXOEntry(amount,
&externalapi.ScriptPublicKey{Script: []byte{}, Version: version}, false, 0),
}
}

func gemOutput(version uint16, amount uint64) *externalapi.DomainTransactionOutput {
return &externalapi.DomainTransactionOutput{
Value:           amount,
ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{}, Version: version},
}
}

func makeTx(inputs []*externalapi.DomainTransactionInput,
outputs []*externalapi.DomainTransactionOutput) *externalapi.DomainTransaction {
return &externalapi.DomainTransaction{Inputs: inputs, Outputs: outputs}
}

func newTestValidator() *transactionValidator {
return &transactionValidator{blocksPerHalving: 100} // halvings chicos para test
}

func TestLevelRules(t *testing.T) {
v := newTestValidator()
// Con blocksPerHalving=100: Diamante abre en 100, Platino 200, Rodio 300, Kings 400
const desbloqueado = uint64(1000) // todos los niveles abiertos

t.Run("gold puro pasa siempre", func(t *testing.T) {
tx := makeTx(
[]*externalapi.DomainTransactionInput{gemInput(constants.LevelGold, 500_000)},
[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelGold, 499_000)})
if err := v.checkLevelRules(tx, 0); err != nil {
t.Fatalf("tx de gold puro rechazada: %v", err)
}
})

t.Run("ascenso valido: 10 diamantes -> 1 platino", func(t *testing.T) {
inputs := []*externalapi.DomainTransactionInput{}
for i := 0; i < 10; i++ {
inputs = append(inputs, gemInput(constants.LevelDiamante, constants.GemAmount))
}
tx := makeTx(inputs,
[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelPlatino, constants.GemAmount)})
if err := v.checkLevelRules(tx, desbloqueado); err != nil {
t.Fatalf("ascenso valido rechazado: %v", err)
}
})

t.Run("ataque: quemar solo 9", func(t *testing.T) {
inputs := []*externalapi.DomainTransactionInput{}
for i := 0; i < 9; i++ {
inputs = append(inputs, gemInput(constants.LevelDiamante, constants.GemAmount))
}
tx := makeTx(inputs,
[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelPlatino, constants.GemAmount)})
if err := v.checkLevelRules(tx, desbloqueado); err == nil {
t.Fatal("ascenso con 9 quemados fue ACEPTADO — debe exigir 10")
}
})

t.Run("ataque: nivel bloqueado (platino antes del halving 2)", func(t *testing.T) {
inputs := []*externalapi.DomainTransactionInput{}
for i := 0; i < 10; i++ {
inputs = append(inputs, gemInput(constants.LevelDiamante, constants.GemAmount))
}
tx := makeTx(inputs,
[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelPlatino, constants.GemAmount)})
err := v.checkLevelRules(tx, 150) // platino abre en 200
if err == nil {
t.Fatal("platino creado ANTES de su halving — debe estar bloqueado")
}
if !strings.Contains(err.Error(), "bloqueado") {
t.Fatalf("error inesperado: %v", err)
}
})

t.Run("ataque: gema fraccionada", func(t *testing.T) {
tx := makeTx(
[]*externalapi.DomainTransactionInput{gemInput(constants.LevelRodio, constants.GemAmount)},
[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelRodio, 5)}) // monto != 1
if err := v.checkLevelRules(tx, desbloqueado); err == nil {
t.Fatal("gema con monto 5 ACEPTADA — las gemas son piezas enteras")
}
})

t.Run("ataque: gemas desaparecen sin ascenso", func(t *testing.T) {
tx := makeTx(
[]*externalapi.DomainTransactionInput{
gemInput(constants.LevelRodio, constants.GemAmount),
gemInput(constants.LevelRodio, constants.GemAmount)},
[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelRodio, constants.GemAmount)})
if err := v.checkLevelRules(tx, desbloqueado); err == nil {
t.Fatal("2 rodios entraron, 1 salio, sin ascenso — destruccion invalida ACEPTADA")
}
})

t.Run("transferencia simple de gema pasa", func(t *testing.T) {
tx := makeTx(
[]*externalapi.DomainTransactionInput{gemInput(constants.LevelKings, constants.GemAmount)},
[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelKings, constants.GemAmount)})
if err := v.checkLevelRules(tx, desbloqueado); err != nil {
t.Fatalf("transferencia de King rechazada: %v", err)
}
})

t.Run("gold->diamante bloqueado por ahora", func(t *testing.T) {
tx := makeTx(
[]*externalapi.DomainTransactionInput{gemInput(constants.LevelGold, 10*constants.RupiaPerRupix)},
[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelDiamante, constants.GemAmount)})
if err := v.checkLevelRules(tx, desbloqueado); err == nil {
t.Fatal("creacion de Diamante ACEPTADA — debe estar bloqueada hasta implementar burn de Gold")
}
})
}
