package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/utils"
)

// confirmar (Rupix) lee una linea y devuelve true solo si la persona escribio si/s/yes/y.
// Cualquier otra cosa, incluido Enter a secas, es "no": lo irreversible se confirma a proposito.
func confirmar(pregunta string) bool {
	fmt.Print(pregunta)
	linea, err := utils.ReadLine(bufio.NewReader(os.Stdin))
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(linea)) {
	case "si", "sí", "s", "yes", "y":
		return true
	}
	return false
}

// miles imprime 1234567 como 1,234,567.
func miles(n uint64) string {
	s := fmt.Sprintf("%d", n)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// tiempoBloques estima cuanto tarda la red en producir n bloques (1 por segundo).
func tiempoBloques(n uint64) string {
	switch {
	case n < 120:
		return fmt.Sprintf(T("tiempo.seg"), n)
	case n < 7200:
		return fmt.Sprintf(T("tiempo.min"), n/60)
	case n < 172800:
		return fmt.Sprintf(T("tiempo.hora"), float64(n)/3600)
	}
	return fmt.Sprintf(T("tiempo.dia"), float64(n)/86400)
}

// traducirErrorNodo (Rupix) explica en palabras los rechazos conocidos del nodo, sin
// esconder el original: el texto del nodo va al final, tal cual, para quien lo necesite.
func traducirErrorNodo(err error) error {
	msg := err.Error()
	var explicacion string
	switch {
	case strings.Contains(msg, "bloqueado"):
		explicacion = T("err.nivel_cerrado")
	case strings.Contains(msg, "no sincronizado"):
		explicacion = T("err.no_sync")
	case strings.Contains(msg, "Gold") && strings.Contains(msg, "insuficiente"):
		explicacion = T("err.gold")
	case strings.Contains(msg, "message authentication failed"):
		explicacion = T("err.clave")
	default:
		return err
	}
	return errors.Errorf(T("err.nodo_dijo"), explicacion, msg)
}

// rupixTxt imprime un monto en RUPIX sin el relleno de columnas del formateador heredado.
func rupixTxt(n uint64) string { return strings.TrimSpace(utils.FormatRupix(n)) }

// verificarFrase (Rupix) pide dos palabras de la frase semilla, elegidas al azar, y
// devuelve true si las dos coinciden. Es lo que hacen las wallets serias: convierte
// "luego la anoto" en "ya la anote".
func verificarFrase(mnemonics []string) bool {
	lector := bufio.NewReader(os.Stdin)
	for i, m := range mnemonics {
		palabras := strings.Fields(m)
		if len(palabras) < 2 {
			continue
		}
		a := rand.Intn(len(palabras))
		b := rand.Intn(len(palabras) - 1)
		if b >= a {
			b++
		}
		for _, pos := range []int{a, b} {
			fmt.Printf(T("crear.verifica"), pos+1, i+1)
			linea, err := utils.ReadLine(lector)
			if err != nil || strings.ToLower(strings.TrimSpace(linea)) != palabras[pos] {
				fmt.Printf(T("crear.mal")+"\n", pos+1, i+1)
				return false
			}
		}
	}
	return true
}

// traducirErrorDaemon (Rupix) reconoce "no me pude conectar al daemon" y dice que hacer.
func traducirErrorDaemon(err error, daemonAddress string, netFlag string) error {
	msg := err.Error()
	if strings.Contains(msg, "daemon is not running") || strings.Contains(msg, "connection refused") || strings.Contains(msg, "Unavailable") || strings.Contains(msg, "connection error") {
		return errors.Errorf(T("daemon.apagado"), daemonAddress, netFlag)
	}
	return err
}
