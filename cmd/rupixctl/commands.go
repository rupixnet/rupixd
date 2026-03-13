package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rupixnet/rupixd/infrastructure/network/netadapter/server/grpcserver/protowire"
)

var commandTypes = []reflect.Type{
	reflect.TypeOf(protowire.RupixdMessage_AddPeerRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetConnectedPeerInfoRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetPeerAddressesRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetCurrentNetworkRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetInfoRequest{}),

	reflect.TypeOf(protowire.RupixdMessage_GetBlockRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetBlocksRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetHeadersRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetBlockCountRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetBlockDagInfoRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetSelectedTipHashRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetVirtualSelectedParentBlueScoreRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetVirtualSelectedParentChainFromBlockRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_ResolveFinalityConflictRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_EstimateNetworkHashesPerSecondRequest{}),

	reflect.TypeOf(protowire.RupixdMessage_GetBlockTemplateRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_SubmitBlockRequest{}),

	reflect.TypeOf(protowire.RupixdMessage_GetMempoolEntryRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetMempoolEntriesRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetMempoolEntriesByAddressesRequest{}),

	reflect.TypeOf(protowire.RupixdMessage_SubmitTransactionRequest{}),

	reflect.TypeOf(protowire.RupixdMessage_GetUtxosByAddressesRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetBalanceByAddressRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_GetCoinSupplyRequest{}),

	reflect.TypeOf(protowire.RupixdMessage_BanRequest{}),
	reflect.TypeOf(protowire.RupixdMessage_UnbanRequest{}),
}

type commandDescription struct {
	name       string
	parameters []*parameterDescription
	typeof     reflect.Type
}

type parameterDescription struct {
	name   string
	typeof reflect.Type
}

func commandDescriptions() []*commandDescription {
	commandDescriptions := make([]*commandDescription, len(commandTypes))

	for i, commandTypeWrapped := range commandTypes {
		commandType := unwrapCommandType(commandTypeWrapped)

		name := strings.TrimSuffix(commandType.Name(), "RequestMessage")
		numFields := commandType.NumField()

		var parameters []*parameterDescription
		for i := 0; i < numFields; i++ {
			field := commandType.Field(i)

			if !isFieldExported(field) {
				continue
			}

			parameters = append(parameters, &parameterDescription{
				name:   field.Name,
				typeof: field.Type,
			})
		}
		commandDescriptions[i] = &commandDescription{
			name:       name,
			parameters: parameters,
			typeof:     commandTypeWrapped,
		}
	}

	return commandDescriptions
}

func (cd *commandDescription) help() string {
	sb := &strings.Builder{}
	sb.WriteString(cd.name)
	for _, parameter := range cd.parameters {
		_, _ = fmt.Fprintf(sb, " [%s]", parameter.name)
	}
	return sb.String()
}

