package proof

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

func TestEthTrie(t *testing.T) {
	cases := map[string]struct {
		items []string
		query string
	}{
		"two levels": {
			items: []string{"a", "B", "7", "ASDF", "    000    ", "fooBAR"},
			query: "fooBAR",
		},
		"short node": {
			items: []string{"aaaaaaa1", "aaaa2", "aaaaaaaaaaaaab", "C"},
			query: "aaaaaaaaaaaaab",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			db := ethdb.NewMemDatabase()
			tr, err := trie.New(common.BytesToHash(nil), trie.NewDatabase(db))
			if err != nil {
				t.Fatalf("cannot create an empty trie: %s", err)
			}
			t.Logf("empty trie root hash: %x", tr.Root())

			for _, s := range tc.items {
				b := []byte(s)
				tr.Update(b, b) // key == value
			}

			if hash, err := tr.Commit(nil); err != nil {
				t.Fatalf("cannot commit: %s", err)
			} else {
				t.Logf("commit hash of the trie: %X", hash)
			}

			val, path, err := ComputeProof(tr, []byte(tc.query))
			if err != nil {
				t.Fatalf("Error: %+v", err)
			}
			if string(val) != tc.query {
				t.Fatalf("invalid value: %s", string(val))
			}
			if len(path) < 2 {
				t.Fatalf("Unexpected path length %d", len(path))
			}
			for _, p := range path {
				fmt.Printf("-> %s\n", p)
			}
		})

	}

}
