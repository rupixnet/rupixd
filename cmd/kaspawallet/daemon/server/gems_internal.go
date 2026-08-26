package server

import (
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
)

// gemCounts cuenta las gemas de la wallet por nivel (Version 1..4).
// Las gemas son posesiones indivisibles (monto 1), no saldo de Gold.
func (s *server) gemCounts() map[uint16]uint64 {
counts := make(map[uint16]uint64)
for _, entry := range s.utxosSortedByAmount {
ver := entry.UTXOEntry.ScriptPublicKey().Version
if ver >= constants.LevelDiamante && ver <= constants.LevelKings {
counts[ver]++
}
}
return counts
}
