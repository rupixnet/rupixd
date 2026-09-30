package dagconfig

import "github.com/rupixnet/rupixd/domain/consensus/model/externalapi"

// mustHash (Rupix) convierte el hash hexadecimal de un checkpoint publicado.
// Si el texto no es un hash valido, el nodo no arranca: mejor fallar al
// arrancar que correr con un checkpoint mal escrito.
func mustHash(hexHash string) *externalapi.DomainHash {
	h, err := externalapi.NewDomainHashFromString(hexHash)
	if err != nil {
		panic("checkpoint con hash invalido: " + hexHash + ": " + err.Error())
	}
	return h
}

// testnetCheckpoints (Rupix): los checkpoints publicados de la testnet.
// Cada uno esta anunciado en CHECKPOINTS.md con red, blue score, hash, fecha y
// version. Regla (v0.6.1): todo bloque con blue score >= X + MergeDepth debe
// tener a H en su pasado. Caducan en TestnetParams.CheckpointsExpireDAAScore.
var testnetCheckpoints = []Checkpoint{
	// #1 — 29-sep-2026, testnet #4, v0.6.1. Punto de poda que el seed y el nodo de
	// la comunidad (JC) ya compartian, verificado desde los dos nodos por separado.
	{BlueScore: 86400, Hash: mustHash("7e2ece393c7d991c86e7ba915276cd85b5fc19e8647d5d197fa26bf116604fad")},
}
