package main

import (
	"testing"
	"time"

	"github.com/rupixnet/rupixd/app/appmessage"
)

// TestHistorial (Rupix, 28-oct-2026): fechas estimadas desde el DAA, minado agrupado por
// dia, gemas y recibidos uno por uno, filtros por fecha y texto, y errores en fechas mal
// escritas (nunca un historial vacio que parezca verdadero).
func TestHistorial(t *testing.T) {
	ahora := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	const virtual = 2_600_000
	utxo := func(daa uint64, coinbase bool, version uint16, monto uint64, dir, tx string) *appmessage.UTXOsByAddressesEntry {
		return &appmessage.UTXOsByAddressesEntry{Address: dir, Outpoint: &appmessage.RPCOutpoint{TransactionID: tx},
			UTXOEntry: &appmessage.RPCUTXOEntry{Amount: monto, BlockDAAScore: daa, IsCoinbase: coinbase, ScriptPublicKey: &appmessage.RPCScriptPublicKey{Version: version}}}
	}
	// Un bloque por segundo: 86,400 DAA = un dia.
	entradas := []*appmessage.UTXOsByAddressesEntry{
		utxo(virtual-10, true, 0, 3125000, "rupixtest:mina", "tx1"),         // minado hoy
		utxo(virtual-20, true, 0, 3125000, "rupixtest:mina", "tx2"),         // minado hoy
		utxo(virtual-86400*2-5, true, 0, 3125000, "rupixtest:mina", "tx3"),  // minado hace 2 dias
		utxo(virtual-3600, false, 0, 5000000000, "rupixtest:yo", "txRecib"), // recibido hace 1 h
		utxo(virtual-86400, false, 1, 100000000, "rupixtest:yo", "txGema"),  // Diamante hace 1 dia
		{Address: "rupixtest:roto"},                                         // entrada sin UTXOEntry: se ignora
	}
	ms := movimientosDesdeUTXOs(entradas, virtual, ahora)
	if len(ms) != 5 {
		t.Fatalf("5 movimientos, hay %d", len(ms))
	}
	if ms[0].TxID != "tx1" || ms[4].TxID != "tx3" {
		t.Fatalf("orden del mas reciente al mas viejo: %+v", ms)
	}
	if got := ms[0].Cuando; !got.Equal(ahora.Add(-10 * time.Second)) {
		t.Fatalf("fecha estimada: %s", got)
	}
	// Un DAA por encima del virtual no viaja al futuro.
	if f := fechaDesdeDAA(virtual+5, virtual, ahora); !f.Equal(ahora) {
		t.Fatalf("DAA futuro debe dar ahora: %s", f)
	}

	dias, resto := agruparMinado(ms)
	if len(dias) != 2 || dias[0].Dia != "2026-10-28" || dias[0].Bloques != 2 || dias[0].Monto != 6250000 || dias[1].Dia != "2026-10-26" {
		t.Fatalf("minado por dia: %+v", dias)
	}
	if len(resto) != 2 || resto[0].TxID != "txRecib" || resto[1].Nivel != 1 {
		t.Fatalf("resto: recibido y gema: %+v", resto)
	}

	// Filtros.
	f, err := filtrar(ms, "2026-10-27", "", "")
	if err != nil || len(f) != 4 {
		t.Fatalf("--desde 27-oct deja 4 (fuera el minado del 26): %d %v", len(f), err)
	}
	f, err = filtrar(ms, "", "2026-10-26", "")
	if err != nil || len(f) != 1 || f[0].TxID != "tx3" {
		t.Fatalf("--hasta 26-oct deja solo el minado del 26: %+v %v", f, err)
	}
	f, err = filtrar(ms, "", "", "GEMA")
	if err != nil || len(f) != 1 || f[0].TxID != "txGema" {
		t.Fatalf("--buscar sin distinguir mayusculas: %+v %v", f, err)
	}
	if _, err := filtrar(ms, "28/10/2026", "", ""); err == nil {
		t.Fatalf("una fecha mal escrita debe dar error, no un historial vacio")
	}
	if acortar("rupixtest:qq8w9v4k2s9c3m7h6x5j4n2p") != "rupixtest:…5j4n2p" {
		t.Fatalf("acortar: %q", acortar("rupixtest:qq8w9v4k2s9c3m7h6x5j4n2p"))
	}
}
