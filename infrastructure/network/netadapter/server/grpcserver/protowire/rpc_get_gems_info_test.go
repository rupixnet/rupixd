package protowire

import (
	"testing"

	"github.com/rupixnet/rupixd/app/appmessage"
	"google.golang.org/protobuf/proto"
)

// TestGetGemsInfoIdaYVuelta (Rupix, 7-oct-2026): el RPC nuevo sobrevive la ida y vuelta
// por el cable con todos sus campos, y los casos nil dan error limpio, nunca panico
// (la leccion del 3-oct aplicada antes de que el fuzzer tenga que recordarnosla).
func TestGetGemsInfoIdaYVuelta(t *testing.T) {
	orig := &appmessage.GetGemsInfoResponseMessage{
		DiamantesNacidos: 22, PlatinosNacidos: 2, RodiosNacidos: 1, KingsNacidos: 1,
		DiamantesVivos: 2, PlatinosVivos: 1, RodiosVivos: 0, KingsVivos: 1,
		TopeDiamantes: 2_100_000, TopePlatinos: 210_000, TopeRodios: 21_000, TopeKings: 2_100,
		VirtualDAAScore: 1_033_755,
	}
	km, err := FromAppMessage(orig)
	if err != nil {
		t.Fatalf("FromAppMessage: %v", err)
	}
	bytes, err := proto.Marshal(km)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var km2 KaspadMessage
	if err := proto.Unmarshal(bytes, &km2); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	back, err := km2.ToAppMessage()
	if err != nil {
		t.Fatalf("ToAppMessage: %v", err)
	}
	got, ok := back.(*appmessage.GetGemsInfoResponseMessage)
	if !ok {
		t.Fatalf("tipo inesperado %T", back)
	}
	if *got != *orig {
		t.Fatalf("ida y vuelta cambio el mensaje:\n  antes %+v\n  despues %+v", *orig, *got)
	}

	// La peticion tambien.
	kmReq, err := FromAppMessage(&appmessage.GetGemsInfoRequestMessage{})
	if err != nil {
		t.Fatalf("FromAppMessage(request): %v", err)
	}
	if _, err := kmReq.ToAppMessage(); err != nil {
		t.Fatalf("request ToAppMessage: %v", err)
	}

	// Nil por dentro: error, no panico.
	if _, err := (&KaspadMessage_GetGemsInfoResponse{}).toAppMessage(); err == nil {
		t.Fatalf("una respuesta sin cuerpo debe dar error")
	}
	var nilResp *GetGemsInfoResponseMessage
	if _, err := nilResp.toAppMessage(); err == nil {
		t.Fatalf("un GetGemsInfoResponseMessage nil debe dar error")
	}
	// Con error RPC dentro, el error viaja.
	conErr := &appmessage.GetGemsInfoResponseMessage{Error: appmessage.RPCErrorf("sin utxoindex")}
	kmErr, err := FromAppMessage(conErr)
	if err != nil {
		t.Fatalf("FromAppMessage(con error): %v", err)
	}
	back2, err := kmErr.ToAppMessage()
	if err != nil {
		t.Fatalf("ToAppMessage(con error): %v", err)
	}
	if back2.(*appmessage.GetGemsInfoResponseMessage).Error == nil {
		t.Fatalf("el error RPC se perdio en el camino")
	}
}

// TestTodaRespuestaRecibibleSePuedeMandar (hallazgo de FuzzKaspadMessage, 7-oct-2026):
// StopNotifyingPruningPointUTXOSetOverrideResponse se podia deserializar pero
// FromAppMessage no la conocia; el nodo nunca podia contestar ese RPC. Aqui se afirma
// la simetria para ese mensaje, y el fuzz la vigila para todos los demas.
func TestTodaRespuestaRecibibleSePuedeMandar(t *testing.T) {
	km := &KaspadMessage{Payload: &KaspadMessage_StopNotifyingPruningPointUTXOSetOverrideResponse{
		StopNotifyingPruningPointUTXOSetOverrideResponse: &StopNotifyingPruningPointUTXOSetOverrideResponseMessage{}}}
	msg, err := km.ToAppMessage()
	if err != nil {
		t.Fatalf("recibir: %v", err)
	}
	if _, err := FromAppMessage(msg); err != nil {
		t.Fatalf("lo que se recibe se debe poder mandar: %v", err)
	}
}
