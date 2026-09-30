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
		return errors.Errorf("--level debe ser 1 (Diamante), 2 (Platino), 3 (Rodio) o 4 (Kings)")
	}
	// Rupix: resumen y confirmacion antes de la clave. La gema cambia de dueno; no vuelve sola.
	fmt.Printf("Vas a enviar 1 %s a %s. La gema deja de ser tuya.\n", nombresNivel[conf.Level], conf.ToAddress)
	if !conf.Yes && !confirmar("Confirmas? Escribe 'si' para continuar: ") {
		fmt.Println("Cancelado. No se movio nada.")
		return nil
	}

	// La clave se pide con prompt: en la linea de comandos quedaba en el historial.
	if len(conf.Password) == 0 {
		conf.Password = keys.GetPassword("Password:")
	}

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	response, err := daemonClient.TransferGem(ctx, &pb.TransferGemRequest{
		Level:     conf.Level,
		ToAddress: conf.ToAddress,
		Password:  conf.Password,
	})
	if err != nil {
		return traducirErrorNodo(err)
	}

	fmt.Printf("Gema %s transferida a %s\n", nombresNivel[conf.Level], conf.ToAddress)
	for _, txID := range response.TxIDs {
		fmt.Printf("  tx: %s\n", txID)
	}
	return nil
}
