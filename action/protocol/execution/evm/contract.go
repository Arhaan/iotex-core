// Copyright (c) 2018 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package evm

import (
	"context"

	"github.com/pkg/errors"

	"github.com/iotexproject/iotex-core/db"
	"github.com/iotexproject/iotex-core/db/trie"
	"github.com/iotexproject/iotex-core/pkg/hash"
	"github.com/iotexproject/iotex-core/pkg/util/byteutil"
	"github.com/iotexproject/iotex-core/state"
)

const (
	// CodeKVNameSpace is the bucket name for code
	CodeKVNameSpace = "Code"

	// ContractKVNameSpace is the bucket name for contract data storage
	ContractKVNameSpace = "Contract"
)

type (
	// Contract is a special type of account with code and storage trie.
	Contract interface {
		GetState(hash.Hash32B) ([]byte, error)
		SetState(hash.Hash32B, []byte) error
		GetCode() ([]byte, error)
		SetCode(hash.Hash32B, []byte)
		SelfState() *state.Account
		Commit() error
		RootHash() hash.Hash32B
		LoadRoot() error
		Iterator() (trie.Iterator, error)
		Snapshot() Contract
	}

	contract struct {
		*state.Account
		dirtyCode  bool   // contract's code has been set
		dirtyState bool   // contract's account state has changed
		code       []byte // contract byte-code
		root       hash.Hash32B
		dao        db.KVStore
		trie       trie.Trie // storage trie of the contract
	}
)

func (c *contract) Iterator() (trie.Iterator, error) {
	return trie.NewLeafIterator(c.trie)
}

// GetState get the value from contract storage
func (c *contract) GetState(key hash.Hash32B) ([]byte, error) {
	v, err := c.trie.Get(key[:])
	if err != nil {
		return nil, err
	}
	return v, nil
}

// SetState set the value into contract storage
func (c *contract) SetState(key hash.Hash32B, value []byte) error {
	c.dirtyState = true
	err := c.trie.Upsert(key[:], value)
	c.Account.Root = byteutil.BytesTo32B(c.trie.RootHash())
	return err
}

// GetCode gets the contract's byte-code
func (c *contract) GetCode() ([]byte, error) {
	if c.code != nil {
		return c.code, nil
	}
	return c.dao.Get(CodeKVNameSpace, c.Account.CodeHash)
}

// SetCode sets the contract's byte-code
func (c *contract) SetCode(hash hash.Hash32B, code []byte) {
	c.Account.CodeHash = hash[:]
	c.code = code
	c.dirtyCode = true
}

// account returns this contract's account
func (c *contract) SelfState() *state.Account {
	return c.Account
}

// Commit writes the changes into underlying trie
func (c *contract) Commit() error {
	if c.dirtyState {
		// record the new root hash, global account trie will Commit all pending writes to DB
		c.Account.Root = byteutil.BytesTo32B(c.trie.RootHash())
		c.dirtyState = false
	}
	if c.dirtyCode {
		// put the code into storage DB
		if err := c.dao.Put(CodeKVNameSpace, c.Account.CodeHash, c.code); err != nil {
			return errors.Wrapf(err, "Failed to store code for new contract, codeHash %x", c.Account.CodeHash[:])
		}
		c.dirtyCode = false
	}
	return nil
}

// RootHash returns storage trie's root hash
func (c *contract) RootHash() hash.Hash32B {
	return c.Account.Root
}

// LoadRoot loads storage trie's root
func (c *contract) LoadRoot() error {
	return c.trie.SetRootHash(c.Account.Root[:])
}

// Snapshot takes a snapshot of the contract object
func (c *contract) Snapshot() Contract {
	return &contract{
		Account:    c.Account.Clone(),
		dirtyCode:  c.dirtyCode,
		dirtyState: c.dirtyState,
		code:       c.code,
		root:       c.Account.Root,
		dao:        c.dao,
		// note we simply save the trie (which is an interface/pointer)
		// later Revert() call needs to reset the saved trie root
		trie: c.trie,
	}
}

// NewContract returns a Contract instance
func newContract(state *state.Account, dao db.KVStore, batch db.CachedBatch) (Contract, error) {
	dbForTrie, err := db.NewKVStoreForTrie(ContractKVNameSpace, dao, db.CachedBatchOption(batch))
	if err != nil {
		return nil, err
	}
	options := []trie.Option{
		trie.KVStoreOption(dbForTrie),
		trie.KeyLengthOption(hash.HashSize),
	}
	if state.Root != hash.ZeroHash32B {
		options = append(options, trie.RootHashOption(state.Root[:]))
	}

	tr, err := trie.NewTrie(options...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create storage trie for new contract")
	}
	if err := tr.Start(context.Background()); err != nil {
		return nil, err
	}

	return &contract{
		Account: state,
		root:    state.Root,
		dao:     dao,
		trie:    tr,
	}, nil
}
