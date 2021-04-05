// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by consensys/gnark-crypto DO NOT EDIT

// Package mimc provides MiMC hash function using Miyaguchi–Preneel construction.
package mimc

import (
	"hash"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"golang.org/x/crypto/sha3"
)

const mimcNbRounds = 91

// BlockSize size that mimc consumes
const BlockSize = 32 + 16

// Params constants for the mimc hash function
type Params []fr.Element

// NewParams creates new mimc object
func NewParams(seed string) Params {

	// set the constants
	res := make(Params, mimcNbRounds)

	rnd := sha3.Sum256([]byte(seed))
	value := new(big.Int).SetBytes(rnd[:])

	for i := 0; i < mimcNbRounds; i++ {
		rnd = sha3.Sum256(value.Bytes())
		value.SetBytes(rnd[:])
		res[i].SetBigInt(value)
	}

	return res
}

// digest represents the partial evaluation of the checksum
// along with the params of the mimc function
type digest struct {
	Params Params
	h      fr.Element
	data   []byte // data to hash
}

// NewMiMC returns a MiMCImpl object, pure-go reference implementation
func NewMiMC(seed string) hash.Hash {
	d := new(digest)
	params := NewParams(seed)
	//d.Reset()
	d.Params = params
	d.Reset()
	return d
}

// Reset resets the Hash to its initial state.
func (d *digest) Reset() {
	d.data = nil
	d.h = fr.Element{0, 0, 0, 0}
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (d *digest) Sum(b []byte) []byte {
	buffer := d.checksum()
	d.data = nil // flush the data already hashed
	hash := buffer.Bytes()
	b = append(b, hash[:]...)
	return b
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (d *digest) Size() int {
	return BlockSize
}

// BlockSize returns the number of bytes Sum will return.
func (d *digest) BlockSize() int {
	return BlockSize
}

// Write (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (d *digest) Write(p []byte) (n int, err error) {
	n = len(p)
	d.data = append(d.data, p...)
	return
}

// Hash hash using Miyaguchi–Preneel:
// https://en.wikipedia.org/wiki/One-way_compression_function
// The XOR operation is replaced by field addition, data is in Montgomery form
func (d *digest) checksum() fr.Element {

	var buffer [BlockSize]byte
	var x fr.Element

	// if data size is not multiple of BlockSizes we padd:
	// .. || 0xaf8 -> .. || 0x0000...0af8
	if len(d.data)%BlockSize != 0 {
		q := len(d.data) / BlockSize
		r := len(d.data) % BlockSize
		sliceq := make([]byte, q*BlockSize)
		copy(sliceq, d.data)
		slicer := make([]byte, r)
		copy(slicer, d.data[q*BlockSize:])
		sliceremainder := make([]byte, BlockSize-r)
		d.data = append(sliceq, sliceremainder...)
		d.data = append(d.data, slicer...)
	}

	if len(d.data) == 0 {
		d.data = make([]byte, 32)
	}

	nbChunks := len(d.data) / BlockSize

	for i := 0; i < nbChunks; i++ {
		copy(buffer[:], d.data[i*BlockSize:(i+1)*BlockSize])
		x.SetBytes(buffer[:])
		d.encrypt(x)
		d.h.Add(&x, &d.h)
	}

	return d.h
}

// plain execution of a mimc run
// m: message
// k: encryption key
func (d *digest) encrypt(m fr.Element) {

	for i := 0; i < len(d.Params); i++ {
		// m = (m+k+c)^5
		var tmp fr.Element
		tmp.Add(&m, &d.h).Add(&tmp, &d.Params[i])
		m.Square(&tmp).
			Square(&m).
			Mul(&m, &tmp)
	}
	m.Add(&m, &d.h)
	d.h = m
}

// Sum computes the mimc hash of msg from seed
func Sum(seed string, msg []byte) ([]byte, error) {
	params := NewParams(seed)
	var d digest
	d.Params = params
	if _, err := d.Write(msg); err != nil {
		return nil, err
	}
	h := d.checksum()
	bytes := h.Bytes()
	return bytes[:], nil
}
