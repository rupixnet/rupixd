package gemscommitment

import (
"encoding/binary"

"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/utils/hashes"
)

// CalculateGemsCommitment (Rupix) calcula el sello del conteo historico de gemas.
// DEBE ser identico en el minado (blockbuilder) y en la validacion (consensusstatemanager),
// por eso vive aqui, compartido. Serializa los 4 conteos (Diamante/Platino/Rodio/Kings)
// a 32 bytes big-endian, deterministas, y los hashea. Mismo conteo -> mismo sello, siempre.
func CalculateGemsCommitment(gemsHistory *externalapi.GemsHistory, kingsCount uint64) *externalapi.DomainHash {
diamante, platino, rodio := uint64(0), uint64(0), uint64(0)
if gemsHistory != nil {
diamante = gemsHistory.Diamante
platino = gemsHistory.Platino
rodio = gemsHistory.Rodio
}
b := make([]byte, 32)
binary.BigEndian.PutUint64(b[0:8], diamante)
binary.BigEndian.PutUint64(b[8:16], platino)
binary.BigEndian.PutUint64(b[16:24], rodio)
binary.BigEndian.PutUint64(b[24:32], kingsCount)
writer := hashes.NewBlockHashWriter()
writer.InfallibleWrite(b)
return writer.Finalize()
}

// GenesisGemsCommitment (Rupix) es el sello del genesis: cero gemas de todos los
// niveles. Se calcula con la misma funcion CalculateGemsCommitment, asi el genesis
// NO se salta ninguna validacion — su sello es correcto y verificable como cualquier bloque.
func GenesisGemsCommitment() *externalapi.DomainHash {
return CalculateGemsCommitment(nil, 0)
}
