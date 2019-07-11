package proof

import (
	"bytes"
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

type Step struct {
	// This is the next step, FullNode or ShortNode
	Step PathStep
	// Index is set if Step is FullNode and refers to which subnode we followed
	Index int
}

type Proof struct {
	Steps        []Step
	Key          []byte
	Value        []byte
	HexRemainder []byte
}

func (p *Proof) RecoverKey() []byte {
	var hexKey []byte
	for _, step := range p.Steps {
		switch t := step.Step.(type) {
		case *shortNode:
			hexKey = append(hexKey, t.Key...)
		case *fullNode:
			hexKey = append(hexKey, byte(step.Index))
		default:
			panic(fmt.Sprintf("Unknown type: %T", step.Step))
		}
	}

	hexKey = append(hexKey, p.HexRemainder...)
	return hexToKeybytes(hexKey)
}

// ComputeProof returns the proof value for a key in given trie. Returned path
// is the way from the value to the root of the tree.
func ComputeProof(tr *trie.Trie, key []byte) (*Proof, error) {
	record := ProofRecorder{}

	value := tr.Get(key)
	if value == nil {
		return nil, fmt.Errorf("No value found for key %X", key)
	}

	if err := tr.Prove(key, 0, &record); err != nil {
		return nil, err
	}

	steps, remainder, err := AnnotatePath(key, record.Path())
	if err != nil {
		return nil, err
	}

	proof := Proof{
		Steps:        steps,
		Key:          key,
		Value:        value,
		HexRemainder: remainder,
	}

	return &proof, nil
}

// AnnotatePath annotates the path of proofs, with the child we followed at each step
func AnnotatePath(key []byte, path []PathStep) ([]Step, []byte, error) {
	hexkey := keybytesToHex(key)
	fmt.Printf("hexkey: %X (%s)\n", hexkey, string(key))
	var steps []Step

	for _, p := range path {
		switch t := p.(type) {
		case *shortNode:
			// remove the prefix and continue
			if len(hexkey) < len(t.Key) || !bytes.Equal(t.Key, hexkey[:len(t.Key)]) {
				return nil, nil, fmt.Errorf("Shortnode prefix %X doesn't match key %X", t.Key, hexkey)
			}
			fmt.Printf("short: %X\n", t.Key)
			hexkey = hexkey[len(t.Key):]
			steps = append(steps, Step{Step: p})
		case *fullNode:
			fmt.Printf("next: %X\n", hexkey[0])
			idx := int(hexkey[0])
			hexkey = hexkey[1:]
			steps = append(steps, Step{Step: p, Index: idx})
		default:
			return nil, nil, fmt.Errorf("Unknown type: %T", p)
		}
	}

	// if len(hexkey) != 0 {
	// 	return nil, fmt.Errorf("Complete annotation, but path remains: %X", hexkey)
	// }

	return steps, hexkey, nil
}

// ProofRecorder is used to help us grab proofs
type ProofRecorder struct {
	path []PathStep
}

var _ ethdb.Putter = (*ProofRecorder)(nil)

func (p *ProofRecorder) Put(hash, value []byte) error {
	step, err := decodeNode(hash, value, 0)
	if err != nil {
		return err
	}
	p.path = append(p.path, step)
	return nil
}

func (p *ProofRecorder) Path() []PathStep {
	return p.path
}
