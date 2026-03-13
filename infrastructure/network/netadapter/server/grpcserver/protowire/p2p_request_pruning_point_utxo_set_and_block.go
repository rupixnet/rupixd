package protowire

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/pkg/errors"
)

func (x *RupixdMessage_RequestPruningPointUTXOSet) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RupixdMessage_RequestPruningPointUTXOSet is nil")
	}
	return x.RequestPruningPointUTXOSet.toAppMessage()
}

func (x *RequestPruningPointUTXOSetMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RequestPruningPointUTXOSetMessage is nil")
	}
	pruningPointHash, err := x.PruningPointHash.toDomain()
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgRequestPruningPointUTXOSet{PruningPointHash: pruningPointHash}, nil
}

func (x *RupixdMessage_RequestPruningPointUTXOSet) fromAppMessage(
	msgRequestPruningPointUTXOSet *appmessage.MsgRequestPruningPointUTXOSet) error {

	x.RequestPruningPointUTXOSet = &RequestPruningPointUTXOSetMessage{}
	x.RequestPruningPointUTXOSet.PruningPointHash = domainHashToProto(msgRequestPruningPointUTXOSet.PruningPointHash)
	return nil
}

