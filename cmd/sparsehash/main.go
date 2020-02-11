// sparsehash is a sample application using sparsehash. It will calculate and report
// file hashes in a format similar to md5sum, etc.
package main

import (
	"flag"
	"fmt"
	"hash"
	"log"
	"os"
	"path/filepath"

	"github.com/oliverpool/sparsehash"
	"github.com/spaolacci/murmur3"
)

func newMurmur3() hash.Hash {
	return murmur3.New128()
}

func main() {
	// defer profile.Start(profile.MemProfile).Stop()
	flag.Parse()
	files := flag.Args()

	if len(files) == 0 {
		fmt.Println("Usage: sparsehash [filenames]")
		return
	}

	h := sparsehash.New(newMurmur3)
	for _, file := range files {
		err := filepath.Walk(file, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if f.IsDir() {
				return nil
			}

			hash, err := h.SumFile(path)
			if err != nil {
				return err
			}
			_, err = fmt.Printf("%016x  %s\n", hash, path)
			return err
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}
