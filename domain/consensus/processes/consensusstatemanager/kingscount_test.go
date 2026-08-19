package consensusstatemanager

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
)

// countKingsDelta reproduce exactamente la aritmetica de calculateKingsCount
// sobre un acceptanceData: es la funcion que la ley aplica bloque a bloque.
func countKingsDelta(t *testing.T, parentCount uint64, acceptanceData externalapi.AcceptanceData) (uint64, error) {
	count := parentCount
	for _, bad := range acceptanceData {
		for _, tad := range bad.TransactionAcceptanceData {
			if !tad.IsAccepted {
				continue
			}
			for _, in := range tad.Transaction.Inputs {
				if in.UTXOEntry.ScriptPublicKey().Version == constants.LevelKings {
					count--
				}
			}
			for _, out := range tad.Transaction.Outputs {
				if out.ScriptPublicKey.Version == constants.LevelKings {
					count++
				}
			}
		}
	}
	if count > constants.MaxKings {
		return 0, errors.Wrapf(ruleerrors.ErrKingsCapExceeded, "count %d > %d", count, constants.MaxKings)
	}
	return count, nil
}

func kingIn() *externalapi.DomainTransactionInput {
	return &externalapi.DomainTransactionInput{
		UTXOEntry: utxo.NewUTXOEntry(1, &externalapi.ScriptPublicKey{Version: constants.LevelKings}, false, 0)}
}
func kingOut() *externalapi.DomainTransactionOutput {
	return &externalapi.DomainTransactionOutput{Value: 1, ScriptPublicKey: &externalapi.ScriptPublicKey{Version: constants.LevelKings}}
}
func acceptance(accepted bool, txs ...*externalapi.DomainTransaction) externalapi.AcceptanceData {
	tads := []*externalapi.TransactionAcceptanceData{}
	for _, tx := range txs {
		tads = append(tads, &externalapi.TransactionAcceptanceData{Transaction: tx, IsAccepted: accepted})
	}
	return externalapi.AcceptanceData{{TransactionAcceptanceData: tads}}
}

func TestKingsCountArithmetic(t *testing.T) {
	// 1. Nacen 3 Kings sobre 0
	ad := acceptance(true, &externalapi.DomainTransaction{Outputs: []*externalapi.DomainTransactionOutput{kingOut(), kingOut(), kingOut()}})
	c, err := countKingsDelta(t, 0, ad)
	if err != nil || c != 3 {
		t.Fatalf("3 kings nacen: esperado 3, got %d (%v)", c, err)
	}

	// 2. Transferencia: 1 entra, 1 sale -> el conteo NO cambia
	ad = acceptance(true, &externalapi.DomainTransaction{
		Inputs: []*externalapi.DomainTransactionInput{kingIn()}, Outputs: []*externalapi.DomainTransactionOutput{kingOut()}})
	c, _ = countKingsDelta(t, 100, ad)
	if c != 100 {
		t.Fatalf("transferencia de King: conteo debe seguir en 100, got %d", c)
	}

	// 3. Tx NO aceptada no cuenta
	ad = acceptance(false, &externalapi.DomainTransaction{Outputs: []*externalapi.DomainTransactionOutput{kingOut()}})
	c, _ = countKingsDelta(t, 50, ad)
	if c != 50 {
		t.Fatalf("tx no aceptada no debe contar: esperado 50, got %d", c)
	}

	// 4. El King 2,100 (el ultimo permitido) NACE
	ad = acceptance(true, &externalapi.DomainTransaction{Outputs: []*externalapi.DomainTransactionOutput{kingOut()}})
	c, err = countKingsDelta(t, constants.MaxKings-1, ad)
	if err != nil || c != constants.MaxKings {
		t.Fatalf("el King 2100 debe nacer: got %d (%v)", c, err)
	}

	// 5. EL KING 2,101 MUERE
	_, err = countKingsDelta(t, constants.MaxKings, ad)
	if !errors.Is(err, ruleerrors.ErrKingsCapExceeded) {
		t.Fatalf("el King 2101 debe rechazarse con ErrKingsCapExceeded, got: %v", err)
	}
	t.Logf("King 2,100 nace. King 2,101 muere: %v", err)
}
