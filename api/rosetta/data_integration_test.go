// Copyright 2021 Optakt Labs OÜ
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

// +build integration

package rosetta_test

import (
	"encoding/hex"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/model/flow"

	"github.com/optakt/flow-dps/api/rosetta"
	"github.com/optakt/flow-dps/models/dps"
	"github.com/optakt/flow-dps/rosetta/configuration"
	"github.com/optakt/flow-dps/rosetta/identifier"
	"github.com/optakt/flow-dps/rosetta/invoker"
	"github.com/optakt/flow-dps/rosetta/retriever"
	"github.com/optakt/flow-dps/rosetta/scripts"
	"github.com/optakt/flow-dps/rosetta/validator"
	"github.com/optakt/flow-dps/service/dictionaries"
	"github.com/optakt/flow-dps/service/index"
	"github.com/optakt/flow-dps/testing/snapshots"
)

const (
	invalidBlockchain = "invalid-blockchain"
	invalidNetwork    = "invalid-network"
	invalidToken      = "invalid-token"

	invalidBlockHash = "af528bb047d6cd1400a326bb127d689607a096f5ccd81d8903dfebbac26afb2z" // invalid hex value

	validBlockHashLen = 64
)

func setupDB(t *testing.T) *badger.DB {
	t.Helper()

	opts := badger.DefaultOptions("").
		WithInMemory(true).
		WithLogger(nil)

	db, err := badger.Open(opts)
	require.NoError(t, err)

	reader := hex.NewDecoder(strings.NewReader(snapshots.Rosetta))
	dict, _ := hex.DecodeString(dictionaries.Payload)

	decompressor, err := zstd.NewReader(reader,
		zstd.WithDecoderDicts(dict),
	)
	require.NoError(t, err)

	err = db.Load(decompressor, runtime.GOMAXPROCS(0))
	require.NoError(t, err)

	return db
}

func setupAPI(t *testing.T, db *badger.DB) *rosetta.Data {
	t.Helper()

	index := index.NewReader(db)

	params := dps.FlowParams[dps.FlowTestnet]
	config := configuration.New(params.ChainID)
	validate := validator.New(params, index)
	generate := scripts.NewGenerator(params)
	invoke, err := invoker.New(index)
	require.NoError(t, err)
	retrieve := retriever.New(params, index, validate, generate, invoke)
	controller := rosetta.NewData(config, retrieve)

	return controller
}

// defaultNetwork returns the Network identifier common for all requests.
func defaultNetwork() identifier.Network {
	return identifier.Network{
		Blockchain: dps.FlowBlockchain,
		Network:    dps.FlowTestnet.String(),
	}
}

// defaultCurrency returns the Currency spec common for all requests.
// At the moment only get the FLOW tokens, perhaps in the future it will support multiple.
func defaultCurrency() []identifier.Currency {
	return []identifier.Currency{
		{
			Symbol:   dps.FlowSymbol,
			Decimals: dps.FlowDecimals,
		},
	}
}

func validateBlock(t *testing.T, height uint64, hash string) validateBlockFunc {
	t.Helper()

	return func(blockID identifier.Block) {
		assert.Equal(t, height, blockID.Index)
		assert.Equal(t, hash, blockID.Hash)
	}
}

func validateByHeader(t *testing.T, header flow.Header) validateBlockFunc {
	return validateBlock(t, header.Height, header.ID().String())
}

