package main

import (
"context"
"fmt"

"github.com/rupixnet/rupixd/cmd/kaspawallet/daemon/client"
"github.com/rupixnet/rupixd/cmd/kaspawallet/daemon/pb"
)

func transferGem(conf *transferGemConfig) error {
daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
if err != nil {
return err
}
defer tearDown()

ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
defer cancel()

response, err := daemonClient.TransferGem(ctx, &pb.TransferGemRequest{
Level:     conf.Level,
ToAddress: conf.ToAddress,
Password:  conf.Password,
})
if err != nil {
return err
}

nombres := map[uint32]string{1: "Diamante", 2: "Platino", 3: "Rodio", 4: "Kings"}
fmt.Printf("Gema %s transferida a %s\n", nombres[conf.Level], conf.ToAddress)
for _, txID := range response.TxIDs {
fmt.Printf("  tx: %s\n", txID)
}
return nil
}
