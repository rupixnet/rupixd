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
// tener a H en su pasado. Caducan todos juntos en TestnetParams.CheckpointsExpireDAAScore (una sola caducidad para la lista).
var testnetCheckpoints = []Checkpoint{
	// #1 — 29-sep-2026, testnet #4, v0.6.1. Punto de poda que el seed y el nodo de
	// la comunidad (JC) ya compartian, verificado desde los dos nodos por separado.
	{BlueScore: 86400, Hash: mustHash("7e2ece393c7d991c86e7ba915276cd85b5fc19e8647d5d197fa26bf116604fad")},
	// #2 — 7-oct-2026, testnet #4, v0.6.4. Punto de poda del seed (269,575 bloques encima al
	// publicarlo). Confirmacion externa: pendiente (ningun nodo ajeno encendido ese dia);
	// se anota en CHECKPOINTS.md cuando un nodo externo reporte el mismo pruningPointHash.
	{BlueScore: 691200, Hash: mustHash("ef79c0b3382489f8d4859608a86fe05411e60978e30fa9f3350272f26f8b9669")},
}
