package transactionvalidator

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
)

// ---- helpers ----
func gemInput(version uint16, amount uint64) *externalapi.DomainTransactionInput {
	return &externalapi.DomainTransactionInput{
		UTXOEntry: utxo.NewUTXOEntry(amount,
			&externalapi.ScriptPublicKey{Script: []byte{}, Version: version}, false, 0),
	}
}
func gemOutput(version uint16, amount uint64) *externalapi.DomainTransactionOutput {
	return &externalapi.DomainTransactionOutput{
		Value:           amount,
		ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{}, Version: version},
	}
}
func burnOutput(amount uint64) *externalapi.DomainTransactionOutput {
	return &externalapi.DomainTransactionOutput{
		Value:           amount,
		ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{0x6a}, Version: constants.LevelGold},
	}
}
func makeTx(in []*externalapi.DomainTransactionInput, out []*externalapi.DomainTransactionOutput) *externalapi.DomainTransaction {
	return &externalapi.DomainTransaction{Inputs: in, Outputs: out}
}
func gems(version uint16, n int) []*externalapi.DomainTransactionInput {
	r := []*externalapi.DomainTransactionInput{}
	for i := 0; i < n; i++ {
		r = append(r, gemInput(version, constants.GemAmount))
	}
	return r
}
func newTestValidator() *transactionValidator {
	// halvings chicos (D:100 P:200 R:300 K:400), burn con valores reales de mainnet
	return &transactionValidator{blocksPerHalving: 100, burnBase: constants.BurnBase, burnPerByte: constants.BurnPerByte}
}

// txBurn: quema generosa que cubre el burn por tx de cualquier tx de test (0.001 RUPIX)
const txBurn = uint64(100_000)

