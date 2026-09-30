package main

import (
	"context"
	"fmt"

	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/client"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/utils"
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

	// Rupix: una linea clara para quien empieza; con -v, la tabla por direccion.
	pendingSuffix := ""
	if response.Pending > 0 {
		pendingSuffix = fmt.Sprintf(T("saldo.pendiente"), rupixTxt(response.Pending))
	}
	if conf.Verbose {
		println(T("saldo.cabecera"))
		println("-----------------------------------------------------------------------------------------------------------")
		for _, addressBalance := range response.AddressBalances {
			fmt.Printf("%s %s %s\n", addressBalance.Address, utils.FormatRupix(addressBalance.Available), utils.FormatRupix(addressBalance.Pending))
		}
		println("-----------------------------------------------------------------------------------------------------------")
	}
	fmt.Printf(T("saldo.total")+"%s\n", rupixTxt(response.Available), pendingSuffix)

	return nil
}
