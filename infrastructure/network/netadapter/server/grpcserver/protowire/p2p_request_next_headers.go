package protowire

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/pkg/errors"
)

func (x *RupixdMessage_RequestNextHeaders) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RupixdMessage_RequestNextHeaders is nil")
	}
	return &appmessage.MsgRequestNextHeaders{}, nil
}

func (x *RupixdMessage_RequestNextHeaders) fromAppMessage(_ *appmessage.MsgRequestNextHeaders) error {
	return nil
}

