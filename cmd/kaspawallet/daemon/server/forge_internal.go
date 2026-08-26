package server

import (
"github.com/pkg/errors"
"github.com/rupixnet/rupixd/cmd/kaspawallet/libkaspawallet"
	"github.com/rupixnet/rupixd/cmd/kaspawallet/libkaspawallet/serialization"
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
"github.com/rupixnet/rupixd/domain/consensus/utils/txscript"
"github.com/rupixnet/rupixd/util"
)

func (s *server) forgeGem(level uint16, gemAddressString string, password string) ([]string, error) {
s.lock.Lock()
defer s.lock.Unlock()

if !s.isSynced() {
return nil, errors.Errorf("wallet daemon no sincronizado: %s", s.formatSyncStateReport())
}
if level < constants.LevelDiamante || level > constants.LevelKings {
return nil, errors.Errorf("nivel %d invalido: entre %d (Diamante) y %d (Kings)",
level, constants.LevelDiamante, constants.LevelKings)
}

gemAddress, err := util.DecodeAddress(gemAddressString, s.params.Prefix)
if err != nil {
return nil, err
}
changeAddress, _, err := s.changeAddress(false, nil)
if err != nil {
return nil, err
}

feeRate := minFeeRate
maxFee := constants.RupiaPerRupix * uint64(20)

burnExacto := uint64(constants.BurnRatio) * uint64(constants.RupiaPerRupix)
necesario := burnExacto + uint64(constants.GemAmount) + uint64(constants.RupiaPerRupix)

selectedUTXOs, _, changeSompi, err := s.selectUTXOs(necesario, false, feeRate, maxFee, nil)
if err != nil {
return nil, err
}
if len(selectedUTXOs) == 0 {
return nil, errors.Errorf("sin Gold suficiente para forjar (requiere >= %d rupias)", necesario)
}

const feeMargin = 100_000
if changeSompi < uint64(constants.GemAmount)+feeMargin {
return nil, errors.Errorf("cambio insuficiente para gema + fee del ascenso")
}
changeSompi -= uint64(constants.GemAmount) + feeMargin

gemScript, err := txscript.PayToAddrScript(gemAddress)
if err != nil {
return nil, err
}
gemScript.Version = level

payments := []*libkaspawallet.Payment{
{ScriptPublicKey: gemScript, Amount: uint64(constants.GemAmount)},
{ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{0x6a}, Version: 0}, Amount: burnExacto},
}
if changeSompi > 0 {
payments = append(payments, &libkaspawallet.Payment{Address: changeAddress, Amount: changeSompi})
}

unsignedTx, err := libkaspawallet.CreateUnsignedTransaction(
s.keysFile.ExtendedPublicKeys, s.keysFile.MinimumSignatures, payments, selectedUTXOs)
if err != nil {
return nil, err
}

unsignedTxBytes, err := serialization.SerializePartiallySignedTransaction(unsignedTx)
if err != nil {
return nil, err
}
signedTxs, err := s.signTransactions([][]byte{unsignedTxBytes}, password)
if err != nil {
return nil, err
}
txIDs, err := s.broadcast(signedTxs, false)
if err != nil {
return nil, errors.Wrapf(err, "error transmitiendo el ascenso")
}
return txIDs, nil
}
