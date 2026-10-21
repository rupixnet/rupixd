package main

// nodos.go (Rupix, 14-oct-2026): el mapa de nodos. El explorador pregunta al seed quien
// esta conectado (GetConnectedPeerInfo) y publica, por nodo, SOLO lo que no identifica
// a nadie: el nombre que el propio nodo eligio (`rupixd --uacomment=Nombre`, viaja en el
// saludo P2P), su version, desde cuando esta conectado, y una ubicacion aproximada
// (pais y ciudad, con coordenadas redondeadas a un decimal, ~10 km) sacada de una base
// de datos local libre (DB-IP lite, formato CSV de ciudad). La IP nunca sale de este
// proceso ni se guarda en ningun lado. Si no hay base de datos, los nodos salen sin
// ubicacion y el mapa los lista aparte.

import (
	"bufio"
	"encoding/binary"
	"math"
	"net"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rupixnet/rupixd/app/appmessage"
)

// Nodo es lo que el mapa publica de cada peer. Sin IP, sin puerto, sin ID.
type Nodo struct {
	Nombre    string  `json:"nombre"`    // de --uacomment; "" si el nodo no eligio uno
	Version   string  `json:"version"`   // de /kaspad:X.Y.Z/
	Pais      string  `json:"pais"`      // codigo ISO-2, "" si no hay base
	Ciudad    string  `json:"ciudad"`    // "" si no hay base
	Lat       float64 `json:"lat"`       // redondeada a 1 decimal; 0 si no hay base
	Lon       float64 `json:"lon"`       // idem
	Ubicado   bool    `json:"ubicado"`   // true si Lat/Lon vienen de la base
	Saliente  bool    `json:"saliente"`  // true: el seed lo llamo; false: el nodo llamo al seed
	Segundos  int64   `json:"segundos"`  // conectado desde hace tantos segundos
	ReportaDe string  `json:"reportaDe"` // quien lo ve: "seed" (por ahora, siempre)
}

// RespuestaNodos es /api/peers.
type RespuestaNodos struct {
	Propio *Nodo  `json:"propio,omitempty"` // el nodo del explorador (el seed), si se configuro
	Nodos  []Nodo `json:"nodos"`
	Total  int    `json:"total"`
	Paises int    `json:"paises"`
	ConGeo bool   `json:"conGeo"` // true si hay base de datos de ubicacion cargada
	Nota   string `json:"nota"`
}

// partirUserAgent saca version y nombre del saludo real del nodo (BIP 14, como lo arma
// MsgVersion.AddUserAgent): "/rupixd:0.6.6/" o "/rupixd:0.6.6(Nombre)/"; varios --uacomment
// van separados por "; " dentro del parentesis. Corregido el 20-oct-2026: la version anterior
// esperaba "/rupixd:0.6.6/Nombre/", un formato que nunca existio, y el nombre se perdia.
// Devuelve ("", "") si no tiene la forma esperada. El nombre se recorta y se limpia
// de caracteres de control para que nadie pinte HTML en el mapa con su uacomment.
func partirUserAgent(ua string) (version, nombre string) {
	// No se corta por "/" porque un comentario puede traer barras; se buscan los tramos
	// "nombre:version(comentario)" directamente. El ultimo tramo con comentario manda.
	for _, m := range reUserAgent.FindAllStringSubmatch(ua, -1) {
		version = m[1]
		if m[2] != "" {
			nombre = limpiarNombre(strings.ReplaceAll(m[2], "; ", " "))
		}
	}
	return version, nombre
}

var reUserAgent = regexp.MustCompile(`(?:rupixd|kaspad):([0-9][^/()]*)(?:\(([^)]*)\))?`)

func limpiarNombre(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 32 || r == 127 || r == '<' || r == '>' || r == '&' || r == '"' {
			continue
		}
		b.WriteRune(r)
	}
	n := strings.TrimSpace(b.String())
	if len(n) > 32 {
		n = n[:32]
	}
	return n
}

// geoBase es la base IP→lugar cargada en memoria. Solo IPv4 (rangos ordenados).
type geoBase struct {
	inicio []uint32
	fin    []uint32
	lugar  []lugar
}

type lugar struct {
	pais, ciudad string
	lat, lon     float64
}

