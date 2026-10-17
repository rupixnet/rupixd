// Rupix (17-oct-2026): /api/estado, la salud de la red en un solo JSON para la portada
// de rupix.network. Todo sale del nodo y del libro; nada se inventa. Si el nodo no
// contesta, el error va en el JSON y la portada lo dice tal cual.
package main

import (
	"sync"
	"time"

	"github.com/rupixnet/rupixd/app/appmessage"
)

// testnet4Inicio: relanzamiento de la testnet #4 con v0.6.0 (CONTEXTO: DAA ~48.6k la
// manana del 26-sep a ~1 bloque/s => arranque la tarde del 25-sep-2026). El timestamp del
// genesis (4-mar-2026) es simbolico y no sirve para contar dias de red.
var testnet4Inicio = time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)

// Estado es lo que publica /api/estado.
type Estado struct {
	Red           string `json:"red"`           // "testnet-4"
	Version       string `json:"version"`       // del nodo que alimenta al explorador
	Sincronizado  bool   `json:"sincronizado"`  // IsSynced del nodo
	DAA           uint64 `json:"daa"`           // puntaje DAA virtual
	Bloques       uint64 `json:"bloques"`       // BlockCount
	UltimoBloque  int64  `json:"ultimoBloque"`  // unix (s) del encabezado de la punta
	HaceSegundos  int64  `json:"haceSegundos"`  // ahora - ultimoBloque
	InicioTestnet int64  `json:"inicioTestnet"` // unix (s) del relanzamiento
	DiasTestnet   int    `json:"diasTestnet"`   // dias completos desde el relanzamiento
	NodosAhora    int    `json:"nodosAhora"`    // libro: encendidos en la ultima pasada
	NodosHistoria int    `json:"nodosHistoria"` // libro: distintos en toda la historia
	Paises        int    `json:"paises"`        // libro: paises distintos
	Diamantes     uint64 `json:"diamantes"`     // nacidos (conteo sellado)
	Platinos      uint64 `json:"platinos"`
	Rodios        uint64 `json:"rodios"`
	Kings         uint64 `json:"kings"`
	Calculado     int64  `json:"calculado"`       // unix (s) en que se armo esta respuesta
	Error         string `json:"error,omitempty"` // si el nodo no contesto; los demas campos pueden venir a medias
}

// fuenteEstado es lo que /api/estado necesita del nodo; se abstrae para el test.
type fuenteEstado interface {
	GetInfo() (*appmessage.GetInfoResponseMessage, error)
	GetBlockDAGInfo() (*appmessage.GetBlockDAGInfoResponseMessage, error)
	GetBlock(hash string, includeTransactions bool) (*appmessage.GetBlockResponseMessage, error)
	GetGemsInfo() (*appmessage.GetGemsInfoResponseMessage, error)
}

// armarEstado consulta el nodo y el libro. `ahora` entra por parametro para el test.
func armarEstado(f fuenteEstado, libro *Libro, ahora time.Time) *Estado {
	e := &Estado{Red: "testnet-4", InicioTestnet: testnet4Inicio.Unix(), Calculado: ahora.Unix()}
	e.DiasTestnet = int(ahora.Sub(testnet4Inicio).Hours() / 24)
	if libro != nil {
		r := libro.respuesta()
		e.NodosAhora, e.NodosHistoria, e.Paises = r.Ahora, r.Total, r.Paises
	}
	info, err := f.GetInfo()
	if err != nil {
		e.Error = "el nodo no contesta: " + err.Error()
		return e
	}
	e.Version = quitarV(info.ServerVersion)
	e.Sincronizado = info.IsSynced
	dag, err := f.GetBlockDAGInfo()
	if err != nil {
		e.Error = "sin informacion del DAG: " + err.Error()
		return e
	}
	e.DAA, e.Bloques = dag.VirtualDAAScore, dag.BlockCount
	// La punta: el encabezado mas reciente entre los tips (puede haber varios en un DAG).
	var ultimo int64
	for _, h := range dag.TipHashes {
		b, err := f.GetBlock(h, false)
		if err != nil || b.Block == nil || b.Block.Header == nil {
			continue
		}
		if t := b.Block.Header.Timestamp / 1000; t > ultimo {
			ultimo = t
		}
	}
	if ultimo > 0 {
		e.UltimoBloque = ultimo
		e.HaceSegundos = ahora.Unix() - ultimo
		if e.HaceSegundos < 0 {
			e.HaceSegundos = 0
		}
	}
	if g, err := f.GetGemsInfo(); err == nil {
		e.Diamantes, e.Platinos, e.Rodios, e.Kings = g.DiamantesNacidos, g.PlatinosNacidos, g.RodiosNacidos, g.KingsNacidos
	}
	return e
}

func quitarV(s string) string {
	if len(s) > 0 && s[0] == 'v' {
		return s[1:]
	}
	return s
}

// cacheEstado: la portada de la web puede tener muchos visitantes; el nodo se consulta
// como mucho una vez cada `vida`.
type cacheEstado struct {
	mu     sync.Mutex
	vida   time.Duration
	valido time.Time
	ultimo *Estado
}

func (c *cacheEstado) obtener(ahora time.Time, armar func() *Estado) *Estado {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ultimo != nil && ahora.Before(c.valido) {
		return c.ultimo
	}
	c.ultimo = armar()
	c.valido = ahora.Add(c.vida)
	return c.ultimo
}
