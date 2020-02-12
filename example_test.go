package sparsehash_test

import (
	"fmt"
	"hash"

	"github.com/oliverpool/sparsehash"
	"github.com/spaolacci/murmur3"
)

func Example() {
	hasher := sparsehash.New(func() hash.Hash {
		// From github.com/spaolacci/murmur3
		// can be anything implementing the hash.Hash interface
		return murmur3.New128()
	})
	hash, err := hasher.SumFile("Makefile")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%016x\n", hash)
	// Output: 15cd63930005d0b6ee5ebc8a3f6483f2
}
