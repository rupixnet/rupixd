package protowire

import (
	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/app/appmessage"
)

// Rupix: GetGemsInfo. Cada acceso a un puntero se comprueba antes (leccion del 3-oct:
// un conversor que derreferencia un campo opcional nil puede tirar un nodo).

func (x *KaspadMessage_GetGemsInfoRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetGemsInfoRequestMessage{}, nil
}

func (x *KaspadMessage_GetGemsInfoRequest) fromAppMessage(_ *appmessage.GetGemsInfoRequestMessage) error {
	x.GetGemsInfoRequest = &GetGemsInfoRequestMessage{}
	return nil
}

func (x *KaspadMessage_GetGemsInfoResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetGemsInfoResponse is nil")
	}
	return x.GetGemsInfoResponse.toAppMessage()
}

func (x *KaspadMessage_GetGemsInfoResponse) fromAppMessage(message *appmessage.GetGemsInfoResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetGemsInfoResponse = &GetGemsInfoResponseMessage{
		DiamantesNacidos: message.DiamantesNacidos, PlatinosNacidos: message.PlatinosNacidos,
		RodiosNacidos: message.RodiosNacidos, KingsNacidos: message.KingsNacidos,
		DiamantesVivos: message.DiamantesVivos, PlatinosVivos: message.PlatinosVivos,
		RodiosVivos: message.RodiosVivos, KingsVivos: message.KingsVivos,
		TopeDiamantes: message.TopeDiamantes, TopePlatinos: message.TopePlatinos,
		TopeRodios: message.TopeRodios, TopeKings: message.TopeKings,
		VirtualDaaScore: message.VirtualDAAScore,
		Error:           err,
	}
	return nil
}

func (x *GetGemsInfoResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetGemsInfoResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.GetGemsInfoResponseMessage{
		DiamantesNacidos: x.DiamantesNacidos, PlatinosNacidos: x.PlatinosNacidos,
		RodiosNacidos: x.RodiosNacidos, KingsNacidos: x.KingsNacidos,
		DiamantesVivos: x.DiamantesVivos, PlatinosVivos: x.PlatinosVivos,
		RodiosVivos: x.RodiosVivos, KingsVivos: x.KingsVivos,
		TopeDiamantes: x.TopeDiamantes, TopePlatinos: x.TopePlatinos,
		TopeRodios: x.TopeRodios, TopeKings: x.TopeKings,
		VirtualDAAScore: x.VirtualDaaScore,
		Error:           rpcErr,
	}, nil
}
