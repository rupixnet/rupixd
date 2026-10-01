package main

import (
	"strings"
	"testing"
)

// Cada clave tiene las dos versiones, sin vacios, y con los mismos verbos de formato
// (%s, %d, %.1f) en el mismo orden: si no, un Printf en ingles reventaria en produccion.
func TestTextosCompletos(t *testing.T) {
	for clave, par := range textos {
		if strings.TrimSpace(par[0]) == "" || strings.TrimSpace(par[1]) == "" {
			t.Errorf("%s: falta una version", clave)
		}
		if verbos(par[0]) != verbos(par[1]) {
			t.Errorf("%s: los formatos no coinciden: es=%q en=%q", clave, verbos(par[0]), verbos(par[1]))
		}
	}
}

func verbos(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+1 < len(s) {
			j := i + 1
			for j < len(s) && (s[j] == '.' || (s[j] >= '0' && s[j] <= '9')) {
				j++
			}
			if j < len(s) {
				b.WriteByte(s[j])
			}
			i = j
		}
	}
	return b.String()
}

func TestTCaeAEspanolYDelataClavesRaras(t *testing.T) {
	idiomaActivo = "zz"
	if T("nivel.1") != "Diamante" {
		t.Fatalf("idioma desconocido debe caer a espanol")
	}
	idiomaActivo = "en"
	if T("nivel.1") != "Diamond" {
		t.Fatalf("en ingles Diamante es Diamond")
	}
	if T("clave.inexistente") != "clave.inexistente" {
		t.Fatalf("una clave sin texto debe verse tal cual, no esconderse")
	}
	idiomaActivo = "es"
}

func TestElegirIdioma(t *testing.T) {
	t.Setenv("RUPIX_LANG", "en")
	if v, origen := elegirIdioma(""); v != "en" || origen != "RUPIX_LANG" {
		t.Fatalf("RUPIX_LANG=en debe dar en: %s/%s", v, origen)
	}
	if v, origen := elegirIdioma("es"); v != "es" || origen != "--lang" {
		t.Fatalf("--lang manda sobre todo: %s/%s", v, origen)
	}
	idiomaActivo = "es"
}
