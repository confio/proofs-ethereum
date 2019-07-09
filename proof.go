package proof

import (
	"github.com/ethereum/go-ethereum/trie"
)

// ComputeProof returns the proof value for a key in given trie. Returned path
// is the way from the value to the root of the tree.
func ComputeProof(tr *trie.Trie, key []byte) (value []byte, path []Node, err error) {
	panic("todo")
}

type Node interface {
	// FString returns the string representation of the node.
	FString(string) string

	// Cache returns the hash of the node.
	Cache() (hashNode []byte, dirty bool)
}
