package transactionvalidator

import (
	"strings"
"encoding/json"
"os"
"path/filepath"
"testing"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
)

type corpusCase struct {
Tx struct {
Inputs  []uint16 `json:"inputs"`
Outputs []struct {
Version     uint16 `json:"version"`
Value       uint64 `json:"value"`
Unspendable bool   `json:"unspendable"`
} `json:"outputs"`
TxBytes  uint64 `json:"tx_bytes"`
PovDaa   uint64 `json:"pov_daa"`
Coinbase bool   `json:"coinbase"`
} `json:"tx"`
ExpectValid bool   `json:"expect_valid"`
Reason      string `json:"reason"`
}

func buildCorpusTx(c corpusCase) *externalapi.DomainTransaction {
inputs := make([]*externalapi.DomainTransactionInput, 0, len(c.Tx.Inputs))
for _, ver := range c.Tx.Inputs {
amount := uint64(constants.GemAmount)
if ver == constants.LevelGold {
amount = 1_000_000
}
inputs = append(inputs, gemInput(ver, amount))
}
outputs := make([]*externalapi.DomainTransactionOutput, 0, len(c.Tx.Outputs))
for _, o := range c.Tx.Outputs {
if o.Unspendable {
// output imposible de gastar (OpReturn 0x6a): quema de Gold o
// "tumba" de gema. Se respeta la version para que checkLevelRules
// detecte una gema mandada a un script imposible.
outputs = append(outputs, &externalapi.DomainTransactionOutput{
Value:           o.Value,
ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{0x6a}, Version: o.Version},
})
} else {
outputs = append(outputs, gemOutput(o.Version, o.Value))
}
}
return makeTx(inputs, outputs)
}

func TestLadderCorpus(t *testing.T) {
path := filepath.Join("testdata", "rupix_ladder_corpus.json")
raw, err := os.ReadFile(path)
if err != nil {
t.Skipf("corpus no encontrado en %s: %v", path, err)
return
}
var cases []corpusCase
if err := json.Unmarshal(raw, &cases); err != nil {
t.Fatalf("no se pudo parsear el corpus: %v", err)
}
// El corpus asume tx_bytes fijo; nuestro codigo usa el tamaño real (txmass).
// Para validar la LOGICA de la escalera (no el burn-por-tx que depende de
// bytes), ponemos burnBase/burnPerByte en 0. El burn se valida por separado.
v := &transactionValidator{blocksPerHalving: 42_000_000, burnBase: 0, burnPerByte: 0}
desacuerdos, criticos := 0, 0
saltados := 0
for i, c := range cases {
// Saltar casos cuyo veredicto depende del burn-por-tx (I5): pusimos
// burnBase/burnPerByte en 0 para validar la LOGICA de la escalera.
// El burn-por-tx se valida aparte en TestLevelRules con bytes reales.
if strings.Contains(c.Reason, "I5") || strings.Contains(c.Reason, "burn por tx") {
saltados++
continue
}
tx := buildCorpusTx(c)
err := v.checkLevelRules(tx, c.Tx.PovDaa, c.Tx.Coinbase)
got := err == nil
if got != c.ExpectValid {
desacuerdos++
if !c.ExpectValid && got {
criticos++
}
if desacuerdos <= 20 {
t.Errorf("caso %d: esperado=%v got=%v | %q | err=%v", i, c.ExpectValid, got, c.Reason, err)
}
}
}
t.Logf("corpus: %d casos, %d saltados (I5/burn), %d desacuerdos (%d criticos)", len(cases), saltados, desacuerdos, criticos)
if desacuerdos > 0 {
t.Fatalf("corpus: %d desacuerdos (%d criticos)", desacuerdos, criticos)
}
}
