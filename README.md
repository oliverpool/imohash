# sparsehash [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/oliverpool/sparsehash)

sparsehash is a fast, constant-time hashing library for Go. It uses sampling to calculate hashes quickly, regardless of file size.

It works by hashing fixed-size chunks of data from the beginning, middle and end of a file using a provided hasher.

[sparsehash](https://github.com/oliverpool/sparsehash/blob/master/cmd/sparsehash/main.go) is
a sample application to hash files from the command line, similar to md5sum.

sparsehash is forked from [imohash](https://github.com/kalafut/imohash).

**The file size is not integrated in the hash** (you should compare it yourself).

### Uses

Because sparsehash only reads a small portion of a file's data, it is very fast and
well suited to file synchronization and deduplication, especially over a fairly
slow but reliable network. A need to manage media (photos and video) over Wi-Fi between a NAS
and multiple family computers is how the library was born.

If you just need to check whether two files are the same, and understand the
limitations that sampling imposes (see below), sparsehash may be a good fit.

### Misuses

Because sparsehash only reads a small portion of a file's data, it is not suitable
for:

- file verification or integrity monitoring (in case of unreliable transmission)
- cases where specific bits are manipulated in a file
- anything cryptographic

## Installation

`go get github.com/oliverpool/sparsehash/...`

The API is described in the [package documentation](https://pkg.go.dev/github.com/oliverpool/sparsehash).


## Small file exemption
Small files are more likely to collide on size than large ones. They're also
probably more likely to change in subtle ways that sampling will miss (e.g.
editing a large text file). For this reason, sparsehash will simply hash the entire
file if it is less than 128K. This parameter is also configurable.

## Performance
The standard hash performance metrics make no sense for sparsehash since it's only
reading a limited set of the data. That said, the real-world performance is
very good. If you are working with large files and/or a slow network,
expect huge speedups. (**spoiler**: reading 48K is quicker than reading 500MB.)

## Credits
*  [imohash](https://github.com/kalafut/imohash) project, from which this module is a fork
