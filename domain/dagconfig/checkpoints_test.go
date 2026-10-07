package dagconfig

import "testing"

// checkpointsPublicadosTestnet: copia a mano de la tabla "Registro de checkpoints
// publicados" de CHECKPOINTS.md. Se escribe aqui a proposito (no se lee del .md)
// para que un cambio en cualquiera de los tres lugares (codigo, este test, el
// registro publico) rompa la prueba.
var checkpointsPublicadosTestnet = []struct {
	blueScore uint64
	hash      string
}{
	{86400, "7e2ece393c7d991c86e7ba915276cd85b5fc19e8647d5d197fa26bf116604fad"},  // #1, 29-sep-2026, v0.6.1
	{691200, "ef79c0b3382489f8d4859608a86fe05411e60978e30fa9f3350272f26f8b9669"}, // #2, 7-oct-2026, v0.6.4
}

// TestCheckpointsPublicados (Rupix): lo que dice el codigo debe ser lo que dice
// CHECKPOINTS.md. Si alguien cambia un checkpoint sin actualizar el registro
// publico (o al reves), esta prueba lo delata.
func TestCheckpointsPublicados(t *testing.T) {
	if len(TestnetParams.Checkpoints) != len(checkpointsPublicadosTestnet) {
		t.Fatalf("testnet: se esperaban %d checkpoints publicados, hay %d", len(checkpointsPublicadosTestnet), len(TestnetParams.Checkpoints))
	}
	var anterior uint64
	for i, esperado := range checkpointsPublicadosTestnet {
		cp := TestnetParams.Checkpoints[i]
		if cp.BlueScore != esperado.blueScore || cp.Hash.String() != esperado.hash {
			t.Fatalf("testnet: checkpoint #%d no coincide con CHECKPOINTS.md: blue score %d hash %s", i+1, cp.BlueScore, cp.Hash)
		}
		if i > 0 && cp.BlueScore <= anterior {
			t.Fatalf("testnet: el checkpoint #%d (blue %d) no es mas reciente que el anterior (%d)", i+1, cp.BlueScore, anterior)
		}
		anterior = cp.BlueScore
		if TestnetParams.CheckpointsExpireDAAScore <= cp.BlueScore+TestnetParams.MergeDepth {
			t.Fatalf("testnet: la caducidad (%d) debe quedar despues del umbral del checkpoint #%d", TestnetParams.CheckpointsExpireDAAScore, i+1)
		}
	}
	for _, p := range []*Params{&MainnetParams, &SimnetParams, &DevnetParams} {
		if len(p.Checkpoints) != 0 {
			t.Fatalf("%s: no debe tener checkpoints publicados", p.Name)
		}
	}
}
