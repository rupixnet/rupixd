package grpcserver

import (
	"strings"
	"testing"

	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/rupixnet/rupixd/infrastructure/network/netadapter/server/grpcserver/protowire"
)

type conversorQueExplota struct{}

func (conversorQueExplota) ToAppMessage() (appmessage.Message, error) {
	var nada *protowire.RpcFeerateBucket
	return nil, nilDeref(nada)
}

func nilDeref(b *protowire.RpcFeerateBucket) error {
	_ = b.Feerate // panico: nil pointer, como el que encontro el fuzzer
	return nil
}

// TestConversionDeRedNoTiraElNodo (hueco #6, 3-oct-2026): la red de seguridad del bucle
// de recepcion. Si el conversor de un mensaje entra en panico, toAppMessageSinCaer lo
// convierte en error (la conexion se cierra) en vez de dejar que mate al proceso.
func TestConversionDeRedNoTiraElNodo(t *testing.T) {
	msg, err := toAppMessageSinCaer(conversorQueExplota{})
	if err == nil || msg != nil {
		t.Fatalf("un conversor en panico debe volver error y mensaje nil; err=%v msg=%v", err, msg)
	}
	if !strings.Contains(err.Error(), "malformed network message") {
		t.Fatalf("el error debe decir que el mensaje es malformado: %v", err)
	}
	// Control: un mensaje real bien formado pasa igual que antes.
	km, err := protowire.FromAppMessage(&appmessage.MsgPing{Nonce: 7})
	if err != nil {
		t.Fatalf("FromAppMessage: %v", err)
	}
	if _, err := toAppMessageSinCaer(km); err != nil {
		t.Fatalf("un ping bien formado debe convertir: %v", err)
	}
}

// TestListaBlancaPorTipoDeConexion (auditor, 7-oct-2026): un peer P2P no puede mandar
// payloads de RPC y un cliente RPC no puede mandar payloads P2P. Se decide por el numero
// de campo del oneof, antes de convertir nada.
func TestListaBlancaPorTipoDeConexion(t *testing.T) {
	ping, _ := protowire.FromAppMessage(appmessage.NewMsgPing(7))
	dag, _ := protowire.FromAppMessage(&appmessage.GetBlockDAGInfoRequestMessage{})
	fee := &protowire.KaspadMessage{Payload: &protowire.KaspadMessage_GetFeeEstimateResponse{}}
	vacio := &protowire.KaspadMessage{}

	casos := []struct {
		servidor string
		msg      *protowire.KaspadMessage
		permite  bool
		nombre   string
	}{
		{"P2P", ping, true, "P2P acepta Ping"},
		{"P2P", dag, false, "P2P rechaza GetBlockDagInfoRequest"},
		{"P2P", fee, false, "P2P rechaza GetFeeEstimateResponse (el de los 9 bytes)"},
		{"P2P", vacio, false, "P2P rechaza un mensaje sin payload"},
		{"RPC", dag, true, "RPC acepta GetBlockDagInfoRequest"},
		{"RPC", ping, false, "RPC rechaza Ping"},
		{"RPC", vacio, false, "RPC rechaza un mensaje sin payload"},
	}
	for _, c := range casos {
		err := cargaPermitida(c.servidor, c.msg)
		if c.permite && err != nil {
			t.Errorf("%s: debia permitirse, dio %v", c.nombre, err)
		}
		if !c.permite && err == nil {
			t.Errorf("%s: debia rechazarse", c.nombre)
		}
	}
	if n := ping.NumeroDeCarga(); n <= 0 || n >= 1000 {
		t.Fatalf("Ping debe tener numero P2P (<1000), tiene %d", n)
	}
	if n := dag.NumeroDeCarga(); n < 1000 {
		t.Fatalf("GetBlockDagInfoRequest debe tener numero RPC (>=1000), tiene %d", n)
	}
}