// cargarGeoBase lee un CSV con el formato de DB-IP lite de ciudad:
//
//	ip_start,ip_end,continente,pais,region,ciudad,lat,lon
//
// Las lineas IPv6 y las mal formadas se saltan. Devuelve nil si la ruta esta vacia.
func cargarGeoBase(ruta string) (*geoBase, error) {
	if ruta == "" {
		return nil, nil
	}
	f, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	g := &geoBase{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		campos := partirCSV(sc.Text())
		if len(campos) < 8 {
			continue
		}
		a, b := ipv4(campos[0]), ipv4(campos[1])
		if a == 0 && campos[0] != "0.0.0.0" || b == 0 && campos[1] != "0.0.0.0" {
			continue // IPv6 o basura
		}
		lat, err1 := strconv.ParseFloat(campos[6], 64)
		lon, err2 := strconv.ParseFloat(campos[7], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		g.inicio = append(g.inicio, a)
		g.fin = append(g.fin, b)
		g.lugar = append(g.lugar, lugar{pais: campos[3], ciudad: campos[5], lat: lat, lon: lon})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	// DB-IP viene ordenado, pero no se confia: se ordena por inicio.
	idx := make([]int, len(g.inicio))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool { return g.inicio[idx[i]] < g.inicio[idx[j]] })
	ordenado := &geoBase{}
	for _, i := range idx {
		ordenado.inicio = append(ordenado.inicio, g.inicio[i])
		ordenado.fin = append(ordenado.fin, g.fin[i])
		ordenado.lugar = append(ordenado.lugar, g.lugar[i])
	}
	return ordenado, nil
}

// partirCSV parte una linea CSV sencilla (campos entre comillas opcionales, sin comas
// escapadas dentro de comillas mas alla de lo que DB-IP usa).
func partirCSV(linea string) []string {
	var campos []string
	var b strings.Builder
	dentro := false
	for _, r := range linea {
		switch {
		case r == '"':
			dentro = !dentro
		case r == ',' && !dentro:
			campos = append(campos, b.String())
			b.Reset()
		default:
			b.WriteRune(r)
		}
	}
	campos = append(campos, b.String())
	return campos
}

func ipv4(s string) uint32 {
	ip := net.ParseIP(strings.TrimSpace(s))
	if ip == nil {
		return 0
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ip4)
}

// buscar devuelve el lugar de una IP (solo IPv4) o ok=false.
func (g *geoBase) buscar(ipTexto string) (lugar, bool) {
	if g == nil {
		return lugar{}, false
	}
	n := ipv4(ipTexto)
	if n == 0 {
		return lugar{}, false
	}
	i := sort.Search(len(g.inicio), func(i int) bool { return g.inicio[i] > n }) - 1
	if i < 0 || n > g.fin[i] {
		return lugar{}, false
	}
	return g.lugar[i], true
}

// redondear a un decimal (~10 km): ciudad aproximada, nunca una casa.
func redondear(x float64) float64 { return math.Round(x*10) / 10 }

// hostDe saca el host de "1.2.3.4:17211" o "[::1]:17211".
func hostDe(direccion string) string {
	h, _, err := net.SplitHostPort(direccion)
	if err != nil {
		return direccion
	}
	return h
}

// nodosDesdePeers convierte la respuesta del nodo en lo que el mapa publica.
func nodosDesdePeers(peers []*appmessage.GetConnectedPeerInfoMessage, g *geoBase, ahora time.Time) []Nodo {
	nodos := make([]Nodo, 0, len(peers))
	for _, p := range peers {
		version, nombre := partirUserAgent(p.UserAgent)
		n := Nodo{Nombre: nombre, Version: version, Saliente: p.IsOutbound, ReportaDe: "seed"}
		if p.TimeConnected > 0 {
			n.Segundos = ahora.Unix() - p.TimeConnected/1000
			if n.Segundos < 0 {
				n.Segundos = 0
			}
		}
		if l, ok := g.buscar(hostDe(p.Address)); ok {
			n.Pais, n.Ciudad, n.Lat, n.Lon, n.Ubicado = l.pais, l.ciudad, redondear(l.lat), redondear(l.lon), true
		}
		nodos = append(nodos, n)
	}
	// Orden estable para que el mapa no baile: primero los ubicados, luego por nombre.
	sort.SliceStable(nodos, func(i, j int) bool {
		if nodos[i].Ubicado != nodos[j].Ubicado {
			return nodos[i].Ubicado
		}
		return nodos[i].Nombre < nodos[j].Nombre
	})
	return nodos
}

func contarPaises(nodos []Nodo, propio *Nodo) int {
	vistos := map[string]bool{}
	for _, n := range nodos {
		if n.Pais != "" {
			vistos[n.Pais] = true
		}
	}
	if propio != nil && propio.Pais != "" {
		vistos[propio.Pais] = true
	}
	return len(vistos)
}

// nodoPropio parsea --nodo-propio="lat,lon,PAIS,Ciudad,Nombre" (el explorador no sabe
// donde esta su propio seed; se le dice). Devuelve nil si esta vacio o mal formado.
func nodoPropio(s string) *Nodo {
	if s == "" {
		return nil
	}
	c := strings.Split(s, ",")
	if len(c) < 5 {
		return nil
	}
	lat, err1 := strconv.ParseFloat(strings.TrimSpace(c[0]), 64)
	lon, err2 := strconv.ParseFloat(strings.TrimSpace(c[1]), 64)
	if err1 != nil || err2 != nil {
		return nil
	}
	return &Nodo{Nombre: limpiarNombre(strings.Join(c[4:], ",")), Pais: strings.TrimSpace(c[2]), Ciudad: strings.TrimSpace(c[3]),
		Lat: redondear(lat), Lon: redondear(lon), Ubicado: true, ReportaDe: "propio"}
}
