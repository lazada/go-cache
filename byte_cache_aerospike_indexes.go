package cache

import (
	"github.com/aerospike/aerospike-client-go"
)

// AerospikeIndex contains info about aerospike index
type AerospikeIndex struct {
	SetName   string
	IndexName string
	IndexType aerospike.IndexType
}
