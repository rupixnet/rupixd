package main

// libro.go (Rupix, 17-oct-2026): el libro de nodos. Lo que el mapa muestra en vivo se olvida
// al desconectarse; el libro recuerda. A cada nodo que el seed ve por primera vez se le da el
// siguiente numero, para siempre: el #1 es el seed, el #2 el primero que llego, y asi. De cada
// uno se guarda: numero, nombre, pais/ciudad aproximados, primera y ultima vez, segundos
// acumulados vistos por el seed, racha de dias. De cada dia: nodos distintos acumulados y
// pico de conectados a la vez. Y una historia de hitos verificables (nodo #N, primer nodo en
// un pais, primer dia con N encendidos).
//
// Identidad, con honestidad: hoy un nodo se reconoce por el NOMBRE que anuncia (--uacomment);
// el primero que llega con un nombre se queda con ese numero. Si no anuncia nombre, se usa una
// huella privada de su conexion (sha256 de la IP con una sal aleatoria del libro) que NUNCA se
// publica ni permite recuperar la IP. Cuando el nodo guarde una identidad fija (v0.7), esa
// identidad heredara la entrada. Es un libro que lleva ESTE explorador: historia, no consenso;
// cualquier otro explorador puede llevar el suyo.
//
// Memoria: ~200 bytes por nodo y dos numeros por dia. Se guarda en un JSON (escritura atomica).

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

