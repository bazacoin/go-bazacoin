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

// Contains the metrics collected by the fetcher.

package fetcher

import (
	"github.com/bazacoin/go-bazacoin/metrics"
)

var (
	propAnnounceInMeter   = metrics.NewMeter("bzc/fetcher/prop/announces/in")
	propAnnounceOutTimer  = metrics.NewTimer("bzc/fetcher/prop/announces/out")
	propAnnounceDropMeter = metrics.NewMeter("bzc/fetcher/prop/announces/drop")
	propAnnounceDOSMeter  = metrics.NewMeter("bzc/fetcher/prop/announces/dos")

	propBroadcastInMeter   = metrics.NewMeter("bzc/fetcher/prop/broadcasts/in")
	propBroadcastOutTimer  = metrics.NewTimer("bzc/fetcher/prop/broadcasts/out")
	propBroadcastDropMeter = metrics.NewMeter("bzc/fetcher/prop/broadcasts/drop")
	propBroadcastDOSMeter  = metrics.NewMeter("bzc/fetcher/prop/broadcasts/dos")

	headerFetchMeter = metrics.NewMeter("bzc/fetcher/fetch/headers")
	bodyFetchMeter   = metrics.NewMeter("bzc/fetcher/fetch/bodies")

	headerFilterInMeter  = metrics.NewMeter("bzc/fetcher/filter/headers/in")
	headerFilterOutMeter = metrics.NewMeter("bzc/fetcher/filter/headers/out")
	bodyFilterInMeter    = metrics.NewMeter("bzc/fetcher/filter/bodies/in")
	bodyFilterOutMeter   = metrics.NewMeter("bzc/fetcher/filter/bodies/out")
)
