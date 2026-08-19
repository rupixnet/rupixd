package protowire

import (
	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/app/appmessage"
)

func (x *KaspadMessage_GetFeeEstimateRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetFeeEstimateRequestMessage{}, nil
}

func (x *KaspadMessage_GetFeeEstimateRequest) fromAppMessage(_ *appmessage.GetFeeEstimateRequestMessage) error {
	return nil
}

func (x *KaspadMessage_GetFeeEstimateResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetFeeEstimateResponse is nil")
	}
	return x.GetFeeEstimateResponse.toAppMessage()
}

func (x *GetFeeEstimateResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetFeeEstimateResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	estimate, err := x.Estimate.toAppMessage()
	if err != nil {
		return nil, err
	}

	return &appmessage.GetFeeEstimateResponseMessage{
		Error:    rpcErr,
		Estimate: estimate,
	}, nil
}

func (x *RpcFeeEstimate) toAppMessage() (appmessage.RPCFeeEstimate, error) {
	if x == nil {
		return appmessage.RPCFeeEstimate{}, errors.Wrapf(errorNil, "RpcFeeEstimate is nil")
	}
	return appmessage.RPCFeeEstimate{
		PriorityBucket: appmessage.RPCFeeRateBucket{
			Feerate:          x.PriorityBucket.Feerate,
			EstimatedSeconds: x.PriorityBucket.EstimatedSeconds,
		},
		NormalBuckets: feeRateBucketsToAppMessage(x.NormalBuckets),
		LowBuckets:    feeRateBucketsToAppMessage(x.LowBuckets),
	}, nil
}

func feeRateBucketsToAppMessage(protoBuckets []*RpcFeerateBucket) []appmessage.RPCFeeRateBucket {
	appMsgBuckets := make([]appmessage.RPCFeeRateBucket, len(protoBuckets))
	for i, bucket := range protoBuckets {
		appMsgBuckets[i] = appmessage.RPCFeeRateBucket{
			Feerate:          bucket.Feerate,
			EstimatedSeconds: bucket.EstimatedSeconds,
		}
	}
	return appMsgBuckets
}

func (x *KaspadMessage_GetFeeEstimateResponse) fromAppMessage(message *appmessage.GetFeeEstimateResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}
	x.GetFeeEstimateResponse = &GetFeeEstimateResponseMessage{
		Estimate: &RpcFeeEstimate{
			PriorityBucket: feeRateBucketFromAppMessage(&message.Estimate.PriorityBucket),
			NormalBuckets:  feeRateBucketsFromAppMessage(message.Estimate.NormalBuckets),
			LowBuckets:     feeRateBucketsFromAppMessage(message.Estimate.LowBuckets),
		},
		Error: rpcErr,
	}
	return nil
}

func feeRateBucketFromAppMessage(bucket *appmessage.RPCFeeRateBucket) *RpcFeerateBucket {
	return &RpcFeerateBucket{
		Feerate:          bucket.Feerate,
		EstimatedSeconds: bucket.EstimatedSeconds,
	}
}

func feeRateBucketsFromAppMessage(buckets []appmessage.RPCFeeRateBucket) []*RpcFeerateBucket {
	protoBuckets := make([]*RpcFeerateBucket, len(buckets))
	for i := range buckets {
		protoBuckets[i] = feeRateBucketFromAppMessage(&buckets[i])
	}
	return protoBuckets
}
