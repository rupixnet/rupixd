// genesisgen: genera los bloques genesis de Rupix para las 4 redes.
// Construye payload -> coinbase -> merkle root -> header, mina el nonce
// y emite los bytes listos para dagconfig/genesis.go.
// CERO PREMINE: subsidy = 0x00 x8, verificable byte por byte.
package main

import (
"fmt"
"math/big"

"github.com/kaspanet/go-muhash"
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/utils/blockheader"
"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
"github.com/rupixnet/rupixd/domain/consensus/utils/merkle"
"github.com/rupixnet/rupixd/domain/consensus/utils/pow"
"github.com/rupixnet/rupixd/domain/consensus/utils/subnetworks"
"github.com/rupixnet/rupixd/domain/consensus/utils/transactionhelper"
)

// 04/03/2026 00:00:00 UTC — la fecha del mensaje, en el reloj del bloque
const genesisTimestamp = int64(1772582400000)

const mensaje = "04/03/2026 - RUPIX IS ALIVE | We are all Rupix. We all build Rupix. | No confies, verifica. - E.R."

func buildPayload(tail []byte) []byte {
p := []byte{
0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Subsidy = 0 (CERO PREMINE)
0x00, 0x00, // Script version
0x01, // Varint
0x00, // OP-FALSE
}
return append(p, tail...)
}

type netSpec struct {
name string
tail []byte
bits uint32
}

func main() {
nets := []netSpec{
{"MAINNET", []byte(mensaje), 0x1e7fffff},
{"TESTNET", []byte("rupix-testnet"), 0x1e7fffff},
{"SIMNET", []byte("rupix-simnet"), 0x207fffff},
{"DEVNET", []byte("rupix-devnet"), 525264379},
}

for _, n := range nets {
payload := buildPayload(n.tail)
tx := transactionhelper.NewSubnetworkTransaction(0,
[]*externalapi.DomainTransactionInput{},
[]*externalapi.DomainTransactionOutput{},
&subnetworks.SubnetworkIDCoinbase, 0, payload)

merkleRoot := merkle.CalculateHashMerkleRoot([]*externalapi.DomainTransaction{tx})

header := blockheader.NewImmutableBlockHeader(
0,                                  // version
[]externalapi.BlockLevelParents{},  // sin padres
merkleRoot,
&externalapi.DomainHash{},          // acceptedIDMerkleRoot vacio
externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()), // universo vacio
genesisTimestamp,
n.bits,
0,              // nonce inicial
0,              // DAA score 0: la historia empieza aqui
0,              // blue score
big.NewInt(0),  // blue work
&externalapi.DomainHash{}, // pruning point vacio
)

mutable := header.ToMutable()
state := pow.NewState(mutable)
for !state.CheckProofOfWork() {
state.IncrementNonce()
}
mutable.SetNonce(state.Nonce)
final := mutable.ToImmutable()
hash := consensushashing.HeaderHash(final)

fmt.Printf("\n// ================= %s =================\n", n.name)
fmt.Printf("// timestamp: %d | bits: 0x%x | nonce: 0x%x\n", genesisTimestamp, n.bits, state.Nonce)
printBytes("Payload", payload)
printBytes("MerkleRoot", merkleRoot.ByteSlice())
printBytes("Hash", hash.ByteSlice())
}
}

func printBytes(label string, b []byte) {
fmt.Printf("// --- %s (%d bytes) ---\n", label, len(b))
for i, v := range b {
if i%8 == 0 {
fmt.Printf("\t")
}
fmt.Printf("0x%02x, ", v)
if i%8 == 7 {
fmt.Println()
}
}
if len(b)%8 != 0 {
fmt.Println()
}
}
