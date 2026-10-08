package rpchandlers

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/rupixnet/rupixd/app/rpc/rpccontext"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/infrastructure/network/netadapter/router"
)

// HandleGetGemsInfo (Rupix) responde con el estado de la escalera: nacidas en toda la
// historia (el conteo que sella cada encabezado), vivas ahora (indice de UTXOs) y topes.
// Dos fuentes a proposito: si alguna vez discrepan donde no deben (Kings), eso se ve.
func HandleGetGemsInfo(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	if !context.Config.UTXOIndex {
		errorMessage := &appmessage.GetGemsInfoResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when rupixd is run without --utxoindex")
		return errorMessage, nil
	}
	historia, err := context.Domain.Consensus().GetVirtualGemsHistory()
	if err != nil {
		return nil, err
	}
	vivas, err := context.UTXOIndex.GetGemCounts()
	if err != nil {
		return nil, err
	}
	daa, err := context.Domain.Consensus().GetVirtualDAAScore()
	if err != nil {
		return nil, err
	}
	return &appmessage.GetGemsInfoResponseMessage{
		DiamantesNacidos: historia.Diamante, PlatinosNacidos: historia.Platino,
		RodiosNacidos: historia.Rodio, KingsNacidos: historia.Kings,
		DiamantesVivos: vivas[constants.LevelDiamante], PlatinosVivos: vivas[constants.LevelPlatino],
		RodiosVivos: vivas[constants.LevelRodio], KingsVivos: vivas[constants.LevelKings],
		TopeDiamantes: constants.MaxDiamante, TopePlatinos: constants.MaxPlatino,
		TopeRodios: constants.MaxRodio, TopeKings: constants.MaxKings,
		VirtualDAAScore: daa,
	}, nil
}
