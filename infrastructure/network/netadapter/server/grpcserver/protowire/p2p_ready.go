package protowire

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/pkg/errors"
)

func (x *RupixdMessage_Ready) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RupixdMessage_Ready is nil")
	}
	return &appmessage.MsgReady{}, nil
}

func (x *RupixdMessage_Ready) fromAppMessage(_ *appmessage.MsgReady) error {
	return nil
}

