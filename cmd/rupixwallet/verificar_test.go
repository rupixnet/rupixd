package main

import "testing"

// TestHashDeLista: la lista SHA256SUMS-<os>.txt que publica la release tiene el formato
// de `sha256sum` ("<hash>  <nombre>"); verificar busca su propio nombre ahi.
func TestHashDeLista(t *testing.T) {
	lista := "aaaa  rupixd\nBBBB  rupixwallet\ncccc *rupixminer\ndddd  ./bin/rupixctl.exe\n"
	casos := map[string]string{
		"rupixd": "aaaa", "rupixwallet": "bbbb", "rupixminer": "cccc", "rupixctl.exe": "dddd", "otro": "",
	}
	for nombre, esperado := range casos {
		if got := hashDeLista(lista, nombre); got != esperado {
			t.Errorf("%s: esperado %q, obtenido %q", nombre, esperado, got)
		}
	}
	if hashDeLista("", "rupixwallet") != "" {
		t.Errorf("lista vacia debe dar vacio")
	}
}

// TestSha256DeArchivo: el hash del binario se calcula sobre el archivo entero.
func TestSha256DeArchivo(t *testing.T) {
	if _, err := sha256DeArchivo("/no/existe"); err == nil {
		t.Fatalf("un archivo inexistente debe dar error")
	}
	h, err := sha256DeArchivo("verificar_test.go")
	if err != nil || len(h) != 64 {
		t.Fatalf("hash de este archivo: %q, %v", h, err)
	}
}
