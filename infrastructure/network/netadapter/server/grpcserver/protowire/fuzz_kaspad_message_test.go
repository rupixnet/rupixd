package protowire

import (
	"testing"

	"github.com/rupixnet/rupixd/app/appmessage"
	"google.golang.org/protobuf/proto"
)

// FuzzKaspadMessage (auditor, 7-oct-2026): el objetivo con nombre propio para lo que
// FuzzMsgPruningPoints encontro por accidente. Muta el tipo de wire completo
// (KaspadMessage, el oneof con ~150 payloads) desde semillas de muchos tipos distintos,
// P2P y RPC. Invariantes:
//  1. Deserializar + ToAppMessage NUNCA entra en panico, con cualquier byte.
//  2. Si ToAppMessage acepta, ida y vuelta estable (FromAppMessage + ToAppMessage).
//  3. La clasificacion por tipo de conexion es total: todo mensaje con payload es P2P o
//     RPC (nunca ninguno, nunca ambos), y uno sin payload no es ninguno.
//
// Correr: go test -fuzz=FuzzKaspadMessage -fuzztime=2h ./infrastructure/network/netadapter/server/grpcserver/protowire/
func FuzzKaspadMessage(f *testing.F) {
	semillas := []appmessage.Message{
		appmessage.NewMsgPing(1),
		appmessage.NewMsgPong(1),
		&appmessage.MsgRequestAddresses{},
		&appmessage.MsgRequestIBDBlocks{},
		&appmessage.MsgPruningPoints{},
		&appmessage.MsgTrustedData{},
		&appmessage.BlockHeadersMessage{},
		&appmessage.MsgRequestHeaders{},
		&appmessage.MsgDoneHeaders{},
		&appmessage.MsgReady{},
		&appmessage.MsgRequestPruningPointUTXOSet{},
		&appmessage.MsgDonePruningPointUTXOSetChunks{},
		&appmessage.MsgInvRelayBlock{},
		&appmessage.MsgInvTransaction{},
		&appmessage.MsgRequestRelayBlocks{},
		&appmessage.MsgReject{},
		&appmessage.GetBlockDAGInfoRequestMessage{},
		&appmessage.GetBlockDAGInfoResponseMessage{},
		&appmessage.GetCoinSupplyResponseMessage{},
		&appmessage.GetFeeEstimateResponseMessage{},
		&appmessage.GetFeeEstimateRequestMessage{},
		&appmessage.GetGemsInfoResponseMessage{},
		&appmessage.GetGemsInfoRequestMessage{},
		&appmessage.GetBlockRequestMessage{},
		&appmessage.GetBlocksResponseMessage{},
		&appmessage.GetUTXOsByAddressesResponseMessage{},
		&appmessage.GetMempoolEntriesResponseMessage{},
		&appmessage.GetInfoResponseMessage{},
		&appmessage.SubmitBlockRequestMessage{},
		&appmessage.SubmitTransactionRequestMessage{},
		&appmessage.NotifyBlockAddedRequestMessage{},
		&appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage{},
	}
	for _, m := range semillas {
		b, ok := semillaSerializada(m)
		if !ok {
			// Algunos mensajes de aplicacion vacios no se pueden serializar (su conversor
			// de SALIDA espera campos poblados; esa direccion la construye el propio nodo,
			// no un peer). No es un fallo del fuzz; la semilla simplemente se omite.
			f.Logf("semilla %T omitida: su conversor de salida no acepta el mensaje vacio", m)
			continue
		}
		f.Add(b)
	}
	f.Add([]byte("\x9aE\x0600\n\x0000")) // el caso de los 9 bytes del 3-oct
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, datos []byte) {
		var km KaspadMessage
		if err := proto.Unmarshal(datos, &km); err != nil {
			return
		}
		// 3) clasificacion total.
		p2p, rpc := km.EsCargaP2P(), km.EsCargaRPC()
		if km.Payload == nil {
			if p2p || rpc {
				t.Fatalf("sin payload no puede ser P2P ni RPC")
			}
		} else if p2p == rpc {
			t.Fatalf("payload %d (%T): P2P=%v RPC=%v; debe ser exactamente uno", km.NumeroDeCarga(), km.Payload, p2p, rpc)
		}
		// 1) nunca panico (sin recover: si entra en panico, el fuzzer lo reporta con pila).
		msg, err := km.ToAppMessage()
		if err != nil {
			return
		}
		// 2) ida y vuelta.
		km2, err := FromAppMessage(msg)
		if err != nil {
			t.Fatalf("acepto %T pero no lo puede volver a serializar: %v", msg, err)
		}
		msg2, err := km2.ToAppMessage()
		if err != nil {
			t.Fatalf("la segunda conversion de %T fallo: %v", msg, err)
		}
		if msg.Command() != msg2.Command() {
			t.Fatalf("ida y vuelta cambio el comando: %s -> %s", msg.Command(), msg2.Command())
		}
	})
}

// semillaSerializada convierte y serializa un mensaje de aplicacion; si el conversor de
// salida entra en panico o falla con el mensaje vacio, devuelve ok=false.
func semillaSerializada(m appmessage.Message) (b []byte, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			b, ok = nil, false
		}
	}()
	km, err := FromAppMessage(m)
	if err != nil {
		return nil, false
	}
	b, err = proto.Marshal(km)
	if err != nil {
		return nil, false
	}
	return b, true
}
