package topes

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
)

func gemIn(level uint16) *externalapi.DomainTransactionInput {
	return &externalapi.DomainTransactionInput{UTXOEntry: utxo.NewUTXOEntry(1, &externalapi.ScriptPublicKey{Version: level}, false, 0)}
}
func gemOut(level uint16) *externalapi.DomainTransactionOutput {
	return &externalapi.DomainTransactionOutput{Value: 1, ScriptPublicKey: &externalapi.ScriptPublicKey{Version: level}}
}
func tx(ins []*externalapi.DomainTransactionInput, outs []*externalapi.DomainTransactionOutput) *externalapi.DomainTransaction {
	return &externalapi.DomainTransaction{Inputs: ins, Outputs: outs}
}

func TestCabe(t *testing.T) {
	c := Desde(&externalapi.GemsHistory{Diamante: constants.MaxDiamante - 1}, 0)
	if !c.Cabe(tx(nil, []*externalapi.DomainTransactionOutput{gemOut(constants.LevelDiamante)})) {
		t.Fatalf("el Diamante %d (el tope exacto) debe caber", constants.MaxDiamante)
	}
	if c.Diamante != constants.MaxDiamante {
		t.Fatalf("conteo tras caber: %d", c.Diamante)
	}
	if c.Cabe(tx(nil, []*externalapi.DomainTransactionOutput{gemOut(constants.LevelDiamante)})) {
		t.Fatalf("el Diamante %d no debe caber", constants.MaxDiamante+1)
	}
	if c.Diamante != constants.MaxDiamante {
		t.Fatalf("un rechazo no debe tocar el conteo: %d", c.Diamante)
	}
	// Una transferencia (1 entra, 1 sale) no nace nada: cabe aunque el tope este lleno.
	if !c.Cabe(tx([]*externalapi.DomainTransactionInput{gemIn(constants.LevelDiamante)}, []*externalapi.DomainTransactionOutput{gemOut(constants.LevelDiamante)})) {
		t.Fatalf("una transferencia de Diamante debe caber con el tope lleno")
	}
	// Kings: tope sobre vivos; consumir uno que no existe no cabe.
	k := Desde(nil, constants.MaxKings)
	if k.Cabe(tx(nil, []*externalapi.DomainTransactionOutput{gemOut(constants.LevelKings)})) {
		t.Fatalf("el King %d no debe caber", constants.MaxKings+1)
	}
	if !k.Cabe(tx([]*externalapi.DomainTransactionInput{gemIn(constants.LevelKings)}, []*externalapi.DomainTransactionOutput{gemOut(constants.LevelKings)})) {
		t.Fatalf("transferir un King con el tope lleno debe caber")
	}
	z := Desde(nil, 0)
	if z.Cabe(tx([]*externalapi.DomainTransactionInput{gemIn(constants.LevelKings)}, nil)) {
		t.Fatalf("consumir un King con conteo 0 no debe caber")
	}
}
