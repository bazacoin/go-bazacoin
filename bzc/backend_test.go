// Copyright 2015 The go-bazacoin Authors
// This file is part of the go-bazacoin library.
//
// The go-bazacoin library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-bazacoin library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-bazacoin library. If not, see <http://www.gnu.org/licenses/>.

package bzc

import (
	"math/big"
	"testing"

	"github.com/bazacoin/go-bazacoin/common"
	"github.com/bazacoin/go-bazacoin/core"
	"github.com/bazacoin/go-bazacoin/core/types"
	"github.com/bazacoin/go-bazacoin/bzcdb"
	"github.com/bazacoin/go-bazacoin/params"
)

func TestMipmapUpgrade(t *testing.T) {
	db, _ := bzcdb.NewMemDatabase()
	addr := common.BytesToAddress([]byte("jeff"))
	genesis := new(core.Genesis).MustCommit(db)

	chain, receipts := core.GenerateChain(params.TestChainConfig, genesis, db, 10, func(i int, gen *core.BlockGen) {
		var receipts types.Receipts
		switch i {
		case 1:
			receipt := types.NewReceipt(nil, new(big.Int))
			receipt.Logs = []*types.Log{{Address: addr}}
			gen.AddUncheckedReceipt(receipt)
			receipts = types.Receipts{receipt}
		case 2:
			receipt := types.NewReceipt(nil, new(big.Int))
			receipt.Logs = []*types.Log{{Address: addr}}
			gen.AddUncheckedReceipt(receipt)
			receipts = types.Receipts{receipt}
		}

		// store the receipts
		err := core.WriteReceipts(db, receipts)
		if err != nil {
			t.Fatal(err)
		}
	})
	for i, block := range chain {
		core.WriteBlock(db, block)
		if err := core.WriteCanonicalHash(db, block.Hash(), block.NumberU64()); err != nil {
			t.Fatalf("failed to insert block number: %v", err)
		}
		if err := core.WriteHeadBlockHash(db, block.Hash()); err != nil {
			t.Fatalf("failed to insert block number: %v", err)
		}
		if err := core.WriteBlockReceipts(db, block.Hash(), block.NumberU64(), receipts[i]); err != nil {
			t.Fatal("error writing block receipts:", err)
		}
	}

	err := addMipmapBloomBins(db)
	if err != nil {
		t.Fatal(err)
	}

	bloom := core.GetMipmapBloom(db, 1, core.MIPMapLevels[0])
	if (bloom == types.Bloom{}) {
		t.Error("got empty bloom filter")
	}

	data, _ := db.Get([]byte("setting-mipmap-version"))
	if len(data) == 0 {
		t.Error("setting-mipmap-version not written to database")
	}
}
