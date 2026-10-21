package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rupixnet/rupixd/app/appmessage"
)

// TestMapaDeNodosNoFiltraIPs (Rupix, 14-oct-2026): el mapa publica nombre elegido,
// version, antiguedad y ciudad aproximada; nunca la IP. El nombre viene del uacomment y
// se limpia de HTML. La ubicacion sale de la base local; sin base, el nodo queda sin
// ubicar pero sigue contando.
func TestMapaDeNodosNoFiltraIPs(t *testing.T) {
	// Base minima con el formato de DB-IP lite de ciudad (dos rangos IPv4, una IPv6 que se salta).
	csv := "1.0.0.0,1.0.0.255,OC,AU,Queensland,\"Brisbane\",-27.4679,153.028\n" +
		"2001:db8::,2001:db8::ffff,EU,DE,Berlin,Berlin,52.52,13.405\n" +
		"187.0.0.0,187.255.255.255,NA,MX,\"Nuevo León\",\"Monterrey\",25.6866,-100.3161\n"
	ruta := filepath.Join(t.TempDir(), "geo.csv")
	if err := os.WriteFile(ruta, []byte(csv), 0o600); err != nil {
		t.Fatal(err)
	}
	g, err := cargarGeoBase(ruta)
	if err != nil {
		t.Fatalf("cargarGeoBase: %v", err)
	}
	if len(g.inicio) != 2 {
		t.Fatalf("debio cargar 2 rangos IPv4, cargo %d", len(g.inicio))
	}

	ahora := time.Unix(1_760_000_000, 0)
	peers := []*appmessage.GetConnectedPeerInfoMessage{
		{Address: "187.123.45.67:17211", UserAgent: "/kaspad:0.6.6/rupixd:0.6.6(Edu <b>x</b>)/", IsOutbound: false, TimeConnected: (ahora.Unix() - 3600) * 1000},
		{Address: "8.8.8.8:17211", UserAgent: "/kaspad:0.6.5/", IsOutbound: true, TimeConnected: ahora.Unix() * 1000},
		{Address: "[2001:db8::1]:17211", UserAgent: "basura", TimeConnected: 0},
	}
	nodos := nodosDesdePeers(peers, g, ahora)
	if len(nodos) != 3 {
		t.Fatalf("3 peers, %d nodos", len(nodos))
	}
	// El ubicado va primero.
	n := nodos[0]
	if !n.Ubicado || n.Pais != "MX" || n.Ciudad != "Monterrey" || n.Lat != 25.7 || n.Lon != -100.3 {
		t.Fatalf("ubicacion mal: %+v", n)
	}
	if n.Nombre != "Edu bx/b" || n.Version != "0.6.6" || n.Segundos != 3600 || n.Saliente {
		t.Fatalf("campos mal: %+v", n)
	}
	// Los otros dos no se ubican (fuera de la base / IPv6), pero existen.
	for _, m := range nodos[1:] {
		if m.Ubicado || m.Pais != "" {
			t.Fatalf("no debia ubicarse: %+v", m)
		}
	}
	// Ninguna IP en lo que se publica.
	for _, m := range nodos {
		for _, campo := range []string{m.Nombre, m.Version, m.Pais, m.Ciudad, m.ReportaDe} {
			if campo == "187.123.45.67" || campo == "8.8.8.8" || campo == "2001:db8::1" {
				t.Fatalf("se filtro una IP en %+v", m)
			}
		}
	}
	if contarPaises(nodos, nil) != 1 {
		t.Fatalf("paises: %d", contarPaises(nodos, nil))
	}

	// Sin base: nadie se ubica y no hay panico.
	sin := nodosDesdePeers(peers, nil, ahora)
	for _, m := range sin {
		if m.Ubicado {
			t.Fatalf("sin base no se ubica: %+v", m)
		}
	}

	// El nodo propio se declara a mano.
	p := nodoPropio("25.68,-100.31,MX,Monterrey,Seed 1")
	if p == nil || p.Nombre != "Seed 1" || p.Lat != 25.7 || !p.Ubicado {
		t.Fatalf("nodoPropio: %+v", p)
	}
	if nodoPropio("") != nil || nodoPropio("x,y,z") != nil {
		t.Fatalf("nodoPropio debe ser nil con entrada vacia o rota")
	}
	if contarPaises(nodos, p) != 1 {
		t.Fatalf("MX se cuenta una vez")
	}
}

func TestPartirUserAgent(t *testing.T) {
	casos := []struct{ ua, version, nombre string }{
		// Formato real (MsgVersion.AddUserAgent): nombre:version(comentarios).
		// El saludo real de un nodo 0.6.6 (DefaultUserAgent heredado + el propio), visto el 20-oct:
		{"/kaspad:0.6.6/rupixd:0.6.6(Coroping)/", "0.6.6", "Coroping"},
		{"/kaspad:0.6.6/rupixd:0.6.6/", "0.6.6", ""},
		{"/rupixd:0.6.6/", "0.6.6", ""},
		{"/rupixd:0.6.6(Coroping)/", "0.6.6", "Coroping"},
		{"/rupixd:0.6.6(JC; Monterrey)/", "0.6.6", "JC Monterrey"},
		{"/kaspad:0.12.22/", "0.12.22", ""},
		{"/rupixd:0.7.0(x)/otro:1/", "0.7.0", "x"},
		{"/rupixd:0.6.6(<b>Edu</b>)/", "0.6.6", "bEdu/b"}, // barras y etiquetas adentro no rompen nada
		{"/kaspad:0.6.6/rupixd:0.6.6(a/b; c)/", "0.6.6", "a/b c"},
		{"", "", ""},
		{"/otro:1/", "", ""},
	}
	for _, c := range casos {
		v, n := partirUserAgent(c.ua)
		if v != c.version || n != c.nombre {
			t.Errorf("%q: (%q,%q), esperaba (%q,%q)", c.ua, v, n, c.version, c.nombre)
		}
	}
}