// add header for block 165 and 181
func knownHeaders(height uint64) flow.Header {

	switch height {

	case 1:
		return flow.Header{
			ChainID:     dps.FlowTestnet,
			ParentID:    flow.Identifier{0xd4, 0x7b, 0x1b, 0xf7, 0xf3, 0x7e, 0x19, 0x2c, 0xf8, 0x3d, 0x2b, 0xee, 0x3f, 0x63, 0x32, 0xb0, 0xd9, 0xb1, 0x5c, 0xa, 0xa7, 0x66, 0xd, 0x1e, 0x53, 0x22, 0xea, 0x96, 0x46, 0x67, 0xb3, 0x33},
			Height:      1,
			PayloadHash: flow.Identifier{0x7b, 0x3b, 0x31, 0x3b, 0xd8, 0x3e, 0x1, 0xd1, 0x3c, 0x44, 0x9d, 0x4d, 0xd4, 0xba, 0xc0, 0x41, 0x37, 0xf5, 0x9, 0xb, 0xcb, 0x30, 0x5d, 0xdd, 0x75, 0x2, 0x98, 0xbd, 0x16, 0xe5, 0x33, 0x9b},
			Timestamp:   time.Unix(0, 1621337323243086400).UTC(),
			View:        2,
			ParentVoterIDs: []flow.Identifier{
				{0x5, 0x55, 0x33, 0x7e, 0xf, 0x66, 0x1e, 0xc9, 0xb0, 0x7e, 0xbb, 0x69, 0x46, 0x8, 0x13, 0x16, 0xfa, 0x65, 0xc0, 0xba, 0xca, 0x6b, 0xd4, 0x70, 0x5b, 0xf6, 0x9d, 0x56, 0xa9, 0xf5, 0xb8, 0xa3},
				{0x45, 0x51, 0xbe, 0x34, 0xd9, 0xf7, 0xa9, 0x3b, 0x0, 0xd2, 0x87, 0xbd, 0x68, 0x3f, 0x7d, 0xd6, 0x34, 0x5e, 0x65, 0x90, 0x72, 0x40, 0x40, 0x5, 0x54, 0xfb, 0xdf, 0xa1, 0x69, 0x7d, 0x3b, 0xfa},
				{0xd3, 0x5f, 0xac, 0xa6, 0x7a, 0xbc, 0x6, 0xc3, 0x34, 0xb1, 0xe5, 0xa7, 0x88, 0x23, 0x98, 0xda, 0xe9, 0xc1, 0xda, 0xd9, 0x13, 0xe5, 0x60, 0x9e, 0xe1, 0xd4, 0x63, 0xd5, 0x5a, 0x22, 0x44, 0xf7},
			},
			ParentVoterSig: []byte{0xa0, 0x2a, 0xed, 0xa7, 0xc4, 0xe1, 0x40, 0x8e, 0x70, 0xe7, 0xa6, 0x7d, 0x81, 0x99, 0x24, 0xf4, 0x7c, 0x30, 0x42, 0x2, 0xe5, 0xaa, 0xfa, 0x89, 0x89, 0xda, 0x9d, 0x22, 0xb5, 0x45, 0xb0, 0xc2, 0xa4, 0x4c, 0x4b, 0xf3, 0xe1, 0xdf, 0x31, 0x73, 0xa2, 0x3e, 0x48, 0x5, 0xb4, 0xec, 0x5d, 0xcf, 0x8a, 0x6f, 0x42, 0xe7, 0xdd, 0xad, 0x7d, 0x4b, 0x7e, 0xc, 0xcb, 0xc, 0x6, 0x64, 0x10, 0x86, 0x4d, 0xd6, 0x89, 0x3a, 0x6f, 0x1e, 0xc4, 0xef, 0x6c, 0x18, 0xbf, 0xd5, 0x3a, 0x36, 0x25, 0xf0, 0xf9, 0xb0, 0x1f, 0x27, 0x4e, 0x4d, 0x72, 0x34, 0xf4, 0x51, 0xcc, 0x7d, 0x81, 0x86, 0xed, 0xb2},
			ProposerID:     flow.Identifier{0x5, 0x55, 0x33, 0x7e, 0xf, 0x66, 0x1e, 0xc9, 0xb0, 0x7e, 0xbb, 0x69, 0x46, 0x8, 0x13, 0x16, 0xfa, 0x65, 0xc0, 0xba, 0xca, 0x6b, 0xd4, 0x70, 0x5b, 0xf6, 0x9d, 0x56, 0xa9, 0xf5, 0xb8, 0xa3},
			ProposerSig:    []byte{0x8f, 0x7a, 0x5c, 0xbb, 0x4d, 0xfa, 0x46, 0x6, 0xe9, 0x9a, 0xd6, 0xea, 0xa1, 0xa3, 0x1a, 0x3b, 0xb4, 0xa4, 0xc0, 0xa4, 0x4a, 0xa6, 0xe7, 0xe6, 0x8b, 0x5e, 0x5e, 0x8b, 0x3f, 0xf, 0xa3, 0x32, 0x68, 0xee, 0x59, 0x20, 0x97, 0xa2, 0x38, 0xd5, 0x25, 0x64, 0xc3, 0x54, 0x1f, 0x1f, 0x9c, 0x5a, 0xaa, 0xc5, 0x1, 0xf3, 0xff, 0x3f, 0x83, 0x4, 0x9f, 0xed, 0xc3, 0x84, 0xf8, 0x5, 0xe6, 0x15, 0xf, 0x21, 0x50, 0x27, 0x3a, 0x72, 0xe3, 0xa0, 0x35, 0xec, 0x43, 0x48, 0x40, 0x9b, 0xef, 0xfa, 0x1b, 0x20, 0xb6, 0x7, 0x53, 0xcf, 0x38, 0x9f, 0x87, 0xf0, 0x52, 0x47, 0xfc, 0xc4, 0x70, 0x47},
		}

	case 13:
		return flow.Header{
			ChainID:     dps.FlowTestnet,
			ParentID:    flow.Identifier{0x90, 0x35, 0xc5, 0x58, 0x37, 0x9b, 0x20, 0x8e, 0xba, 0x11, 0x13, 0xc, 0x92, 0x85, 0x37, 0xfe, 0x50, 0xad, 0x93, 0xcd, 0xee, 0x31, 0x49, 0x80, 0xfc, 0xcb, 0x69, 0x5a, 0xa3, 0x1d, 0xf7, 0xfc},
			Height:      13,
			PayloadHash: flow.Identifier{0xd0, 0x43, 0x96, 0xd9, 0x5, 0x79, 0xea, 0xe8, 0xfc, 0xf5, 0x90, 0x6f, 0xce, 0xde, 0x4d, 0x26, 0xbe, 0x66, 0x7d, 0x5d, 0x6e, 0x36, 0x8c, 0xd1, 0x99, 0xed, 0x8a, 0x66, 0x17, 0x84, 0x67, 0xf2},
			Timestamp:   time.Unix(0, 1621338403243086400).UTC(),
			View:        14,
			ParentVoterIDs: []flow.Identifier{
				{0x5, 0x55, 0x33, 0x7e, 0xf, 0x66, 0x1e, 0xc9, 0xb0, 0x7e, 0xbb, 0x69, 0x46, 0x8, 0x13, 0x16, 0xfa, 0x65, 0xc0, 0xba, 0xca, 0x6b, 0xd4, 0x70, 0x5b, 0xf6, 0x9d, 0x56, 0xa9, 0xf5, 0xb8, 0xa3},
				{0x45, 0x51, 0xbe, 0x34, 0xd9, 0xf7, 0xa9, 0x3b, 0x0, 0xd2, 0x87, 0xbd, 0x68, 0x3f, 0x7d, 0xd6, 0x34, 0x5e, 0x65, 0x90, 0x72, 0x40, 0x40, 0x5, 0x54, 0xfb, 0xdf, 0xa1, 0x69, 0x7d, 0x3b, 0xfa},
				{0xd3, 0x5f, 0xac, 0xa6, 0x7a, 0xbc, 0x6, 0xc3, 0x34, 0xb1, 0xe5, 0xa7, 0x88, 0x23, 0x98, 0xda, 0xe9, 0xc1, 0xda, 0xd9, 0x13, 0xe5, 0x60, 0x9e, 0xe1, 0xd4, 0x63, 0xd5, 0x5a, 0x22, 0x44, 0xf7},
			},
			ParentVoterSig: []byte{0x92, 0xf6, 0x89, 0x55, 0x76, 0xb, 0x5, 0xe5, 0x89, 0xae, 0x3e, 0x21, 0xa6, 0x4a, 0x4f, 0xb6, 0xd4, 0x40, 0xcc, 0x94, 0x90, 0x8f, 0x40, 0xeb, 0xcd, 0xfd, 0x30, 0x45, 0xd7, 0x94, 0xc8, 0x95, 0xfe, 0xf1, 0x7e, 0xd8, 0x71, 0xce, 0x6c, 0x3, 0xb8, 0x4f, 0x5f, 0x8, 0x30, 0x2, 0x8a, 0x85, 0x90, 0x2a, 0xc5, 0xd, 0x81, 0x49, 0x11, 0xd9, 0x37, 0x35, 0x6f, 0xf9, 0x3f, 0x7b, 0x52, 0x4, 0xdb, 0x5a, 0x36, 0x81, 0xda, 0xa6, 0x47, 0xb5, 0xd9, 0xa7, 0xec, 0x6, 0xda, 0x34, 0x70, 0xdf, 0x8, 0x4a, 0xd5, 0xd0, 0x14, 0xf7, 0x2d, 0xd7, 0x5b, 0x66, 0x39, 0x64, 0x3c, 0xf1, 0xbb, 0xe4},
			ProposerID:     flow.Identifier{0x5, 0x55, 0x33, 0x7e, 0xf, 0x66, 0x1e, 0xc9, 0xb0, 0x7e, 0xbb, 0x69, 0x46, 0x8, 0x13, 0x16, 0xfa, 0x65, 0xc0, 0xba, 0xca, 0x6b, 0xd4, 0x70, 0x5b, 0xf6, 0x9d, 0x56, 0xa9, 0xf5, 0xb8, 0xa3},
			ProposerSig:    []byte{0x98, 0xe2, 0xf9, 0x46, 0xc4, 0xd7, 0x71, 0xc6, 0xf6, 0x56, 0x21, 0xe, 0xe7, 0xa4, 0x9c, 0xaa, 0xc, 0x3f, 0x7a, 0x75, 0xb9, 0x53, 0x95, 0x37, 0xdd, 0xb7, 0x4b, 0x7f, 0xfc, 0x1e, 0x1a, 0xe9, 0xfe, 0xbb, 0x56, 0x2e, 0xb8, 0x6e, 0xf6, 0xd8, 0x25, 0x4d, 0x5f, 0xee, 0x46, 0x1d, 0xd, 0xd4, 0x82, 0xeb, 0x7, 0xde, 0xa2, 0x8, 0x58, 0x13, 0xba, 0xfb, 0xeb, 0x2e, 0xcd, 0x88, 0x2e, 0x7c, 0x1b, 0xe8, 0xc5, 0x1d, 0x84, 0xe, 0xa2, 0x10, 0xbe, 0xe3, 0xb6, 0x26, 0x87, 0x4b, 0x6c, 0xbf, 0xc2, 0xe0, 0x85, 0xf4, 0x7e, 0xf5, 0xf3, 0x55, 0x5d, 0xd3, 0x49, 0xff, 0xc8, 0xe3, 0xb5, 0x53},
		}

	case 43:
		return flow.Header{
			ChainID:     dps.FlowTestnet,
			ParentID:    flow.Identifier{0x91, 0xc0, 0xb, 0x22, 0xdc, 0x9b, 0x84, 0x28, 0x1d, 0x29, 0x3f, 0x6e, 0x1f, 0xf6, 0x80, 0x13, 0x32, 0x39, 0xad, 0xdd, 0x8b, 0x2, 0x20, 0xa2, 0x44, 0x55, 0x4e, 0x1d, 0x96, 0xae, 0xd8, 0xe0},
			Height:      43,
			PayloadHash: flow.Identifier{0x5, 0xae, 0xad, 0x1e, 0xaa, 0x5f, 0x40, 0x85, 0xf0, 0xb4, 0xa2, 0x67, 0x67, 0x3d, 0x13, 0xc4, 0x6, 0x26, 0xbf, 0xe9, 0x3d, 0xf9, 0x90, 0x38, 0x5c, 0xf3, 0xbc, 0x7a, 0xfd, 0x77, 0x15, 0x21},
			Timestamp:   time.Unix(0, 1621341103243086400).UTC(),
			View:        44,
			ParentVoterIDs: []flow.Identifier{
				{0x5, 0x55, 0x33, 0x7e, 0xf, 0x66, 0x1e, 0xc9, 0xb0, 0x7e, 0xbb, 0x69, 0x46, 0x8, 0x13, 0x16, 0xfa, 0x65, 0xc0, 0xba, 0xca, 0x6b, 0xd4, 0x70, 0x5b, 0xf6, 0x9d, 0x56, 0xa9, 0xf5, 0xb8, 0xa3},
				{0x45, 0x51, 0xbe, 0x34, 0xd9, 0xf7, 0xa9, 0x3b, 0x0, 0xd2, 0x87, 0xbd, 0x68, 0x3f, 0x7d, 0xd6, 0x34, 0x5e, 0x65, 0x90, 0x72, 0x40, 0x40, 0x5, 0x54, 0xfb, 0xdf, 0xa1, 0x69, 0x7d, 0x3b, 0xfa},
				{0xd3, 0x5f, 0xac, 0xa6, 0x7a, 0xbc, 0x6, 0xc3, 0x34, 0xb1, 0xe5, 0xa7, 0x88, 0x23, 0x98, 0xda, 0xe9, 0xc1, 0xda, 0xd9, 0x13, 0xe5, 0x60, 0x9e, 0xe1, 0xd4, 0x63, 0xd5, 0x5a, 0x22, 0x44, 0xf7},
			},
			ParentVoterSig: []byte{0x95, 0x8e, 0x2b, 0x47, 0x8, 0xfa, 0x12, 0xa5, 0x3, 0x99, 0xbc, 0xce, 0xb5, 0x82, 0xac, 0x71, 0x7a, 0x9, 0x87, 0x60, 0x17, 0x70, 0x1c, 0x51, 0xa, 0xef, 0x45, 0x9e, 0x7, 0xc1, 0x4, 0x92, 0xa, 0x7b, 0xd6, 0x13, 0xb0, 0x6c, 0x45, 0x4c, 0x2c, 0xba, 0xc4, 0xa3, 0xb0, 0xf6, 0x87, 0x64, 0x93, 0x83, 0xca, 0x2b, 0x48, 0x41, 0x7f, 0x84, 0x2b, 0xf1, 0x84, 0xda, 0x2e, 0xec, 0xd7, 0xc, 0xb6, 0x54, 0x19, 0x8e, 0x20, 0x1e, 0xa8, 0x8c, 0xb0, 0x38, 0xab, 0xc4, 0x40, 0x13, 0x0, 0xfc, 0x55, 0x26, 0xb8, 0xc5, 0x5a, 0xa9, 0xd4, 0xe4, 0x9f, 0xf7, 0x3c, 0x68, 0x68, 0xdf, 0x38, 0x54},
			ProposerID:     flow.Identifier{0x5, 0x55, 0x33, 0x7e, 0xf, 0x66, 0x1e, 0xc9, 0xb0, 0x7e, 0xbb, 0x69, 0x46, 0x8, 0x13, 0x16, 0xfa, 0x65, 0xc0, 0xba, 0xca, 0x6b, 0xd4, 0x70, 0x5b, 0xf6, 0x9d, 0x56, 0xa9, 0xf5, 0xb8, 0xa3},
			ProposerSig:    []byte{0xb0, 0x2, 0x4f, 0xfc, 0x71, 0xba, 0x38, 0xc1, 0x24, 0x79, 0xa0, 0xd2, 0x66, 0xe3, 0xfb, 0x20, 0xe2, 0x2e, 0x27, 0xe4, 0x99, 0x91, 0xc4, 0x44, 0x28, 0x85, 0x87, 0x4, 0x44, 0x54, 0xcb, 0x47, 0x28, 0xd5, 0x8f, 0xc4, 0x89, 0x38, 0xfb, 0xce, 0xf6, 0xba, 0x35, 0x74, 0xf1, 0x52, 0xb4, 0x5a, 0x8e, 0x4e, 0x4a, 0xc3, 0x23, 0xe3, 0xfe, 0xac, 0x9, 0xf9, 0x1, 0x37, 0xd4, 0xa, 0x54, 0x81, 0x63, 0x93, 0x56, 0xeb, 0x7d, 0x23, 0x13, 0x20, 0x7f, 0xb7, 0x47, 0xe3, 0x33, 0xb8, 0x2e, 0x5, 0xc3, 0x96, 0xdd, 0x20, 0x56, 0xca, 0x48, 0xd1, 0x6b, 0x48, 0x10, 0x6, 0x26, 0xb3, 0x84, 0x18},
		}

	case 44:
		return flow.Header{
			ChainID:     dps.FlowTestnet,
			ParentID:    flow.Identifier{0xda, 0xb1, 0x86, 0xb4, 0x51, 0x99, 0xc0, 0xc2, 0x60, 0x60, 0xea, 0x9, 0x28, 0x8b, 0x2f, 0x16, 0x3, 0x2d, 0xa4, 0xf, 0xc5, 0x4c, 0x81, 0xbb, 0x2a, 0x82, 0x67, 0xa5, 0xc1, 0x39, 0x6, 0xe6},
			Height:      44,
			PayloadHash: flow.Identifier{0x80, 0xfe, 0xaf, 0x28, 0x4f, 0x8a, 0x51, 0x6c, 0x8c, 0x8, 0x6a, 0x9f, 0xae, 0xc0, 0xbd, 0xbb, 0x6b, 0xcd, 0xf1, 0xc8, 0x2b, 0x4f, 0xc6, 0xdb, 0x35, 0xff, 0x75, 0x42, 0x11, 0x8, 0x1b, 0xd9},
			Timestamp:   time.Unix(0, 1621341193243086400).UTC(),
			View:        45,
			ParentVoterIDs: []flow.Identifier{
				{0x45, 0x51, 0xbe, 0x34, 0xd9, 0xf7, 0xa9, 0x3b, 0x0, 0xd2, 0x87, 0xbd, 0x68, 0x3f, 0x7d, 0xd6, 0x34, 0x5e, 0x65, 0x90, 0x72, 0x40, 0x40, 0x5, 0x54, 0xfb, 0xdf, 0xa1, 0x69, 0x7d, 0x3b, 0xfa},
				{0x5, 0x55, 0x33, 0x7e, 0xf, 0x66, 0x1e, 0xc9, 0xb0, 0x7e, 0xbb, 0x69, 0x46, 0x8, 0x13, 0x16, 0xfa, 0x65, 0xc0, 0xba, 0xca, 0x6b, 0xd4, 0x70, 0x5b, 0xf6, 0x9d, 0x56, 0xa9, 0xf5, 0xb8, 0xa3},
				{0xd3, 0x5f, 0xac, 0xa6, 0x7a, 0xbc, 0x6, 0xc3, 0x34, 0xb1, 0xe5, 0xa7, 0x88, 0x23, 0x98, 0xda, 0xe9, 0xc1, 0xda, 0xd9, 0x13, 0xe5, 0x60, 0x9e, 0xe1, 0xd4, 0x63, 0xd5, 0x5a, 0x22, 0x44, 0xf7},
			},
			ParentVoterSig: []byte{0x8f, 0x90, 0xd9, 0xf6, 0x9, 0xc9, 0x9, 0xb7, 0x5b, 0x46, 0x7d, 0x4a, 0x17, 0xdb, 0x4e, 0xb7, 0xce, 0xc0, 0x8, 0x7e, 0xcb, 0xf6, 0xde, 0x76, 0xc6, 0xf5, 0x31, 0xbe, 0x13, 0xa3, 0x90, 0xc, 0xc3, 0x8f, 0x33, 0xeb, 0x50, 0xfc, 0x4d, 0x93, 0xb3, 0x64, 0xe8, 0x80, 0x74, 0x51, 0xc4, 0xb3, 0x8c, 0x41, 0xd7, 0xb5, 0xd5, 0x51, 0x72, 0x15, 0x4f, 0x7, 0x95, 0xfb, 0xcf, 0x54, 0xa7, 0x92, 0xa0, 0x90, 0x98, 0x9d, 0x57, 0xc2, 0xb2, 0x7a, 0xe3, 0x1e, 0x5b, 0xbd, 0x2c, 0x5d, 0x71, 0x23, 0x48, 0x87, 0xda, 0xb7, 0x4a, 0x13, 0xf1, 0xea, 0x37, 0x41, 0x53, 0xb7, 0xf4, 0x1f, 0x53, 0x30},
			ProposerID:     flow.Identifier{0x45, 0x51, 0xbe, 0x34, 0xd9, 0xf7, 0xa9, 0x3b, 0x0, 0xd2, 0x87, 0xbd, 0x68, 0x3f, 0x7d, 0xd6, 0x34, 0x5e, 0x65, 0x90, 0x72, 0x40, 0x40, 0x5, 0x54, 0xfb, 0xdf, 0xa1, 0x69, 0x7d, 0x3b, 0xfa},
			ProposerSig:    []byte{0x8f, 0x30, 0x2d, 0x1f, 0xb1, 0x6c, 0x30, 0x24, 0xf0, 0x6, 0x76, 0x95, 0x30, 0xeb, 0xda, 0x22, 0xea, 0x7f, 0x4, 0x8a, 0x2e, 0x76, 0x8a, 0x72, 0xcd, 0x91, 0x29, 0x9b, 0xca, 0x3e, 0xf, 0x78, 0x31, 0xf, 0x79, 0x1, 0x68, 0xb4, 0x26, 0xc1, 0x92, 0x48, 0xf8, 0xaa, 0xb6, 0x41, 0x85, 0x70, 0xb3, 0x3, 0x23, 0x4e, 0x22, 0xf0, 0x1a, 0x69, 0x73, 0x55, 0x4c, 0x91, 0xdb, 0xde, 0x8b, 0x7f, 0xf6, 0xa8, 0xe1, 0x6f, 0xf4, 0xf7, 0xd3, 0x51, 0xd7, 0xd2, 0xf5, 0x90, 0x1e, 0x2a, 0x95, 0xa, 0xd5, 0x11, 0xf3, 0xec, 0x53, 0x87, 0x5, 0xf, 0x21, 0xba, 0xfe, 0x98, 0x97, 0x93, 0xb3, 0xc},
		}

	case 165:
		return flow.Header{
			ChainID:     dps.FlowTestnet,
			ParentID:    flow.Identifier{0x99, 0xe4, 0x79, 0x7, 0x96, 0x13, 0x34, 0x81, 0x82, 0x9c, 0xe, 0xd6, 0x43, 0xcb, 0x21, 0x87, 0xa8, 0xab, 0x3a, 0xbf, 0x23, 0xa1, 0x38, 0x5d, 0xe7, 0xa8, 0xf8, 0x64, 0x40, 0x35, 0xc8, 0x82},
			Height:      165,
			PayloadHash: flow.Identifier{0x96, 0x97, 0xac, 0xda, 0xba, 0x28, 0xfb, 0xe9, 0x75, 0xeb, 0xc, 0x56, 0x21, 0x94, 0x1b, 0x54, 0x7e, 0x71, 0x58, 0x26, 0xbd, 0xa6, 0xa8, 0xce, 0xd6, 0xcf, 0x7d, 0xa9, 0x66, 0xec, 0x37, 0xb7},
			Timestamp:   time.Unix(0, 1621352083243086400).UTC(),
			View:        166,
			ParentVoterIDs: []flow.Identifier{
				{0x5, 0x55, 0x33, 0x7e, 0xf, 0x66, 0x1e, 0xc9, 0xb0, 0x7e, 0xbb, 0x69, 0x46, 0x8, 0x13, 0x16, 0xfa, 0x65, 0xc0, 0xba, 0xca, 0x6b, 0xd4, 0x70, 0x5b, 0xf6, 0x9d, 0x56, 0xa9, 0xf5, 0xb8, 0xa3},
				{0x45, 0x51, 0xbe, 0x34, 0xd9, 0xf7, 0xa9, 0x3b, 0x0, 0xd2, 0x87, 0xbd, 0x68, 0x3f, 0x7d, 0xd6, 0x34, 0x5e, 0x65, 0x90, 0x72, 0x40, 0x40, 0x5, 0x54, 0xfb, 0xdf, 0xa1, 0x69, 0x7d, 0x3b, 0xfa},
				{0xd3, 0x5f, 0xac, 0xa6, 0x7a, 0xbc, 0x6, 0xc3, 0x34, 0xb1, 0xe5, 0xa7, 0x88, 0x23, 0x98, 0xda, 0xe9, 0xc1, 0xda, 0xd9, 0x13, 0xe5, 0x60, 0x9e, 0xe1, 0xd4, 0x63, 0xd5, 0x5a, 0x22, 0x44, 0xf7},
			},
			ParentVoterSig: []byte{0xb2, 0x9f, 0x53, 0x34, 0x5c, 0x98, 0xae, 0xb, 0xc4, 0x2a, 0xa1, 0xc4, 0xc5, 0xeb, 0x14, 0xe9, 0xc3, 0xb9, 0x33, 0xb6, 0x6e, 0xe7, 0xe0, 0x2, 0x51, 0x5b, 0xc5, 0x81, 0xfc, 0xf5, 0xbd, 0x2, 0x61, 0xde, 0xea, 0x87, 0x5a, 0x6, 0xab, 0xd4, 0x1e, 0xf9, 0x28, 0xc3, 0x84, 0x54, 0x2a, 0x77, 0xb5, 0x7a, 0x83, 0xda, 0x6d, 0xb6, 0x34, 0x90, 0x66, 0x8a, 0xa7, 0x93, 0x3a, 0x5, 0xff, 0xa5, 0xcb, 0x88, 0xd8, 0x58, 0x67, 0xd5, 0xdf, 0xe7, 0x8f, 0xea, 0xdc, 0x76, 0xd6, 0x5b, 0xf0, 0x5, 0x53, 0xd3, 0x83, 0x24, 0xb5, 0x4c, 0x94, 0x31, 0x0, 0x93, 0x3c, 0xf0, 0x7, 0x84, 0x5c, 0x72},
			ProposerID:     flow.Identifier{0x45, 0x51, 0xbe, 0x34, 0xd9, 0xf7, 0xa9, 0x3b, 0x0, 0xd2, 0x87, 0xbd, 0x68, 0x3f, 0x7d, 0xd6, 0x34, 0x5e, 0x65, 0x90, 0x72, 0x40, 0x40, 0x5, 0x54, 0xfb, 0xdf, 0xa1, 0x69, 0x7d, 0x3b, 0xfa},
			ProposerSig:    []byte{0x8e, 0x8b, 0xcb, 0xef, 0x35, 0x80, 0xbf, 0x85, 0xd6, 0xb2, 0xaa, 0xa, 0x5d, 0x42, 0x8c, 0xcf, 0x40, 0x95, 0x9c, 0x2d, 0x5f, 0xc9, 0x22, 0x1e, 0x8f, 0xe0, 0x62, 0x14, 0x33, 0x70, 0xe5, 0x8b, 0xaa, 0xaf, 0x13, 0x92, 0x7f, 0xc0, 0x7a, 0x3c, 0x2e, 0xf6, 0x88, 0x8a, 0x35, 0x46, 0x3c, 0xa0, 0xb9, 0x37, 0x75, 0xbd, 0x24, 0xe2, 0x81, 0xa1, 0x4e, 0x40, 0x98, 0x31, 0xfb, 0xa4, 0x2e, 0x2b, 0x59, 0x6c, 0x5, 0x67, 0xd7, 0x68, 0xb9, 0xc, 0x51, 0x7d, 0x48, 0xd2, 0x27, 0xb8, 0xe0, 0x97, 0x93, 0xf2, 0xf2, 0x93, 0x1e, 0xe7, 0xe6, 0x9d, 0x6b, 0xc0, 0x64, 0xce, 0x62, 0xf5, 0xdd, 0xe8},
		}

	case 181:
		return flow.Header{
			ChainID:     dps.FlowTestnet,
			ParentID:    flow.Identifier{0x9e, 0x33, 0xa6, 0x57, 0x8, 0x8f, 0xc6, 0xf1, 0x62, 0xc2, 0x36, 0xee, 0x1d, 0x8f, 0xc4, 0x78, 0x2c, 0x9f, 0xc1, 0xb6, 0x5b, 0xaf, 0x4a, 0x40, 0x6, 0x24, 0xf7, 0x66, 0x36, 0x9c, 0xaa, 0xd3},
			Height:      181,
			PayloadHash: flow.Identifier{0xfb, 0x1a, 0xae, 0xc9, 0x22, 0xeb, 0xbe, 0xa0, 0xb2, 0x36, 0x24, 0x57, 0xda, 0xa6, 0xc0, 0xa, 0x4a, 0x34, 0xe2, 0x11, 0xef, 0x8f, 0x8, 0x7a, 0x4f, 0x71, 0x33, 0xfb, 0xe5, 0x35, 0x45, 0x5a},
			Timestamp:   time.Unix(0, 1621353523243086400).UTC(),
			View:        182,
			ParentVoterIDs: []flow.Identifier{
				{0x5, 0x55, 0x33, 0x7e, 0xf, 0x66, 0x1e, 0xc9, 0xb0, 0x7e, 0xbb, 0x69, 0x46, 0x8, 0x13, 0x16, 0xfa, 0x65, 0xc0, 0xba, 0xca, 0x6b, 0xd4, 0x70, 0x5b, 0xf6, 0x9d, 0x56, 0xa9, 0xf5, 0xb8, 0xa3},
				{0x45, 0x51, 0xbe, 0x34, 0xd9, 0xf7, 0xa9, 0x3b, 0x0, 0xd2, 0x87, 0xbd, 0x68, 0x3f, 0x7d, 0xd6, 0x34, 0x5e, 0x65, 0x90, 0x72, 0x40, 0x40, 0x5, 0x54, 0xfb, 0xdf, 0xa1, 0x69, 0x7d, 0x3b, 0xfa},
				{0xd3, 0x5f, 0xac, 0xa6, 0x7a, 0xbc, 0x6, 0xc3, 0x34, 0xb1, 0xe5, 0xa7, 0x88, 0x23, 0x98, 0xda, 0xe9, 0xc1, 0xda, 0xd9, 0x13, 0xe5, 0x60, 0x9e, 0xe1, 0xd4, 0x63, 0xd5, 0x5a, 0x22, 0x44, 0xf7},
			},
			ParentVoterSig: []byte{0xa4, 0x25, 0xf1, 0xc9, 0x4, 0x55, 0xdb, 0x46, 0x51, 0x7a, 0x15, 0x8a, 0x10, 0x74, 0x9b, 0x1d, 0x1d, 0xbc, 0xb, 0x6d, 0x67, 0x33, 0x60, 0x21, 0x2e, 0x7b, 0xec, 0xae, 0xb3, 0x85, 0x14, 0x2b, 0x99, 0x1b, 0xc2, 0xd2, 0xa3, 0xfd, 0x59, 0x38, 0x13, 0x76, 0x21, 0x50, 0xc8, 0x57, 0xa4, 0xf8, 0xad, 0xd7, 0x2c, 0xce, 0x0, 0x45, 0x61, 0xc6, 0xb8, 0x98, 0x4a, 0x51, 0x92, 0xff, 0x2, 0x3, 0x2a, 0x87, 0xc1, 0x61, 0xbb, 0xda, 0x9, 0xab, 0x93, 0xc6, 0xb0, 0x60, 0x2e, 0xf9, 0x2f, 0xb9, 0x1f, 0x0, 0x58, 0xde, 0xd7, 0x86, 0x95, 0xbb, 0xfc, 0x85, 0x75, 0x1a, 0x52, 0x4c, 0x88, 0x7},
			ProposerID:     flow.Identifier{0x5, 0x55, 0x33, 0x7e, 0xf, 0x66, 0x1e, 0xc9, 0xb0, 0x7e, 0xbb, 0x69, 0x46, 0x8, 0x13, 0x16, 0xfa, 0x65, 0xc0, 0xba, 0xca, 0x6b, 0xd4, 0x70, 0x5b, 0xf6, 0x9d, 0x56, 0xa9, 0xf5, 0xb8, 0xa3},
			ProposerSig:    []byte{0x94, 0x8e, 0x5, 0x8f, 0xbe, 0xa9, 0x6e, 0xf8, 0xbe, 0x43, 0x8, 0x49, 0xbf, 0x8b, 0x9d, 0x62, 0x47, 0xca, 0x57, 0x41, 0xbd, 0x81, 0x55, 0x23, 0x19, 0xa7, 0x5, 0x3a, 0xc8, 0x32, 0x51, 0x55, 0x2f, 0x62, 0x9f, 0xe2, 0xa0, 0x14, 0x30, 0xb0, 0xa9, 0x32, 0xa5, 0x1f, 0xa8, 0x27, 0xc1, 0x8a, 0x88, 0x8e, 0xa0, 0xd2, 0xcc, 0x67, 0xca, 0x18, 0x4d, 0xb4, 0x4b, 0x1a, 0xad, 0x62, 0x15, 0xb1, 0xa, 0xdd, 0x14, 0x1a, 0xf5, 0xa0, 0x1b, 0x56, 0x30, 0x7d, 0x70, 0x7c, 0x5b, 0xc6, 0xa8, 0xad, 0x74, 0x13, 0x1, 0xd, 0x10, 0xe2, 0x42, 0xe6, 0xc5, 0x96, 0xd4, 0x3c, 0xa8, 0xeb, 0x8a, 0x55},
		}

	case 425:
		return flow.Header{
			ChainID:     dps.FlowTestnet,
			ParentID:    flow.Identifier{0x6a, 0xf2, 0x66, 0x21, 0xec, 0xa9, 0x2b, 0xab, 0xda, 0x2d, 0xf3, 0xeb, 0xcd, 0x2f, 0xe2, 0x69, 0x94, 0x6b, 0x3b, 0xf2, 0x8, 0x18, 0x35, 0x69, 0x25, 0x86, 0x30, 0xe6, 0x44, 0x86, 0x83, 0x1d},
			Height:      425,
			PayloadHash: flow.Identifier{0xca, 0x42, 0x4b, 0x9c, 0x56, 0xeb, 0xb7, 0x1d, 0xbf, 0xaa, 0x5d, 0x3b, 0xbd, 0xfa, 0x2f, 0xf6, 0x2b, 0x39, 0x7b, 0xb9, 0xcd, 0x4, 0x5, 0xfe, 0x8c, 0xd8, 0x9d, 0x5e, 0xe7, 0x74, 0x25, 0x7a},
			Timestamp:   time.Unix(0, 1621375483243086400).UTC(),
			View:        427,
			ParentVoterIDs: []flow.Identifier{
				{0x45, 0x51, 0xbe, 0x34, 0xd9, 0xf7, 0xa9, 0x3b, 0x0, 0xd2, 0x87, 0xbd, 0x68, 0x3f, 0x7d, 0xd6, 0x34, 0x5e, 0x65, 0x90, 0x72, 0x40, 0x40, 0x5, 0x54, 0xfb, 0xdf, 0xa1, 0x69, 0x7d, 0x3b, 0xfa},
				{0xd3, 0x5f, 0xac, 0xa6, 0x7a, 0xbc, 0x6, 0xc3, 0x34, 0xb1, 0xe5, 0xa7, 0x88, 0x23, 0x98, 0xda, 0xe9, 0xc1, 0xda, 0xd9, 0x13, 0xe5, 0x60, 0x9e, 0xe1, 0xd4, 0x63, 0xd5, 0x5a, 0x22, 0x44, 0xf7},
				{0x5, 0x55, 0x33, 0x7e, 0xf, 0x66, 0x1e, 0xc9, 0xb0, 0x7e, 0xbb, 0x69, 0x46, 0x8, 0x13, 0x16, 0xfa, 0x65, 0xc0, 0xba, 0xca, 0x6b, 0xd4, 0x70, 0x5b, 0xf6, 0x9d, 0x56, 0xa9, 0xf5, 0xb8, 0xa3},
			},
			ParentVoterSig: []byte{0xa9, 0x69, 0xb8, 0x57, 0xaf, 0xff, 0xd9, 0x4c, 0xd8, 0x94, 0x34, 0xc, 0xd8, 0xad, 0xa4, 0xb2, 0xe6, 0xd2, 0x8f, 0x18, 0x17, 0xa8, 0xa8, 0xc4, 0x80, 0x1e, 0x7a, 0x82, 0x90, 0xed, 0x58, 0x38, 0xf3, 0x2c, 0x4c, 0xa3, 0x31, 0xe3, 0x2b, 0x7e, 0x7c, 0x4b, 0xef, 0x5, 0x22, 0x9, 0xc4, 0xbc, 0xaf, 0x22, 0x1, 0x12, 0x4e, 0x37, 0xbb, 0xb1, 0xeb, 0xbb, 0x26, 0x9, 0x94, 0x6e, 0x7c, 0x49, 0xca, 0xba, 0xdc, 0x4f, 0x68, 0xbf, 0xd9, 0xb9, 0xc, 0x7e, 0x5d, 0xc1, 0x74, 0x1f, 0x4e, 0x5, 0xa2, 0xc5, 0x9f, 0x9b, 0x57, 0xb1, 0xfb, 0x75, 0x20, 0x95, 0xa3, 0x53, 0x7d, 0xa3, 0xa0, 0x3e},
			ProposerID:     flow.Identifier{0xd3, 0x5f, 0xac, 0xa6, 0x7a, 0xbc, 0x6, 0xc3, 0x34, 0xb1, 0xe5, 0xa7, 0x88, 0x23, 0x98, 0xda, 0xe9, 0xc1, 0xda, 0xd9, 0x13, 0xe5, 0x60, 0x9e, 0xe1, 0xd4, 0x63, 0xd5, 0x5a, 0x22, 0x44, 0xf7},
			ProposerSig:    []byte{0xa5, 0x58, 0x17, 0xf4, 0x1c, 0x51, 0x43, 0xb7, 0x89, 0xd5, 0x65, 0xb4, 0x68, 0x21, 0xe2, 0x54, 0x78, 0xf4, 0xc2, 0xe7, 0x51, 0xd5, 0xd3, 0xe8, 0xbd, 0xbb, 0x1a, 0xb5, 0x90, 0xf8, 0xb, 0x5a, 0x32, 0xb8, 0x3d, 0xf6, 0xbc, 0xf, 0x10, 0x11, 0x71, 0x24, 0x6f, 0xe4, 0x77, 0x71, 0x8c, 0x32, 0x84, 0xfd, 0x52, 0xf5, 0x31, 0xc7, 0x1f, 0xe7, 0x4d, 0x43, 0xfe, 0xc9, 0xfa, 0x3b, 0x78, 0x67, 0xcc, 0xfe, 0x77, 0xd4, 0xfc, 0x42, 0x74, 0x36, 0x0, 0x9e, 0xf, 0x99, 0xf7, 0xaa, 0xb4, 0xc3, 0xf3, 0xff, 0x3a, 0x17, 0x6e, 0xc3, 0xd7, 0x3d, 0x41, 0xc2, 0x50, 0x7, 0x33, 0xb5, 0x17, 0x83},
		}

	default:
		return flow.Header{}
	}
}
