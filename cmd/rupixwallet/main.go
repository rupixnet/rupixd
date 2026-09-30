package main

import (
	"os"

	"github.com/pkg/errors"
)

func main() {
	subCmd, config := parseCommandLine()
	elegirIdioma(langFlag)
	var err error
	switch subCmd {
	case languageSubCmd:
		err = language(config.(*languageConfig))
	case createSubCmd:
		err = create(config.(*createConfig))
	case balanceSubCmd:
		err = balance(config.(*balanceConfig))
	case sendSubCmd:
		err = send(config.(*sendConfig))
	case createUnsignedTransactionSubCmd:
		err = createUnsignedTransaction(config.(*createUnsignedTransactionConfig))
	case signSubCmd:
		err = sign(config.(*signConfig))
	case broadcastSubCmd:
		err = broadcast(config.(*broadcastConfig))
	case broadcastReplacementSubCmd:
		err = broadcastReplacement(config.(*broadcastConfig))
	case parseSubCmd:
		err = parse(config.(*parseConfig))
	case showAddressesSubCmd:
		err = showAddresses(config.(*showAddressesConfig))
	case newAddressSubCmd:
		err = newAddress(config.(*newAddressConfig))
	case forgeSubCmd:
		err = forge(config.(*forgeConfig))
	case gemsSubCmd:
		err = gems(config.(*gemsConfig))
	case transferGemSubCmd:
		err = transferGem(config.(*transferGemConfig))
	case dumpUnencryptedDataSubCmd:
		err = dumpUnencryptedData(config.(*dumpUnencryptedDataConfig))
	case startDaemonSubCmd:
		err = startDaemon(config.(*startDaemonConfig))
	case sweepSubCmd:
		err = sweep(config.(*sweepConfig))
	case versionSubCmd:
		showVersion()
	case getDaemonVersionSubCmd:
		err = getDaemonVersion(config.(*getDaemonVersionConfig))
	case bumpFeeSubCmd:
		err = bumpFee(config.(*bumpFeeConfig))
	case bumpFeeUnsignedSubCmd:
		err = bumpFeeUnsigned(config.(*bumpFeeUnsignedConfig))
	default:
		err = errors.Errorf("Unknown sub-command '%s'\n", subCmd)
	}

	if err != nil {
		// Rupix: el error mas comun de un principiante es que el daemon no esta corriendo.
		printErrorAndExit(traducirErrorDaemon(err, defaultListen, banderaRed()))
	}
}

// banderaRed devuelve --testnet/--devnet/--simnet si venia en la linea de comandos, para
// repetirla en los mensajes de ayuda (asi el comando sugerido se puede copiar tal cual).
func banderaRed() string {
	for _, a := range os.Args[1:] {
		if a == "--testnet" || a == "--devnet" || a == "--simnet" {
			return a
		}
	}
	return ""
}
