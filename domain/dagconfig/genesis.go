// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"github.com/kaspanet/go-muhash"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/gemscommitment"
	"github.com/rupixnet/rupixd/domain/consensus/utils/blockheader"
	"github.com/rupixnet/rupixd/domain/consensus/utils/subnetworks"
	"github.com/rupixnet/rupixd/domain/consensus/utils/transactionhelper"
	"math/big"
)

// Genesis de Rupix: generado por cmd/genesisgen, verificable con
// `go run ./cmd/genesisgen` — los bytes deben coincidir exactamente.
// CERO PREMINE: el campo subsidy del payload es 0x00 x8 en las 4 redes.

var genesisTxOuts = []*externalapi.DomainTransactionOutput{}

// Payload: blue score (8) | subsidy = CERO (8) | script version (2) |
// varint (1) | OP-FALSE (1) | mensaje:
// "04/03/2026 - RUPIX IS ALIVE | We are all Rupix. We all build Rupix. | No confies, verifica. - E.R."
var genesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x30, 0x34, 0x2f, 0x30,
	0x33, 0x2f, 0x32, 0x30, 0x32, 0x36, 0x20, 0x2d,
	0x20, 0x52, 0x55, 0x50, 0x49, 0x58, 0x20, 0x49,
	0x53, 0x20, 0x41, 0x4c, 0x49, 0x56, 0x45, 0x20,
	0x7c, 0x20, 0x57, 0x65, 0x20, 0x61, 0x72, 0x65,
	0x20, 0x61, 0x6c, 0x6c, 0x20, 0x52, 0x75, 0x70,
	0x69, 0x78, 0x2e, 0x20, 0x57, 0x65, 0x20, 0x61,
	0x6c, 0x6c, 0x20, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x20, 0x52, 0x75, 0x70, 0x69, 0x78, 0x2e, 0x20,
	0x7c, 0x20, 0x4e, 0x6f, 0x20, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x65, 0x73, 0x2c, 0x20, 0x76, 0x65,
	0x72, 0x69, 0x66, 0x69, 0x63, 0x61, 0x2e, 0x20,
	0x2d, 0x20, 0x45, 0x2e, 0x52, 0x2e,
}

// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network.
var genesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(0, []*externalapi.DomainTransactionInput{}, genesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, genesisTxPayload)

// genesisHash is the hash of the first block in the block DAG for the main
// network (genesis block).
var genesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
		0xfc, 0xaa, 0x82, 0x82, 0x20, 0x0e, 0xfc, 0xd1,
		0x26, 0x87, 0x02, 0x93, 0x07, 0xd6, 0x79, 0x26,
		0xab, 0x6f, 0x5b, 0x3a, 0x64, 0xd0, 0x9a, 0xae,
		0xfa, 0xc8, 0xfe, 0xa0, 0x8f, 0xa6, 0x79, 0x00,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x9c, 0x72, 0x3b, 0x4e, 0x8a, 0x98, 0xad, 0x5b,
	0x0f, 0x1a, 0x86, 0xc3, 0xd5, 0x47, 0x8a, 0x9f,
	0x0b, 0x0c, 0xb0, 0x9a, 0x4e, 0x8f, 0xcd, 0x87,
	0x5f, 0x4b, 0x41, 0xba, 0xc0, 0xbc, 0x73, 0x3e,
})

// genesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the main network.
var genesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]externalapi.BlockLevelParents{},
		genesisMerkleRoot,
		&externalapi.DomainHash{},
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		gemscommitment.GenesisGemsCommitment(), // gemsCommitment: genesis con cero gemas (sello real, verificable)
		1772582400000, // 04/03/2026 00:00:00 UTC — la fecha del mensaje
		0x1e7fffff,
		0x281df,
		0,
		0,
		big.NewInt(0),
		&externalapi.DomainHash{},
	),
	Transactions: []*externalapi.DomainTransaction{genesisCoinbaseTx},
}

var devnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var devnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x72, 0x75, 0x70, 0x69,
	0x78, 0x2d, 0x64, 0x65, 0x76, 0x6e, 0x65, 0x74,
}

// devnetGenesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the development network.
var devnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(0,
	[]*externalapi.DomainTransactionInput{}, devnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, devnetGenesisTxPayload)

// devnetGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var devnetGenesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
		0xd8, 0x43, 0xf3, 0x44, 0x80, 0x75, 0xab, 0x23,
		0xf0, 0x06, 0xe6, 0xba, 0x58, 0x97, 0x9f, 0xb7,
		0x17, 0xda, 0x77, 0x3c, 0xff, 0xa6, 0x57, 0xf5,
		0xc2, 0xca, 0x5b, 0x21, 0x73, 0xe7, 0xcb, 0x9f,
})

// devnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x3c, 0xc3, 0x3e, 0x8b, 0x2c, 0xe7, 0x20, 0x50,
	0x3d, 0xfd, 0x96, 0x26, 0x7c, 0x6c, 0x0f, 0x6b,
	0x3b, 0x4c, 0x41, 0x65, 0xbd, 0x86, 0x2c, 0xb8,
	0xb2, 0xd8, 0x03, 0xca, 0xdc, 0x47, 0x07, 0x6b,
})

// devnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var devnetGenesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]externalapi.BlockLevelParents{},
		devnetGenesisMerkleRoot,
		&externalapi.DomainHash{},
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		gemscommitment.GenesisGemsCommitment(), // gemsCommitment: genesis con cero gemas (sello real, verificable)
		1772582400000, // 04/03/2026 00:00:00 UTC — la fecha del mensaje
		0x1f4ee5fb,
		0x113,
		0,
		0,
		big.NewInt(0),
		&externalapi.DomainHash{},
	),
	Transactions: []*externalapi.DomainTransaction{devnetGenesisCoinbaseTx},
}

var simnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var simnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x72, 0x75, 0x70, 0x69,
	0x78, 0x2d, 0x73, 0x69, 0x6d, 0x6e, 0x65, 0x74,
}

// simnetGenesisCoinbaseTx is the coinbase transaction for the simnet genesis block.
var simnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(0,
	[]*externalapi.DomainTransactionInput{}, simnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, simnetGenesisTxPayload)

// simnetGenesisHash is the hash of the first block in the block DAG for
// the simnet (genesis block).
var simnetGenesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
		0xf8, 0x80, 0xe5, 0x64, 0xe8, 0x3d, 0xe7, 0x92,
		0x24, 0xd3, 0xc5, 0x20, 0xc2, 0x6b, 0x80, 0xc6,
		0x95, 0xe2, 0x4d, 0xf8, 0x42, 0xd1, 0x77, 0x20,
		0xea, 0x0a, 0xba, 0x96, 0x46, 0xe8, 0x62, 0x8f,
})

// simnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x92, 0x5f, 0xd1, 0xe2, 0x3d, 0xc1, 0xb4, 0x64,
	0x95, 0x0b, 0x92, 0xdb, 0x52, 0x3e, 0x92, 0x17,
	0xcf, 0x90, 0xde, 0x7b, 0xd5, 0x8a, 0xe8, 0x6c,
	0xd2, 0xca, 0x32, 0xf6, 0x10, 0xda, 0x44, 0x31,
})

// simnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var simnetGenesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]externalapi.BlockLevelParents{},
		simnetGenesisMerkleRoot,
		&externalapi.DomainHash{},
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		gemscommitment.GenesisGemsCommitment(), // gemsCommitment: genesis con cero gemas (sello real, verificable)
		1772582400000, // 04/03/2026 00:00:00 UTC — la fecha del mensaje
		0x207fffff,
		0x0,
		0,
		0,
		big.NewInt(0),
		&externalapi.DomainHash{},
	),
	Transactions: []*externalapi.DomainTransaction{simnetGenesisCoinbaseTx},
}

var testnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var testnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x72, 0x75, 0x70, 0x69,
	0x78, 0x2d, 0x74, 0x65, 0x73, 0x74, 0x6e, 0x65,
	0x74,
}

// testnetGenesisCoinbaseTx is the coinbase transaction for the testnet genesis block.
var testnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(0,
	[]*externalapi.DomainTransactionInput{}, testnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, testnetGenesisTxPayload)

// testnetGenesisHash is the hash of the first block in the block DAG for the test
// network (genesis block).
var testnetGenesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
		0xc6, 0x75, 0x0f, 0x74, 0xec, 0xcc, 0x6c, 0x71,
		0xa4, 0x90, 0xf7, 0xd9, 0x09, 0x8d, 0x16, 0x0f,
		0xae, 0x93, 0xd6, 0x8e, 0xc1, 0x29, 0xdd, 0xb5,
		0x51, 0x56, 0x48, 0xe9, 0x1b, 0xc9, 0xc9, 0x55,
})

// testnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0xbe, 0xb5, 0x8d, 0x89, 0x08, 0x27, 0x04, 0xd9,
	0x9a, 0xb6, 0xe9, 0x8a, 0x67, 0xf5, 0xad, 0x8f,
	0x15, 0x14, 0x71, 0x58, 0x13, 0x36, 0x8d, 0x4c,
	0x64, 0x4e, 0xd1, 0x6a, 0x7a, 0xe6, 0xc6, 0xbc,
})

// testnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for testnet.
var testnetGenesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]externalapi.BlockLevelParents{},
		testnetGenesisMerkleRoot,
		&externalapi.DomainHash{},
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		gemscommitment.GenesisGemsCommitment(), // gemsCommitment: genesis con cero gemas (sello real, verificable)
		1772582400000, // 04/03/2026 00:00:00 UTC — la fecha del mensaje
		0x1e7fffff,
		0x163e0,
		0,
		0,
		big.NewInt(0),
		&externalapi.DomainHash{},
	),
	Transactions: []*externalapi.DomainTransaction{testnetGenesisCoinbaseTx},
}
