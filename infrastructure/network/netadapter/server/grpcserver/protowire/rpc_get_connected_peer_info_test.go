package protowire

import (
	"testing"

	"github.com/rupixnet/rupixd/app/appmessage"
)

// TestTimeConnectedSobreviveIdaYVuelta (Rupix, 4-nov-2026, bug heredado de kaspad): al
// convertir la respuesta de GetConnectedPeerInfo de protobuf a appmessage, TimeConnected se
// copiaba desde TimeOffset. Todo cliente RPC (el explorador, rupixctl) veia el desfase de
// reloj (≈0) en vez del tiempo conectado; el mapa decia "20761 d" (dias desde 1970) y,
// corregido el primer sintoma, "0 s". Ida y vuelta: cada campo vuelve a su lugar.
func TestTimeConnectedSobreviveIdaYVuelta(t *testing.T) {
	original := &appmessage.GetConnectedPeerInfoResponseMessage{Infos: []*appmessage.GetConnectedPeerInfoMessage{{
		ID: "abc", Address: "203.0.113.5:17211", LastPingDuration: 42, IsOutbound: true, TimeOffset: -7,
		UserAgent: "/rupixd:0.7.0/", AdvertisedProtocolVersion: 5, TimeConnected: 88 * 3600 * 1000, IsIBDPeer: true,
	}}}
	var x KaspadMessage_GetConnectedPeerInfoResponse
	if err := x.fromAppMessage(original); err != nil {
		t.Fatal(err)
	}
	vuelta, err := x.toAppMessage()
	if err != nil {
		t.Fatal(err)
	}
	got := vuelta.(*appmessage.GetConnectedPeerInfoResponseMessage).Infos[0]
	want := original.Infos[0]
	if *got != *want {
		t.Fatalf("ida y vuelta cambio algo:\n got  %+v\n want %+v", *got, *want)
	}
}
