// Copyright (c) 2018 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package action

import (
	"math/big"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/iotexproject/iotex-core/pkg/util/byteutil"
	"github.com/iotexproject/iotex-core/pkg/version"
	"github.com/iotexproject/iotex-core/proto"
)

const (
	// StopSubChainIntrinsicGas is the instrinsic gas for stop sub chain action
	StopSubChainIntrinsicGas = uint64(1000)
)

// StopSubChain defines the action to stop sub chain
type StopSubChain struct {
	AbstractAction
	stopHeight uint64
}

// NewStopSubChain returns a StopSubChain instance
func NewStopSubChain(
	senderAddress string,
	nonce uint64,
	chainAddress string,
	stopHeight uint64,
	gasLimit uint64,
	gasPrice *big.Int,
) *StopSubChain {
	return &StopSubChain{
		AbstractAction: AbstractAction{
			version:  version.ProtocolVersion,
			nonce:    nonce,
			srcAddr:  senderAddress,
			dstAddr:  chainAddress,
			gasLimit: gasLimit,
			gasPrice: gasPrice,
		},
		stopHeight: stopHeight,
	}
}

// ChainAddress returns the address of the sub chain
func (ssc *StopSubChain) ChainAddress() string {
	return ssc.DstAddr()
}

// StopHeight returns the height to stop the sub chain
func (ssc *StopSubChain) StopHeight() uint64 {
	return ssc.stopHeight
}

// TotalSize returns the total size of this instance
func (ssc *StopSubChain) TotalSize() uint32 {
	return ssc.BasicActionSize() + 4 + 8 // chain id size + stop height size
}

// ByteStream returns a raw byte stream of this instance
func (ssc *StopSubChain) ByteStream() []byte {
	return byteutil.Must(proto.Marshal(ssc.Proto()))
}

// Proto converts StopSubChain to protobuf's ActionPb
func (ssc *StopSubChain) Proto() *iproto.StopSubChainPb {
	return &iproto.StopSubChainPb{
		StopHeight:      ssc.stopHeight,
		SubChainAddress: ssc.dstAddr,
	}
}

// LoadProto converts a protobuf's ActionPb to StopSubChain
func (ssc *StopSubChain) LoadProto(pbSSC *iproto.StopSubChainPb) error {
	if ssc == nil {
		return errors.New("nil action to load proto")
	}
	*ssc = StopSubChain{}

	if pbSSC == nil {
		return errors.New("empty action proto to load")
	}

	ssc.stopHeight = pbSSC.StopHeight
	return nil
}

// IntrinsicGas returns the intrinsic gas of a StopSubChain
func (ssc *StopSubChain) IntrinsicGas() (uint64, error) {
	return StopSubChainIntrinsicGas, nil
}

// Cost returns the total cost of a StopSubChain
func (ssc *StopSubChain) Cost() (*big.Int, error) {
	intrinsicGas, err := ssc.IntrinsicGas()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get intrinsic gas for the stop sub-chain action")
	}
	fee := big.NewInt(0).Mul(ssc.GasPrice(), big.NewInt(0).SetUint64(intrinsicGas))
	return fee, nil
}
