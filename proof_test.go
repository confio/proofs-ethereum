package proof

import (
	"crypto/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/trie"
)

// randomEthTrie generates a Ethereum trie structure with random values.
// Returned is trie and an index of all created key-value pairs.
func randomEthTrie(size int) (*trie.Trie, map[string]*kv) {
	var t trie.Trie
	vals := make(map[string]*kv)

	// TODO XXX why is the below creating 200 keys that are build using not
	// a random value but an incrementing value?
	// https://github.com/ethereum/go-ethereum/blob/f5d89cdb72c1e82e9deb54754bef8dd20bf12591/trie/proof_test.go#L203

	for i := byte(0); i < 100; i++ {
		value := &kv{
			k: common.LeftPadBytes([]byte{i}, 32),
			v: []byte{i},
		}
		t.Update(value.k, value.v)
		vals[string(value.k)] = value

		value2 := &kv{
			k: common.LeftPadBytes([]byte{i + 10}, 32),
			v: []byte{i},
		}
		t.Update(value2.k, value2.v)
		vals[string(value2.k)] = value2
	}
	for i := 0; i < size; i++ {
		value := &kv{
			k: randBytes(32),
			v: randBytes(20),
		}
		t.Update(value.k, value.v)
		vals[string(value.k)] = value
	}
	return &t, vals
}

type kv struct {
	k []byte
	v []byte
}

func randBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}
