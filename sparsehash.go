// Package sparsehash implements a fast, constant-time hash for files. It hashes three fixed
// samples of the file (beginning, middle and end) and uses file size to construct the hash.
//
// For more information, including important caveats on usage, consult https://github.com/oliverpool/sparsehash.
package sparsehash

import (
	"bytes"
	"encoding/binary"
	"hash"
	"io"
	"os"
)

// HashSize is the size of byte array produced by the Sum function
const HashSize = 16

var emptyArray = [HashSize]byte{}

// Hasher respresents a sparse hasher
type Hasher struct {
	// SubHasher is the actual hash function through which the samples will be hashed (murmur3.New128() for instance)
	SubHasher func() hash.Hash
	// Size of the 3 samples to actually hash (at the beginning, middle and end of the input)
	// A SampleSize of 0 will hash alls inputs entirely
	SampleSize int64
	// Minimum size of the input to only hash the 3 samples (smaller inputs will be hashed entirely)
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
func (h Hasher) SumBytes(data []byte) ([HashSize]byte, error) {
	sr := io.NewSectionReader(bytes.NewReader(data), 0, int64(len(data)))

	return h.Sum(sr)
}

// SumFile hashes a file sparsely
func (h Hasher) SumFile(filename string) ([HashSize]byte, error) {
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
func (h Hasher) Sum(f *io.SectionReader) ([HashSize]byte, error) {
	var hasher hash.Hash
	var err error

	if f.Size() < h.SizeThreshold || h.SampleSize < 1 {
		hasher, err = h.hashAll(f)
	} else {
		hasher, err = h.hashSamples(f)
	}
	hash := make([]byte, 0, HashSize)
	hash = hasher.Sum(hash)

	binary.PutUvarint(hash, uint64(f.Size()))

	var result [HashSize]byte
	copy(result[:], hash)
	return result, err
}

func (h Hasher) hashAll(f *io.SectionReader) (hash.Hash, error) {
	hasher := h.SubHasher()
	_, err := io.Copy(hasher, f)
	return hasher, err
}

func (h Hasher) hashSamples(f *io.SectionReader) (hash.Hash, error) {
	var err error

	// The following functions do nothing if err != nil
	hasher := h.SubHasher()
	hWrite := func(p []byte) {
		if err != nil {
			return
		}
		_, err = hasher.Write(p)
	}

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

	buffer := make([]byte, h.SampleSize)
	fRead(buffer)
	hWrite(buffer)
	fSeek(f.Size()/2, 0)
	fRead(buffer)
	hWrite(buffer)
	fSeek(-h.SampleSize, 2)
	fRead(buffer)
	hWrite(buffer)

	return hasher, err
}
