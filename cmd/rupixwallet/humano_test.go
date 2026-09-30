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
