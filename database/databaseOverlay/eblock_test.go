// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	. "github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestSaveLoadEBlockHead(t *testing.T) {
	b1, _ := testHelper.CreateTestEntryBlock(nil)

	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	err := dbo.SaveEBlock(b1, false)
	if err != nil {
		t.Error(err)
	}

	head, err := dbo.FetchEBlock(b1.DatabasePrimaryIndex())
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("EBlock head is nil")
	}

	m1, err := b1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err := head.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}

	b2, _ := testHelper.CreateTestEntryBlock(b1)

	err = dbo.SaveEBlock(b2, false)
	if err != nil {
		t.Error(err)
	}

	head, err = dbo.FetchEBlock(b2.DatabaseSecondaryIndex())
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("DBlock head is nil")
	}

	m1, err = b2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err = head.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}
}

func TestSaveLoadEBlockChain(t *testing.T) {
	blocks := []*EBlock{}
	max := 10
	var prev *EBlock = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev, _ = testHelper.CreateTestEntryBlock(prev)
		blocks = append(blocks, prev)
		err := dbo.SaveEBlock(prev, false)
		if err != nil {
			t.Error(err)
		}
	}

	current, err := dbo.FetchEBlock(prev.DatabasePrimaryIndex())
	if err != nil {
		t.Error(err)
	}
	zero := primitives.NewZeroHash()
	fetchedCount := 1
	for {
		keyMR := current.(*EBlock).GetHeader().GetPrevKeyMR()
		if keyMR.IsSameAs(zero) {
			break
		}
		//t.Logf("KeyMR - %v", keyMR.String())
		hash := current.GetHeader().GetPrevFullHash()

		current, err = dbo.FetchEBlockByPrimary(keyMR)
		if err != nil {
			t.Error(err)
		}
		if current == nil {
			t.Fatal("Block not found")
		}
		fetchedCount++

		byHash, err := dbo.FetchEBlockBySecondary(hash)

		same, err := primitives.AreBinaryMarshallablesEqual(current, byHash)
		if err != nil {
			t.Error(err)
		}
		if same == false {
			t.Error("Blocks fetched by keyMR and hash are not identical")
		}
	}
	if fetchedCount != max {
		t.Errorf("Wrong number of entries fetched - %v vs %v", fetchedCount, max)
	}
}

func TestLoadUnknownEBlocks(t *testing.T) {
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()
	for i := 0; i < 10; i++ {
		b := testHelper.IntToByteSlice(i)
		hash, err := primitives.NewShaHash(b)
		if err != nil {
			t.Error(err)
		}
		data, err := dbo.FetchEBlockByPrimary(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Error("Fetched entry while we expected nil - %v", data)
		}
		data, err = dbo.FetchEBlockBySecondary(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Error("Fetched entry while we expected nil - %v", data)
		}
		data, err = dbo.FetchEBlockHead(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Error("Fetched entry while we expected nil - %v", data)
		}
		all, err := dbo.FetchAllEBlocksByChain(hash)
		if err != nil {
			t.Error(err)
		}
		if len(all) != 0 {
			t.Error("Fetched entries while we expected nil - %v", all)
		}
	}
}

func TestFetchAllEBlockChainIDs(t *testing.T) {
	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()
	defer dbo.Close()

	chains, err := dbo.FetchAllEBlockChainIDs()
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(chains) != 2 {
		t.Errorf("Got wrong number of chains - %v", len(chains))
	}
}
