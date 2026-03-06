package multiset

import (
	"github.com/pkg/errors"
	"github.com/kaspanet/go-muhash"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
)

type multiset struct {
	ms *muhash.MuHash
}

func (m *multiset) Add(data []byte) {
	m.ms.Add(data)
}

func (m *multiset) Remove(data []byte) {
	m.ms.Remove(data)
}

func (m *multiset) Hash() *externalapi.DomainHash {
	finalizedHash := m.ms.Finalize()
	return externalapi.NewDomainHashFromByteArray(finalizedHash.AsArray())
}

func (m *multiset) Serialize() []byte {
	return m.ms.Serialize()[:]
}

func (m *multiset) Clone() model.Multiset {
	return &multiset{ms: m.ms.Clone()}
}

func FromBytes(multisetBytes []byte) (model.Multiset, error) {
	// Corregido: Usamos la constante de tamaÃ±o correcta de la librerÃ­a
	if len(multisetBytes) != muhash.SerializedMuHashSize {
		return nil, errors.Errorf("multiset bytes expected to be in length of %d but got %d",
			muhash.SerializedMuHashSize, len(multisetBytes))
	}
	
	var serialized muhash.SerializedMuHash
	copy(serialized[:], multisetBytes)
	ms, err := muhash.DeserializeMuHash(&serialized)
	if err != nil {
		return nil, err
	}

	return &multiset{ms: ms}, nil
}

func New() model.Multiset {
	return &multiset{ms: muhash.NewMuHash()}
}
