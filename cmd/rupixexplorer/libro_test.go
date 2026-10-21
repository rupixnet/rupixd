package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLibroDeNodos (Rupix, 17-oct-2026): numeros de llegada para siempre, horas y rachas
// medidas, hitos verificables, persistencia atomica, y ni una IP en lo que se publica.
func TestLibroDeNodos(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "libro.json")
	l, err := cargarLibro(ruta)
	if err != nil {
		t.Fatal(err)
	}
	t0 := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	propio := &Nodo{Nombre: "Seed1", Pais: "DE", Ciudad: "Nuremberg", Lat: 49.45, Lon: 11.08, Ubicado: true, Version: "0.6.6"}
	l.registrarPropio(propio, t0)
	// Si la bandera cambia el nombre, el #1 cambia de nombre sin perder el numero.
	propio.Nombre = "Seed1-Nuremberg"
	l.registrarPropio(propio, t0)

	jc := Nodo{Nombre: "JC", Pais: "MX", Ciudad: "Monterrey", Lat: 25.7, Lon: -100.3, Ubicado: true, Version: "0.6.6"}
	anon := Nodo{Version: "0.6.5"}
	// Pasada 1: JC y un anonimo llegan.
	l.pasada([]Nodo{jc, anon}, []string{"187.1.2.3", "8.8.8.8"}, t0, 0)
	// Pasada 2, 30 s despues: siguen los dos; el anonimo desde otra conexion del mismo host.
	l.pasada([]Nodo{jc, anon, anon}, []string{"187.1.2.3", "8.8.8.8", "8.8.8.8"}, t0.Add(30*time.Second), 30*time.Second)
	// Pasada 3, al dia siguiente: solo JC. El anonimo se apago.
	t1 := t0.Add(24 * time.Hour)
	l.pasada([]Nodo{jc}, []string{"187.9.9.9"}, t1, 30*time.Second) // JC cambio de IP: el nombre lo identifica
	// Pasada 4, tres dias despues: JC vuelve; racha se rompe y reinicia.
	t4 := t0.Add(4 * 24 * time.Hour)
	l.pasada([]Nodo{jc}, []string{"187.9.9.9"}, t4, 30*time.Second)

	r := l.respuesta()
	if r.Total != 3 {
		t.Fatalf("3 nodos en el libro, hay %d", r.Total)
	}
	if r.Nodos[0].Numero != 1 || r.Nodos[0].Nombre != "Seed1-Nuremberg" {
		t.Fatalf("el #1 es el seed: %+v", r.Nodos[0])
	}
	e := r.Nodos[1]
	if e.Numero != 2 || e.Nombre != "JC" || e.Pais != "MX" {
		t.Fatalf("el #2 es JC: %+v", e)
	}
	if e.Segundos != 90 {
		t.Fatalf("JC acumulo 90 s (3 pasadas con transcurrido), tiene %d", e.Segundos)
	}
	if e.Racha != 1 || e.MejorRacha != 2 {
		t.Fatalf("racha de JC: hoy 1 tras romperse, mejor 2; tiene %d/%d", e.Racha, e.MejorRacha)
	}
	if !e.Conectado {
		t.Fatalf("JC esta conectado en la ultima pasada")
	}
	a := r.Nodos[2]
	if a.Numero != 3 || a.Nombre != "" || a.Conectado || a.Segundos != 30 {
		t.Fatalf("el anonimo es #3, desconectado, con 30 s: %+v", a)
	}
	if r.Ahora != 2 { // seed + JC
		t.Fatalf("ahora conectados: 2, hay %d", r.Ahora)
	}
	if r.Paises != 2 {
		t.Fatalf("paises: 2 (DE, MX), hay %d", r.Paises)
	}
	// Hitos: #2 y primer nodo en MX, #3; el mas reciente primero.
	textos := []string{}
	for _, h := range r.Hitos {
		textos = append(textos, h.Texto)
	}
	todo := strings.Join(textos, "\n")
	for _, esperado := range []string{"Nodo #2 · JC · Monterrey, MX", "Primer nodo en MX: #2", "Nodo #3 · nodo · sin ubicar"} {
		if !strings.Contains(todo, esperado) {
			t.Fatalf("falta el hito %q en:\n%s", esperado, todo)
		}
	}
	// Dias: el primero con pico 3 (seed+JC+anon), distintos 3.
	if len(r.Dias) != 3 || r.Dias[0].Pico != 3 || r.Dias[0].Distintos != 3 {
		t.Fatalf("dias: %+v", r.Dias)
	}
	// Veteranos por numero; constantes por segundos (seed 90, JC 90, anon 30).
	if r.Veteranos[0].Numero != 1 || r.Veteranos[1].Numero != 2 {
		t.Fatalf("veteranos: %+v", r.Veteranos)
	}
	if r.Constante[2].Numero != 3 {
		t.Fatalf("el anonimo es el menos constante: %+v", r.Constante)
	}

	// Lo publicado no trae claves ni IPs.
	pub, _ := json.Marshal(r)
	for _, prohibido := range []string{"187.1.2.3", "8.8.8.8", "187.9.9.9", `"h:`, `"n:`, `"huellas"`, l.Sal} {
		if strings.Contains(string(pub), prohibido) {
			t.Fatalf("se publico %q", prohibido)
		}
	}

	// Persistencia: guardar, recargar, mismo libro, mismo siguiente numero.
	if err := l.guardar(); err != nil {
		t.Fatal(err)
	}
	l2, err := cargarLibro(ruta)
	if err != nil {
		t.Fatal(err)
	}
	if l2.Siguiente != 4 || len(l2.Nodos) != 3 || l2.Sal != l.Sal {
		t.Fatalf("recarga: siguiente %d, nodos %d", l2.Siguiente, len(l2.Nodos))
	}
	// Un nodo nuevo tras recargar recibe el #4, no repite.
	l2.pasada([]Nodo{{Nombre: "JP", Pais: "MX", Ubicado: true}}, []string{"1.1.1.1"}, t4.Add(time.Minute), 30*time.Second)
	if l2.respuesta().Nodos[3].Numero != 4 || l2.respuesta().Nodos[3].Nombre != "JP" {
		t.Fatalf("JP debe ser el #4: %+v", l2.respuesta().Nodos[3])
	}
	// El archivo en disco tampoco trae IPs (si trae las huellas, que son privadas e irreversibles).
	if err := l2.guardar(); err != nil {
		t.Fatal(err)
	}
	enDisco, _ := json.Marshal(l2)
	for _, ip := range []string{"187.1.2.3", "8.8.8.8", "1.1.1.1"} {
		if strings.Contains(string(enDisco), ip) {
			t.Fatalf("el libro en disco guardo la IP %s", ip)
		}
	}
}

