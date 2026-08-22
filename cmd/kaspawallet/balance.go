package main

import (
	"context"
	"fmt"

	"github.com/rupixnet/rupixd/cmd/kaspawallet/daemon/client"
	"github.com/rupixnet/rupixd/cmd/kaspawallet/daemon/pb"
	"github.com/rupixnet/rupixd/cmd/kaspawallet/utils"
)

func balance(conf *balanceConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()
	response, err := daemonClient.GetBalance(ctx, &pb.GetBalanceRequest{})
	if err != nil {
		return err
	}

	pendingSuffix := ""
	if response.Pending > 0 {
		pendingSuffix = " (pending)"
	}
	if conf.Verbose {
		pendingSuffix = ""
		println("Address                                                                       Available             Pending")
		println("-----------------------------------------------------------------------------------------------------------")
		for _, addressBalance := range response.AddressBalances {
			fmt.Printf("%s %s %s\n", addressBalance.Address, utils.FormatRupix(addressBalance.Available), utils.FormatRupix(addressBalance.Pending))
		}
		println("-----------------------------------------------------------------------------------------------------------")
		print("                                                 ")
	}
	fmt.Printf("Total balance, RUPIX %s %s%s\n", utils.FormatRupix(response.Available), utils.FormatRupix(response.Pending), pendingSuffix)

	return nil
}
