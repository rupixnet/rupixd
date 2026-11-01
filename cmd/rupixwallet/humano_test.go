package main

import (
	"errors"
	"strings"
	"testing"
)

func TestMiles(t *testing.T) {
	casos := map[uint64]string{0: "0", 999: "999", 1000: "1,000", 86400: "86,400", 400000: "400,000", 41999994: "41,999,994"}
	for n, esperado := range casos {
		if got := miles(n); got != esperado {
			t.Errorf("miles(%d) = %q, esperado %q", n, got, esperado)
		}
	}
}

func TestTiempoBloques(t *testing.T) {
	casos := map[uint64]string{30: "30 segundos", 600: "10 minutos", 7200: "2.0 horas", 86400: "24.0 horas", 172800: "2.0 dias"}
	for n, esperado := range casos {
		if got := tiempoBloques(n); got != esperado {
			t.Errorf("tiempoBloques(%d) = %q, esperado %q", n, got, esperado)
		}
	}
}

func TestTraducirErrorNodo(t *testing.T) {
	// Un rechazo conocido se explica y conserva el texto original.
	err := traducirErrorNodo(errors.New("nivel 2 bloqueado: se desbloquea en DAA score 200000 (actual: 185738)"))
	if !strings.Contains(err.Error(), "todavia no se abre") || !strings.Contains(err.Error(), "actual: 185738") {
		t.Fatalf("traduccion incompleta: %v", err)
	}
	// Uno desconocido pasa tal cual.
	orig := errors.New("algo raro")
	if traducirErrorNodo(orig) != orig {
		t.Fatalf("un error desconocido debe pasar sin tocar")
	}
}

func TestTraducirErrorDaemon(t *testing.T) {
	err := traducirErrorDaemon(errors.New("rpc error: code = Unavailable desc = connection error: connection refused"), "localhost:8082", "--testnet")
	if !strings.Contains(err.Error(), "start-daemon") || !strings.Contains(err.Error(), "--testnet") {
		t.Fatalf("debe decir como arrancar el daemon con la misma red: %v", err)
	}
	otro := errors.New("otra cosa")
	if traducirErrorDaemon(otro, "x", "") != otro {
		t.Fatalf("un error distinto pasa sin tocar")
	}
}

// TestHumanizarError (Rupix, 1-nov-2026): los errores que ve una persona nueva salen en
// lenguaje llano, con el comando a copiar, y el error original debajo. Lo desconocido pasa
// tal cual.
func TestHumanizarError(t *testing.T) {
	idiomaActivo = "es"
	casos := []struct{ entra, contiene string }{
		{"rpc error: code = Unavailable desc = connection error: connection refused", "daemon"},
		{"wallet daemon is not synced yet, 3 of 10 addresses", "poniendose al dia"},
		{"Insufficient funds for send: 10.000000 required, while only 2.000000 available", "No alcanza el saldo"},
		{"decoded address is of unknown format: checksum failed", "no es valida para esta red"},
		{"cipher: message authentication failed", "contrasena"},
		{"open /root/.rupixwallet/rupix-testnet/keys.json: no such file or directory", "No hay wallet creada"},
		{"/root/.rupixwallet/rupix-testnet/keys.json is locked and cannot be used. Make sure that no other active wallet command is using it.", "Otro comando de la wallet"},
	}
	for _, c := range casos {
		out := humanizarError(errors.New(c.entra), "localhost:8082", "--testnet").Error()
		if !strings.Contains(out, c.contiene) {
			t.Errorf("%q: esperaba %q en:\n%s", c.entra, c.contiene, out)
		}
		if c.contiene != "daemon" && !strings.Contains(out, c.entra) {
			t.Errorf("%q: el error original debe quedar debajo", c.entra)
		}
		if strings.Contains(out, "%!") {
			t.Errorf("%q: formato roto: %s", c.entra, out)
		}
	}
	desconocido := errors.New("algo rarisimo")
	if humanizarError(desconocido, "x", "--testnet") != desconocido {
		t.Errorf("un error desconocido pasa tal cual")
	}
}
