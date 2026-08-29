package server

import (
"context"

"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
)

func (s *server) TransferGem(_ context.Context, request *pb.TransferGemRequest) (*pb.TransferGemResponse, error) {
txIDs, err := s.transferGem(uint16(request.GetLevel()), request.GetToAddress(), request.GetPassword())
if err != nil {
return nil, err
}
return &pb.TransferGemResponse{TxIDs: txIDs}, nil
}
