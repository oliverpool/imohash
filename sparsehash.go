// Package sparsehash implements a fast, constant-time hash for files. It is based atop
// murmurhash3 and uses file size and sample data to construct the hash.
//
// For more information, including important caveats on usage, consult https://github.com/kalafut/sparsehash.
package sparsehash

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/spaolacci/murmur3"
)

// HashSize is the size of the resulting array
const HashSize = 16

// Files smaller than this will be hashed in their entirety.
const SampleThreshold = 128 * 1024
const SampleSize = 16 * 1024

var emptyArray = [HashSize]byte{}

type sparsehash struct {
	hasher          murmur3.Hash128
	sampleSize      int
	sampleThreshold int
	bytesAdded      int
}

// New returns a new sparsehash using the default sample size
// and sample threshhold values.
func New() sparsehash {
	return NewCustom(SampleSize, SampleThreshold)
}

// NewCustom returns a new sparsehash using the provided sample size
// and sample threshhold values. The entire file will be hashed
// (i.e. no sampling), if sampleSize < 1.
func NewCustom(sampleSize, sampleThreshold int) sparsehash {
	h := sparsehash{
		hasher:          murmur3.New128(),
		sampleSize:      sampleSize,
		sampleThreshold: sampleThreshold,
	}

	return h
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
func (imo *sparsehash) Sum(data []byte) [HashSize]byte {
	sr := io.NewSectionReader(bytes.NewReader(data), 0, int64(len(data)))

	return imo.hashCore(sr)
}

// SumFile hashes a file using using the sparsehash parameters.
func (imo *sparsehash) SumFile(filename string) ([HashSize]byte, error) {
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
func (imo *sparsehash) hashCore(f *io.SectionReader) [HashSize]byte {
	var result [HashSize]byte

	imo.hasher.Reset()

	if f.Size() < int64(imo.sampleThreshold) || imo.sampleSize < 1 {
		buffer := make([]byte, f.Size())
		f.Read(buffer)
		imo.hasher.Write(buffer)
	} else {
		buffer := make([]byte, imo.sampleSize)
		f.Read(buffer)
		imo.hasher.Write(buffer)
		f.Seek(f.Size()/2, 0)
		f.Read(buffer)
		imo.hasher.Write(buffer)
		f.Seek(int64(-imo.sampleSize), 2)
		f.Read(buffer)
		imo.hasher.Write(buffer)
	}

	hash := imo.hasher.Sum(nil)
	fmt.Println(len(hash), hash)

	binary.PutUvarint(hash, uint64(f.Size()))
	fmt.Println("2", len(hash), hash)
	copy(result[:], hash)

	return result
}
