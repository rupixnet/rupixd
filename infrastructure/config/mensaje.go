package config

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// Rupix (15-nov-2026): el muro de nodos. Idea de JC (nodo #2). Un nodo puede llevar en su saludo
// un mensaje corto que el explorador muestra 24 h junto a su luz. Reglas, todas aqui o en el
// explorador, ninguna en consenso:
//   - solo nodos: el mensaje viaja en el saludo de la red como el nombre (comentario "~texto");
//   - 15 caracteres como maximo, solo letras (con acentos), numeros y espacios: sin enlaces,
//     sin arrobas, sin simbolos; el nodo lo rechaza antes de salir a la red;
//   - lo demas (uno por hora, 24 h de vida, nodos con >= 24 h de historia) lo aplica el explorador.
// No es un canal de mensajes: no puede decir mas que 15 caracteres y desaparece solo.

// MensajeMaxRunas es el largo maximo del mensaje del muro, en caracteres (runas).
const MensajeMaxRunas = 15

// MarcaMensaje antecede al mensaje dentro del saludo para distinguirlo del nombre.
const MarcaMensaje = "~"

// ValidarMensaje devuelve el mensaje limpio (espacios de los bordes fuera, espacios internos
// colapsados) o un error que explica por que no vale. Vacio es valido y significa "sin mensaje".
func ValidarMensaje(m string) (string, error) {
	m = strings.Join(strings.Fields(m), " ")
	if m == "" {
		return "", nil
	}
	if n := utf8.RuneCountInString(m); n > MensajeMaxRunas {
		return "", errors.Errorf("--mensaje: %d caracteres; el maximo es %d / %d characters; the maximum is %d", n, MensajeMaxRunas, n, MensajeMaxRunas)
	}
	for _, r := range m {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ') {
			return "", errors.Errorf("--mensaje: solo letras, numeros y espacios (no vale %q) / letters, digits and spaces only (%q is not allowed)", r, r)
		}
	}
	return m, nil
}

// ComentarioMensaje es como viaja el mensaje en el saludo: "~texto".
func ComentarioMensaje(m string) string {
	return MarcaMensaje + m
}
