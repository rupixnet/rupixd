package constants

import "testing"

// TestLevelUnlockDaaScore verifica el calendario de desbloqueos:
// cada gema nace exactamente en su halving.
func TestLevelUnlockDaaScore(t *testing.T) {
const blocksPerHalving = 42_000_000

casos := []struct {
nivel    uint16
nombre   string
esperado uint64
}{
{LevelGold, "Gold", 0},
{LevelDiamante, "Diamante", 42_000_000},   // halving 1
{LevelPlatino, "Platino", 84_000_000},     // halving 2
{LevelRodio, "Rodio", 126_000_000},        // halving 3
{LevelKings, "Kings", 168_000_000},        // halving 4
}
for _, c := range casos {
got := LevelUnlockDaaScore(c.nivel, blocksPerHalving)
if got != c.esperado {
t.Errorf("%s: desbloqueo esperado en DAA %d, got %d", c.nombre, c.esperado, got)
}
t.Logf("%-8s se desbloquea en DAA score %d", c.nombre, got)
}

// Las invariantes selladas del diseño
if BurnRatio != 10 {
t.Errorf("BurnRatio debe ser 10, es %d", BurnRatio)
}
if MaxKings != 2_100 {
t.Errorf("MaxKings debe ser 2100, es %d", MaxKings)
}
if GemAmount != 1 {
t.Errorf("GemAmount debe ser 1 (piezas enteras), es %d", GemAmount)
}
}
