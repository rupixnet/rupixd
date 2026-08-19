package rpchandlers

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/rupixnet/rupixd/app/rpc/rpccontext"
	"github.com/rupixnet/rupixd/infrastructure/network/netadapter/router"
)

// HandleGetFeeEstimate handles the respectively named RPC command.
//
// Rupix: estimacion honesta para el estado actual de la red — sin congestion,
// la fee minima estandar (1 rupia por gramo de masa) basta en cualquier
// prioridad. Cuando la red tenga trafico real, este handler puede evolucionar
// a leer la mempool y estimar por percentiles, como hace Kaspa moderno.
// Nota: la fee NO incluye el burn por transaccion (BurnBase + BurnPerByte),
// que es un output OpReturn aparte y responsabilidad de quien arma la tx.
func HandleGetFeeEstimate(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	response := appmessage.NewGetFeeEstimateResponseMessage()
	response.Estimate = appmessage.RPCFeeEstimate{
		PriorityBucket: appmessage.RPCFeeRateBucket{Feerate: 1, EstimatedSeconds: 1},
		NormalBuckets:  []appmessage.RPCFeeRateBucket{{Feerate: 1, EstimatedSeconds: 1}},
		LowBuckets:     []appmessage.RPCFeeRateBucket{{Feerate: 1, EstimatedSeconds: 1}},
	}
	return response, nil
}
