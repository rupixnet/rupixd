package server

import (
"github.com/pkg/errors"
"github.com/rupixnet/rupixd/cmd/rupixwallet/librupixwallet"
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
"github.com/rupixnet/rupixd/domain/consensus/utils/txscript"
"github.com/rupixnet/rupixd/util"
)

func (s *server) transferGem(level uint16, toAddressString string, password string) ([]string, error) {
s.lock.Lock()
defer s.lock.Unlock()

if !s.isSynced() {
return nil, errors.Errorf("wallet daemon no sincronizado: %s", s.formatSyncStateReport())
}
if level < constants.LevelDiamante || level > constants.LevelKings {
return nil, errors.Errorf("nivel %d invalido", level)
}

toAddress, err := util.DecodeAddress(toAddressString, s.params.Prefix)
if err != nil {
return nil, err
}
changeAddress, _, err := s.changeAddress(false, nil)
if err != nil {
return nil, err
}

var gemInput *librupixwallet.UTXO
for _, entry := range s.utxosSortedByAmount {
if entry.UTXOEntry.ScriptPublicKey().Version == level {
gemInput = &librupixwallet.UTXO{
Outpoint:       entry.Outpoint,
UTXOEntry:      entry.UTXOEntry,
DerivationPath: s.walletAddressPath(entry.address),
}
break
}
}
if gemInput == nil {
return nil, errors.Errorf("no tienes ninguna gema de nivel %d para transferir", level)
}

feeRate := minFeeRate
maxFee := constants.RupiaPerRupix * uint64(20)
goldNecesario := uint64(2) * uint64(constants.RupiaPerRupix)
goldUTXOs, _, changeSompi, err := s.selectUTXOs(goldNecesario, false, feeRate, maxFee, nil)
if err != nil {
return nil, err
}
if len(goldUTXOs) == 0 {
return nil, errors.Errorf("sin Gold para el burn-por-tx de la transferencia")
}

burnTx := uint64(constants.BurnBase) + uint64(constants.BurnPerByte)*uint64(2000)
const feeMargin = 100_000
if changeSompi < burnTx+feeMargin {
return nil, errors.Errorf("Gold insuficiente para burn-por-tx + fee")
}
changeSompi -= burnTx + feeMargin

gemScript, err := txscript.PayToAddrScript(toAddress)
if err != nil {
return nil, err
}
gemScript.Version = level

allUTXOs := append([]*librupixwallet.UTXO{gemInput}, goldUTXOs...)

payments := []*librupixwallet.Payment{
{ScriptPublicKey: gemScript, Amount: uint64(constants.GemAmount)},
{ScriptPublicKey: &externalapi.ScriptPublicKey{Script: []byte{0x6a}, Version: 0}, Amount: burnTx},
}
if changeSompi > 0 {
payments = append(payments, &librupixwallet.Payment{Address: changeAddress, Amount: changeSompi})
}

return s.signAndBroadcastForge(payments, allUTXOs, password)
}
