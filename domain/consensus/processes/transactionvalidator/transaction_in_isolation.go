package transactionvalidator

import (
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/consensus/utils/subnetworks"
	"github.com/rupixnet/rupixd/domain/consensus/utils/transactionhelper"
	"github.com/pkg/errors"
        "github.com/rupixnet/rupixd/domain/consensus/processes/burnmanager" 
)

// ValidateTransactionInIsolation validates the parts of the transaction that can be validated context-free
func (v *transactionValidator) ValidateTransactionInIsolation(tx *externalapi.DomainTransaction, povDAAScore uint64) error {
	err := v.checkTransactionInputCount(tx)
	if err != nil {
		return err
	}
	err = v.checkTransactionAmountRanges(tx)
	if err != nil {
		return err
	}
	err = v.checkDuplicateTransactionInputs(tx)
	if err != nil {
		return err
	}
	err = v.checkCoinbaseInIsolation(tx)
	if err != nil {
		return err
	}
	err = v.checkGasInBuiltInOrNativeTransactions(tx)
	if err != nil {
		return err
	}
	err = v.checkSubnetworkRegistryTransaction(tx)
	if err != nil {
		return err
	}

	err = v.checkNativeTransactionPayload(tx)
	if err != nil {
		return err
	}

	// TODO: fill it with the node's subnetwork id.
	err = v.checkTransactionSubnetwork(tx, nil)
	if err != nil {
		return err
	}

	if tx.Version > constants.MaxTransactionVersion {
		return errors.Wrapf(ruleerrors.ErrTransactionVersionIsUnknown, "validation failed: unknown transaction version. ")
	}
	// RUPIX: validar TX de quema si es tipo TxTypeBurn
	if len(tx.Payload) >= 3 && tx.Payload[0] == constants.TxTypeBurn && tx.Payload[1] >= 1 && tx.Payload[2] >= 2 {
		bm := burnmanager.New()
		err = bm.ValidateBurnTransaction(tx)
		if err != nil {
			return err
		}
	}
	return nil
}
func (v *transactionValidator) checkTransactionInputCount(tx *externalapi.DomainTransaction) error {
	if !transactionhelper.IsCoinBase(tx) && len(tx.Inputs) == 0 {
		return errors.Wrapf(ruleerrors.ErrNoTxInputs, "transaction has no inputs")
	}
	// RUPIX: limite maximo de inputs anti-spam
	if uint64(len(tx.Inputs)) > constants.MaxTxInputs {
		return errors.Wrapf(ruleerrors.ErrNoTxInputs, "transaction has too many inputs: got %d max %d",
			len(tx.Inputs), constants.MaxTxInputs)
	}
	// RUPIX: limite maximo de outputs anti-spam
	if !transactionhelper.IsCoinBase(tx) && uint64(len(tx.Outputs)) > constants.MaxTxOutputs {
		return errors.Wrapf(ruleerrors.ErrNoTxInputs, "transaction has too many outputs: got %d max %d",
			len(tx.Outputs), constants.MaxTxOutputs)
	}
	// RUPIX: limite maximo de payload anti-spam
	if !transactionhelper.IsCoinBase(tx) && uint64(len(tx.Payload)) > constants.MaxTxPayloadSize {
		return errors.Wrapf(ruleerrors.ErrInvalidPayload, "transaction payload too large: got %d max %d",
			len(tx.Payload), constants.MaxTxPayloadSize)
	}
	return nil
}

func (v *transactionValidator) checkTransactionAmountRanges(tx *externalapi.DomainTransaction) error {
	// Ensure the transaction amounts are in range. Each transaction
	// output must not be negative or more than the max allowed per
	// transaction. Also, the total of all outputs must abide by the same
	// restrictions. All amounts in a transaction are in a unit value known
	// as a rupia. One rupix is a quantity of rupia as defined by the
	// RupiaPerRupix constant.
	var totalrupia uint64
	for _, txOut := range tx.Outputs {
		rupia := txOut.Value
		if rupia == 0 {
			return errors.Wrap(ruleerrors.ErrTxOutValueZero, "zero value outputs are forbidden")
		}

		if rupia > constants.MaxRupia {
			return errors.Wrapf(ruleerrors.ErrBadTxOutValue, "transaction output value of %d is "+
				"higher than max allowed value of %d", rupia, constants.MaxRupia)
		}

		// Binary arithmetic guarantees that any overflow is detected and reported.
		// This is impossible for rupix, but perhaps possible if an alt increases
		// the total money supply.
		newTotalrupia := totalrupia + rupia
		if newTotalrupia < totalrupia {
			return errors.Wrapf(ruleerrors.ErrBadTxOutValue, "total value of all transaction "+
				"outputs exceeds max allowed value of %d",
				constants.MaxRupia)
		}
		totalrupia = newTotalrupia
		if totalrupia > constants.MaxRupia {
			return errors.Wrapf(ruleerrors.ErrBadTxOutValue, "total value of all transaction "+
				"outputs is %d which is higher than max "+
				"allowed value of %d", totalrupia,
				constants.MaxRupia)
		}
	}

	return nil
}

