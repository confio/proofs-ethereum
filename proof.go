package proof

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

type PathStep interface {
	node
	isPathStep()
}

func (fullNode) isPathStep()  {}
func (shortNode) isPathStep() {}

// Seems like this can also be full/short node???
// I only see hash/value Node in tests
type Link interface {
	node
}

// format is ProofStep - FullNode or ShortNode
// child link either ValueNode or HashNode
//  - hashNode links to next step
//  - valueNode is the end

// ComputeProof returns the proof value for a key in given trie. Returned path
// is the way from the value to the root of the tree.
func ComputeProof(tr *trie.Trie, key []byte) (value []byte, path []PathStep, err error) {
	record := ProofComposer{}

	val := tr.Get(key)
	if val == nil {
		return nil, nil, fmt.Errorf("No value found for key %X", key)
	}

	if err := tr.Prove(key, 0, &record); err != nil {
		return nil, nil, err
	}

	return val, record.Path(), nil
}

type ProofComposer struct {
	path []PathStep
}

var _ ethdb.Putter = (*ProofComposer)(nil)

func (p *ProofComposer) Put(hash, value []byte) error {
	step, err := decodeNode(hash, value, 0)
	if err != nil {
		return err
	}
	p.path = append(p.path, step)
	return nil
}

func (p *ProofComposer) Path() []PathStep {
	return p.path
}
