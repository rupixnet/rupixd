package protowire

import (
	"testing"

	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/blockheader"
	"github.com/rupixnet/rupixd/util/mstime"
	"google.golang.org/protobuf/proto"
	"math/big"
)

// FuzzMsgPruningPoints (hueco #6 de ESPECIFICACION.md): bytes arbitrarios que llegan por
// la red como un KaspadMessage. Es la primera pieza nuestra que toca un peer hostil
// durante la sincronizacion (la lista de pruning points que v0.6.2 valida contra los
// checkpoints). Invariantes:
//  1. Deserializar + ToAppMessage NUNCA entra en panico, con cualquier byte.
//  2. Si ToAppMessage acepta, convertir el mensaje de vuelta (FromAppMessage) y
//     volver a convertirlo da un mensaje de aplicacion identico (ida y vuelta estable),
//     y cada encabezado se puede pasar al dominio sin panico.
//
// La semilla es un MsgPruningPoints valido con dos encabezados, serializado como lo
// mandaria un peer honesto, para que el fuzzer mute desde algo con forma.
//
// Correr: go test -fuzz=FuzzMsgPruningPoints -fuzztime=1h ./infrastructure/network/netadapter/server/grpcserver/protowire/
func FuzzMsgPruningPoints(f *testing.F) {
	semilla := func(n int) []byte {
		headers := make([]*appmessage.MsgBlockHeader, 0, n)
		for i := 0; i < n; i++ {
			h := externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{byte(i + 1)})
			dh := blockheader.NewImmutableBlockHeader(
				1, []externalapi.BlockLevelParents{{h}}, h, h, h, h,
				mstime.Now().UnixMilliseconds(), 0x1e7fffff, uint64(i), uint64(i*10), uint64(i*7),
				big.NewInt(int64(i+1)), h)
			headers = append(headers, appmessage.DomainBlockHeaderToBlockHeader(dh))
		}
		km, err := FromAppMessage(&appmessage.MsgPruningPoints{Headers: headers})
		if err != nil {
			f.Fatalf("semilla: %+v", err)
		}
		b, err := proto.Marshal(km)
		if err != nil {
			f.Fatalf("semilla marshal: %+v", err)
		}
		return b
	}
	f.Add(semilla(0))
	f.Add(semilla(2))
	f.Add([]byte{})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})

	f.Fuzz(func(t *testing.T, datos []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("PANICO con %d bytes: %v", len(datos), r)
			}
		}()
		var km KaspadMessage
		if err := proto.Unmarshal(datos, &km); err != nil {
			return // bytes que no son protobuf valido: rechazo limpio
		}
		msg, err := km.ToAppMessage()
		if err != nil {
			return // mensaje con forma pero invalido: rechazo limpio
		}
		pp, ok := msg.(*appmessage.MsgPruningPoints)
		if !ok {
			return // el fuzzer mudo el tipo de mensaje; no es lo que probamos aqui
		}
		for _, h := range pp.Headers {
			_ = appmessage.BlockHeaderToDomainBlockHeader(h) // no debe entrar en panico
		}
		km2, err := FromAppMessage(pp)
		if err != nil {
			t.Fatalf("un mensaje aceptado no se pudo volver a serializar: %v", err)
		}
		msg2, err := km2.ToAppMessage()
		if err != nil {
			t.Fatalf("ida y vuelta fallo: %v", err)
		}
		pp2 := msg2.(*appmessage.MsgPruningPoints)
		if len(pp2.Headers) != len(pp.Headers) {
			t.Fatalf("ida y vuelta cambio el numero de encabezados: %d -> %d", len(pp.Headers), len(pp2.Headers))
		}
		for i := range pp.Headers {
			a := appmessage.BlockHeaderToDomainBlockHeader(pp.Headers[i])
			b := appmessage.BlockHeaderToDomainBlockHeader(pp2.Headers[i])
			if !a.Equal(b) {
				t.Fatalf("ida y vuelta cambio el encabezado %d", i)
			}
		}
	})
}
