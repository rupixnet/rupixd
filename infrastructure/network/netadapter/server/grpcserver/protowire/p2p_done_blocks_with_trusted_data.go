package protowire

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/pkg/errors"
)

func (x *RupixdMessage_DoneBlocksWithTrustedData) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RupixdMessage_DoneBlocksWithTrustedData is nil")
	}
	return &appmessage.MsgDoneBlocksWithTrustedData{}, nil
}

func (x *RupixdMessage_DoneBlocksWithTrustedData) fromAppMessage(_ *appmessage.MsgDoneBlocksWithTrustedData) error {
	return nil
}

