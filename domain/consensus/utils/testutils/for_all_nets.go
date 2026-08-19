package testutils

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/dagconfig"
)

// ForAllNets runs the passed testFunc with all available networks
// if setDifficultyToMinumum = true - will modify the net params to have minimal difficulty, like in SimNet
func ForAllNets(t *testing.T, skipPow bool, testFunc func(*testing.T, *consensus.Config)) {
	allParams := []dagconfig.Params{
		dagconfig.MainnetParams,
		dagconfig.TestnetParams,
		dagconfig.SimnetParams,
		dagconfig.DevnetParams,
	}

	for _, params := range allParams {
		consensusConfig := consensus.Config{Params: params}
		t.Run(consensusConfig.Name, func(t *testing.T) {
			t.Parallel()
			consensusConfig.SkipProofOfWork = skipPow
			// Rupix: los tests de mecanica del DAG (reorgs, pruning, doble gasto...)
			// no prueban el burn por transaccion; sus txs son incidentales. Se apaga
			// aqui para las 4 redes. El burn tiene su arsenal propio con valores reales
			// de mainnet en transactionvalidator (TestLevelRules).
			consensusConfig.BurnBase = 0
			consensusConfig.BurnPerByte = 0
			t.Logf("Running test for %s", consensusConfig.Name)
			testFunc(t, &consensusConfig)
		})
	}
}
