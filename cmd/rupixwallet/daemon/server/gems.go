package server

import (
"context"

"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
)

// Gems reporta el inventario de gemas de la wallet, por nivel.
func (s *server) Gems(_ context.Context, _ *pb.GemsRequest) (*pb.GemsResponse, error) {
s.lock.RLock()
defer s.lock.RUnlock()
counts := s.gemCounts()
return &pb.GemsResponse{
Diamante: counts[constants.LevelDiamante],
Platino:  counts[constants.LevelPlatino],
Rodio:    counts[constants.LevelRodio],
Kings:    counts[constants.LevelKings],
}, nil
}
