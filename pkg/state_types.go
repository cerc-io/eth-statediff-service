package statediff

// Wrapper types for state trie

import (
	"github.com/ethereum/go-ethereum/core/state"
)

// StateNode holds the data for a single state diff node
type StateNode struct {
	NodeType     NodeType      `json:"nodeType"        gencodec:"required"`
	Path         []byte        `json:"path"            gencodec:"required"`
	NodeValue    []byte        `json:"value"           gencodec:"required"`
	StorageNodes []StorageNode `json:"storage"`
	LeafKey      []byte        `json:"leafKey"`
}

// StorageNode holds the data for a single storage diff node
type StorageNode struct {
	NodeType  NodeType `json:"nodeType"        gencodec:"required"`
	Path      []byte   `json:"path"            gencodec:"required"`
	NodeValue []byte   `json:"value"           gencodec:"required"`
	LeafKey   []byte   `json:"leafKey"`
}

// AccountMap is a mapping of hex encoded path => account wrapper
type AccountMap map[string]accountWrapper

// accountWrapper is used to temporary associate the unpacked node with its raw values
type accountWrapper struct {
	Account   *state.Account
	NodeType  NodeType
	Path      []byte
	NodeValue []byte
	LeafKey   []byte
}

// NodeType for explicitly setting type of node
type NodeType string

const (
	Unknown   NodeType = "Unknown"
	Leaf      NodeType = "Leaf"
	Extension NodeType = "Extension"
	Branch    NodeType = "Branch"
	Removed   NodeType = "Removed" // used to represent pathes which have been emptied
)
