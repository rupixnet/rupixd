package main

import (
	"context"
	"fmt"

	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/client"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
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
	fmt.Println(T("gemas.titulo"))
	fmt.Printf("  💎 %-9s: %d\n", nombreNivel(1), response.Diamante)
	fmt.Printf("  ⬜ %-9s: %d\n", nombreNivel(2), response.Platino)
	fmt.Printf("  ◼ %-9s: %d\n", nombreNivel(3), response.Rodio)
	fmt.Printf("  👑 %-9s: %d\n", nombreNivel(4), response.Kings)
	fmt.Printf(T("gemas.total")+"\n", total)
	return nil
}
