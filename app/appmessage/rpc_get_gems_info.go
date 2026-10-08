package appmessage

// GetGemsInfoRequestMessage (Rupix) pide el estado de la escalera: cuantas gemas de
// cada nivel han nacido en toda la historia (lo que sella GemsCommitment), cuantas
// viven ahora en el UTXO set, y los topes.
type GetGemsInfoRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetGemsInfoRequestMessage) Command() MessageCommand {
	return CmdGetGemsInfoRequestMessage
}

// NewGetGemsInfoRequestMessage returns a instance of the message
func NewGetGemsInfoRequestMessage() *GetGemsInfoRequestMessage {
	return &GetGemsInfoRequestMessage{}
}

// GetGemsInfoResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetGemsInfoResponseMessage struct {
	baseMessage
	// Nacidas en toda la historia (del conteo sellado del virtual). Kings no se consumen,
	// asi que nacidos == vivos.
	DiamantesNacidos, PlatinosNacidos, RodiosNacidos, KingsNacidos uint64
	// Vivas ahora mismo en el UTXO set (del indice de UTXOs).
	DiamantesVivos, PlatinosVivos, RodiosVivos, KingsVivos uint64
	// Topes historicos.
	TopeDiamantes, TopePlatinos, TopeRodios, TopeKings uint64
	VirtualDAAScore                                    uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetGemsInfoResponseMessage) Command() MessageCommand {
	return CmdGetGemsInfoResponseMessage
}
