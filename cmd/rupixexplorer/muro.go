// muro (Rupix, 15-nov-2026): el muro de nodos. Idea de JC (nodo #2): un chat donde solo hablan
// los nodos, sin moderador, que se borra solo.
//
// Reglas (las del nodo estan en infrastructure/config/mensaje.go):
//   - solo nodos: el mensaje llega en el saludo de la red (comentario "~texto"); nadie escribe
//     desde la web y no hay cuentas;
//   - habla quien ya sostuvo: solo nodos con >= 24 h de historia en el libro;
//   - uno por hora: un texto nuevo del mismo nodo se acepta si paso >= 1 h desde el anterior;
//   - cada mensaje vive 24 h desde que se vio por primera vez y desaparece solo; repetir el
//     mismo texto no lo revive (hay que decir algo distinto);
//   - sin moderador, sin archivo: vive solo en memoria del explorador, nunca en libro.json.
package main

import (
	"sort"
	"time"
)

const (
	muroVida          = 24 * time.Hour // cuanto vive un mensaje
	muroEntreMensajes = time.Hour      // minimo entre dos mensajes del mismo nodo
	muroHistoriaMin   = int64(86400)   // segundos de historia que un nodo necesita para hablar
)

// MensajeMuro es lo que publica /api/muro.
type MensajeMuro struct {
	Numero int    `json:"numero"`
	Nombre string `json:"nombre"` // "" si el nodo no tiene nombre
	Pais   string `json:"pais"`
	Texto  string `json:"texto"`
	Desde  int64  `json:"desde"` // unix (s) en que se vio por primera vez
	Hasta  int64  `json:"hasta"` // unix (s) en que desaparece
}

type entradaMuro struct {
	clave string
	MensajeMuro
}

// Muro vive dentro del Libro pero no se guarda: solo memoria.
type Muro struct {
	vivos  map[string]*entradaMuro // clave -> mensaje vigente
	ultimo map[string]entradaMuro  // clave -> ultimo mensaje aceptado (aunque haya caducado)
}

func nuevoMuro() *Muro {
	return &Muro{vivos: map[string]*entradaMuro{}, ultimo: map[string]entradaMuro{}}
}

// ofrecer recibe lo que un nodo anuncia en esta pasada. `historia` son sus segundos en el
// libro. Devuelve true si el mensaje entro al muro como nuevo.
func (m *Muro) ofrecer(clave string, e *EntradaLibro, texto string, ahora time.Time) bool {
	if texto == "" || e == nil || e.Segundos < muroHistoriaMin {
		return false
	}
	if u, ok := m.ultimo[clave]; ok {
		if u.Texto == texto {
			return false // repetir no revive ni renueva
		}
		if ahora.Unix()-u.Desde < int64(muroEntreMensajes.Seconds()) {
			return false // uno por hora
		}
	}
	n := entradaMuro{clave: clave, MensajeMuro: MensajeMuro{Numero: e.Numero, Nombre: e.Nombre, Pais: e.Pais, Texto: texto,
		Desde: ahora.Unix(), Hasta: ahora.Add(muroVida).Unix()}}
	m.vivos[clave] = &n
	m.ultimo[clave] = n
	return true
}

// barrer quita lo que ya vivio sus 24 h.
func (m *Muro) barrer(ahora time.Time) {
	for k, v := range m.vivos {
		if ahora.Unix() >= v.Hasta {
			delete(m.vivos, k)
		}
	}
}

// lista devuelve los mensajes vivos, el mas nuevo primero.
func (m *Muro) lista(ahora time.Time) []MensajeMuro {
	m.barrer(ahora)
	out := make([]MensajeMuro, 0, len(m.vivos))
	for _, v := range m.vivos {
		out = append(out, v.MensajeMuro)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Desde != out[j].Desde {
			return out[i].Desde > out[j].Desde
		}
		return out[i].Numero < out[j].Numero
	})
	return out
}
