package dagconfig

// Rupix (10-nov-2026): el estado de la lista de checkpoints respecto al DAA actual, para que
// el nodo y `rupixwallet verificar` avisen ANTES de que caduque y digan la verdad cuando ya
// caduco. Nacio de un hueco real: la lista de v0.6.9 caducaba en el DAA 3,900,000 y el #4 no
// se publico a tiempo (el plazo dependia de que una persona abriera un chat; ahora lo dice
// el propio nodo, todos los dias).
//
// MargenAvisoCheckpoints: con menos de este margen (~5 dias a un bloque por segundo) el nodo
// avisa que hay que publicar el siguiente checkpoint y una version nueva.
const MargenAvisoCheckpoints uint64 = 450_000

// EstadoCheckpoints devuelve, para el DAA virtual actual:
//   - caducada: la lista ya no se aplica (daa >= CheckpointsExpireDAAScore);
//   - pronto:   todavia vale pero le queda menos de MargenAvisoCheckpoints;
//   - margen:   DAA que faltan para la caducidad (0 si ya caduco).
//
// Sin checkpoints en la red, o sin caducidad definida, no hay nada que avisar.
func (p *Params) EstadoCheckpoints(daa uint64) (caducada, pronto bool, margen uint64) {
	if len(p.Checkpoints) == 0 || p.CheckpointsExpireDAAScore == 0 {
		return false, false, 0
	}
	if daa >= p.CheckpointsExpireDAAScore {
		return true, false, 0
	}
	margen = p.CheckpointsExpireDAAScore - daa
	return false, margen < MargenAvisoCheckpoints, margen
}
