package core

import (
	"bytes"

	"github.com/cypherium/go-cypherium/core/types"
	"github.com/cypherium/go-cypherium/hotstuff"
	"github.com/cypherium/go-cypherium/params"
	"github.com/cypherium/go-cypherium/reconfig/bftview"
)

type KeyBlockValidator struct {
	config *params.ChainConfig // Chain configuration options
	kbc    *KeyBlockChain      // Canonical block chain
}

// NewBlockValidator returns a new block validator which is safe for re-use
func NewKeyBlockValidator(config *params.ChainConfig, blockchain *KeyBlockChain) *KeyBlockValidator {
	validator := &KeyBlockValidator{
		config: config,
		kbc:    blockchain,
	}
	return validator
}

//ValidateKeyBlock,verify new keyblock
//All node rotations:1.Normal reconfig,witness=prvCommittee+new leader(input[0]);2.viewchange ,witness=prvCommittee
//2f+1 fixed，f node rotations:1.Normal reconfig,witness=prvCommittee;2.viewchange ,witness=prvCommittee
//Manual reconfig:witness= input
func (kbv *KeyBlockValidator) ValidateKeyBlock(block *types.KeyBlock) error {
	if block.Signatrue() == nil {
		return types.ErrEmptySignature
	}
	blockNumber := block.NumberU64()
	if kbv.kbc.HasBlock(block.Hash(), blockNumber) {
		return types.ErrKnownBlock
	}

	if !kbv.kbc.HasBlock(block.ParentHash(), blockNumber-1) {
		return types.ErrUnknownAncestor
	}

	//TxHash  verify

	mycommittee := &bftview.Committee{List: kbv.kbc.GetCommitteeByNumber(blockNumber - 1)}
	if mycommittee == nil || len(mycommittee.List) < 2 {
		return types.ErrInvalidCommittee
	}
	pubs := mycommittee.ToBlsPublicKeys(blockNumber - 1)

	tmpBlock := block.WithSignatrue(nil, nil)
	m := make([]byte, 0)
	buff := bytes.NewBuffer(m)
	err := tmpBlock.EncodeRLP(buff)
	if err != nil {
		return err
	}

	if !hotstuff.VerifySignature(block.Signatrue(), block.Exceptions(), buff.Bytes(), pubs) {
		return types.ErrInvalidSignature
	}

	return nil
}
