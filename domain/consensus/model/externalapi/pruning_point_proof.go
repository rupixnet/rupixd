package externalapi

// PruningPointProof is the data structure holding the pruning point proof
type PruningPointProof struct {
	Headers [][]BlockHeader
// GemsHistory: conteo historico de gemas (Diamante/Platino/Rodio) hasta el
// pruning point. Rupix: el nodo nuevo verifica los topes sin bajar toda la historia.
GemsHistory *GemsHistory
}
