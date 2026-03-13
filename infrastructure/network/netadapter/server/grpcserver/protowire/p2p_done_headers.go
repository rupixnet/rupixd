package protowire

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/pkg/errors"
)

func (x *RupixdMessage_DoneHeaders) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RupixdMessage_DoneHeaders is nil")
	}
	return &appmessage.MsgDoneHeaders{}, nil
}

func (x *RupixdMessage_DoneHeaders) fromAppMessage(_ *appmessage.MsgDoneHeaders) error {
	return nil
}

