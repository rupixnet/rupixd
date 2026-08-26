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

gemScript, err := txscript.PayToAddrScript(gemAddress)
if err != nil {
return nil, err
}
gemScript.Version = level

if level == constants.LevelDiamante {
return s.forgeDiamante(gemScript, changeAddress, password)
}
return s.forgeGemFromGems(level, gemScript, changeAddress, password)
}

func (s *server) forgeDiamante(gemScript *externalapi.ScriptPublicKey, changeAddress util.Address, password string) ([]string, error) {
feeRate := minFeeRate
maxFee := constants.RupiaPerRupix * uint64(20)

burnExacto := uint64(constants.BurnRatio) * uint64(constants.RupiaPerRupix)
necesario := burnExacto + uint64(constants.GemAmount) + uint64(constants.RupiaPerRupix)

selectedUTXOs, _, changeSompi, err := s.selectUTXOs(necesario, false, feeRate, maxFee, nil)
if err != nil {
return nil, err
}
if len(selectedUTXOs) == 0 {
return nil, errors.Errorf("sin Gold suficiente para forjar Diamante (requiere >= %d rupias)", necesario)
}

const feeMargin = 100_000
if changeSompi < uint64(constants.GemAmount)+feeMargin {
return nil, errors.Errorf("cambio insuficiente para gema + fee")
}
changeSompi -= uint64(constants.GemAmount) + feeMargin

payments := []*libkaspawallet.Payment{
{ScriptPublicKey: gemScript, Amount: uint64(constants.GemAmount)},
{ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{0x6a}, Version: 0}, Amount: burnExacto},
}
if changeSompi > 0 {
payments = append(payments, &libkaspawallet.Payment{Address: changeAddress, Amount: changeSompi})
}
return s.signAndBroadcastForge(payments, selectedUTXOs, password)
}

func (s *server) forgeGemFromGems(level uint16, gemScript *externalapi.ScriptPublicKey, changeAddress util.Address, password string) ([]string, error) {
inferior := level - 1

gemInputs := make([]*libkaspawallet.UTXO, 0, constants.BurnRatio)
for _, entry := range s.utxosSortedByAmount {
if entry.UTXOEntry.ScriptPublicKey().Version == inferior {
gemInputs = append(gemInputs, &libkaspawallet.UTXO{
Outpoint:       entry.Outpoint,
UTXOEntry:      entry.UTXOEntry,
DerivationPath: s.walletAddressPath(entry.address),
})
if len(gemInputs) == constants.BurnRatio {
break
}
}
}
if len(gemInputs) < constants.BurnRatio {
return nil, errors.Errorf("se requieren %d gemas de nivel %d para ascender; solo hay %d",
constants.BurnRatio, inferior, len(gemInputs))
}

feeRate := minFeeRate
maxFee := constants.RupiaPerRupix * uint64(20)
goldNecesario := uint64(2) * uint64(constants.RupiaPerRupix)
goldUTXOs, _, changeSompi, err := s.selectUTXOs(goldNecesario, false, feeRate, maxFee, nil)
if err != nil {
return nil, err
}
if len(goldUTXOs) == 0 {
return nil, errors.Errorf("sin Gold para cubrir el burn-por-tx del ascenso")
}

burnTx := uint64(constants.BurnBase) + uint64(constants.BurnPerByte)*uint64(2000)
const feeMargin = 100_000
if changeSompi < burnTx+feeMargin {
return nil, errors.Errorf("Gold insuficiente para burn-por-tx + fee del ascenso")
}
changeSompi -= burnTx + feeMargin

allUTXOs := append(gemInputs, goldUTXOs...)

payments := []*libkaspawallet.Payment{
{ScriptPublicKey: gemScript, Amount: uint64(constants.GemAmount)},
{ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{0x6a}, Version: 0}, Amount: burnTx},
}
if changeSompi > 0 {
payments = append(payments, &libkaspawallet.Payment{Address: changeAddress, Amount: changeSompi})
}
return s.signAndBroadcastForge(payments, allUTXOs, password)
}

func (s *server) signAndBroadcastForge(payments []*libkaspawallet.Payment, utxos []*libkaspawallet.UTXO, password string) ([]string, error) {
unsignedTx, err := libkaspawallet.CreateUnsignedTransaction(
s.keysFile.ExtendedPublicKeys, s.keysFile.MinimumSignatures, payments, utxos)
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
