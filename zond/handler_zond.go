// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package zond

import (
	"fmt"

	"github.com/theQRL/go-zond/core"
	"github.com/theQRL/go-zond/p2p/enode"
	"github.com/theQRL/go-zond/zond/protocols/zond"
)

// zondHandler implements the zond.Backend interface to handle the various network
// packets that are sent as replies or broadcasts.
type zondHandler handler

func (h *zondHandler) Chain() *core.BlockChain { return h.chain }
func (h *zondHandler) TxPool() zond.TxPool     { return h.txpool }

// RunPeer is invoked when a peer joins on the `zond` protocol.
func (h *zondHandler) RunPeer(peer *zond.Peer, hand zond.Handler) error {
	return (*handler)(h).runZondPeer(peer, hand)
}

// PeerInfo retrieves all known `zond` information about a peer.
func (h *zondHandler) PeerInfo(id enode.ID) interface{} {
	if p := h.peers.peer(id.String()); p != nil {
		return p.info()
	}
	return nil
}

// AcceptTxs retrieves whether transaction processing is enabled on the node
// or if inbound transactions should simply be dropped.
func (h *zondHandler) AcceptTxs() bool {
	return h.synced.Load()
}

// Handle is invoked from a peer's message handler when it receives a new remote
// message that the handler couldn't consume and serve itself.
func (h *zondHandler) Handle(peer *zond.Peer, packet zond.Packet) error {
	// Consume any broadcasts and announces, forwarding the rest to the downloader
	switch packet := packet.(type) {
	case *zond.NewPooledTransactionHashesPacket:
		return h.txFetcher.Notify(peer.ID(), packet.Types, packet.Sizes, packet.Hashes)

	case *zond.TransactionsPacket:
		return h.txFetcher.Enqueue(peer.ID(), *packet, false)

	case *zond.PooledTransactionsResponse:
		return h.txFetcher.Enqueue(peer.ID(), *packet, true)

	default:
		return fmt.Errorf("unexpected zond packet type: %T", packet)
	}
}
