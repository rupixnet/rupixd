package rpchandlers

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/rupixnet/rupixd/app/rpc/rpccontext"
	"github.com/rupixnet/rupixd/domain/consensus/processes/burnmanager"
	"github.com/rupixnet/rupixd/infrastructure/network/netadapter/router"
)

// HandleGetLevelByAddress maneja la consulta del nivel L1-L5 de una direccion
func HandleGetLevelByAddress(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getLevelRequest := request.(*appmessage.GetLevelByAddressRequestMessage)
	address := getLevelRequest.Address

	if address == "" {
		errorMessage := &appmessage.GetLevelByAddressResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Address is required")
		return errorMessage, nil
	}

	level, err := context.Domain.Consensus().GetAddressLevel(address)
	if err != nil {
		errorMessage := &appmessage.GetLevelByAddressResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Failed to get address level: %s", err)
		return errorMessage, nil
	}

	levelName := burnmanager.LevelName(level)

	response := &appmessage.GetLevelByAddressResponseMessage{
		Address:   address,
		Level:     uint32(level),
		LevelName: levelName,
	}
	return response, nil
}