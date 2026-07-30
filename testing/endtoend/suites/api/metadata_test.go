// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"

	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/crypto"
	"github.com/theQRL/go-qrl/p2p"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertNodeMetadata(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()

	var modules map[string]string
	gomega.Expect(raw.CallContext(ctx, &modules, "rpc_modules")).To(gomega.Succeed())
	for _, namespace := range []string{"admin", "debug", "net", "qrl", "txpool", "web3"} {
		gomega.Expect(modules).To(gomega.HaveKey(namespace))
	}

	var clientVersion string
	gomega.Expect(raw.CallContext(ctx, &clientVersion, "web3_clientVersion")).To(gomega.Succeed())
	gomega.Expect(clientVersion).NotTo(gomega.BeEmpty())

	var digest hexutil.Bytes
	gomega.Expect(raw.CallContext(ctx, &digest, "web3_sha3", hexutil.Bytes("api"))).To(gomega.Succeed())
	gomega.Expect(digest).To(gomega.Equal(hexutil.Bytes(crypto.Keccak256([]byte("api")))))

	var networkVersion string
	gomega.Expect(raw.CallContext(ctx, &networkVersion, "net_version")).To(gomega.Succeed())
	gomega.Expect(networkVersion).To(gomega.Equal(suite.chainID.String()))

	var listening bool
	gomega.Expect(raw.CallContext(ctx, &listening, "net_listening")).To(gomega.Succeed())
	gomega.Expect(listening).To(gomega.BeTrue())

	var peerCount hexutil.Uint
	gomega.Expect(raw.CallContext(ctx, &peerCount, "net_peerCount")).To(gomega.Succeed())

	var nodeInfo p2p.NodeInfo
	gomega.Expect(raw.CallContext(ctx, &nodeInfo, "admin_nodeInfo")).To(gomega.Succeed())
	gomega.Expect(nodeInfo.ID).NotTo(gomega.BeEmpty())

	var peers []*p2p.PeerInfo
	gomega.Expect(raw.CallContext(ctx, &peers, "admin_peers")).To(gomega.Succeed())

	var datadir string
	gomega.Expect(raw.CallContext(ctx, &datadir, "admin_datadir")).To(gomega.Succeed())
	gomega.Expect(datadir).NotTo(gomega.BeEmpty())
}
