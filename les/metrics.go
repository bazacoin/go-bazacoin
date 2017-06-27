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

package les

import (
	"github.com/bazacoin/go-bazacoin/metrics"
	"github.com/bazacoin/go-bazacoin/p2p"
)

var (
	/*	propTxnInPacketsMeter     = metrics.NewMeter("bzc/prop/txns/in/packets")
		propTxnInTrafficMeter     = metrics.NewMeter("bzc/prop/txns/in/traffic")
		propTxnOutPacketsMeter    = metrics.NewMeter("bzc/prop/txns/out/packets")
		propTxnOutTrafficMeter    = metrics.NewMeter("bzc/prop/txns/out/traffic")
		propHashInPacketsMeter    = metrics.NewMeter("bzc/prop/hashes/in/packets")
		propHashInTrafficMeter    = metrics.NewMeter("bzc/prop/hashes/in/traffic")
		propHashOutPacketsMeter   = metrics.NewMeter("bzc/prop/hashes/out/packets")
		propHashOutTrafficMeter   = metrics.NewMeter("bzc/prop/hashes/out/traffic")
		propBlockInPacketsMeter   = metrics.NewMeter("bzc/prop/blocks/in/packets")
		propBlockInTrafficMeter   = metrics.NewMeter("bzc/prop/blocks/in/traffic")
		propBlockOutPacketsMeter  = metrics.NewMeter("bzc/prop/blocks/out/packets")
		propBlockOutTrafficMeter  = metrics.NewMeter("bzc/prop/blocks/out/traffic")
		reqHashInPacketsMeter     = metrics.NewMeter("bzc/req/hashes/in/packets")
		reqHashInTrafficMeter     = metrics.NewMeter("bzc/req/hashes/in/traffic")
		reqHashOutPacketsMeter    = metrics.NewMeter("bzc/req/hashes/out/packets")
		reqHashOutTrafficMeter    = metrics.NewMeter("bzc/req/hashes/out/traffic")
		reqBlockInPacketsMeter    = metrics.NewMeter("bzc/req/blocks/in/packets")
		reqBlockInTrafficMeter    = metrics.NewMeter("bzc/req/blocks/in/traffic")
		reqBlockOutPacketsMeter   = metrics.NewMeter("bzc/req/blocks/out/packets")
		reqBlockOutTrafficMeter   = metrics.NewMeter("bzc/req/blocks/out/traffic")
		reqHeaderInPacketsMeter   = metrics.NewMeter("bzc/req/headers/in/packets")
		reqHeaderInTrafficMeter   = metrics.NewMeter("bzc/req/headers/in/traffic")
		reqHeaderOutPacketsMeter  = metrics.NewMeter("bzc/req/headers/out/packets")
		reqHeaderOutTrafficMeter  = metrics.NewMeter("bzc/req/headers/out/traffic")
		reqBodyInPacketsMeter     = metrics.NewMeter("bzc/req/bodies/in/packets")
		reqBodyInTrafficMeter     = metrics.NewMeter("bzc/req/bodies/in/traffic")
		reqBodyOutPacketsMeter    = metrics.NewMeter("bzc/req/bodies/out/packets")
		reqBodyOutTrafficMeter    = metrics.NewMeter("bzc/req/bodies/out/traffic")
		reqStateInPacketsMeter    = metrics.NewMeter("bzc/req/states/in/packets")
		reqStateInTrafficMeter    = metrics.NewMeter("bzc/req/states/in/traffic")
		reqStateOutPacketsMeter   = metrics.NewMeter("bzc/req/states/out/packets")
		reqStateOutTrafficMeter   = metrics.NewMeter("bzc/req/states/out/traffic")
		reqReceiptInPacketsMeter  = metrics.NewMeter("bzc/req/receipts/in/packets")
		reqReceiptInTrafficMeter  = metrics.NewMeter("bzc/req/receipts/in/traffic")
		reqReceiptOutPacketsMeter = metrics.NewMeter("bzc/req/receipts/out/packets")
		reqReceiptOutTrafficMeter = metrics.NewMeter("bzc/req/receipts/out/traffic")*/
	miscInPacketsMeter  = metrics.NewMeter("les/misc/in/packets")
	miscInTrafficMeter  = metrics.NewMeter("les/misc/in/traffic")
	miscOutPacketsMeter = metrics.NewMeter("les/misc/out/packets")
	miscOutTrafficMeter = metrics.NewMeter("les/misc/out/traffic")
)

// meteredMsgReadWriter is a wrapper around a p2p.MsgReadWriter, capable of
// accumulating the above defined metrics based on the data stream contents.
type meteredMsgReadWriter struct {
	p2p.MsgReadWriter     // Wrapped message stream to meter
	version           int // Protocol version to select correct meters
}

// newMeteredMsgWriter wraps a p2p MsgReadWriter with metering support. If the
// metrics system is disabled, this function returns the original object.
func newMeteredMsgWriter(rw p2p.MsgReadWriter) p2p.MsgReadWriter {
	if !metrics.Enabled {
		return rw
	}
	return &meteredMsgReadWriter{MsgReadWriter: rw}
}

// Init sets the protocol version used by the stream to know which meters to
// increment in case of overlapping message ids between protocol versions.
func (rw *meteredMsgReadWriter) Init(version int) {
	rw.version = version
}

func (rw *meteredMsgReadWriter) ReadMsg() (p2p.Msg, error) {
	// Read the message and short circuit in case of an error
	msg, err := rw.MsgReadWriter.ReadMsg()
	if err != nil {
		return msg, err
	}
	// Account for the data traffic
	packets, traffic := miscInPacketsMeter, miscInTrafficMeter
	packets.Mark(1)
	traffic.Mark(int64(msg.Size))

	return msg, err
}

func (rw *meteredMsgReadWriter) WriteMsg(msg p2p.Msg) error {
	// Account for the data traffic
	packets, traffic := miscOutPacketsMeter, miscOutTrafficMeter
	packets.Mark(1)
	traffic.Mark(int64(msg.Size))

	// Send the packet to the p2p layer
	return rw.MsgReadWriter.WriteMsg(msg)
}
