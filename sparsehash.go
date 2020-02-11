// Package sparsehash implements a fast, constant-time hash for files. It is based atop
// murmurhash3 and uses file size and sample data to construct the hash.
//
// For more information, including important caveats on usage, consult https://github.com/kalafut/sparsehash.
package sparsehash

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/spaolacci/murmur3"
)

// HashSize is the size of the resulting array
const HashSize = 16

var emptyArray = [HashSize]byte{}

// Hasher respresents a sparse hasher
type Hasher struct {
	SubHasher  func() hash.Hash
	sampleSize int
	// Files smaller than sampleThreshold this will be hashed in their entirety.
	sampleThreshold int
	bytesAdded      int
}

func newMurmur3() hash.Hash {
	return murmur3.New128()
}

// New returns a new sparsehash using murmur3 as hasher, 16K as sample size
// and 128K as threshhold values.
func New() Hasher {
	return Hasher{
		SubHasher:       newMurmur3,
		sampleSize:      16 * 1024,
		sampleThreshold: 128 * 1024,
	}
}

// SumFile hashes a file using default sample parameters.
func SumFile(filename string) ([HashSize]byte, error) {
	imo := New()
	return imo.SumFile(filename)
}

// Sum hashes a byte slice using default sample parameters.
func Sum(data []byte) [HashSize]byte {
	imo := New()
	return imo.Sum(data)
}

// Sum hashes a byte slice using the sparsehash parameters.
func (imo *Hasher) Sum(data []byte) [HashSize]byte {
	sr := io.NewSectionReader(bytes.NewReader(data), 0, int64(len(data)))

	return imo.hashCore(sr)
}

// SumFile hashes a file using using the sparsehash parameters.
func (imo *Hasher) SumFile(filename string) ([HashSize]byte, error) {
	f, err := os.Open(filename)
	defer f.Close()

	if err != nil {
		return emptyArray, err
	}

	fi, err := f.Stat()
	if err != nil {
		return emptyArray, err
	}
	sr := io.NewSectionReader(f, 0, fi.Size())
	return imo.hashCore(sr), nil
}

// hashCore hashes a SectionReader using the sparsehash parameters.
func (imo *Hasher) hashCore(f *io.SectionReader) [HashSize]byte {
	var result [HashSize]byte

	hasher := imo.SubHasher()

	if f.Size() < int64(imo.sampleThreshold) || imo.sampleSize < 1 {
		buffer := make([]byte, f.Size())
		f.Read(buffer)
		hasher.Write(buffer)
	} else {
		buffer := make([]byte, imo.sampleSize)
		f.Read(buffer)
		hasher.Write(buffer)
		f.Seek(f.Size()/2, 0)
		f.Read(buffer)
		hasher.Write(buffer)
		f.Seek(int64(-imo.sampleSize), 2)
		f.Read(buffer)
		hasher.Write(buffer)
	}

	hash := hasher.Sum(nil)
	fmt.Println(len(hash), hash)

	binary.PutUvarint(hash, uint64(f.Size()))
	fmt.Println("2", len(hash), hash)
	copy(result[:], hash)

	return result
}
