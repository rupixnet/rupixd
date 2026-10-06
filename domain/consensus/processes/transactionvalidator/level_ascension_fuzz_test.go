package transactionvalidator

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/util/txmass"
)

// FuzzCheckLevelRules (hueco #6 de ESPECIFICACION.md, ultimo objetivo): transacciones
// generadas con cualquier combinacion de entradas y salidas por nivel, quema de Gold,
// monto de gema, DAA y coinbase. El oraculo es una segunda implementacion de las
// reglas de la seccion 3 de ESPECIFICACION.md, escrita aparte y sin mirar
// checkLevelRules; si las dos discrepan en "valida / no valida", el fuzzer lo reporta.
// Invariantes:
//  1. Nunca entra en panico, con cualquier entrada.
//  2. Veredicto igual al del oraculo.
//  3. Si es valida y no es coinbase: ninguna gema desaparece salvo exactamente
//     BurnRatio por cada gema que nace un nivel arriba, y ninguna gema nace sin esa
//     quema (o, para Diamante, sin la quema exacta de Gold en OpReturn).
//  4. Determinista.
//
// Correr: go test -fuzz=FuzzCheckLevelRules -fuzztime=2h ./domain/consensus/processes/transactionvalidator/
func FuzzCheckLevelRules(f *testing.F) {
	// Semillas: forja de un Diamante correcta, ascenso Platino correcto, transferencia,
	// una gema que desaparece, una gema fraccionada, coinbase con gema, nivel 5.
	f.Add(uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(1), uint8(0), uint8(0), uint8(0), uint64(10*constants.RupiaPerRupix), uint64(constants.GemAmount), uint64(1000), false, false, uint8(0))
	f.Add(uint8(1), uint8(10), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0), uint8(0), uint64(txBurn), uint64(constants.GemAmount), uint64(1000), false, false, uint8(0))
	f.Add(uint8(1), uint8(0), uint8(0), uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0), uint64(txBurn), uint64(constants.GemAmount), uint64(1000), false, false, uint8(0))
	f.Add(uint8(1), uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint64(txBurn), uint64(constants.GemAmount), uint64(1000), false, false, uint8(0))
	f.Add(uint8(1), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint64(txBurn), uint64(2), uint64(1000), false, false, uint8(0))
	f.Add(uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(1), uint8(0), uint8(0), uint8(0), uint64(0), uint64(constants.GemAmount), uint64(1000), true, false, uint8(0))
	f.Add(uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint64(txBurn), uint64(constants.GemAmount), uint64(1000), false, false, uint8(5))

	f.Fuzz(func(t *testing.T,
		inG, inD, inP, inR, inK uint8, // entradas por nivel
		outG, outD, outP, outR, outK uint8, // salidas gastables por nivel
		quema uint64, // rupias de Gold en OpReturn
		montoGema uint64, // monto de cada salida de gema
		povDaa uint64,
		coinbase bool,
		tumba bool, // una salida de Diamante a un script imposible de gastar
		extraNivel uint8, // >4: se agrega una salida de ese nivel (no existe)
	) {
		v := newTestValidator()
		// Montos absurdos de gema no aportan: lo interesante es 1 contra todo lo demas.
		montoGema %= 4
		if coinbase {
			quema = 0
		}
		tx := construirTxFuzz(inG, inD, inP, inR, inK, outG, outD, outP, outR, outK, quema, montoGema, tumba, extraNivel)

		err := v.checkLevelRules(tx, povDaa, coinbase) // invariante 1: no panico
		valida := err == nil

		esperado := oraculoEscalera(v, tx, povDaa, coinbase, inG, inD, inP, inR, inK, outD, outP, outR, outK, quema, montoGema, tumba, extraNivel)
		if valida != esperado {
			t.Fatalf("checkLevelRules dice valida=%v, la especificacion dice %v (err=%v)\n in G%d D%d P%d R%d K%d · out G%d D%d P%d R%d K%d · quema %d · monto %d · daa %d · coinbase %v · tumba %v · extra %d",
				valida, esperado, err, inG, inD, inP, inR, inK, outG, outD, outP, outR, outK, quema, montoGema, povDaa, coinbase, tumba, extraNivel)
		}

		// Invariante 3: conservacion de gemas en toda tx valida.
		if valida && !coinbase {
			in := [5]int{int(inG), int(inD), int(inP), int(inR), int(inK)}
			out := [5]int{int(outG), int(outD), int(outP), int(outR), int(outK)}
			for n := int(constants.LevelDiamante); n <= int(constants.LevelKings); n++ {
				nacenArriba := 0
				if n < int(constants.LevelKings) {
					nacenArriba = out[n+1] - in[n+1]
					if nacenArriba < 0 {
						nacenArriba = 0
					}
				}
				desaparecen := in[n] - out[n]
				if desaparecen > 0 && desaparecen != constants.BurnRatio*nacenArriba {
					t.Fatalf("INVARIANTE 3 ROTA: valida y desaparecen %d gemas de nivel %d sin que nazcan %d arriba", desaparecen, n, desaparecen/constants.BurnRatio)
				}
				nacen := out[n] - in[n]
				if nacen > 0 {
					if n == int(constants.LevelDiamante) {
						if quema != uint64(nacen)*constants.BurnRatio*constants.RupiaPerRupix {
							t.Fatalf("INVARIANTE 3 ROTA: nacen %d Diamantes con quema %d", nacen, quema)
						}
					} else if in[n-1]-out[n-1] != constants.BurnRatio*nacen {
						t.Fatalf("INVARIANTE 3 ROTA: nacen %d gemas de nivel %d quemando %d del inferior", nacen, n, in[n-1]-out[n-1])
					}
				}
			}
		}

		// Invariante 4: determinismo.
		tx2 := construirTxFuzz(inG, inD, inP, inR, inK, outG, outD, outP, outR, outK, quema, montoGema, tumba, extraNivel)
		if (v.checkLevelRules(tx2, povDaa, coinbase) == nil) != valida {
			t.Fatalf("checkLevelRules no es determinista")
		}
	})
}