func TestLevelRules(t *testing.T) {
	v := newTestValidator()
	const abierto = uint64(1000)

	// ===== BURN POR TRANSACCION =====
	t.Run("gold puro con quema pasa", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(0, 500_000)},
			[]*externalapi.DomainTransactionOutput{gemOutput(0, 399_000), burnOutput(txBurn)})
		if err := v.checkLevelRules(tx, 0, false); err != nil {
			t.Fatalf("gold puro con quema rechazado: %v", err)
		}
	})
	t.Run("ataque: tx sin quema muere", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(0, 500_000)},
			[]*externalapi.DomainTransactionOutput{gemOutput(0, 499_000)})
		err := v.checkLevelRules(tx, abierto, false)
		if !errors.Is(err, ruleerrors.ErrInsufficientBurn) {
			t.Fatalf("tx sin quema debe morir con ErrInsufficientBurn, got: %v", err)
		}
	})
	t.Run("ataque: quema por debajo del minimo", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(0, 500_000)},
			[]*externalapi.DomainTransactionOutput{gemOutput(0, 499_000), burnOutput(500)}) // < 1000 base
		if err := v.checkLevelRules(tx, abierto, false); !errors.Is(err, ruleerrors.ErrInsufficientBurn) {
			t.Fatalf("quema de 500 (< base 1000) debe morir, got: %v", err)
		}
	})
	t.Run("coinbase exenta del burn", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{},
			[]*externalapi.DomainTransactionOutput{gemOutput(0, 50_000_000)})
		if err := v.checkLevelRules(tx, abierto, true); err != nil {
			t.Fatalf("la coinbase no debe quemar: %v", err)
		}
	})

	// ===== ESCALERA: ASCENSOS ENTRE GEMAS =====
	t.Run("ascenso valido: 10 diamantes -> 1 platino", func(t *testing.T) {
		in := append(gems(constants.LevelDiamante, 10), gemInput(0, 200_000))
		tx := makeTx(in, []*externalapi.DomainTransactionOutput{gemOutput(constants.LevelPlatino, 1), burnOutput(txBurn)})
		if err := v.checkLevelRules(tx, abierto, false); err != nil {
			t.Fatalf("ascenso valido rechazado: %v", err)
		}
	})
	t.Run("ataque: quemar solo 9", func(t *testing.T) {
		in := append(gems(constants.LevelDiamante, 9), gemInput(0, 200_000))
		tx := makeTx(in, []*externalapi.DomainTransactionOutput{gemOutput(constants.LevelPlatino, 1), burnOutput(txBurn)})
		if err := v.checkLevelRules(tx, abierto, false); err == nil {
			t.Fatal("ascenso con 9 quemados ACEPTADO — deben ser 10")
		}
	})
	t.Run("ataque: nivel bloqueado (platino antes del halving 2)", func(t *testing.T) {
		in := append(gems(constants.LevelDiamante, 10), gemInput(0, 200_000))
		tx := makeTx(in, []*externalapi.DomainTransactionOutput{gemOutput(constants.LevelPlatino, 1), burnOutput(txBurn)})
		err := v.checkLevelRules(tx, 150, false) // platino abre en 200
		if err == nil || !strings.Contains(err.Error(), "bloqueado") {
			t.Fatalf("platino antes de su halving debe estar bloqueado, got: %v", err)
		}
	})
	t.Run("ataque: gema fraccionada", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(constants.LevelRodio, 1), gemInput(0, 200_000)},
			[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelRodio, 5), burnOutput(txBurn)})
		if err := v.checkLevelRules(tx, abierto, false); err == nil {
			t.Fatal("gema con monto 5 ACEPTADA — piezas enteras")
		}
	})
	t.Run("ataque: gemas desaparecen sin ascenso", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(constants.LevelRodio, 1), gemInput(constants.LevelRodio, 1), gemInput(0, 200_000)},
			[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelRodio, 1), burnOutput(txBurn)})
		if err := v.checkLevelRules(tx, abierto, false); err == nil {
			t.Fatal("2 rodios entran, 1 sale, sin ascenso — ACEPTADO")
		}
	})
	t.Run("transferencia de King pasa", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(constants.LevelKings, 1), gemInput(0, 200_000)},
			[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelKings, 1), burnOutput(txBurn)})
		if err := v.checkLevelRules(tx, abierto, false); err != nil {
			t.Fatalf("transferencia de King rechazada: %v", err)
		}
	})

	t.Run("ataque: gema a OpReturn (destruccion disfrazada de ascenso)", func(t *testing.T) {
		// 10 diamantes entran, "nace" un platino... con script OpReturn: una tumba.
		// El UTXO set lo excluiria: los 10 diamantes desaparecen sin que nazca nada.
		in := append(gems(constants.LevelDiamante, 10), gemInput(0, 200_000))
		tumba := &externalapi.DomainTransactionOutput{Value: 1,
			ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{0x6a}, Version: constants.LevelPlatino}}
		tx := makeTx(in, []*externalapi.DomainTransactionOutput{tumba, burnOutput(txBurn)})
		if err := v.checkLevelRules(tx, abierto, false); err == nil {
			t.Fatal("gema hacia OpReturn ACEPTADA — destruccion disfrazada de ascenso")
		}
	})

	t.Run("ataque: nivel fantasma (Version 99)", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(0, 500_000)},
			[]*externalapi.DomainTransactionOutput{gemOutput(99, 1), burnOutput(txBurn)})
		if err := v.checkLevelRules(tx, abierto, false); err == nil {
			t.Fatal("output Version 99 ACEPTADO — no existe ese nivel")
		}
	})
	t.Run("ataque: la coinbase se regala un King", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{},
			[]*externalapi.DomainTransactionOutput{gemOutput(0, 50_000_000), gemOutput(constants.LevelKings, 1)})
		if err := v.checkLevelRules(tx, abierto, true); err == nil {
			t.Fatal("coinbase con un King ACEPTADA — la recompensa solo es Gold")
		}
	})

	// ===== LA PUERTA: GOLD -> DIAMANTE =====
	diez := uint64(10 * constants.RupiaPerRupix)
	t.Run("ascenso valido: quema 10 gold -> 1 diamante (con cambio)", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(0, 12*constants.RupiaPerRupix)},
			[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelDiamante, 1), burnOutput(diez), gemOutput(0, 2*constants.RupiaPerRupix-txBurn)})
		if err := v.checkLevelRules(tx, abierto, false); err != nil {
			t.Fatalf("ascenso a diamante rechazado: %v", err)
		}
	})
	t.Run("ascenso multiple: 30 gold -> 3 diamantes", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(0, 30*constants.RupiaPerRupix)},
			[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelDiamante, 1), gemOutput(constants.LevelDiamante, 1), gemOutput(constants.LevelDiamante, 1), burnOutput(3 * diez)})
		if err := v.checkLevelRules(tx, abierto, false); err != nil {
			t.Fatalf("ascenso multiple rechazado: %v", err)
		}
	})
	t.Run("ataque: quema insuficiente para diamante (9 gold)", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(0, 10*constants.RupiaPerRupix)},
			[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelDiamante, 1), burnOutput(9 * constants.RupiaPerRupix)})
		if err := v.checkLevelRules(tx, abierto, false); err == nil {
			t.Fatal("diamante con 9 gold ACEPTADO — deben ser 10 exactos")
		}
	})
	t.Run("ataque: diamante sin OpReturn (gold iria al minero)", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(0, 11*constants.RupiaPerRupix)},
			[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelDiamante, 1)})
		if err := v.checkLevelRules(tx, abierto, false); err == nil {
			t.Fatal("diamante sin quema ACEPTADO — el gold iria al minero")
		}
	})
	t.Run("ataque: diamante antes del halving 1", func(t *testing.T) {
		tx := makeTx([]*externalapi.DomainTransactionInput{gemInput(0, 10*constants.RupiaPerRupix)},
			[]*externalapi.DomainTransactionOutput{gemOutput(constants.LevelDiamante, 1), burnOutput(diez)})
		if err := v.checkLevelRules(tx, 50, false); err == nil { // diamante abre en 100
			t.Fatal("diamante ANTES del halving 1 — debe estar bloqueado")
		}
	})
}

