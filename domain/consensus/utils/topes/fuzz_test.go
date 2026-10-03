package topes

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
)

// FuzzCabe (hueco #6 de ESPECIFICACION.md): entradas arbitrarias de conteo inicial y de
// gemas que entran y salen por nivel. Invariantes, en este orden de importancia:
//  1. Si Cabe devuelve true, ningun nivel queda sobre su tope. Nunca.
//  2. Si devuelve false, el conteo no se toco.
//  3. Lo que devuelve coincide con la aritmetica de la especificacion calculada aparte:
//     nacidos = max(0, out-in) por nivel para D/P/R; Kings vivos = k + outK - inK,
//     y consumir mas Kings de los que el conteo conoce no cabe.
//  4. Es determinista: la misma entrada da lo mismo dos veces.
//
// Correr: go test -fuzz=FuzzCabe -fuzztime=1h ./domain/consensus/utils/topes/
func FuzzCabe(f *testing.F) {
	f.Add(uint64(0), uint64(0), uint64(0), uint64(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0), uint8(0), uint8(0))
	f.Add(uint64(constants.MaxDiamante-1), uint64(0), uint64(0), uint64(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0), uint8(0), uint8(0))
	f.Add(uint64(constants.MaxDiamante), uint64(0), uint64(0), uint64(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0), uint8(0), uint8(0))
	f.Add(uint64(constants.MaxDiamante), uint64(constants.MaxPlatino), uint64(constants.MaxRodio), uint64(constants.MaxKings), uint8(10), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0), uint8(0))
	f.Add(uint64(0), uint64(0), uint64(0), uint64(constants.MaxKings), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0), uint8(0), uint8(0), uint8(1))
	f.Add(uint64(0), uint64(0), uint64(0), uint64(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0), uint8(0), uint8(0), uint8(0))

	f.Fuzz(func(t *testing.T, d, p, r, k uint64, inD, inP, inR, inK, outD, outP, outR, outK uint8) {
		// Conteos iniciales cerca de los topes son los interesantes; se recortan a
		// [0, tope+3] para que el fuzzer no gaste el tiempo en numeros absurdos.
		d %= constants.MaxDiamante + 4
		p %= constants.MaxPlatino + 4
		r %= constants.MaxRodio + 4
		k %= constants.MaxKings + 4

		tx := &externalapi.DomainTransaction{}
		add := func(n uint8, level uint16, in bool) {
			for i := uint8(0); i < n; i++ {
				if in {
					tx.Inputs = append(tx.Inputs, gemIn(level))
				} else {
					tx.Outputs = append(tx.Outputs, gemOut(level))
				}
			}
		}
		add(inD, constants.LevelDiamante, true)
		add(inP, constants.LevelPlatino, true)
		add(inR, constants.LevelRodio, true)
		add(inK, constants.LevelKings, true)
		add(outD, constants.LevelDiamante, false)
		add(outP, constants.LevelPlatino, false)
		add(outR, constants.LevelRodio, false)
		add(outK, constants.LevelKings, false)
		// Ruido que no debe contar: una entrada sin UTXOEntry y una salida Gold.
		tx.Inputs = append(tx.Inputs, &externalapi.DomainTransactionInput{})
		tx.Outputs = append(tx.Outputs, &externalapi.DomainTransactionOutput{Value: 5, ScriptPublicKey: &externalapi.ScriptPublicKey{Version: constants.LevelGold}})
		_ = utxo.NewUTXOEntry // (gemIn usa utxo.NewUTXOEntry; el import queda explicito)

		c := Desde(&externalapi.GemsHistory{Diamante: d, Platino: p, Rodio: r}, k)
		antes := *c
		ok := c.Cabe(tx)

		// Aritmetica independiente de la especificacion.
		nac := func(out, in uint8) uint64 {
			if out > in {
				return uint64(out - in)
			}
			return 0
		}
		ed, ep, er := d+nac(outD, inD), p+nac(outP, inP), r+nac(outR, inR)
		esperado := uint64(inK) <= k+uint64(outK)
		var ek uint64
		if esperado {
			ek = k + uint64(outK) - uint64(inK)
			esperado = ed <= constants.MaxDiamante && ep <= constants.MaxPlatino && er <= constants.MaxRodio && ek <= constants.MaxKings
		}
		if ok != esperado {
			t.Fatalf("Cabe=%v, la especificacion dice %v (inicio %+v, in D%d P%d R%d K%d, out D%d P%d R%d K%d)", ok, esperado, antes, inD, inP, inR, inK, outD, outP, outR, outK)
		}
		if ok {
			if c.Diamante > constants.MaxDiamante || c.Platino > constants.MaxPlatino || c.Rodio > constants.MaxRodio || c.Kings > constants.MaxKings {
				t.Fatalf("INVARIANTE 1 ROTA: Cabe dijo true y dejo el conteo sobre un tope: %+v", *c)
			}
			if c.Diamante != ed || c.Platino != ep || c.Rodio != er || c.Kings != ek {
				t.Fatalf("conteo tras caber %+v, esperado D%d P%d R%d K%d", *c, ed, ep, er, ek)
			}
		} else if *c != antes {
			t.Fatalf("INVARIANTE 2 ROTA: Cabe dijo false y toco el conteo: antes %+v, despues %+v", antes, *c)
		}
		// Determinismo.
		c2 := Desde(&externalapi.GemsHistory{Diamante: d, Platino: p, Rodio: r}, k)
		if c2.Cabe(tx) != ok {
			t.Fatalf("Cabe no es determinista")
		}
	})
}
