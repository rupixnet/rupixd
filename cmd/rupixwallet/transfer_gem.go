package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"

	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/client"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/keys"
)

func transferGem(conf *transferGemConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	if conf.Level < 1 || conf.Level > 4 {
		return errors.New(T("forjar.nivel_invalido"))
	}
	// Rupix: resumen y confirmacion antes de la clave. La gema cambia de dueno; no vuelve sola.
	fmt.Printf(T("gema.enviar")+"\n", nombreNivel(conf.Level), conf.ToAddress)
	if !conf.Yes && !confirmar(T("confirmar")) {
		fmt.Println(T("cancelado"))
		return nil
	}

	// La clave se pide siempre con prompt (v0.6.3: ya no existe --password).
	password := keys.GetPassword(T("clave.prompt"))

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	response, err := daemonClient.TransferGem(ctx, &pb.TransferGemRequest{
		Level:     conf.Level,
		ToAddress: conf.ToAddress,
		Password:  password,
	})
	if err != nil {
		return traducirErrorNodo(err)
	}

	fmt.Printf(T("gema.enviada")+"\n", nombreNivel(conf.Level), conf.ToAddress)
	for _, txID := range response.TxIDs {
		fmt.Printf("  tx: %s\n", txID)
	}
	return nil
}
