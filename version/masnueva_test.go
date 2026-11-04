package version

import "testing"

// TestMasNueva (Rupix, 4-nov-2026): el aviso solo salta con MAYORIA de pares en una version
// mayor; un solo par no basta; los saludos ilegibles no votan; se anuncia la mas alta.
func TestMasNueva(t *testing.T) {
	casos := []struct {
		nombre string
		mia    string
		pares  []string
		quiero string
		hay    bool
	}{
		{"mayoria en 0.7.0", "0.6.9", []string{"/rupixd:0.7.0/rupixd:0.7.0(A)/", "/kaspad:0.7.0/rupixd:0.7.0(B)/", "/rupixd:0.6.9/"}, "0.7.0", true},
		{"empate no es mayoria", "0.6.9", []string{"/rupixd:0.7.0/", "/rupixd:0.6.9/"}, "", false},
		{"un solo par no basta", "0.6.9", []string{"/rupixd:0.7.0/"}, "", false},
		{"todos iguales", "0.6.9", []string{"/rupixd:0.6.9/", "/rupixd:0.6.9/"}, "", false},
		{"pares mas viejos", "0.7.0", []string{"/rupixd:0.6.9/", "/rupixd:0.6.6/"}, "", false},
		{"ilegibles no votan", "0.6.9", []string{"basura", "/rupixd:0.7.0/", "/rupixd:0.7.1/", "", "/rupixd:0.6.9/"}, "0.7.1", true},
		{"mi version con sufijo", "0.6.9-abc", []string{"/rupixd:0.7.0/", "/rupixd:0.7.0/"}, "0.7.0", true},
		{"mi version ilegible", "", []string{"/rupixd:0.7.0/", "/rupixd:0.7.0/"}, "", false},
		{"comparacion numerica, no de texto", "0.6.9", []string{"/rupixd:0.6.10/", "/rupixd:0.6.10/"}, "0.6.10", true},
		{"saludo heredado con kaspad delante", "0.6.9", []string{"/kaspad:0.6.9/rupixd:0.7.0(X)/", "/kaspad:0.6.9/rupixd:0.7.0(Y)/"}, "0.7.0", true},
	}
	for _, c := range casos {
		got, hay := MasNueva(c.mia, c.pares)
		if got != c.quiero || hay != c.hay {
			t.Errorf("%s: MasNueva(%q)=%q,%v; quiero %q,%v", c.nombre, c.mia, got, hay, c.quiero, c.hay)
		}
	}
	if DeUserAgent("/kaspad:0.6.9/rupixd:0.7.0(Nombre)/") != "0.7.0" {
		t.Errorf("DeUserAgent debe tomar la ultima version anunciada")
	}
	if p, ok := Partes("v1.2.3"); !ok || p != [3]int{1, 2, 3} {
		t.Errorf("Partes(v1.2.3) = %v %v", p, ok)
	}
}
