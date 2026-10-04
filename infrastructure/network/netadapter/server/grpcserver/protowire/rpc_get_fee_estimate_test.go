package protowire

import (
	"testing"

	"google.golang.org/protobuf/proto"
)

// TestMensajeDe9BytesNoTiraElNodo (hueco #6, fuzzing, 3-oct-2026): el fuzzer encontro que
// un KaspadMessage de 9 bytes (una GetFeeEstimateResponse con el PriorityBucket vacio)
// hacia panico en RpcFeeEstimate.toAppMessage. Como el bucle de recepcion convierte
// cualquier mensaje antes de saber que es, y el unico recover del nodo apaga el
// proceso, un peer podia tirar un nodo con ese paquete. El caso exacto vive en
// testdata/fuzz/FuzzMsgPruningPoints/80ec04c90cc2259b y lo corre el fuzz test como
// test normal; aqui se afirma la raiz con los structs, sin depender del corpus:
// conversion = error limpio, nunca panico.
func TestMensajeDe9BytesNoTiraElNodo(t *testing.T) {
	// 1) Los 9 bytes tal cual los guardo el fuzzer.
	datos := []byte("\x9aE\x0600\n\x0000")
	var km KaspadMessage
	if err := proto.Unmarshal(datos, &km); err != nil {
		t.Fatalf("los 9 bytes deben deserializar (asi llego al conversor): %v", err)
	}
	if _, ok := km.Payload.(*KaspadMessage_GetFeeEstimateResponse); !ok {
		t.Fatalf("el caso debe ser una GetFeeEstimateResponse, es %T", km.Payload)
	}
	if _, err := km.ToAppMessage(); err == nil {
		t.Fatalf("una GetFeeEstimateResponse sin PriorityBucket debe dar error, no mensaje")
	}

	// 2) La raiz, con structs: PriorityBucket nil y una cubeta nil.
	if _, err := (&RpcFeeEstimate{}).toAppMessage(); err == nil {
		t.Fatalf("RpcFeeEstimate sin PriorityBucket debe dar error")
	}
	if _, err := (&RpcFeeEstimate{PriorityBucket: &RpcFeerateBucket{}, NormalBuckets: []*RpcFeerateBucket{nil}}).toAppMessage(); err == nil {
		t.Fatalf("una cubeta nil debe dar error")
	}

	// 3) Control: el mensaje bien formado sigue convirtiendo.
	bueno := &KaspadMessage{Payload: &KaspadMessage_GetFeeEstimateResponse{
		GetFeeEstimateResponse: &GetFeeEstimateResponseMessage{
			Estimate: &RpcFeeEstimate{
				PriorityBucket: &RpcFeerateBucket{Feerate: 2, EstimatedSeconds: 1},
				NormalBuckets:  []*RpcFeerateBucket{{Feerate: 1, EstimatedSeconds: 10}},
			},
		},
	}}
	if _, err := bueno.ToAppMessage(); err != nil {
		t.Fatalf("la respuesta bien formada debe convertir: %v", err)
	}
}
