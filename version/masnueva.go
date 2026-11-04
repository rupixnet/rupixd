// masnueva (Rupix, 4-nov-2026): aviso de version nueva SIN red nueva.
//
// Cada nodo ya recibe la version de sus pares en el saludo ("/rupixd:0.7.0/..."). Si la
// MAYORIA de los pares anuncia una version mayor que la propia, el nodo lo dice en su log
// (una vez al dia) y `rupixwallet verificar` lo marca como AVISO. Nada se envia, nada se
// consulta a ningun servidor, no se expone IP, y el aviso no puede decir otra cosa que un
// numero de version: un unico par malicioso no basta (hace falta la mayoria), y aun asi lo
// peor que consigue es que leas rupix.network. No es un canal de mensajes: Rupix no tiene
// "mensajes del fundador" y no los va a tener.
package version

import (
	"regexp"
	"strconv"
	"strings"
)

// reVersionUA extrae "X.Y.Z" de un user agent de la red: /rupixd:0.7.0/ o /kaspad:0.6.9/.
var reVersionUA = regexp.MustCompile(`(?:rupixd|kaspad):([0-9]+)\.([0-9]+)\.([0-9]+)`)

// Partes convierte "0.7.0" (o "v0.7.0", o "0.7.0-abc") en [0 7 0]. ok=false si no es una version.
func Partes(v string) (p [3]int, ok bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+ "); i >= 0 {
		v = v[:i]
	}
	campos := strings.Split(v, ".")
	if len(campos) != 3 {
		return p, false
	}
	for i, c := range campos {
		n, err := strconv.Atoi(c)
		if err != nil || n < 0 {
			return p, false
		}
		p[i] = n
	}
	return p, true
}

// Mayor dice si a > b.
func Mayor(a, b [3]int) bool {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

// DeUserAgent saca la ultima version rupixd/kaspad anunciada en un user agent ("" si no hay).
func DeUserAgent(ua string) string {
	m := reVersionUA.FindAllStringSubmatch(ua, -1)
	if len(m) == 0 {
		return ""
	}
	u := m[len(m)-1]
	return u[1] + "." + u[2] + "." + u[3]
}

// MasNueva devuelve la version mas alta que anuncia la MAYORIA de los pares por encima de
// `mia`, y true, o "" y false si no la hay. Reglas:
//   - hacen falta al menos 2 pares con version legible (con uno solo, nadie es mayoria);
//   - cuenta como "mas nueva" cualquier version mayor que la mia; la mayoria se mide sobre
//     los pares con version legible (los saludos raros no votan);
//   - se devuelve la mas alta entre las mayores, para que el aviso diga a donde ir.
func MasNueva(mia string, userAgents []string) (string, bool) {
	yo, ok := Partes(mia)
	if !ok {
		return "", false
	}
	legibles, mayores := 0, 0
	var tope [3]int
	var topeTxt string
	for _, ua := range userAgents {
		v := DeUserAgent(ua)
		p, ok := Partes(v)
		if !ok {
			continue
		}
		legibles++
		if Mayor(p, yo) {
			mayores++
			if topeTxt == "" || Mayor(p, tope) {
				tope, topeTxt = p, v
			}
		}
	}
	if legibles < 2 || mayores*2 <= legibles {
		return "", false
	}
	return topeTxt, true
}
