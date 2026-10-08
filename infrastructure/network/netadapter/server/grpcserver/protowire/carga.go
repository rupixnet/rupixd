package protowire

import "google.golang.org/protobuf/reflect/protoreflect"

// Rupix (7-oct-2026, pedido del auditor tras el fallo de los 9 bytes): en messages.proto
// los payloads del protocolo P2P tienen numeros de campo menores a 1000 y los del RPC,
// 1000 o mas. Con eso una conexion puede rechazar, ANTES de convertir nada, cualquier
// payload que no pertenezca a su protocolo: un peer P2P no tiene por que mandar una
// respuesta de RPC, y un cliente RPC no tiene por que mandar bloques.

// primerNumeroRPC es el primer numero de campo del oneof `payload` que pertenece al RPC.
const primerNumeroRPC = 1000

// NumeroDeCarga devuelve el numero de campo del payload presente en el oneof, o 0 si no
// hay payload.
func (x *KaspadMessage) NumeroDeCarga() int32 {
	if x == nil {
		return 0
	}
	m := x.ProtoReflect()
	oneof := m.Descriptor().Oneofs().ByName("payload")
	if oneof == nil {
		return 0
	}
	fd := m.WhichOneof(oneof)
	if fd == nil {
		return 0
	}
	return int32(fd.Number())
}

// EsCargaP2P dice si el payload pertenece al protocolo entre nodos.
func (x *KaspadMessage) EsCargaP2P() bool {
	n := x.NumeroDeCarga()
	return n > 0 && n < primerNumeroRPC
}

// EsCargaRPC dice si el payload pertenece al RPC.
func (x *KaspadMessage) EsCargaRPC() bool {
	return x.NumeroDeCarga() >= primerNumeroRPC
}

var _ protoreflect.ProtoMessage = (*KaspadMessage)(nil)
