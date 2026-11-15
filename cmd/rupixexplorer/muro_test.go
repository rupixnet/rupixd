package main

import (
	"testing"
	"time"
)

// TestMuro (Rupix, 15-nov-2026): solo nodos con >= 24 h de historia; uno por hora; vive 24 h;
// repetir el mismo texto no lo revive; el mas nuevo primero.
func TestMuro(t *testing.T) {
	t0 := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	m := nuevoMuro()
	nuevo := &EntradaLibro{Numero: 9, Nombre: "Recien", Pais: "AR", Segundos: 3600}
	veterano := &EntradaLibro{Numero: 4, Nombre: "", Pais: "FR", Segundos: 400_000}

	if m.ofrecer("n:Recien", nuevo, "hola", t0) {
		t.Fatalf("un nodo con 1 h de historia no habla todavia")
	}
	if !m.ofrecer("h:fr", veterano, "hola desde FR", t0) {
		t.Fatalf("un veterano con mensaje valido entra")
	}
	if m.ofrecer("h:fr", veterano, "hola desde FR", t0.Add(10*time.Minute)) {
		t.Fatalf("repetir el mismo texto no cuenta")
	}
	if m.ofrecer("h:fr", veterano, "otro texto", t0.Add(30*time.Minute)) {
		t.Fatalf("antes de 1 h no se acepta un texto nuevo")
	}
	if !m.ofrecer("h:fr", veterano, "otro texto", t0.Add(61*time.Minute)) {
		t.Fatalf("tras 1 h si")
	}
	l := m.lista(t0.Add(62 * time.Minute))
	if len(l) != 1 || l[0].Texto != "otro texto" || l[0].Numero != 4 || l[0].Pais != "FR" {
		t.Fatalf("lista: %+v", l)
	}
	// Vive 24 h desde que se vio; luego desaparece aunque el nodo siga anunciandolo.
	if l := m.lista(t0.Add(61*time.Minute + 24*time.Hour)); len(l) != 0 {
		t.Fatalf("a las 24 h debe desaparecer: %+v", l)
	}
	if m.ofrecer("h:fr", veterano, "otro texto", t0.Add(26*time.Hour)) {
		t.Fatalf("el mismo texto no revive tras caducar")
	}
	if !m.ofrecer("h:fr", veterano, "algo nuevo", t0.Add(26*time.Hour)) {
		t.Fatalf("un texto distinto si entra")
	}
	// Orden: el mas nuevo primero.
	otro := &EntradaLibro{Numero: 2, Nombre: "Coroking", Pais: "MX", Segundos: 100_000}
	m.ofrecer("n:Coroking", otro, "saludos", t0.Add(27*time.Hour))
	l = m.lista(t0.Add(27*time.Hour + time.Minute))
	if len(l) != 2 || l[0].Nombre != "Coroking" || l[1].Numero != 4 {
		t.Fatalf("orden: %+v", l)
	}
	if m.ofrecer("h:fr", veterano, "", t0.Add(30*time.Hour)) {
		t.Fatalf("vacio no es mensaje")
	}
}
