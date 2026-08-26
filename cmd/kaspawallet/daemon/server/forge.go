package server

import (
"context"

"github.com/rupixnet/rupixd/cmd/kaspawallet/daemon/pb"
)

// Forge es el endpoint RPC del ascenso de nivel: recibe nivel, direccion de
// la gema y password, y delega en forgeGem (que arma la tx, quema el Gold
// exacto, crea la gema, firma y transmite).
func (s *server) Forge(_ context.Context, request *pb.ForgeRequest) (*pb.ForgeResponse, error) {
txIDs, err := s.forgeGem(uint16(request.GetLevel()), request.GetGemAddress(), request.GetPassword())
if err != nil {
return nil, err
}
return &pb.ForgeResponse{TxIDs: txIDs}, nil
}
