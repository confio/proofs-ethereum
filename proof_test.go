package proof

import (
	"bytes"
	"fmt"
	// predictable "random" is good for testing :)
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

func TestEthTrie(t *testing.T) {
	cases := map[string]struct {
		items    []string
		query    string
		numSteps int
		isError  bool
	}{
		"two levels": {
			items:    []string{"a", "B", "7", "ASDF", "    000    ", "fooBAR"},
			query:    "fooBAR",
			numSteps: 2,
		},
		"short node": {
			items:    []string{"aaaaaaa1", "aaaa2", "aaaaaaaaaaaaab", "C"},
			query:    "aaaaaaaaaaaaab",
			numSteps: 5,
		},
		"embeded full node": {
			// bytes 41, 42, 61, 62 (full node include full node with value refs)
			items:    []string{"a", "b", "A", "B"},
			query:    "a",
			numSteps: 1,
		},
		"ends with value node": {
			// bytes 41, 42, 61, 62 first nibble fullnode, follow 4x leads to lots of data, so A is value node
			// correction - we only get shortnode with key = 16 to repr value node
			items:    []string{"a", "b", "A", "BBB", "CDUHIUHIUH", "DJOIOIHFW", "EHFKHEHOHWOHF", "BDED"},
			query:    "A",
			numSteps: 2,
		},
		"only short node": {
			items:    []string{"1"},
			query:    "1",
			numSteps: 1,
		},
		"longer short node": {
			items:    []string{"fooled"},
			query:    "fooled",
			numSteps: 1,
		},
		"longest short node": {
			items:    []string{"more than 16 bytes here..."},
			query:    "more than 16 bytes here...",
			numSteps: 1,
		},
		"invalid query": {
			items:   []string{"aaaaaaa1", "aaaa2", "aaaaaaaaaaaaab", "C"},
			query:   "aaaaaaaaaa",
			isError: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			db := ethdb.NewMemDatabase()
			tr, err := trie.New(common.BytesToHash(nil), trie.NewDatabase(db))
			if err != nil {
				t.Fatalf("cannot create an empty trie: %s", err)
			}

			for _, s := range tc.items {
				b := []byte(s)
				tr.Update(b, b) // key == value
			}

			hash, err := tr.Commit(nil)
			if err != nil {
				t.Fatalf("cannot commit: %s", err)
			} else {
				t.Logf("commit hash of the trie: %X", hash)
			}

			proof, err := ComputeProof(tr, []byte(tc.query))
			if tc.isError {
				if err == nil {
					t.Fatalf("Expected error, but was <nil>")
				}
				return
			}

			if err != nil {
				t.Fatalf("Error: %+v", err)
			}
			for _, p := range proof.Steps {
				t.Logf("-> %X\n", p.Hash)
				t.Logf("---> (%d) %s\n", p.Index, p.Step)
			}

			val := string(proof.Value)
			if val != tc.query {
				t.Fatalf("invalid value: %s", val)
			}
			if len(proof.Steps) != tc.numSteps {
				t.Fatalf("Unexpected path length %d (expected %d)", len(proof.Steps), tc.numSteps)
			}

			recovered := proof.RecoverKey()
			if string(recovered) != tc.query {
				t.Fatalf("Recovered key %s doesn't match query %s\n", string(recovered), tc.query)
			}

			// see if we can verify this proof
			err = VerifyProof(proof, hash)
			if err != nil {
				t.Fatalf("Invalid proof %+v", err)
			}
		})

	}
}

// TestRandomTrie is basically a fuzz-tester
// If there is an error, it should dump out enough info to create a targetted case above
func TestRandomTries(t *testing.T) {
	runs := 20
	size := 5000

	for i := 0; i<runs; i++ {
		t.Run(fmt.Sprintf("Run %d", i), func (t *testing.T) {
			tr, keys := randomTrie(t, size)
			// grab one of the last ones (random) to query
			query := keys[len(keys)-3]

			proof, err := ComputeProof(tr, query.k)
			if err != nil {
				t.Fatalf("ComputeProof: %+v", err)
			}
			t.Logf("Path length %d", len(proof.Steps))
			if !bytes.Equal(query.v, proof.Value) {
				t.Fatalf("invalid value: %X (expected %X)", proof.Value, query.v)
			}

			recovered := proof.RecoverKey()
			if !bytes.Equal(query.k, recovered) {
				t.Fatalf("Recovered key %X doesn't match query %X\n", recovered, query.k)
			}

			err = VerifyProof(proof, tr.Hash())
			if err != nil {
				t.Fatalf("Invalid proof %+v", err)
			}
		})
	}
}

type kv struct {
	k []byte
	v []byte
}

func randomTrie(t *testing.T, n int) (*trie.Trie, []kv) {
	db := ethdb.NewMemDatabase()
	tr, err := trie.New(common.BytesToHash(nil), trie.NewDatabase(db))
	if err != nil {
		t.Fatalf("cannot create an empty trie: %s", err)
	}

	var vals []kv
	for i := byte(0); i < 100; i++ {
		value := kv{k: common.LeftPadBytes([]byte{i}, 32), v: []byte{i}}
		tr.Update(value.k, value.v)

		value2 := kv{k: common.LeftPadBytes([]byte{i + 10}, 32), v: []byte{i}}
		tr.Update(value2.k, value2.v)

		vals = append(vals, value, value2)
	}
	for i := 0; i < n; i++ {
		value := kv{k: randBytes(32), v: randBytes(20)}
		tr.Update(value.k, value.v)
		vals = append(vals, value)
	}

	_, err = tr.Commit(nil)
	if err != nil {
		t.Fatalf("cannot commit: %s", err)
	}

	return tr, vals
}

func randBytes(n int) []byte {
	r := make([]byte, n)
	rand.Read(r)
	return r
}
