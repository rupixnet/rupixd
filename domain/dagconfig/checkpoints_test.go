package dagconfig

import "testing"

// TestCheckpointsPublicados (Rupix): lo que dice el codigo debe ser lo que dice
// CHECKPOINTS.md. Si alguien cambia un checkpoint sin actualizar el registro
// publico (o al reves), esta prueba lo delata.
func TestCheckpointsPublicados(t *testing.T) {
	if len(TestnetParams.Checkpoints) != 1 {
		t.Fatalf("testnet: se esperaba 1 checkpoint publicado, hay %d", len(TestnetParams.Checkpoints))
	}
	cp := TestnetParams.Checkpoints[0]
	if cp.BlueScore != 86400 || cp.Hash.String() != "7e2ece393c7d991c86e7ba915276cd85b5fc19e8647d5d197fa26bf116604fad" {
		t.Fatalf("testnet: checkpoint #1 no coincide con CHECKPOINTS.md: blue score %d hash %s", cp.BlueScore, cp.Hash)
	}
	if TestnetParams.CheckpointsExpireDAAScore <= cp.BlueScore+TestnetParams.MergeDepth {
		t.Fatalf("testnet: la caducidad (%d) debe quedar despues del umbral del checkpoint", TestnetParams.CheckpointsExpireDAAScore)
	}
	for _, p := range []*Params{&MainnetParams, &SimnetParams, &DevnetParams} {
		if len(p.Checkpoints) != 0 {
			t.Fatalf("%s: no debe tener checkpoints publicados", p.Name)
		}
	}
}
