package main

import (
	"strings"
	"testing"
)

// TestTarjetaSaldo (Rupix, 3-nov-2026): la tarjeta de `balance` dice Gold, gemas y estado del
// nodo; sin color es texto plano; con color lleva el dorado de Rupix en el Gold; nunca
// inventa el estado del nodo si no se pudo consultar.
func TestTarjetaSaldo(t *testing.T) {
	base := tarjetaSaldo{Red: "testnet", Gold: 7510652741813, Gemas: [5]uint64{0, 1, 0, 0, 0}, NodoConsultado: true, Sincronizado: true, DAA: 3293235}

	plano := strings.Join(lineasTarjeta(base, false), "\n")
	if strings.Contains(plano, "\x1b[") {
		t.Fatalf("sin color no debe haber secuencias ANSI:\n%s", plano)
	}
	for _, quiero := range []string{"Rupix · testnet", "Gold      75,106.52741813 RUPIX", "Diamante  1", "Platino   0", "Kings     0", "DAA 3,293,235"} {
		if !strings.Contains(plano, quiero) {
			t.Fatalf("falta %q en:\n%s", quiero, plano)
		}
	}
	if strings.Contains(plano, "llegando") || strings.Contains(plano, "arriving") {
		t.Fatalf("sin pendiente no debe salir la linea de 'llegando':\n%s", plano)
	}

	color := strings.Join(lineasTarjeta(base, true), "\n")
	if !strings.Contains(color, "\x1b[38;2;212;175;55m") { // #d4af37, el dorado de rupix.network
		t.Fatalf("el Gold debe ir en dorado Rupix:\n%q", color)
	}
	if sinColor(color) != plano {
		t.Fatalf("quitar el color debe dar exactamente la version plana")
	}

	// Pendiente > 0: aparece la linea de lo que llega.
	conPendiente := base
	conPendiente.Pendiente = 150000000
	if s := strings.Join(lineasTarjeta(conPendiente, false), "\n"); !strings.Contains(s, "+1.50 RUPIX") {
		t.Fatalf("pendiente debe mostrarse como '+1.50 RUPIX':\n%s", s)
	}

	// Nodo atrasado: aviso, con el DAA que tenga.
	atras := base
	atras.Sincronizado = false
	if s := strings.Join(lineasTarjeta(atras, false), "\n"); !strings.Contains(s, T("tarjeta.nodo.atras")) {
		t.Fatalf("nodo atrasado debe avisar:\n%s", s)
	}
	// Nodo sin consultar: se dice, no se inventa "al dia".
	sin := base
	sin.NodoConsultado = false
	if s := strings.Join(lineasTarjeta(sin, false), "\n"); !strings.Contains(s, T("tarjeta.nodo.sin")) || strings.Contains(s, "DAA") {
		t.Fatalf("nodo sin consultar: ni 'al dia' ni DAA:\n%s", s)
	}

	if rupixBonito(7510652741813) != "75,106.52741813" || rupixBonito(0) != "0.00" || rupixBonito(312500000) != "3.125" {
		t.Fatalf("rupixBonito: %q %q %q", rupixBonito(7510652741813), rupixBonito(0), rupixBonito(312500000))
	}
	// pintar con hex invalido no rompe: devuelve el texto.
	if pintar("rojo", "x", true) != "x" || pintar(colorGold, "x", false) != "x" {
		t.Fatalf("pintar debe devolver el texto si no hay color valido")
	}
}
