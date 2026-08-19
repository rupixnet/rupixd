package utxo

import (
"testing"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
)

// TestBurnOutputsExcludedFromUTXOSet verifica que los outputs de quema
// (OpReturn) jamas entran al UTXO set: la quema es visible en la historia
// del bloque pero no existe en el inventario de lo gastable.
func TestBurnOutputsExcludedFromUTXOSet(t *testing.T) {
normalScript := &externalapi.ScriptPublicKey{Script: []byte{0x76, 0xa9}, Version: 0}
burnScript := &externalapi.ScriptPublicKey{Script: []byte{0x6a}, Version: 0} // OpReturn

tx := &externalapi.DomainTransaction{
Inputs: []*externalapi.DomainTransactionInput{},
Outputs: []*externalapi.DomainTransactionOutput{
{Value: 5_000, ScriptPublicKey: normalScript},  // indice 0: normal
{Value: 10_000, ScriptPublicKey: burnScript},   // indice 1: QUEMA
{Value: 2_000, ScriptPublicKey: normalScript},  // indice 2: normal
},
}

diff := NewMutableUTXODiff()
err := diff.AddTransaction(tx, 500)
if err != nil {
t.Fatalf("AddTransaction fallo: %v", err)
}

toAdd := diff.ToImmutable().ToAdd()
if toAdd.Len() != 2 {
t.Fatalf("el UTXO set tiene %d entradas; deben ser 2 (la quema excluida)", toAdd.Len())
}

// Verificar que la entrada ausente es exactamente la quema (indice 1)
iterator := toAdd.Iterator()
defer iterator.Close()
for ok := iterator.First(); ok; ok = iterator.Next() {
outpoint, entry, err := iterator.Get()
if err != nil {
t.Fatal(err)
}
if outpoint.Index == 1 {
t.Fatalf("la QUEMA (indice 1, %d rupias) entro al UTXO set — debe ser ceniza", entry.Amount())
}
}
t.Log("quema excluida del set: 2 outputs normales dentro, la ceniza fuera")
}
