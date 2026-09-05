package appmessage

// MsgPruningPointProof represents a kaspa PruningPointProof message
type MsgPruningPointProof struct {
	baseMessage

	Headers [][]*MsgBlockHeader
GemsHistory *MsgGemsHistory
}

// MsgGemsHistory (Rupix): conteo de gemas para el pruning proof.
type MsgGemsHistory struct {
	Diamante uint64
	Platino  uint64
	Rodio    uint64
Kings    uint64
}

// Command returns the protocol command string for the message
func (msg *MsgPruningPointProof) Command() MessageCommand {
	return CmdPruningPointProof
}

// NewMsgPruningPointProof returns a new MsgPruningPointProof.
func NewMsgPruningPointProof(headers [][]*MsgBlockHeader) *MsgPruningPointProof {
	return &MsgPruningPointProof{
		Headers: headers,
	}
}
