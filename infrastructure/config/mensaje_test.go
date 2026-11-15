package config

import "testing"

// TestValidarMensaje (Rupix, 15-nov-2026): 15 caracteres, solo letras/numeros/espacios; sin
// enlaces ni simbolos; vacio = sin mensaje; acentos y otros alfabetos valen.
func TestValidarMensaje(t *testing.T) {
	ok := map[string]string{
		"":                 "",
		"hola desde FR":    "hola desde FR",
		"  viva   Rupix  ": "viva Rupix",
		"día 46 testnet":   "día 46 testnet",
		"привет мир":       "привет мир",
		"123456789012345":  "123456789012345",
		"ñandú y camión":   "ñandú y camión", // 14 caracteres con acentos (bytes son mas)
	}
	for in, want := range ok {
		got, err := ValidarMensaje(in)
		if err != nil || got != want {
			t.Errorf("%q: got %q, %v; want %q", in, got, err, want)
		}
	}
	mal := []string{
		"1234567890123456", // 16
		"ñandú y camión 7", // 16 (el primer borrador del test lo contaba como 15)
		"rupix.network",    // punto
		"hola@rupix",       // arroba
		"https://x",        // enlace
		"hola, mundo",      // coma
		"viva Rupix!",      // simbolo
		"<b>hola</b>",      // html
		"gracias 🙏",        // emoji (no es letra ni numero)
	}
	for _, in := range mal {
		if got, err := ValidarMensaje(in); err == nil {
			t.Errorf("%q debia rechazarse, dio %q", in, got)
		}
	}
	if ComentarioMensaje("hola") != "~hola" {
		t.Errorf("ComentarioMensaje")
	}
}
