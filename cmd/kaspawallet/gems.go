package main

import (
"context"
"fmt"

"github.com/rupixnet/rupixd/cmd/kaspawallet/daemon/client"
"github.com/rupixnet/rupixd/cmd/kaspawallet/daemon/pb"
)

func gems(conf *gemsConfig) error {
daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
if err != nil {
return err
}
defer tearDown()

ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
defer cancel()

response, err := daemonClient.Gems(ctx, &pb.GemsRequest{})
if err != nil {
return err
}

total := response.Diamante + response.Platino + response.Rodio + response.Kings
fmt.Println("La escalera — tus gemas:")
fmt.Printf("  💎 Diamante : %d\n", response.Diamante)
fmt.Printf("  ⬜ Platino  : %d\n", response.Platino)
fmt.Printf("  ◼ Rodio    : %d\n", response.Rodio)
fmt.Printf("  👑 Kings    : %d\n", response.Kings)
fmt.Printf("  Total: %d gema(s)\n", total)
return nil
}
