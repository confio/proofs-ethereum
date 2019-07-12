package proof

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

type PathStep interface {
	node
	isPathStep()
}

func (fullNode) isPathStep()  {}
func (shortNode) isPathStep() {}

// Link in intermediate instances is a hashNode (reference to next step)
// At the end, it is often shortNode(valueNode) to capture all remaining key (or key = 16 for direct value)
// Sometimes this is an embedded fullnode if there is little data
type Link interface {
	node
}

type Step struct {
	// This is the next step, FullNode or ShortNode
	Step PathStep
	// Index is set if Step is FullNode and refers to which subnode we followed
	Index int
	// Hash is set to the expected hash of this level
	Hash []byte
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

	proof, err := buildProof(key, value, record.Path())
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func VerifyProof(proof *Proof, rootHash common.Hash) error {
	// let's make sure this is consistent with the claim - key and value should match
	recovered := proof.RecoverKey()
	if !bytes.Equal(recovered, proof.Key) {
		return fmt.Errorf("Proof.Key doesn't match key recovered from the steps")
	}

	// TODO: grab value and hexremainer from the last step

	// first approach: let's go from top to bottom validating the hash matches expectations at each step
	// expected := rootHash[:]
	// for _, step := range proof.Steps {
	// 	// calculate hash of this level, make sure it is expected

	// 	// find hash of next link and set expected
	// }
	return nil
}

// buildProof annotates the path of proofs, with the child we followed at each step
func buildProof(key, value []byte, path []Step) (*Proof, error) {
	hexkey := keybytesToHex(key)
	fmt.Printf("hexkey: %X (%s)\n", hexkey, string(key))

	for i, p := range path {
		switch t := p.Step.(type) {
		case *shortNode:
			// remove the prefix and continue
			if len(hexkey) < len(t.Key) || !bytes.Equal(t.Key, hexkey[:len(t.Key)]) {
				return nil, fmt.Errorf("Shortnode prefix %X doesn't match key %X", t.Key, hexkey)
			}
			fmt.Printf("short: %X\n", t.Key)
			hexkey = hexkey[len(t.Key):]
		case *fullNode:
			fmt.Printf("next: %X\n", hexkey[0])
			idx := int(hexkey[0])
			hexkey = hexkey[1:]
			path[i].Index = idx
		default:
			return nil, fmt.Errorf("Unknown type: %T", p)
		}
	}

	// let's do a sanity check here (for me understanding)
	// if hexkey is empty, last ref is a valueNode
	// if it is non-empty, last ref is a shortNode with Key=hexkey, ref to valueNode
	// in both cases, the valueNode should contain value

	proof := Proof{
		Steps:        path,
		Key:          key,
		Value:        value,
		HexRemainder: hexkey,
	}

	return &proof, nil
}

// ProofRecorder is used to help us grab proofs
type ProofRecorder struct {
	path []Step
}

var _ ethdb.Putter = (*ProofRecorder)(nil)

func (p *ProofRecorder) Put(hash, value []byte) error {
	step, err := decodeNode(hash, value, 0)
	if err != nil {
		return err
	}
	p.path = append(p.path, Step{Step: step, Hash: hash})
	return nil
}

func (p *ProofRecorder) Path() []Step {
	return p.path
}