func construirTxFuzz(inG, inD, inP, inR, inK, outG, outD, outP, outR, outK uint8, quema, montoGema uint64, tumba bool, extraNivel uint8) *externalapi.DomainTransaction {
	var in []*externalapi.DomainTransactionInput
	var out []*externalapi.DomainTransactionOutput
	add := func(n uint8, nivel uint16) {
		for i := uint8(0); i < n; i++ {
			if nivel == constants.LevelGold {
				in = append(in, gemInput(nivel, 1_000_000_000))
			} else {
				in = append(in, gemInput(nivel, constants.GemAmount))
			}
		}
	}
	add(inG, constants.LevelGold)
	add(inD, constants.LevelDiamante)
	add(inP, constants.LevelPlatino)
	add(inR, constants.LevelRodio)
	add(inK, constants.LevelKings)
	addOut := func(n uint8, nivel uint16) {
		for i := uint8(0); i < n; i++ {
			if nivel == constants.LevelGold {
				out = append(out, gemOutput(nivel, 1_000))
			} else {
				out = append(out, gemOutput(nivel, montoGema))
			}
		}
	}
	addOut(outG, constants.LevelGold)
	addOut(outD, constants.LevelDiamante)
	addOut(outP, constants.LevelPlatino)
	addOut(outR, constants.LevelRodio)
	addOut(outK, constants.LevelKings)
	if quema > 0 {
		out = append(out, burnOutput(quema))
	}
	if tumba {
		out = append(out, &externalapi.DomainTransactionOutput{Value: constants.GemAmount,
			ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{0x6a}, Version: constants.LevelDiamante}})
	}
	if extraNivel > uint8(constants.LevelKings) {
		out = append(out, gemOutput(uint16(extraNivel), constants.GemAmount))
	}
	return makeTx(in, out)
}

// oraculoEscalera: las reglas de la seccion 3 de ESPECIFICACION.md, escritas de nuevo.
// Devuelve si la transaccion es valida. No comparte codigo con checkLevelRules.
func oraculoEscalera(v *transactionValidator, tx *externalapi.DomainTransaction, povDaa uint64, coinbase bool,
	inG, inD, inP, inR, inK, outD, outP, outR, outK uint8, quema, montoGema uint64, tumba bool, extraNivel uint8) bool {
	// Un nivel que no existe, una gema a una tumba, una coinbase con gema: nunca.
	if extraNivel > uint8(constants.LevelKings) {
		return false
	}
	if tumba {
		return false
	}
	hayGemaOut := outD+outP+outR+outK > 0
	if coinbase && hayGemaOut {
		return false
	}
	// Toda gema es una pieza entera.
	if hayGemaOut && montoGema != constants.GemAmount {
		return false
	}
	// Burn por transaccion: al menos BurnBase + BurnPerByte * bytes, en Gold OpReturn.
	if !coinbase && quema < v.burnBase+v.burnPerByte*txmass.TransactionSerializedSize(tx) {
		return false
	}
	in := [5]int{int(inG), int(inD), int(inP), int(inR), int(inK)}
	out := [5]int{0, int(outD), int(outP), int(outR), int(outK)}
	for n := int(constants.LevelDiamante); n <= int(constants.LevelKings); n++ {
		nacen := out[n] - in[n]
		switch {
		case nacen > 0:
			// Ascenso: nivel desbloqueado y quema exacta del inferior.
			if povDaa < constants.LevelUnlockDaaScore(uint16(n), v.blocksPerHalving) {
				return false
			}
			if n == int(constants.LevelDiamante) {
				if quema != uint64(nacen)*constants.BurnRatio*constants.RupiaPerRupix {
					return false
				}
			} else if in[n-1]-out[n-1] != constants.BurnRatio*nacen {
				return false
			}
		case nacen < 0:
			// Desaparecen: solo como quema exacta de un ascenso al nivel de arriba.
			if n == int(constants.LevelKings) {
				return false
			}
			arriba := out[n+1] - in[n+1]
			if arriba <= 0 || -nacen != constants.BurnRatio*arriba {
				return false
			}
		}
	}
	return true
}
