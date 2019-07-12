package proof

import (
	"fmt"

	proofs "github.com/confio/proofs/go"

	"github.com/ethereum/go-ethereum/rlp"

)

func ConvertProof(p *Proof) (*proofs.ExistenceProof, error) {
	// convert back to forth.. starting with leaf...

	l := len(p.Steps)-1
	path, last := p.Steps[:l], p.Steps[l]

	ref := &proofs.ExistenceProof{
		Key: p.Key,
		Value: p.Value,		
	}

	err := addLeaf(ref, last)
	if err != nil {
		return nil, err
	}

	// convert last to 
	if len(path) > 0 {
		return nil, fmt.Errorf("not implemented")
	}
	return ref, nil
}

func addLeaf(p *proofs.ExistenceProof, step Step) error {
	switch n := step.Step.(type) {
	case *fullNode:
		return fmt.Errorf("Not implemented for fullNode")
	case *shortNode:
		bz, err := rlp.EncodeToBytes(collapseShortNode(n))
		if err != nil {
			return err
		}
		valLen := len(p.Value) + 1
		prefix := bz[:len(bz)-valLen]
		p.Leaf = &proofs.LeafOp{ 
			// Hash: proofs.HashOp_NO_HASH,
			Hash: proofs.HashOp_KECCAK,
			// TODO: ignore key completely
			PrehashKey: proofs.HashOp_NO_HASH,
			PrehashValue: proofs.HashOp_NO_HASH,
			Length: proofs.LengthOp_VAR_RLP,
			Prefix: prefix,
		}
		return nil
	default:
		return fmt.Errorf("Unexpected ending node: %T", step.Step)
	}
}