// EntradaLibro es un nodo en el libro. Lo que se publica es exactamente esto (sin la clave).
type EntradaLibro struct {
	Numero     int     `json:"numero"`
	Nombre     string  `json:"nombre"`
	Pais       string  `json:"pais"`
	Ciudad     string  `json:"ciudad"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Ubicado    bool    `json:"ubicado"`
	Version    string  `json:"version"`
	PrimeraVez int64   `json:"primeraVez"` // unix
	UltimaVez  int64   `json:"ultimaVez"`  // unix
	Segundos   int64   `json:"segundos"`   // acumulados, vistos por el seed
	Racha      int     `json:"racha"`      // dias seguidos vistos
	MejorRacha int     `json:"mejorRacha"`
	UltimoDia  string  `json:"ultimoDia"` // YYYY-MM-DD
	Conectado  bool    `json:"conectado"` // en la ultima pasada
}

// DiaLibro es el resumen de un dia.
type DiaLibro struct {
	Fecha     string `json:"fecha"`
	Distintos int    `json:"distintos"` // acumulados hasta ese dia
	Pico      int    `json:"pico"`      // conectados a la vez, maximo del dia
}

// Hito es una linea de la historia de la red.
type Hito struct {
	Fecha string `json:"fecha"`
	Texto string `json:"texto"`
	Unix  int64  `json:"unix"`
}

// Libro es el estado completo. Las claves (identidad) se guardan en disco pero no se publican.
type Libro struct {
	mu        sync.Mutex
	ruta      string
	Sal       string                   `json:"sal"`
	Siguiente int                      `json:"siguiente"`
	Nodos     map[string]*EntradaLibro `json:"nodos"` // clave -> entrada
	Dias      map[string]*DiaLibro     `json:"dias"`  // fecha -> resumen
	Hitos     []Hito                   `json:"hitos"`
	Paises    map[string]bool          `json:"paises"` // paises ya vistos (para "primer nodo en X")
	cambiado  bool
}

// RespuestaLibro es /api/libro: todo lo publicable.
type RespuestaLibro struct {
	Total     int             `json:"total"`      // nodos distintos en toda la historia
	Ahora     int             `json:"ahora"`      // conectados en la ultima pasada
	Paises    int             `json:"paises"`     // paises distintos en toda la historia
	Nodos     []*EntradaLibro `json:"nodos"`      // por numero
	Dias      []*DiaLibro     `json:"dias"`       // por fecha
	Hitos     []Hito          `json:"hitos"`      // mas reciente primero
	Veteranos []*EntradaLibro `json:"veteranos"`  // primeros 10 por numero
	Constante []*EntradaLibro `json:"constantes"` // 10 con mas segundos
	Racha     []*EntradaLibro `json:"rachas"`     // 10 con mejor racha
	Nota      string          `json:"nota"`
}

func nuevoLibro(ruta string) *Libro {
	l := &Libro{ruta: ruta, Siguiente: 1, Nodos: map[string]*EntradaLibro{}, Dias: map[string]*DiaLibro{}, Paises: map[string]bool{}}
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	l.Sal = hex.EncodeToString(b)
	return l
}

// cargarLibro lee el JSON si existe; si no, empieza uno nuevo.
func cargarLibro(ruta string) (*Libro, error) {
	l := nuevoLibro(ruta)
	if ruta == "" {
		return l, nil
	}
	datos, err := os.ReadFile(ruta)
	if os.IsNotExist(err) {
		return l, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(datos, l); err != nil {
		return nil, fmt.Errorf("libro %s ilegible: %w", ruta, err)
	}
	if l.Nodos == nil {
		l.Nodos = map[string]*EntradaLibro{}
	}
	if l.Dias == nil {
		l.Dias = map[string]*DiaLibro{}
	}
	if l.Paises == nil {
		l.Paises = map[string]bool{}
	}
	if l.Siguiente < 1 {
		l.Siguiente = 1
	}
	l.ruta = ruta
	return l, nil
}

// guardar escribe el JSON de forma atomica (tmp + rename). Sin ruta, no persiste.
func (l *Libro) guardar() error {
	if l.ruta == "" || !l.cambiado {
		return nil
	}
	datos, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		return err
	}
	tmp := l.ruta + ".tmp"
	if err := os.WriteFile(tmp, datos, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, l.ruta); err != nil {
		return err
	}
	l.cambiado = false
	return nil
}

// claveDe decide la identidad de un peer: nombre anunciado, o huella privada de su IP.
func (l *Libro) claveDe(nombre, host string) string {
	if nombre != "" {
		return "n:" + nombre
	}
	h := sha256.Sum256([]byte(l.Sal + "|" + host))
	return "h:" + hex.EncodeToString(h[:8])
}

func fechaDe(t time.Time) string { return t.UTC().Format("2006-01-02") }

// registrarPropio asegura que el seed propio sea el #1 (clave fija "propio").
func (l *Libro) registrarPropio(p *Nodo, ahora time.Time) {
	if p == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.Nodos["propio"]
	if e == nil {
		e = l.alta("propio", p.Nombre, ahora)
	}
	e.Pais, e.Ciudad, e.Lat, e.Lon, e.Ubicado = p.Pais, p.Ciudad, p.Lat, p.Lon, p.Ubicado
	if p.Version != "" {
		e.Version = p.Version
	}
	l.verPais(e, ahora)
	l.cambiado = true
}

// alta crea la entrada con el siguiente numero y anota el hito. Se llama con el lock tomado.
func (l *Libro) alta(clave, nombre string, ahora time.Time) *EntradaLibro {
	e := &EntradaLibro{Numero: l.Siguiente, Nombre: nombre, PrimeraVez: ahora.Unix(), UltimaVez: ahora.Unix()}
	l.Siguiente++
	l.Nodos[clave] = e
	l.cambiado = true
	return e
}

func (l *Libro) hito(ahora time.Time, formato string, args ...interface{}) {
	l.Hitos = append(l.Hitos, Hito{Fecha: fechaDe(ahora), Unix: ahora.Unix(), Texto: fmt.Sprintf(formato, args...)})
	l.cambiado = true
}

// verPais anota el "primer nodo en X" si el pais es nuevo. Con el lock tomado.
func (l *Libro) verPais(e *EntradaLibro, ahora time.Time) {
	if e.Pais == "" || l.Paises[e.Pais] {
		return
	}
	l.Paises[e.Pais] = true
	if len(l.Paises) > 1 { // el primero del libro no es noticia aparte del "#1"
		l.hito(ahora, "Primer nodo en %s: #%d · %s · %s", e.Pais, e.Numero, nombreOnodo(e.Nombre), lugarDe(e))
	}
}

func nombreOnodo(n string) string {
	if n == "" {
		return "nodo"
	}
	return n
}

func lugarDe(e *EntradaLibro) string {
	switch {
	case e.Ciudad != "" && e.Pais != "":
		return e.Ciudad + ", " + e.Pais
	case e.Pais != "":
		return e.Pais
	}
	return "sin ubicar"
}

// pasada registra lo que el seed ve conectado ahora. `peers` son los nodos ya convertidos (sin
// IP) y `hosts` sus hosts en el mismo orden (solo para la huella privada; no se guardan).
// `transcurrido` es cuanto tiempo paso desde la pasada anterior (lo que se suma a cada conectado).
func (l *Libro) pasada(peers []Nodo, hosts []string, ahora time.Time, transcurrido time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	hoy := fechaDe(ahora)
	ayer := fechaDe(ahora.Add(-24 * time.Hour))
	for _, e := range l.Nodos {
		e.Conectado = false
	}
	if p := l.Nodos["propio"]; p != nil {
		p.Conectado = true
	}
	vistos := map[string]bool{}
	for i, n := range peers {
		host := ""
		if i < len(hosts) {
			host = hosts[i]
		}
		clave := l.claveDe(n.Nombre, host)
		if vistos[clave] {
			continue // dos conexiones del mismo nodo cuentan una vez
		}
		vistos[clave] = true
		e := l.Nodos[clave]
		nuevo := e == nil
		if nuevo {
			e = l.alta(clave, n.Nombre, ahora)
		}
		e.Conectado = true
		e.UltimaVez = ahora.Unix()
		if n.Version != "" {
			e.Version = n.Version
		}
		if n.Ubicado {
			e.Pais, e.Ciudad, e.Lat, e.Lon, e.Ubicado = n.Pais, n.Ciudad, n.Lat, n.Lon, true
		}
		if !nuevo && transcurrido > 0 {
			e.Segundos += int64(transcurrido.Seconds())
		}
		switch e.UltimoDia {
		case hoy:
		case ayer:
			e.Racha++
			e.UltimoDia = hoy
		default:
			e.Racha = 1
			e.UltimoDia = hoy
		}
		if e.Racha > e.MejorRacha {
			e.MejorRacha = e.Racha
		}
		if nuevo {
			l.hito(ahora, "Nodo #%d · %s · %s", e.Numero, nombreOnodo(e.Nombre), lugarDe(e))
			switch e.Numero {
			case 10, 25, 50, 100, 250, 500, 1000:
				l.hito(ahora, "%d nodos han sostenido Rupix", e.Numero)
			}
		}
		l.verPais(e, ahora)
		l.cambiado = true
	}
	// el propio tambien acumula
	if p := l.Nodos["propio"]; p != nil {
		p.UltimaVez = ahora.Unix()
		if transcurrido > 0 {
			p.Segundos += int64(transcurrido.Seconds())
		}
		if p.UltimoDia != hoy {
			if p.UltimoDia == ayer {
				p.Racha++
			} else {
				p.Racha = 1
			}
			p.UltimoDia = hoy
			if p.Racha > p.MejorRacha {
				p.MejorRacha = p.Racha
			}
		}
	}
	// el dia
	conectados := 0
	for _, e := range l.Nodos {
		if e.Conectado {
			conectados++
		}
	}
	d := l.Dias[hoy]
	if d == nil {
		d = &DiaLibro{Fecha: hoy}
		l.Dias[hoy] = d
	}
	d.Distintos = len(l.Nodos)
	if conectados > d.Pico {
		for _, umbral := range []int{5, 10, 20, 50, 100, 500, 1000} {
			if d.Pico < umbral && conectados >= umbral && !l.yaHuboPico(umbral) {
				l.hito(ahora, "Primer dia con %d encendidos a la vez", umbral)
			}
		}
		d.Pico = conectados
	}
	l.cambiado = true
}

func (l *Libro) yaHuboPico(umbral int) bool {
	for _, d := range l.Dias {
		if d.Pico >= umbral {
			return true
		}
	}
	return false
}

// respuesta arma lo publicable. Nunca incluye las claves.
func (l *Libro) respuesta() *RespuestaLibro {
	l.mu.Lock()
	defer l.mu.Unlock()
	r := &RespuestaLibro{Nota: "Libro que lleva este explorador: a cada nodo que el seed ve por primera vez se le da el siguiente numero, para siempre. Horas y rachas medidas mientras estuvo conectado al seed. Identidad: el nombre anunciado o una huella privada de la conexion (nunca la IP). Historia, no consenso."}
	for _, e := range l.Nodos {
		r.Nodos = append(r.Nodos, e)
		if e.Conectado {
			r.Ahora++
		}
	}
	sort.Slice(r.Nodos, func(i, j int) bool { return r.Nodos[i].Numero < r.Nodos[j].Numero })
	r.Total = len(r.Nodos)
	r.Paises = len(l.Paises)
	for _, d := range l.Dias {
		r.Dias = append(r.Dias, d)
	}
	sort.Slice(r.Dias, func(i, j int) bool { return r.Dias[i].Fecha < r.Dias[j].Fecha })
	r.Hitos = append(r.Hitos, l.Hitos...)
	sort.SliceStable(r.Hitos, func(i, j int) bool { return r.Hitos[i].Unix > r.Hitos[j].Unix })
	top := func(menor func(a, b *EntradaLibro) bool) []*EntradaLibro {
		c := append([]*EntradaLibro(nil), r.Nodos...)
		sort.SliceStable(c, func(i, j int) bool { return menor(c[i], c[j]) })
		if len(c) > 10 {
			c = c[:10]
		}
		return c
	}
	r.Veteranos = top(func(a, b *EntradaLibro) bool { return a.Numero < b.Numero })
	r.Constante = top(func(a, b *EntradaLibro) bool { return a.Segundos > b.Segundos })
	r.Racha = top(func(a, b *EntradaLibro) bool { return a.MejorRacha > b.MejorRacha })
	return r
}
