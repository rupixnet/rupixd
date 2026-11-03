package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/client"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/utils"
	"github.com/rupixnet/rupixd/infrastructure/network/rpcclient"
)

// balance (Rupix, 3-nov-2026): la tarjeta con Gold, gemas y estado del nodo, en los colores
// de Rupix. Antes era una sola linea "Tienes X RUPIX"; las gemas habia que pedirlas aparte.
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
	t := tarjetaSaldo{Red: nombreRed(conf.NetParams().Name), Gold: response.Available, Pendiente: response.Pending}
	// Gemas: del daemon; si falla, la tarjeta sale sin ellas y se dice al final.
	gemasErr := error(nil)
	if g, err := daemonClient.Gems(ctx, &pb.GemsRequest{}); err == nil {
		t.Gemas = [5]uint64{0, g.Diamante, g.Platino, g.Rodio, g.Kings}
	} else {
		gemasErr = err
	}
	// Estado del nodo: consulta corta y opcional. Si el nodo no contesta en 5 s, la tarjeta
	// dice que no se pudo comprobar; nunca "al dia" por defecto.
	rpcServer := conf.RPCServer
	if rpcServer == "" {
		rpcServer = "127.0.0.1:" + conf.NetParams().RPCPort
	}
	if c, err := rpcclient.NewRPCClient(rpcServer); err == nil {
		c.SetTimeout(5 * time.Second)
		if info, err := c.GetInfo(); err == nil {
			if dag, err := c.GetBlockDAGInfo(); err == nil {
				t.NodoConsultado, t.Sincronizado, t.DAA = true, info.IsSynced, dag.VirtualDAAScore
			}
		}
		c.Close()
	}

	color := quiereColor(conf.SinColor)
	for _, l := range lineasTarjeta(t, color) {
		fmt.Println(l)
	}
	if gemasErr != nil {
		fmt.Println(pintar(colorAviso, fmt.Sprintf(T("tarjeta.gemas.sin"), gemasErr), color))
	}
	if conf.Verbose {
		fmt.Println()
		fmt.Println(T("saldo.cabecera"))
		fmt.Println("-----------------------------------------------------------------------------------------------------------")
		for _, addressBalance := range response.AddressBalances {
			fmt.Printf("%s %s %s\n", addressBalance.Address, utils.FormatRupix(addressBalance.Available), utils.FormatRupix(addressBalance.Pending))
		}
		fmt.Println("-----------------------------------------------------------------------------------------------------------")
	}
	return nil
}

// nombreRed convierte "rupix-testnet" en "testnet" para la tarjeta.
func nombreRed(nombreParams string) string {
	const prefijo = "rupix-"
	if len(nombreParams) > len(prefijo) && nombreParams[:len(prefijo)] == prefijo {
		return nombreParams[len(prefijo):]
	}
	return nombreParams
}
