package main

import (
"context"
"fmt"

"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/client"
"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
)

func forge(conf *forgeConfig) error {
daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
if err != nil {
return err
}
defer tearDown()

ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
defer cancel()

response, err := daemonClient.Forge(ctx, &pb.ForgeRequest{
Level:      conf.Level,
GemAddress: conf.GemAddress,
Password:   conf.Password,
})
if err != nil {
return err
}

nombres := map[uint32]string{1: "Diamante", 2: "Platino", 3: "Rodio", 4: "Kings"}
fmt.Printf("Ascenso forjado — gema %s creada.\n", nombres[conf.Level])
for _, txID := range response.TxIDs {
fmt.Printf("  tx: %s\n", txID)
}
return nil
}
