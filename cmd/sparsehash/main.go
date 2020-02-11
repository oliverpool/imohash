// sparsehash is a sample application using sparsehash. It will calculate and report
// file hashes in a format similar to md5sum, etc.
package main

import (
	"fmt"
	"hash"
	"log"
	"os"

	"github.com/kalafut/sparsehash"
	"github.com/spaolacci/murmur3"
)

func newMurmur3() hash.Hash {
	return murmur3.New128()
}

func main() {
	files := os.Args[1:]

	if len(files) == 0 {
		fmt.Println("Usage: sparsehash [filenames]")
		return
	}

	h := sparsehash.New(newMurmur3)
	for _, file := range files {
		hash, err := h.SumFile(file)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%016x  %s\n", hash, file)
	}
}
