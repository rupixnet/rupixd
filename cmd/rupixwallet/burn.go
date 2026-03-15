package main

import (
	"context"
	"fmt"

	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/client"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/keys"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/librupixwallet"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/infrastructure/config"
	"github.com/pkg/errors"
)


type burnConfig struct {
	DaemonAddress string `long:"daemonaddress" short:"d" description:"Wallet daemon server to connect to"`
	KeysFile      string `long:"keys-file" short:"k" description:"Keys file location"`
	Password      string `long:"password" short:"p" description:"Wallet password"`
	FromLevel     uint8  `long:"from-level" short:"l" description:"Nivel origen: 1=Gold 2=Diamante 3=Platino 4=Rodio" required:"true"`
	config.NetworkFlags
}

func burn(cfg *burnConfig) error {
	if cfg.FromLevel < 1 || cfg.FromLevel > 4 {
		return errors.Errorf("from-level debe ser 1, 2, 3 o 4 (recibido: %d)", cfg.FromLevel)
	}

	toLevel := cfg.FromLevel + 1
	payload := []byte{constants.TxTypeBurn, cfg.FromLevel, toLevel}
	burnAddress := "rupixsim:qpdk00fpv7n7wfkkv7y5k7t2f7qmu8krj9049lmxt6mlg3r8j0wy6nkl5s0mu"
	amountSompi := uint64(10 * 100_000_000)

	levelNames := map[uint8]string{1: "Gold", 2: "Diamante", 3: "Platino", 4: "Rodio", 5: "Kings Rupix"}
	fmt.Printf("Quemando 10x %s (L%d) para obtener 1x %s (L%d)\n",
		levelNames[cfg.FromLevel], cfg.FromLevel,
		levelNames[toLevel], toLevel)

	keysFile, err := keys.ReadKeysFile(cfg.NetParams(), cfg.KeysFile)
	if err != nil {
		return err
	}

	daemonClient, tearDown, err := client.Connect(cfg.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	unsignedResp, err := daemonClient.CreateUnsignedTransactions(ctx,
		&pb.CreateUnsignedTransactionsRequest{
			Address: burnAddress,
			Amount:   amountSompi,
                Payload: payload,
            })
	if err != nil {
		return errors.Wrap(err, "error creando TX de quema")
	}

	if len(cfg.Password) == 0 {
		cfg.Password = keys.GetPassword("Password:")
	}

	mnemonics, err := keysFile.DecryptMnemonics(cfg.Password)
	if err != nil {
		return err
	}

	signedTransactions := make([][]byte, len(unsignedResp.UnsignedTransactions))
	for i, unsigned := range unsignedResp.UnsignedTransactions {
		signed, err := librupixwallet.Sign(cfg.NetParams(), mnemonics, unsigned, keysFile.ECDSA)
		if err != nil {
			return err
		}
		signedTransactions[i] = signed
	}

	broadcastCtx, broadcastCancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer broadcastCancel()

	response, err := daemonClient.Broadcast(broadcastCtx,
		&pb.BroadcastRequest{Transactions: signedTransactions})
	if err != nil {
		return errors.Wrap(err, "error en broadcast de quema")
	}

	fmt.Printf("\nQuema exitosa! L%d -> L%d\n", cfg.FromLevel, toLevel)
	fmt.Println("Transaction ID(s):")
	for _, txID := range response.TxIDs {
		fmt.Printf("\t%s\n", txID)
	}
	fmt.Printf("Tu token %s (L%d) aparecera tras la confirmacion.\n",
		levelNames[toLevel], toLevel)

	return nil
}
