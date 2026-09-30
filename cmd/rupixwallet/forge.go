package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/client"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/keys"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/dagconfig"
	"github.com/rupixnet/rupixd/infrastructure/network/rpcclient"
)

var nombresNivel = map[uint32]string{0: "Gold", 1: "Diamante", 2: "Platino", 3: "Rodio", 4: "Kings"}

// forge (Rupix, v0.6.2 "wallet para principiantes"): antes de mandar nada a la red,
// la wallet revisa lo que hay contra lo que hace falta, dice en que bloque va la red
// y en cual se abre el nivel, resume lo que va a pasar y pide confirmacion. La forja
// quema para siempre; no debe ocurrir por un dedo de mas. Con --yes no pregunta.
func forge(conf *forgeConfig) error {
	if conf.Level < 1 || conf.Level > 4 {
		return errors.Errorf("--level debe ser 1 (Diamante), 2 (Platino), 3 (Rodio) o 4 (Kings)")
	}
	nombre := nombresNivel[conf.Level]

	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	// Direccion de la gema: si no se da, la primera de la wallet.
	if conf.GemAddress == "" {
		addrs, err := daemonClient.ShowAddresses(ctx, &pb.ShowAddressesRequest{})
		if err != nil {
			return err
		}
		if len(addrs.Address) == 0 {
			return errors.New("la wallet no tiene direcciones todavia; crea una con new-address")
		}
		conf.GemAddress = addrs.Address[0]
	}

	// Revisar ANTES de mandar: gemas y Gold que hay, contra lo que pide la escalera.
	gemas, err := daemonClient.Gems(ctx, &pb.GemsRequest{})
	if err != nil {
		return err
	}
	saldo, err := daemonClient.GetBalance(ctx, &pb.GetBalanceRequest{})
	if err != nil {
		return err
	}
	tengo := map[uint32]uint64{1: gemas.Diamante, 2: gemas.Platino, 3: gemas.Rodio, 4: gemas.Kings}
	const rupix = uint64(constants.RupiaPerRupix)
	quema := uint64(constants.BurnRatio)
	var faltas []string
	if conf.Level == 1 {
		necesario := quema*rupix + rupix // 10 Gold que se queman + margen para comision y quema por tx
		if saldo.Available < necesario {
			faltas = append(faltas, fmt.Sprintf("Gold: tienes %s y necesitas al menos %s (10 que se queman + comision)",
				rupixTxt(saldo.Available), rupixTxt(necesario)))
		}
	} else {
		inferior := conf.Level - 1
		if tengo[inferior] < quema {
			faltas = append(faltas, fmt.Sprintf("%s: tienes %d y necesitas %d", nombresNivel[inferior], tengo[inferior], quema))
		}
		if saldo.Available < 2*rupix {
			faltas = append(faltas, fmt.Sprintf("Gold para la comision: tienes %s y necesitas al menos 2", rupixTxt(saldo.Available)))
		}
	}

	// Esta abierto el nivel? Se le pregunta al nodo. Si no responde, se avisa y la red decide.
	params := conf.NetParams()
	abreEn := uint64(conf.Level) * params.BlocksPerHalving
	daa, errNodo := daaDelNodo(conf.RPCServer, params)
	if errNodo != nil {
		fmt.Printf("Aviso: no pude preguntarle al nodo en que bloque va (%s). La red revisara el nivel de todos modos.\n", errNodo)
	} else if daa < abreEn {
		faltan := abreEn - daa
		faltas = append(faltas, fmt.Sprintf("el %s se abre en el bloque %s y la red va en el %s: faltan %s bloques (~%s)",
			nombre, miles(abreEn), miles(daa), miles(faltan), tiempoBloques(faltan)))
	}

	if len(faltas) > 0 {
		fmt.Println("Todavia no se puede forjar:")
		for _, f := range faltas {
			fmt.Printf("  - %s\n", f)
		}
		return errors.New("forja cancelada: no se mando nada a la red")
	}

	// Resumen y confirmacion: lo que se quema no vuelve.
	fmt.Printf("Vas a forjar 1 %s en %s\n", nombre, conf.GemAddress)
	if conf.Level == 1 {
		fmt.Printf("Se queman 10 Gold PARA SIEMPRE, mas una comision pequena.\n")
	} else {
		fmt.Printf("Se queman %d %s PARA SIEMPRE (y una comision pequena en Gold). Quedaran %d %s.\n",
			quema, nombresNivel[conf.Level-1], tengo[conf.Level-1]-quema, nombresNivel[conf.Level-1])
	}
	if errNodo == nil {
		fmt.Printf("La red va en el bloque %s; el %s esta abierto desde el %s.\n", miles(daa), nombre, miles(abreEn))
	}
	if !conf.Yes && !confirmar("Confirmas? Escribe 'si' para continuar: ") {
		fmt.Println("Cancelado. No se movio nada.")
		return nil
	}

	// La clave se pide con prompt: en la linea de comandos quedaba en el historial.
	if len(conf.Password) == 0 {
		conf.Password = keys.GetPassword("Password:")
	}

	forgeCtx, forgeCancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer forgeCancel()
	response, err := daemonClient.Forge(forgeCtx, &pb.ForgeRequest{
		Level:      conf.Level,
		GemAddress: conf.GemAddress,
		Password:   conf.Password,
	})
	if err != nil {
		return traducirErrorNodo(err)
	}

	fmt.Printf("Ascenso forjado: gema %s creada.\n", nombre)
	for _, txID := range response.TxIDs {
		fmt.Printf("  tx: %s\n", txID)
	}
	fmt.Printf("Verificalo desde cualquier nodo: rupixctl GetUtxosByAddresses %s (busca version = %d).\n", conf.GemAddress, conf.Level)
	return nil
}

// daaDelNodo pregunta al nodo en que bloque (DAA score) va la red.
func daaDelNodo(rpcServer string, params *dagconfig.Params) (uint64, error) {
	if rpcServer == "" {
		rpcServer = "127.0.0.1:" + params.RPCPort
	}
	c, err := rpcclient.NewRPCClient(rpcServer)
	if err != nil {
		return 0, err
	}
	defer c.Close()
	info, err := c.GetBlockDAGInfo()
	if err != nil {
		return 0, err
	}
	return info.VirtualDAAScore, nil
}
