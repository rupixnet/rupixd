// tarjeta (Rupix, 3-nov-2026): la pantalla de `balance` como una tarjeta con los colores de
// Rupix (los mismos de rupix.network): Gold dorado, marco violeta, cada gema con su color.
//
// Decisiones:
//   - Sin borde derecho. Los emojis miden distinto en cada terminal y un borde cerrado se
//     desalinea; un marco solo a la izquierda se ve bien en todas y en pantallas angostas.
//   - Color solo si la salida es una terminal, no hay NO_COLOR y no se pidio --sin-color.
//     Un script que hace `balance | grep` recibe texto plano.
//   - Color exacto (24 bits). Windows Terminal, PowerShell 7, Linux y Mac lo entienden; una
//     consola vieja lo aproxima al color mas cercano o lo ignora.
//   - El color no reemplaza al texto: "al dia" / "poniendose al dia" se lee igual sin color.
package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Paleta de rupix.network (index.html :root).
const (
	colorGold     = "#d4af37"
	colorVioleta  = "#b794f6"
	colorDiamante = "#b9e2f5"
	colorPlatino  = "#c9cdd4"
	colorRodio    = "#8a8f98"
	colorKings    = "#b06ef7"
	colorApagado  = "#a89fbd"
	colorBien     = "#4ade80" // verde: al dia
	colorAviso    = "#facc15" // amarillo: atencion
)

// pintar envuelve s en el color hex si color es true; si no, devuelve s tal cual.
func pintar(hex, s string, color bool) string {
	if !color || len(hex) != 7 || hex[0] != '#' {
		return s
	}
	var r, g, b int
	if _, err := fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b); err != nil {
		return s
	}
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, s)
}

// quiereColor decide si la salida lleva color: terminal real, sin NO_COLOR y sin --sin-color.
func quiereColor(sinColor bool) bool {
	if sinColor || os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// tarjetaSaldo es lo que la tarjeta muestra. Nodo* viene del nodo rupixd; si no contesto,
// NodoConsultado es false y la tarjeta lo dice en vez de inventar.
type tarjetaSaldo struct {
	Red            string // "testnet", "mainnet"...
	Gold           uint64 // rupias disponibles
	Pendiente      uint64 // rupias llegando (sin confirmar)
	Gemas          [5]uint64
	NodoConsultado bool
	Sincronizado   bool
	DAA            uint64
}

// lineasTarjeta arma la tarjeta como lineas de texto. Es una funcion pura para poder
// probarla; `color` entra por parametro.
func lineasTarjeta(t tarjetaSaldo, color bool) []string {
	marco := func(s string) string { return pintar(colorVioleta, s, color) }
	var out []string
	out = append(out, marco("┌─ ")+pintar(colorVioleta, "Rupix · "+t.Red, color))
	out = append(out, marco("│  ")+pintar(colorGold, fmt.Sprintf("%-10s%s RUPIX", "Gold", rupixBonito(t.Gold)), color))
	if t.Pendiente > 0 {
		out = append(out, marco("│  ")+pintar(colorAviso, fmt.Sprintf("%-10s+%s RUPIX %s", "", rupixBonito(t.Pendiente), T("tarjeta.llegando")), color))
	}
	out = append(out, marco("│"))
	gemas := []struct {
		icono, hex string
		nivel      uint32
	}{{"💎", colorDiamante, 1}, {"⬜", colorPlatino, 2}, {"◼ ", colorRodio, 3}, {"👑", colorKings, 4}}
	for _, g := range gemas {
		out = append(out, marco("│  ")+g.icono+" "+pintar(g.hex, fmt.Sprintf("%-10s%d", nombreNivel(g.nivel), t.Gemas[g.nivel]), color))
	}
	out = append(out, marco("│"))
	switch {
	case !t.NodoConsultado:
		out = append(out, marco("│  ")+pintar(colorApagado, T("tarjeta.nodo.sin"), color))
	case t.Sincronizado:
		out = append(out, marco("│  ")+pintar(colorBien, T("tarjeta.nodo.aldia"), color)+pintar(colorApagado, " · DAA "+miles(t.DAA), color))
	default:
		out = append(out, marco("│  ")+pintar(colorAviso, T("tarjeta.nodo.atras"), color)+pintar(colorApagado, " · DAA "+miles(t.DAA), color))
	}
	out = append(out, marco("└─"))
	return out
}

// rupixBonito imprime rupias como RUPIX con separador de miles y sin ceros de relleno
// (minimo dos decimales): 7510652741813 -> "75,106.52741813"; 150000000 -> "1.50".
func rupixBonito(rupias uint64) string {
	entero := rupias / 100_000_000
	dec := fmt.Sprintf("%08d", rupias%100_000_000)
	for len(dec) > 2 && dec[len(dec)-1] == '0' {
		dec = dec[:len(dec)-1]
	}
	return miles(entero) + "." + dec
}

// sinColor quita las secuencias de color (para tests y para medir texto).
func sinColor(s string) string {
	for {
		i := strings.Index(s, "\x1b[")
		if i < 0 {
			return s
		}
		j := strings.Index(s[i:], "m")
		if j < 0 {
			return s
		}
		s = s[:i] + s[i+j+1:]
	}
}
