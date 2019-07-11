package proof

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

type Node interface{}

// ComputeProof returns the proof value for a key in given trie. Returned path
// is the way from the value to the root of the tree.
func ComputeProof(tr *trie.Trie, key []byte) (value []byte, path []Node, err error) {
	db := ethdb.NewMemDatabase()

	val := tr.Get(key)
	if val == nil {
		return nil, nil, fmt.Errorf("No value found for key %X", key)
	}

	if err := tr.Prove(key, 0, db); err != nil {
		return nil, nil, err
	}

	rootHash := tr.Root()
	fmt.Printf("Root:  %X\n", rootHash)

	for _, key := range db.Keys() {
		fmt.Printf("  %X\n", key)
	}

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
		keyrest, cld := get(n, keyHex)
		fmt.Printf("Got node %T cld %T\n", n, cld)

		if cld != nil {
			path = append(path, cld)
		}

		switch cld := cld.(type) {
		case nil:
			// The trie doesn't contain the key.
			return nil, nil, fmt.Errorf("Key missing in proof")
		case hashNode:
			keyHex = keyrest
			copy(wantHash[:], cld)
		case valueNode:
			return val, path, nil
		}
	}

	return val, path, nil
}

// 	// FString returns the string representation of the node.
// 	FString(string) string

// 	// Cache returns the hash of the node.
// 	Cache() (hashNode []byte, dirty bool)
// }

func get(tn Node, key []byte) ([]byte, Node) {
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
