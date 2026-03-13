package protowire

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/pkg/errors"
)

func (x *RupixdMessage_RequestPruningPointProof) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RupixdMessage_RequestPruningPointProof is nil")
	}
	return &appmessage.MsgRequestPruningPointProof{}, nil
}

func (x *RupixdMessage_RequestPruningPointProof) fromAppMessage(_ *appmessage.MsgRequestPruningPointProof) error {
	return nil
}

