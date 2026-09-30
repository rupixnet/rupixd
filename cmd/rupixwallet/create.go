package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/librupixwallet"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/librupixwallet/bip32"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/utils"

	"github.com/rupixnet/rupixd/cmd/rupixwallet/keys"
)

func create(conf *createConfig) error {
	var encryptedMnemonics []*keys.EncryptedMnemonic
	var signerExtendedPublicKeys []string
	var mnemonics []string
	var err error
	isMultisig := conf.NumPublicKeys > 1
	if !conf.Import {
		encryptedMnemonics, signerExtendedPublicKeys, mnemonics, err = keys.CreateMnemonicsRevealing(conf.NetParams(), conf.NumPrivateKeys, conf.Password, isMultisig)
	} else {
		encryptedMnemonics, signerExtendedPublicKeys, err = keys.ImportMnemonics(conf.NetParams(), conf.NumPrivateKeys, conf.Password, isMultisig)
	}
	if err != nil {
		return err
	}

	for i, extendedPublicKey := range signerExtendedPublicKeys {
		fmt.Printf("Extended public key of mnemonic #%d:\n%s\n\n", i+1, extendedPublicKey)
	}

	fmt.Printf("Notice the above is neither a secret key to your wallet " +
		"(your secret seed phrase is shown below, once the keys file is saved) " +
		"nor a wallet public address (use \"rupixwallet new-address\" to create and see one)\n\n")

	extendedPublicKeys := make([]string, conf.NumPrivateKeys, conf.NumPublicKeys)
	copy(extendedPublicKeys, signerExtendedPublicKeys)
	reader := bufio.NewReader(os.Stdin)
	for i := conf.NumPrivateKeys; i < conf.NumPublicKeys; i++ {
		fmt.Printf("Enter public key #%d here:\n", i+1)
		extendedPublicKey, err := utils.ReadLine(reader)
		if err != nil {
			return err
		}

		_, err = bip32.DeserializeExtendedKey(string(extendedPublicKey))
		if err != nil {
			return errors.Wrapf(err, "%s is invalid extended public key", string(extendedPublicKey))
		}

		fmt.Println()

		extendedPublicKeys = append(extendedPublicKeys, string(extendedPublicKey))
	}

	// For a read only wallet the cosigner index is 0
	cosignerIndex := uint32(0)
	if len(signerExtendedPublicKeys) > 0 {
		cosignerIndex, err = librupixwallet.MinimumCosignerIndex(signerExtendedPublicKeys, extendedPublicKeys)
		if err != nil {
			return err
		}
	}

	file := keys.File{
		Version:            keys.LastVersion,
		EncryptedMnemonics: encryptedMnemonics,
		ExtendedPublicKeys: extendedPublicKeys,
		MinimumSignatures:  conf.MinimumSignatures,
		CosignerIndex:      cosignerIndex,
		ECDSA:              conf.ECDSA,
	}

	err = file.SetPath(conf.NetParams(), conf.KeysFile, conf.Yes)
	if err != nil {
		return err
	}

	err = file.TryLock()
	if err != nil {
		return err
	}

	err = file.Save()
	if err != nil {
		return err
	}

	fmt.Printf("Wrote the keys into %s\n", file.Path())

	// Rupix: la frase semilla se muestra UNA vez, ya con el archivo guardado.
	// Es la unica forma de recuperar la wallet si se pierde el archivo o la
	// contrasena. Antes no se mostraba nunca al crear.
	if len(mnemonics) > 0 {
		fmt.Printf("\n========================================================\n")
		fmt.Printf("FRASE SEMILLA / SEED PHRASE — copiala en PAPEL, ahora.\n")
		fmt.Printf("Es la unica forma de recuperar esta wallet si pierdes el\n")
		fmt.Printf("archivo o la contrasena. No la guardes en el telefono ni\n")
		fmt.Printf("en el chat. Se muestra UNA sola vez.\n")
		fmt.Printf("========================================================\n")
		for i, mnemonic := range mnemonics {
			fmt.Printf("Frase #%d:\n%s\n\n", i+1, mnemonic)
		}
		fmt.Printf("Cuando la tengas en papel, limpia la pantalla (clear / cls).\n")
	}
	return nil
}
