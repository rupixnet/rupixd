package transactionvalidator

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
)

// TestQuemaDeDiamanteEsExacta (auditor, 7-oct-2026): la tabla decia que el Diamante quema
// 10 Gold "ademas de la quema por transaccion", y el oraculo del fuzz exigia exactamente
// 10. Uno de los dos mentia. Este test fija la verdad del codigo: la quema del Diamante es
// EXACTAMENTE 10 RUPIX por Diamante en OpReturn, y esa misma quema cubre la quema por
// transaccion (no se suma). 10 RUPIX + la quema de tx se rechaza. La tabla se corrigio.
func TestQuemaDeDiamanteEsExacta(t *testing.T) {
	v := newTestValidator()
	const abierto = uint64(1000)
	diez := uint64(constants.BurnRatio * constants.RupiaPerRupix)
	forja := func(quema uint64) *externalapi.DomainTransaction {
		return makeTx([]*externalapi.DomainTransactionInput{gemInput(0, 20*constants.RupiaPerRupix)},
			[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelDiamante, 1), burnOutput(quema), gemOutput(0, constants.RupiaPerRupix)})
	}
	if err := v.checkLevelRules(forja(diez), abierto, false); err != nil {
		t.Fatalf("exactamente 10 RUPIX debe pasar: %v", err)
	}
	if err := v.checkLevelRules(forja(diez+txBurn), abierto, false); err == nil {
		t.Fatalf("10 RUPIX + la quema por tx NO es la regla: la quema del Diamante es exacta y ya cubre la quema por tx")
	}
	if err := v.checkLevelRules(forja(diez-1), abierto, false); err == nil {
		t.Fatalf("una rupia menos de 10 RUPIX debe rechazarse")
	}
	// Y la quema por tx sigue exigiendose cuando NO hay forja: una tx de Gold sin quema muere.
	sinQuema := makeTx([]*externalapi.DomainTransactionInput{gemInput(0, 20*constants.RupiaPerRupix)},
		[]*externalapi.DomainTransactionOutput{gemOutput(0, 19*constants.RupiaPerRupix)})
	if err := v.checkLevelRules(sinQuema, abierto, false); err == nil {
		t.Fatalf("una tx sin quema debe rechazarse")
	}
}
