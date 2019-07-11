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

			if hash, err := tr.Commit(nil); err != nil {
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
				t.Logf("-> (%d) %s\n", p.Index, p.Step)
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
		})

	}

}
