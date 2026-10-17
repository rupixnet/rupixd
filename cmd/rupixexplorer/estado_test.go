package main

import (
	"errors"
	"testing"
	"time"

	"github.com/rupixnet/rupixd/app/appmessage"
)

type fuenteFalsa struct {
	info  *appmessage.GetInfoResponseMessage
	dag   *appmessage.GetBlockDAGInfoResponseMessage
	bloq  map[string]int64 // hash -> timestamp ms
	gems  *appmessage.GetGemsInfoResponseMessage
	caido bool
	ll    int
}

func (f *fuenteFalsa) GetInfo() (*appmessage.GetInfoResponseMessage, error) {
	f.ll++
	if f.caido {
		return nil, errors.New("connection refused")
	}
	return f.info, nil
}
func (f *fuenteFalsa) GetBlockDAGInfo() (*appmessage.GetBlockDAGInfoResponseMessage, error) {
	return f.dag, nil
}
func (f *fuenteFalsa) GetBlock(h string, _ bool) (*appmessage.GetBlockResponseMessage, error) {
	ts, ok := f.bloq[h]
	if !ok {
		return nil, errors.New("no existe")
	}
	return &appmessage.GetBlockResponseMessage{Block: &appmessage.RPCBlock{Header: &appmessage.RPCBlockHeader{Timestamp: ts}}}, nil
}
func (f *fuenteFalsa) GetGemsInfo() (*appmessage.GetGemsInfoResponseMessage, error) {
	return f.gems, nil
}

// TestEstadoDeLaRed (Rupix, 17-oct-2026): /api/estado dice la verdad del nodo y del libro,
// toma la punta mas reciente entre varios tips, cuenta los dias desde el relanzamiento
// real (no desde el genesis simbolico) y, si el nodo esta caido, lo dice en vez de inventar.
func TestEstadoDeLaRed(t *testing.T) {
	ahora := time.Date(2026, 10, 17, 20, 0, 0, 0, time.UTC)
	f := &fuenteFalsa{
		info: &appmessage.GetInfoResponseMessage{ServerVersion: "v0.6.6", IsSynced: true},
		dag:  &appmessage.GetBlockDAGInfoResponseMessage{VirtualDAAScore: 1_900_000, BlockCount: 1_950_000, TipHashes: []string{"a", "b", "rota"}},
		bloq: map[string]int64{"a": ahora.Add(-9 * time.Second).UnixMilli(), "b": ahora.Add(-3 * time.Second).UnixMilli()},
		gems: &appmessage.GetGemsInfoResponseMessage{DiamantesNacidos: 25, PlatinosNacidos: 5, RodiosNacidos: 0, KingsNacidos: 0},
	}
	l := nuevoLibro("")
	l.registrarPropio(&Nodo{Nombre: "Seed1-Nuremberg", Pais: "DE", Ubicado: true}, ahora)
	l.pasada([]Nodo{{Nombre: "JC", Pais: "MX", Ubicado: true}}, []string{"1.2.3.4"}, ahora, 0)

	e := armarEstado(f, l, ahora)
	if e.Error != "" {
		t.Fatalf("sin error esperado: %s", e.Error)
	}
	if e.Red != "testnet-4" || e.Version != "0.6.6" || !e.Sincronizado || e.DAA != 1_900_000 || e.Bloques != 1_950_000 {
		t.Fatalf("cabecera mal: %+v", e)
	}
	if e.HaceSegundos != 3 || e.UltimoBloque != ahora.Add(-3*time.Second).Unix() {
		t.Fatalf("la punta es el tip mas reciente (3 s), tengo %d s", e.HaceSegundos)
	}
	if e.DiasTestnet != 22 { // 25-sep -> 17-oct
		t.Fatalf("dias de testnet #4 desde el 25-sep: 22, tengo %d", e.DiasTestnet)
	}
	if e.NodosAhora != 2 || e.NodosHistoria != 2 || e.Paises != 2 {
		t.Fatalf("libro: 2 ahora, 2 historia, 2 paises; tengo %d/%d/%d", e.NodosAhora, e.NodosHistoria, e.Paises)
	}
	if e.Diamantes != 25 || e.Platinos != 5 || e.Rodios != 0 || e.Kings != 0 {
		t.Fatalf("gemas: %+v", e)
	}

	// Nodo caido: error dicho, libro y dias presentes, nada inventado.
	f.caido = true
	e = armarEstado(f, l, ahora)
	if e.Error == "" || e.DAA != 0 || e.Version != "" || e.NodosAhora != 2 || e.DiasTestnet != 22 {
		t.Fatalf("con el nodo caido: %+v", e)
	}

	// Cache: dos pedidos dentro de la vida consultan el nodo una sola vez.
	f.caido, f.ll = false, 0
	c := &cacheEstado{vida: 10 * time.Second}
	armar := func() *Estado { return armarEstado(f, l, ahora) }
	c.obtener(ahora, armar)
	c.obtener(ahora.Add(5*time.Second), armar)
	if f.ll != 1 {
		t.Fatalf("la cache debia ahorrar la segunda consulta; llamadas: %d", f.ll)
	}
	c.obtener(ahora.Add(11*time.Second), armar)
	if f.ll != 2 {
		t.Fatalf("pasada la vida se vuelve a consultar; llamadas: %d", f.ll)
	}
}
