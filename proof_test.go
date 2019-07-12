package proof

import (
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
