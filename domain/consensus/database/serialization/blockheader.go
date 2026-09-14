package serialization

import (
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/blockheader"
	"github.com/pkg/errors"
	"math"
	"math/big"
)

// DomainBlockHeaderToDbBlockHeader converts BlockHeader to DbBlockHeader
func DomainBlockHeaderToDbBlockHeader(domainBlockHeader externalapi.BlockHeader) *DbBlockHeader {
	return &DbBlockHeader{
		Version:              uint32(domainBlockHeader.Version()),
		Parents:              DomainParentsToDbParents(domainBlockHeader.Parents()),
		HashMerkleRoot:       DomainHashToDbHash(domainBlockHeader.HashMerkleRoot()),
		AcceptedIDMerkleRoot: DomainHashToDbHash(domainBlockHeader.AcceptedIDMerkleRoot()),
		UtxoCommitment:       DomainHashToDbHash(domainBlockHeader.UTXOCommitment()),
		GemsCommitment:       DomainHashToDbHash(gemsCommitmentOrZero(domainBlockHeader.GemsCommitment())),
		TimeInMilliseconds:   domainBlockHeader.TimeInMilliseconds(),
		Bits:                 domainBlockHeader.Bits(),
		Nonce:                domainBlockHeader.Nonce(),
		DaaScore:             domainBlockHeader.DAAScore(),
		BlueScore:            domainBlockHeader.BlueScore(),
		BlueWork:             domainBlockHeader.BlueWork().Bytes(),
		PruningPoint:         DomainHashToDbHash(domainBlockHeader.PruningPoint()),
	}
}

// DbBlockHeaderToDomainBlockHeader converts DbBlockHeader to BlockHeader
func DbBlockHeaderToDomainBlockHeader(dbBlockHeader *DbBlockHeader) (externalapi.BlockHeader, error) {
	parents, err := DbParentsToDomainParents(dbBlockHeader.Parents)
	if err != nil {
		return nil, err
	}
	hashMerkleRoot, err := DbHashToDomainHash(dbBlockHeader.HashMerkleRoot)
	if err != nil {
		return nil, err
	}
	acceptedIDMerkleRoot, err := DbHashToDomainHash(dbBlockHeader.AcceptedIDMerkleRoot)
	if err != nil {
		return nil, err
	}
	utxoCommitment, err := DbHashToDomainHash(dbBlockHeader.UtxoCommitment)
	if err != nil {
		return nil, err
	}
gemsCommitment, err := DbHashToDomainHash(dbBlockHeader.GemsCommitment)
if err != nil {
return nil, err
}
	if dbBlockHeader.Version > math.MaxUint16 {
		return nil, errors.Errorf("Invalid header version - bigger then uint16")
	}

	pruningPoint, err := DbHashToDomainHash(dbBlockHeader.PruningPoint)
	if err != nil {
		return nil, err
	}

	return blockheader.NewImmutableBlockHeader(
		uint16(dbBlockHeader.Version),
		parents,
		hashMerkleRoot,
		acceptedIDMerkleRoot,
		utxoCommitment,
		gemsCommitment,
		dbBlockHeader.TimeInMilliseconds,
		dbBlockHeader.Bits,
		dbBlockHeader.Nonce,
		dbBlockHeader.DaaScore,
		dbBlockHeader.BlueScore,
		new(big.Int).SetBytes(dbBlockHeader.BlueWork),
		pruningPoint,
	), nil
}

// gemsCommitmentOrZero (Rupix) devuelve un hash de ceros si el gemsCommitment es
// nil, para no petar al serializar headers sin gems (creados a mano en tests, o
// cualquier camino que lo deje nil). Un nil = "cero gemas", nunca panic.
func gemsCommitmentOrZero(h *externalapi.DomainHash) *externalapi.DomainHash {
if h == nil {
return &externalapi.DomainHash{}
}
return h
}
