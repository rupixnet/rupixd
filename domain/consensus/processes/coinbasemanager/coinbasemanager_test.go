package coinbasemanager

import (
"testing"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
"github.com/rupixnet/rupixd/domain/dagconfig"
)

// baseSubsidy es la emision por bloque al inicio: 0.5 RUPIX
const baseSubsidy = 50_000_000

func newTestCoinbaseManager(deflationaryPhaseDaaScore, baseSub uint64) *coinbaseManager {
iface := New(nil, 0, 0, 0, &externalapi.DomainHash{},
deflationaryPhaseDaaScore, baseSub,
nil, nil, nil, nil, nil, nil, nil)
return iface.(*coinbaseManager)
}

func TestCalcDeflationaryPeriodBlockSubsidy(t *testing.T) {
cbm := newTestCoinbaseManager(0, baseSubsidy)

tests := []struct {
name          string
blockDaaScore uint64
expected      uint64
}{
{"bloque 1", 1, baseSubsidy},
{"ultimo bloque antes del primer halving", blocksPerHalving - 1, baseSubsidy},
{"primer halving", blocksPerHalving, baseSubsidy / 2},
{"segundo halving", blocksPerHalving * 2, baseSubsidy / 4},
{"quinto halving", blocksPerHalving * 5, baseSubsidy / 32},
{"halving 25 - ultimo con emision", blocksPerHalving * 25, 1},
{"halving 26 - emision agotada", blocksPerHalving * 26, 0},
{"halving 64 - guarda contra overflow", blocksPerHalving * 64, 0},
{"halving 100 - muy despues del final", blocksPerHalving * 100, 0},
}

for _, test := range tests {
got := cbm.calcDeflationaryPeriodBlockSubsidy(test.blockDaaScore)
if got != test.expected {
t.Errorf("%s: esperado %d, obtenido %d", test.name, test.expected, got)
}
}
}

// TestTotalSupply verifica la emision total de Rupix sumando cada periodo
// de halving. Este test es la prueba verificable de la tokenomics.
func TestTotalSupply(t *testing.T) {
var total uint64
for halving := uint64(0); halving < 64; halving++ {
subsidy := uint64(baseSubsidy) >> halving
if subsidy == 0 {
break
}
total += subsidy * blocksPerHalving
}

if total > constants.MaxRupia {
t.Errorf("la emision total (%d rupias) supera MaxRupia (%d)", total, constants.MaxRupia)
}

const esperado = 4_199_999_496_000_000 // 41,999,994.96 RUPIX
if total != esperado {
t.Errorf("emision total: esperado %d rupias, obtenido %d", esperado, total)
}

t.Logf("Emision total Rupix: %d rupias = %.2f RUPIX",
total, float64(total)/float64(constants.RupiaPerRupix))
t.Logf("MaxRupia (techo):    %d rupias = %.2f RUPIX",
constants.MaxRupia, float64(constants.MaxRupia)/float64(constants.RupiaPerRupix))
}

// TestNoPremine verifica que el bloque genesis no emite nada.
func TestNoPremine(t *testing.T) {
for _, params := range []*dagconfig.Params{
&dagconfig.MainnetParams, &dagconfig.TestnetParams,
&dagconfig.SimnetParams, &dagconfig.DevnetParams,
} {
if params.SubsidyGenesisReward != 0 {
t.Errorf("%s: el genesis emite %d rupias, debe ser 0 (sin premine)",
params.Name, params.SubsidyGenesisReward)
}
}
}
