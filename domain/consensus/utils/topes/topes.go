// Package topes (Rupix, v0.6.2) lleva el conteo corriente de gemas durante la
// aceptacion de transacciones, para que una forja que rompa un tope historico quede
// NO ACEPTADA, igual que una transaccion que falla en contexto, en vez de invalidar al
// bloque honesto que la mergea. Antes (hasta v0.6.1) el tope se comprobaba DESPUES de
// aceptar, sobre el bloque entero: un bloque con la gema 2,100,001 mataba al bloque
// que lo mergeaba, dos mineros honestos forjando la ultima gema en bloques hermanos
// mataban al que los juntara, y una sola transaccion sobre el tope en el mempool
// dejaba a los mineros sin plantilla. Hallazgo del auditor externo, 30-sep-2026.
package topes

import (
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
)

// Conteo es el estado corriente: gemas NACIDAS en toda la historia por nivel
// (Diamante, Platino, Rodio; nunca bajan) y Kings VIVOS. Como Kings es el ultimo nivel,
// la escalera no permite consumirlos (checkLevelRules rechaza que un King desaparezca):
// en toda transaccion valida vivos == nacidos. La resta por entradas de Kings de abajo
// es defensiva (y si alguna vez se ejecuta, grita INVARIANTE ROTA). Corregido el
// 7-oct-2026 a pedido del auditor: antes decia "bajan al consumirse", y eso no pasa.
type Conteo struct {
	Diamante, Platino, Rodio, Kings uint64
}

// Desde arma el conteo corriente a partir de lo guardado para el padre.
func Desde(historia *externalapi.GemsHistory, kings uint64) *Conteo {
	c := &Conteo{Kings: kings}
	if historia != nil {
		c.Diamante, c.Platino, c.Rodio = historia.Diamante, historia.Platino, historia.Rodio
	}
	return c
}

// Cabe intenta aplicar la transaccion al conteo. Si algun nivel superaria su tope (o
// un King se consumiria sin existir), NO toca el conteo y devuelve false: la
// transaccion no se acepta. Si cabe, actualiza el conteo y devuelve true.
// Las entradas necesitan UTXOEntry poblado (asi llegan en la aceptacion y en el mempool).
func (c *Conteo) Cabe(tx *externalapi.DomainTransaction) bool {
	var inD, inP, inR, inK, outD, outP, outR, outK uint64
	for _, in := range tx.Inputs {
		if in.UTXOEntry == nil {
			continue
		}
		switch in.UTXOEntry.ScriptPublicKey().Version {
		case constants.LevelDiamante:
			inD++
		case constants.LevelPlatino:
			inP++
		case constants.LevelRodio:
			inR++
		case constants.LevelKings:
			inK++
		}
	}
	for _, out := range tx.Outputs {
		switch out.ScriptPublicKey.Version {
		case constants.LevelDiamante:
			outD++
		case constants.LevelPlatino:
			outP++
		case constants.LevelRodio:
			outR++
		case constants.LevelKings:
			outK++
		}
	}
	nacidos := func(out, in uint64) uint64 {
		if out > in {
			return out - in
		}
		return 0
	}
	d := c.Diamante + nacidos(outD, inD)
	p := c.Platino + nacidos(outP, inP)
	r := c.Rodio + nacidos(outR, inR)
	if inK > c.Kings+outK {
		// INVARIANTE: el UTXO set ya valido que esos Kings existen (las entradas
		// llegan con UTXOEntry). Si el conteo no los conoce, el conteo de Kings y el
		// UTXO set divergieron: es un bug, no un ataque, y tiene que gritar.
		log.Errorf("INVARIANTE ROTA: la transaccion consume %d Kings con un conteo de %d vivos (+%d que crea); "+
			"el conteo de Kings y el UTXO set divergieron", inK, c.Kings, outK)
		return false
	}
	k := c.Kings + outK - inK
	if d > constants.MaxDiamante || p > constants.MaxPlatino || r > constants.MaxRodio || k > constants.MaxKings {
		return false
	}
	c.Diamante, c.Platino, c.Rodio, c.Kings = d, p, r, k
	return true
}
