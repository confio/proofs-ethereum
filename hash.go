package proof

import "github.com/ethereum/go-ethereum/rlp"

func hashShortNode(n *shortNode) []byte {
	h := newHasher(0, 0, nil)
	// h.tmp.Reset()

	if _, ok := n.Val.(valueNode); !ok {
		panic("only implemented for value nodes")
	}

	collapsed := n.copy()
	collapsed.Key = hexToCompact(n.Key)

	if err := rlp.Encode(&h.tmp, collapsed); err != nil {
		panic("encode error: " + err.Error())
	}
	hash := h.makeHashNode(h.tmp)
	return hash
}
