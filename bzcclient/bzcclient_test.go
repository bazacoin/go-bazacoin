// Copyright 2016 The go-bazacoin Authors
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

package bzcclient

import "github.com/bazacoin/go-bazacoin"

// Verify that Client implements the bazacoin interfaces.
var (
	_ = bazacoin.ChainReader(&Client{})
	_ = bazacoin.TransactionReader(&Client{})
	_ = bazacoin.ChainStateReader(&Client{})
	_ = bazacoin.ChainSyncReader(&Client{})
	_ = bazacoin.ContractCaller(&Client{})
	_ = bazacoin.GasEstimator(&Client{})
	_ = bazacoin.GasPricer(&Client{})
	_ = bazacoin.LogFilterer(&Client{})
	_ = bazacoin.PendingStateReader(&Client{})
	// _ = bazacoin.PendingStateEventer(&Client{})
	_ = bazacoin.PendingContractCaller(&Client{})
)