func (v *transactionValidator) checkDuplicateTransactionInputs(tx *externalapi.DomainTransaction) error {
	existingTxOut := make(map[externalapi.DomainOutpoint]struct{})
	for _, txIn := range tx.Inputs {
		if _, exists := existingTxOut[txIn.PreviousOutpoint]; exists {
			return errors.Wrapf(ruleerrors.ErrDuplicateTxInputs, "transaction "+
				"contains duplicate inputs")
		}
		existingTxOut[txIn.PreviousOutpoint] = struct{}{}
	}
	return nil
}

func (v *transactionValidator) checkCoinbaseInIsolation(tx *externalapi.DomainTransaction) error {
	if !transactionhelper.IsCoinBase(tx) {
		return nil
	}

	// Coinbase payload length must not exceed the max length.
	payloadLen := len(tx.Payload)
	if uint64(payloadLen) > v.maxCoinbasePayloadLength {
		return errors.Wrapf(ruleerrors.ErrBadCoinbasePayloadLen, "coinbase transaction payload length "+
			"of %d is out of range (max: %d)",
			payloadLen, v.maxCoinbasePayloadLength)
	}

	if len(tx.Inputs) != 0 {
		return errors.Wrap(ruleerrors.ErrCoinbaseWithInputs, "coinbase has inputs")
	}

	outputsLimit := uint64(v.ghostdagK) + 2
	if uint64(len(tx.Outputs)) > outputsLimit {
		return errors.Wrapf(ruleerrors.ErrCoinbaseTooManyOutputs, "coinbase has too many outputs: got %d where the limit is %d", len(tx.Outputs), outputsLimit)
	}

	for i, output := range tx.Outputs {
		if len(output.ScriptPublicKey.Script) > int(v.coinbasePayloadScriptPublicKeyMaxLength) {
			return errors.Wrapf(ruleerrors.ErrCoinbaseTooLongScriptPublicKey, "coinbase output %d has a too long script public key", i)

		}
	}

	return nil
}

func (v *transactionValidator) checkGasInBuiltInOrNativeTransactions(tx *externalapi.DomainTransaction) error {
	// Transactions in native, registry and coinbase subnetworks must have Gas = 0
	if subnetworks.IsBuiltInOrNative(tx.SubnetworkID) && tx.Gas > 0 {
		return errors.Wrapf(ruleerrors.ErrInvalidGas, "transaction in the native or "+
			"registry subnetworks has gas > 0 ")
	}
	return nil
}

func (v *transactionValidator) checkSubnetworkRegistryTransaction(tx *externalapi.DomainTransaction) error {
	if tx.SubnetworkID != subnetworks.SubnetworkIDRegistry {
		return nil
	}

	if len(tx.Payload) != 8 {
		return errors.Wrapf(ruleerrors.ErrSubnetworkRegistry, "validation failed: subnetwork registry "+
			"tx has an invalid payload")
	}
	return nil
}

func (v *transactionValidator) checkNativeTransactionPayload(tx *externalapi.DomainTransaction) error {
	if tx.SubnetworkID == subnetworks.SubnetworkIDNative && len(tx.Payload) > 0 {
		return errors.Wrapf(ruleerrors.ErrInvalidPayload, "transaction in the native subnetwork "+
			"includes a payload")
	}
	return nil
}

func (v *transactionValidator) checkTransactionSubnetwork(tx *externalapi.DomainTransaction,
	localNodeSubnetworkID *externalapi.DomainSubnetworkID) error {
	if !v.enableNonNativeSubnetworks && tx.SubnetworkID != subnetworks.SubnetworkIDNative &&
		tx.SubnetworkID != subnetworks.SubnetworkIDCoinbase {
		return errors.Wrapf(ruleerrors.ErrSubnetworksDisabled, "transaction has non native or coinbase "+
			"subnetwork ID")
	}

	// If we are a partial node, only transactions on built in subnetworks
	// or our own subnetwork may have a payload
	isLocalNodeFull := localNodeSubnetworkID == nil
	shouldTxBeFull := subnetworks.IsBuiltIn(tx.SubnetworkID) || tx.SubnetworkID.Equal(localNodeSubnetworkID)
	if !isLocalNodeFull && !shouldTxBeFull && len(tx.Payload) > 0 {
		return errors.Wrapf(ruleerrors.ErrInvalidPayload,
			"transaction that was expected to be partial has a payload "+
				"with length > 0")
	}
	return nil
}



