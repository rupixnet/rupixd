package utxoindex

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
	"github.com/rupixnet/rupixd/infrastructure/db/database/ldb"
)

// TestGemCountsSiguenAlIndice (Rupix, 14-oct-2026): el explorador mostro durante unos
// segundos "creados 23 · vivos 2 · quemados 21" tras forjar un Diamante: "creados" sale
// del conteo sellado (al instante) y "vivos" salia de una cache de 30 s del indice.
// Dos numeros de momentos distintos, y la resta mentia. Desde este test, GetGemCounts
// refleja cada Update del indice en el acto: suma al anadir, resta al quitar, y solo
// recorre el indice entero la primera vez o tras un Reset.
func TestGemCountsSiguenAlIndice(t *testing.T) {
	db, err := ldb.NewLevelDB(t.TempDir(), 8)
	if err != nil {
		t.Fatalf("leveldb: %v", err)
	}
	defer db.Close()
	ui := &UTXOIndex{store: newUTXOIndexStore(db)}
	if err := ui.store.initializeCirculatingSompiSupply(); err != nil {
		t.Fatalf("supply: %v", err)
	}

	gema := func(version uint16, n byte) (externalapi.DomainOutpoint, externalapi.UTXOEntry) {
		var bytes [externalapi.DomainHashSize]byte
		bytes[0], bytes[1] = byte(version), n
		id := externalapi.NewDomainTransactionIDFromByteArray(&bytes)
		spk := &externalapi.ScriptPublicKey{Script: []byte{0x20, n}, Version: version}
		return externalapi.DomainOutpoint{TransactionID: *id, Index: 0}, utxo.NewUTXOEntry(1, spk, false, 1)
	}
	aplicar := func(toAdd, toRemove map[externalapi.DomainOutpoint]externalapi.UTXOEntry) {
		t.Helper()
		diff, err := utxo.NewUTXODiffFromCollections(utxo.NewUTXOCollection(toAdd), utxo.NewUTXOCollection(toRemove))
		if err != nil {
			t.Fatalf("diff: %v", err)
		}
		if _, err := ui.Update(&externalapi.VirtualChangeSet{VirtualUTXODiff: diff}); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	esperar := func(paso string, want [5]uint64) {
		t.Helper()
		got, err := ui.GetGemCounts()
		if err != nil {
			t.Fatalf("%s: GetGemCounts: %v", paso, err)
		}
		if got != want {
			t.Fatalf("%s: conteo %v, esperaba %v", paso, got, want)
		}
	}

	// Antes de nada: vacio.
	esperar("vacio", [5]uint64{})

	// Nacen 10 Diamantes y 1 Gold suelto.
	add := map[externalapi.DomainOutpoint]externalapi.UTXOEntry{}
	for i := byte(0); i < 10; i++ {
		o, e := gema(1, i)
		add[o] = e
	}
	o0, e0 := gema(0, 99)
	add[o0] = e0
	aplicar(add, nil)
	esperar("10 diamantes", [5]uint64{1, 10, 0, 0, 0})

	// Suben a 1 Platino: se consumen los 10, nace 1. En el mismo instante del Update,
	// sin esperar ninguna cache.
	rm := map[externalapi.DomainOutpoint]externalapi.UTXOEntry{}
	for i := byte(0); i < 10; i++ {
		o, e := gema(1, i)
		rm[o] = e
	}
	oP, eP := gema(2, 0)
	aplicar(map[externalapi.DomainOutpoint]externalapi.UTXOEntry{oP: eP}, rm)
	esperar("1 platino", [5]uint64{1, 0, 1, 0, 0})

	// Un King llega y nunca se va.
	oK, eK := gema(4, 0)
	aplicar(map[externalapi.DomainOutpoint]externalapi.UTXOEntry{oK: eK}, nil)
	esperar("1 king", [5]uint64{1, 0, 1, 0, 1})

	// El conteo incremental coincide con el recorrido completo del indice.
	completo, err := ui.store.countGems()
	if err != nil {
		t.Fatalf("countGems: %v", err)
	}
	if completo != [5]uint64{1, 0, 1, 0, 1} {
		t.Fatalf("recorrido completo %v no coincide", completo)
	}
}
