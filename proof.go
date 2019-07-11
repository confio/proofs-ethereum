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

// format is ProofStep - FullNode or ShortNode
// child link either ValueNode or HashNode
//  - hashNode links to next step
//  - valueNode is the end

// ComputeProof returns the proof value for a key in given trie. Returned path
// is the way from the value to the root of the tree.
func ComputeProof(tr *trie.Trie, key []byte) (value []byte, path []PathStep, err error) {
	db := ethdb.NewMemDatabase()

	val := tr.Get(key)
	if val == nil {
		return nil, nil, fmt.Errorf("No value found for key %X", key)
	}

	if err := tr.Prove(key, 0, db); err != nil {
		return nil, nil, err
	}

	rootHash := tr.Root()
	fmt.Printf("Query: %X\n", key)
	fmt.Printf("Root:  %X\n", rootHash)

	wantHash := rootHash
	keyHex := keybytesToHex(key)
	for i := 0; ; i++ {
		buf, _ := db.Get(wantHash[:])
		if buf == nil {
			return nil, nil, fmt.Errorf("proof node %d (hash %064x) missing", i, wantHash)
		}
		n, err := decodeNode(wantHash[:], buf, 0)
		if err != nil {
			return nil, nil, fmt.Errorf("bad proof node %d: %v", i, err)
		}
		keyrest, child := get(n, keyHex)
		fmt.Printf("Got node %T child %T\n", n, child)

		if child == nil {
			// The trie doesn't contain the key.
			return nil, nil, fmt.Errorf("Key missing in proof")
		}
		path = append(path, n)

		// both of these are hex
		switch child := child.(type) {
		case hashNode:
			fmt.Printf("-> Following hash to %s\n", child)
			keyHex = keyrest
			copy(wantHash[:], child)
		case valueNode:
			fmt.Printf("-> Got value %s\n", child)
			return child, path, nil
		}
	}

	return val, path, nil
}

// 	// FString returns the string representation of the node.
// 	FString(string) string

// 	// Cache returns the hash of the node.
// 	Cache() (hashNode []byte, dirty bool)
// }

func get(step PathStep, key []byte) ([]byte, Link) {
	var tn Link = step
	for {
		switch n := tn.(type) {
		case *shortNode:
			if len(key) < len(n.Key) || !bytes.Equal(n.Key, key[:len(n.Key)]) {
				return nil, nil
			}
			tn = n.Val
			key = key[len(n.Key):]
		case *fullNode:
			tn = n.Children[key[0]]
			key = key[1:]
		case hashNode:
			return key, n
		case nil:
			return key, nil
		case valueNode:
			return nil, n
		default:
			panic(fmt.Sprintf("%T: invalid node: %v", tn, tn))
		}
	}
}
