// Package sparsehash implements a fast, constant-time hash for files. It is based atop
// murmurhash3 and uses file size and sample data to construct the hash.
//
// For more information, including important caveats on usage, consult https://github.com/kalafut/sparsehash.
package sparsehash

import (
	"bytes"
	"encoding/binary"
	"hash"
	"io"
	"os"
)

// HashSize is the size of the resulting array
const HashSize = 16

var emptyArray = [HashSize]byte{}

// Hasher respresents a sparse hasher
type Hasher struct {
	SubHasher func() hash.Hash
	// 3 samples of SampleSize will be hashed (at the beginning, middle and end of the input file)
	SampleSize int64
	// Files smaller than SizeThreshold will be hashed in their entirety.
	SizeThreshold int64
}

// New returns a new sparsehash using the specified subhasher as hasher, 16K as sample size
// and 128K as threshhold values.
func New(subhasher func() hash.Hash) Hasher {
	return Hasher{
		SubHasher:     subhasher,
		SampleSize:    16 * 1024,
		SizeThreshold: 128 * 1024,
	}
}

// SumBytes hashes a byte slice using the sparsehash parameters.
func (h *Hasher) SumBytes(data []byte) [HashSize]byte {
	sr := io.NewSectionReader(bytes.NewReader(data), 0, int64(len(data)))

	return h.Sum(sr)
}

// SumFile hashes a file sparsely
func (h *Hasher) SumFile(filename string) ([HashSize]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return emptyArray, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return emptyArray, err
	}
	sr := io.NewSectionReader(f, 0, fi.Size())
	return h.Sum(sr), nil
}

// Sum hashes a SectionReader using the sparsehash parameters.
func (h *Hasher) Sum(f *io.SectionReader) [HashSize]byte {
	hasher := h.SubHasher()

	if f.Size() < h.SizeThreshold || h.SampleSize < 1 {
		buffer := make([]byte, f.Size())
		f.Read(buffer)
		hasher.Write(buffer)
	} else {
		buffer := make([]byte, h.SampleSize)
		f.Read(buffer)
		hasher.Write(buffer)
		f.Seek((f.Size())/2, 0)
		f.Read(buffer)
		hasher.Write(buffer)
		f.Seek(-h.SampleSize, 2)
		f.Read(buffer)
		hasher.Write(buffer)
	}

	r := make([]byte, 0, HashSize)
	hash := hasher.Sum(r)

	binary.PutUvarint(hash, uint64(f.Size()))

	var result [HashSize]byte
	copy(result[:], hash)
	return result
}
