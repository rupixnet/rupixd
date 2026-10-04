package main

import (
	"fmt"

	"github.com/rupixnet/rupixd/cmd/rupixwallet/keys"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/librupixwallet"
)

func dumpUnencryptedData(conf *dumpUnencryptedDataConfig) error {
	// v0.6.3: este comando muestra la frase semilla. Nunca se salta la confirmacion
	// (no hay --yes) y la clave no viaja por la linea de comandos (no hay --password):
	// mismo criterio que en create, pedido por el auditor externo.
	fmt.Println(T("dump.aviso"))
	if !confirmar(T("dump.confirmar")) {
		fmt.Println(T("dump.cancelado"))
		return nil
	}

	keysFile, err := keys.ReadKeysFile(conf.NetParams(), conf.KeysFile)
	if err != nil {
		return err
	}

	password := keys.GetPassword(T("clave.prompt"))
	mnemonics, err := keysFile.DecryptMnemonics(password)
	if err != nil {
		return err
	}

	mnemonicPublicKeys := make(map[string]struct{})
	for i, mnemonic := range mnemonics {
		fmt.Printf(T("dump.frase"), i+1, mnemonic)
		publicKey, err := librupixwallet.MasterPublicKeyFromMnemonic(conf.NetParams(), mnemonic, len(keysFile.ExtendedPublicKeys) > 1)
		if err != nil {
			return err
		}

		mnemonicPublicKeys[publicKey] = struct{}{}
	}

	i := 1
	for _, extendedPublicKey := range keysFile.ExtendedPublicKeys {
		if _, exists := mnemonicPublicKeys[extendedPublicKey]; exists {
			continue
		}

		fmt.Printf("Extended Public key #%d:\n%s\n\n", i, extendedPublicKey)
		i++
	}

	fmt.Printf("Minimum number of signatures: %d\n", keysFile.MinimumSignatures)
	fmt.Println(T("dump.limpia"))
	return nil
}
