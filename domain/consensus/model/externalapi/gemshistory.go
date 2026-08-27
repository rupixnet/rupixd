package externalapi

// GemsHistory guarda, por bloque, cuantas gemas de cada nivel han NACIDO
// en toda la historia de la cadena hasta ese bloque (inclusive). Son topes
// de EMISION historica: solo suben, jamas bajan. Kings tiene su propio store
// (kingscount) por razones historicas; aqui viven Diamante, Platino y Rodio.
type GemsHistory struct {
Diamante uint64
Platino  uint64
Rodio    uint64
}

// Clone devuelve una copia del GemsHistory.
func (gh *GemsHistory) Clone() *GemsHistory {
if gh == nil {
return nil
}
return &GemsHistory{
Diamante: gh.Diamante,
Platino:  gh.Platino,
Rodio:    gh.Rodio,
}
}
