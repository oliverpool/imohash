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
func (h *Hasher) SumBytes(data []byte) ([HashSize]byte, error) {
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
	return h.Sum(sr)
}

// Sum hashes a SectionReader using the sparsehash parameters.
func (h *Hasher) Sum(f *io.SectionReader) ([HashSize]byte, error) {
	var err error

	// The following functions do nothing if err != nil
	fRead := func(p []byte) {
		if err != nil {
			return
		}
		_, err = f.Read(p)
		if err == io.EOF {
			err = nil
		}
	}
	fSeek := func(offset int64, whence int) {
		if err != nil {
			return
		}
		_, err = f.Seek(offset, whence)
	}

	hasher := h.SubHasher()
	hWrite := func(p []byte) {
		if err != nil {
			return
		}
		_, err = hasher.Write(p)
	}

	if f.Size() < h.SizeThreshold || h.SampleSize < 1 {
		buffer := make([]byte, f.Size())
		fRead(buffer)
		hWrite(buffer)
	} else {
		buffer := make([]byte, h.SampleSize)
		fRead(buffer)
		hWrite(buffer)
		fSeek((f.Size())/2, 0)
		fRead(buffer)
		hWrite(buffer)
		fSeek(-h.SampleSize, 2)
		fRead(buffer)
		hWrite(buffer)
	}

	r := make([]byte, 0, HashSize)
	hash := hasher.Sum(r)

	binary.PutUvarint(hash, uint64(f.Size()))

	var result [HashSize]byte
	copy(result[:], hash)
	return result, err
}
