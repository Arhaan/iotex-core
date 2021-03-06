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

	"github.com/iotexproject/iotex-core/pkg/keypair"
	"github.com/iotexproject/iotex-core/pkg/util/byteutil"
	"github.com/iotexproject/iotex-core/pkg/version"
	"github.com/iotexproject/iotex-core/proto"
)

const (
	// VoteIntrinsicGas represents the intrinsic gas for vote
	VoteIntrinsicGas = uint64(10000)
)

// Vote defines the struct of account-based vote
type Vote struct {
	AbstractAction
}

// NewVote returns a Vote instance
func NewVote(nonce uint64, voterAddress string, voteeAddress string, gasLimit uint64, gasPrice *big.Int) (*Vote, error) {
	if voterAddress == "" {
		return nil, errors.Wrap(ErrAddress, "address of the voter is empty")
	}
	return &Vote{
		AbstractAction: AbstractAction{
			version:  version.ProtocolVersion,
			nonce:    nonce,
			srcAddr:  voterAddress,
			dstAddr:  voteeAddress,
			gasLimit: gasLimit,
			gasPrice: gasPrice,
		},
	}, nil
}

// Voter returns the voter's address
func (v *Vote) Voter() string {
	return v.SrcAddr()
}

// VoterPublicKey returns the voter's public key
func (v *Vote) VoterPublicKey() keypair.PublicKey {
	return v.SrcPubkey()
}

// Votee returns the votee's address
func (v *Vote) Votee() string {
	return v.DstAddr()
}

// TotalSize returns the total size of this Vote
func (v *Vote) TotalSize() uint32 {
	return v.BasicActionSize() + uint32(8) // TimestampSizeInBytes
}

// ByteStream returns a raw byte stream of this Transfer
func (v *Vote) ByteStream() []byte {
	// TODO: remove pbVote.Timestamp from the proto because we never set it
	return byteutil.Must(proto.Marshal(v.Proto()))
}

// Proto converts Vote to protobuf's ActionPb
func (v *Vote) Proto() *iproto.VotePb {
	return &iproto.VotePb{
		VoteeAddress: v.dstAddr,
	}
}

// LoadProto converts a protobuf's ActionPb to Vote
func (v *Vote) LoadProto(pbAct *iproto.VotePb) error { return nil }

// IntrinsicGas returns the intrinsic gas of a vote
func (v *Vote) IntrinsicGas() (uint64, error) {
	return VoteIntrinsicGas, nil
}

// Cost returns the total cost of a vote
func (v *Vote) Cost() (*big.Int, error) {
	intrinsicGas, err := v.IntrinsicGas()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get intrinsic gas for the vote")
	}
	voteFee := big.NewInt(0).Mul(v.GasPrice(), big.NewInt(0).SetUint64(intrinsicGas))
	return voteFee, nil
}
