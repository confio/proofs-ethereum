package proof

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

func TestEthTrie(t *testing.T) {
	db := ethdb.NewMemDatabase()
	tr, err := trie.New(common.BytesToHash(nil), trie.NewDatabase(db))
	if err != nil {
		t.Fatalf("cannot create an empty trie: %s", err)
	}
	t.Logf("empty trie root hash: %x", tr.Root())

	items := []string{"a", "B", "7", "ASDF", "    000    ", "fooBAR"}

	for _, s := range items {
		b := []byte(s)
		tr.Update(b, b) // key == value
	}

	if hash, err := tr.Commit(nil); err != nil {
		t.Fatalf("cannot commit: %s", err)
	} else {
		t.Logf("commit hash of the trie: %X", hash)
	}

	k := "fooBAR"
	val, path, err := ComputeProof(tr, []byte(k))
	if err != nil {
		t.Fatalf("Error: %+v", err)
	}
	if string(val) != k {
		t.Fatalf("invalid value: %s", string(val))
	}
	if len(path) != 2 {
		t.Fatalf("Unexpected path length")
	}

	//for _, s := range items {
	//	b := []byte(s)
	//	val, _, err := trie.VerifyProof(tr.Hash(), b, db)
	//	if err != nil {
	//		t.Fatalf("cannot verify proof: %s", err)
	//	}
	//	t.Logf("value of the key %q: %q", s, val)
	//}

	// it := tr.NodeIterator(nil)
	// for {
	// 	if err := it.Error(); err != nil {
	// 		t.Fatalf("iterator failed: %s", err)
	// 	}

	// 	if it.Leaf() {
	// 		for i, p := range it.LeafProof() {
	// 			t.Logf("%x: leaf %q proof: %2d %x", it.Path(), it.LeafKey(), i, p)
	// 		}
	// 	}
	// 	t.Logf("%x: node hash: %x", it.Path(), it.Hash())

	// 	if !it.Next(true) {
	// 		break
	// 	}
	// }
}
