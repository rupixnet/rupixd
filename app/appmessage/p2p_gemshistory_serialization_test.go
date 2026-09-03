package appmessage

import (
"testing"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
)

// TestGemsHistorySerialization (Rupix) verifica que el conteo de gemas
// sobrevive el viaje domain -> appmessage -> domain sin corromperse.
// Es la prueba del pruning verificable: el conteo viaja intacto.
func TestGemsHistorySerialization(t *testing.T) {
// 1. Crear un proof en domain con un conteo de gemas conocido
original := &externalapi.PruningPointProof{
Headers: [][]externalapi.BlockHeader{},
GemsHistory: &externalapi.GemsHistory{
Diamante: 5,
Platino:  2,
Rodio:    1,
},
}

// 2. domain -> appmessage
msg := DomainPruningPointProofToMsgPruningPointProof(original)
if msg.GemsHistory == nil {
t.Fatal("GemsHistory se perdio en domain->appmessage")
}
if msg.GemsHistory.Diamante != 5 || msg.GemsHistory.Platino != 2 || msg.GemsHistory.Rodio != 1 {
t.Fatalf("conteo corrupto en domain->appmessage: %+v", msg.GemsHistory)
}

// 3. appmessage -> domain (viaje de vuelta)
back := MsgPruningPointProofToDomainPruningPointProof(msg)
if back.GemsHistory == nil {
t.Fatal("GemsHistory se perdio en appmessage->domain")
}
if back.GemsHistory.Diamante != 5 || back.GemsHistory.Platino != 2 || back.GemsHistory.Rodio != 1 {
t.Fatalf("conteo corrupto en appmessage->domain: %+v", back.GemsHistory)
}

// 4. Confirmar que es identico al original
if back.GemsHistory.Diamante != original.GemsHistory.Diamante ||
back.GemsHistory.Platino != original.GemsHistory.Platino ||
back.GemsHistory.Rodio != original.GemsHistory.Rodio {
t.Fatal("el conteo NO sobrevivio el viaje completo")
}
t.Logf("OK: conteo de gemas sobrevivio intacto (Diamante=%d Platino=%d Rodio=%d)",
back.GemsHistory.Diamante, back.GemsHistory.Platino, back.GemsHistory.Rodio)
}

// TestGemsHistoryNil verifica que un proof sin gemas (nil) no rompe nada.
func TestGemsHistoryNil(t *testing.T) {
original := &externalapi.PruningPointProof{
Headers:     [][]externalapi.BlockHeader{},
GemsHistory: nil,
}
msg := DomainPruningPointProofToMsgPruningPointProof(original)
back := MsgPruningPointProofToDomainPruningPointProof(msg)
if back.GemsHistory != nil {
t.Fatal("GemsHistory deberia seguir nil pero no lo esta")
}
t.Log("OK: proof sin gemas (nil) se maneja sin romper")
}