// TestLibroPicoHito: el hito "Primer dia con 5 encendidos" sale una sola vez.
func TestLibroPicoHito(t *testing.T) {
	l := nuevoLibro("")
	t0 := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	var peers []Nodo
	var hosts []string
	for i := 0; i < 5; i++ {
		peers = append(peers, Nodo{Nombre: string(rune('A' + i))})
		hosts = append(hosts, "10.0.0.1")
	}
	l.pasada(peers, hosts, t0, 0)
	l.pasada(peers, hosts, t0.Add(time.Minute), time.Minute)
	l.pasada(peers, hosts, t0.Add(48*time.Hour), time.Minute)
	n := 0
	for _, h := range l.respuesta().Hitos {
		if strings.Contains(h.Texto, "Primer dia con 5 encendidos") {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("el hito de 5 encendidos debe salir una vez, salio %d", n)
	}
}

// TestLibroCambioDeNombre (20-oct-2026): JC arranco sin que el explorador leyera su nombre
// (quedo anonimo, #2), luego con --uacomment=Coroping y luego Corokingping, siempre desde la
// misma maquina. Debe ser UN nodo, el #2, con el nombre mas reciente; otra maquina con nombre
// si es un nodo nuevo. Y el archivo del libro recuerda las huellas tras recargar.
func TestLibroCambioDeNombre(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "libro.json")
	l, err := cargarLibro(ruta)
	if err != nil {
		t.Fatal(err)
	}
	t0 := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	l.registrarPropio(&Nodo{Nombre: "Seed1-Nuremberg", Pais: "DE", Ubicado: true}, t0)
	// 1) anonimo desde la casa de JC
	l.pasada([]Nodo{{Version: "0.6.6"}}, []string{"201.1.1.1"}, t0, 0)
	// 2) se va, vuelve con nombre
	l.pasada(nil, nil, t0.Add(time.Minute), time.Minute)
	l.pasada([]Nodo{{Nombre: "Coroping", Version: "0.6.6"}}, []string{"201.1.1.1"}, t0.Add(2*time.Minute), time.Minute)
	// 3) se va, vuelve con otro nombre
	l.pasada([]Nodo{{Nombre: "Corokingping", Version: "0.6.6"}}, []string{"201.1.1.1"}, t0.Add(3*time.Minute), time.Minute)
	// 4) otra maquina, otro nombre: nodo nuevo
	l.pasada([]Nodo{{Nombre: "Corokingping", Version: "0.6.6"}, {Nombre: "JP", Version: "0.6.6"}}, []string{"201.1.1.1", "201.2.2.2"}, t0.Add(4*time.Minute), time.Minute)

	r := l.respuesta()
	if r.Total != 3 {
		t.Fatalf("seed, JC y JP: 3 nodos, hay %d: %+v", r.Total, r.Nodos)
	}
	jc := r.Nodos[1]
	if jc.Numero != 2 || jc.Nombre != "Corokingping" {
		t.Fatalf("JC debe seguir siendo el #2 con su nombre nuevo: %+v", jc)
	}
	if jc.Segundos != 180 {
		t.Fatalf("JC acumulo 3 min (pasadas 2, 3 y 4): %d s", jc.Segundos)
	}
	if r.Nodos[2].Numero != 3 || r.Nodos[2].Nombre != "JP" {
		t.Fatalf("JP es el #3: %+v", r.Nodos[2])
	}
	todo := ""
	for _, h := range r.Hitos {
		todo += h.Texto + "\n"
	}
	for _, esperado := range []string{"Nodo #2 ahora se llama Coroping", "Nodo #2 cambió de nombre: Coroping → Corokingping", "Nodo #3 · JP"} {
		if !strings.Contains(todo, esperado) {
			t.Fatalf("falta el hito %q en:\n%s", esperado, todo)
		}
	}
	// Dos maquinas distintas con el mismo nombre al mismo tiempo: la segunda NO roba el
	// numero (el primero en llegar con un nombre se lo queda) y no se abre uno nuevo por
	// la huella, porque la clave con nombre ya existe. Limitacion documentada.
	// Persistencia de las huellas.
	if err := l.guardar(); err != nil {
		t.Fatal(err)
	}
	l2, err := cargarLibro(ruta)
	if err != nil {
		t.Fatal(err)
	}
	l2.pasada([]Nodo{{Nombre: "JC-final", Version: "0.6.6"}}, []string{"201.1.1.1"}, t0.Add(time.Hour), time.Minute)
	r2 := l2.respuesta()
	if r2.Total != 3 || r2.Nodos[1].Numero != 2 || r2.Nodos[1].Nombre != "JC-final" {
		t.Fatalf("tras recargar, el cambio de nombre debe seguir reconociendo la maquina: %+v", r2.Nodos)
	}
}