// ============================================================================
// TestGemAscensionAttacksRound2 — ataques del auditor "mas cabron", ronda 2.
// Angulos nuevos no cubiertos por TestLevelRules: robo multi-pieza, timing
// exacto del halving (off-by-one), descenso fantasma, cambio excesivo, y
// doble ascenso contradictorio. Todos DEBEN ser rechazados.
// ============================================================================
func TestGemAscensionAttacksRound2(t *testing.T) {
v := newTestValidator()
const abierto = uint64(1000) // todos los niveles desbloqueados (K abre en 400)

// --- ATAQUE 1: 2 Kings quemando solo 10 Rodios (deben ser 20) ---
t.Run("ataque: 2 kings con solo 10 rodios", func(t *testing.T) {
in := append(gems(constants.LevelRodio, 10), gemInput(0, 200_000))
tx := makeTx(in, []*externalapi.DomainTransactionOutput{
gemOutput(constants.LevelKings, 1), gemOutput(constants.LevelKings, 1), burnOutput(txBurn)})
if err := v.checkLevelRules(tx, abierto, false); err == nil {
t.Fatal("2 kings con 10 rodios ACEPTADO — deben ser 20 rodios")
}
})

// --- ATAQUE 2: timing EXACTO del halving (off-by-one) ---
// Kings abre en nivel*blocksPerHalving = 4*100 = 400.
t.Run("timing: king en unlock-1 (399) debe fallar", func(t *testing.T) {
in := append(gems(constants.LevelRodio, 10), gemInput(0, 200_000))
tx := makeTx(in, []*externalapi.DomainTransactionOutput{
gemOutput(constants.LevelKings, 1), burnOutput(txBurn)})
if err := v.checkLevelRules(tx, 399, false); err == nil {
t.Fatal("king en DAA 399 (unlock-1) ACEPTADO — abre en 400")
}
})
t.Run("timing: king en unlock exacto (400) debe pasar", func(t *testing.T) {
in := append(gems(constants.LevelRodio, 10), gemInput(0, 200_000))
tx := makeTx(in, []*externalapi.DomainTransactionOutput{
gemOutput(constants.LevelKings, 1), burnOutput(txBurn)})
if err := v.checkLevelRules(tx, 400, false); err != nil {
t.Fatalf("king en DAA 400 (unlock exacto) rechazado: %v", err)
}
})

// --- ATAQUE 3: robo con cambio excesivo (11 rodio in, 2 out, 1 king) ---
// Consume solo 9 rodios reales pero crea 1 king (exige 10).
t.Run("ataque: cambio excesivo esconde quema insuficiente", func(t *testing.T) {
in := append(gems(constants.LevelRodio, 11), gemInput(0, 200_000))
tx := makeTx(in, []*externalapi.DomainTransactionOutput{
gemOutput(constants.LevelRodio, 1), gemOutput(constants.LevelRodio, 1),
gemOutput(constants.LevelKings, 1), burnOutput(txBurn)})
if err := v.checkLevelRules(tx, abierto, false); err == nil {
t.Fatal("11 rodio - 2 cambio = 9 quemados, 1 king ACEPTADO — faltan rodios")
}
})

// --- ATAQUE 4: descenso fantasma / doble ascenso contradictorio ---
// Intento crear Platino Y Kings desde un set de Rodios insuficiente.
t.Run("ataque: doble ascenso desde rodios insuficientes", func(t *testing.T) {
in := append(gems(constants.LevelRodio, 10), gemInput(0, 200_000))
tx := makeTx(in, []*externalapi.DomainTransactionOutput{
gemOutput(constants.LevelKings, 1), gemOutput(constants.LevelPlatino, 1), burnOutput(txBurn)})
if err := v.checkLevelRules(tx, abierto, false); err == nil {
t.Fatal("king + platino desde solo 10 rodios ACEPTADO — imposible")
}
})

// --- ATAQUE 5: transferir rodios Y crear king con los mismos ---
// in: 10 rodio; out: 10 rodio (transferidos) + 1 king. No queda nada quemado.
t.Run("ataque: transferir y ascender con las mismas piezas", func(t *testing.T) {
out := []*externalapi.DomainTransactionOutput{}
for i := 0; i < 10; i++ {
out = append(out, gemOutput(constants.LevelRodio, 1))
}
out = append(out, gemOutput(constants.LevelKings, 1), burnOutput(txBurn))
in := append(gems(constants.LevelRodio, 10), gemInput(0, 200_000))
if err := v.checkLevelRules(makeTx(in, out), abierto, false); err == nil {
t.Fatal("transferir 10 rodios Y crear king ACEPTADO — no se quemo nada")
}
})

// --- ATAQUE 6: king legitimo con cambio correcto (debe PASAR) ---
t.Run("control: king legitimo con cambio (11 rodio, 1 cambio)", func(t *testing.T) {
in := append(gems(constants.LevelRodio, 11), gemInput(0, 200_000))
tx := makeTx(in, []*externalapi.DomainTransactionOutput{
gemOutput(constants.LevelRodio, 1), gemOutput(constants.LevelKings, 1), burnOutput(txBurn)})
if err := v.checkLevelRules(tx, abierto, false); err != nil {
t.Fatalf("king legitimo con cambio rechazado: %v", err)
}
})
}
