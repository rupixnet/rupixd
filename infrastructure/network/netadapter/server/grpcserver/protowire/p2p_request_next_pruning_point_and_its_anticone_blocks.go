package protowire

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/pkg/errors"
)

func (x *RupixdMessage_RequestNextPruningPointAndItsAnticoneBlocks) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RupixdMessage_DonePruningPointAndItsAnticoneBlocks is nil")
	}
	return &appmessage.MsgRequestNextPruningPointAndItsAnticoneBlocks{}, nil
}

func (x *RupixdMessage_RequestNextPruningPointAndItsAnticoneBlocks) fromAppMessage(_ *appmessage.MsgRequestNextPruningPointAndItsAnticoneBlocks) error {
	return nil
}

