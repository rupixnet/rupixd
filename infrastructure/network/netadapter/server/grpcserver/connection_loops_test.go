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
