// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package itmod

import "errors"

// A little-endia bitstream.
type bitstream struct {
	source  []byte
	readPos int

	// A buffer of 64 bits.
	buffer uint64

	// Number of bits in the buffer.
	buffered int
}

var ErrEndOfStream = errors.New("end of stream")
var ErrBadParam = errors.New("bad param")

func createBitstream(source []byte) bitstream {
	return bitstream{source: source}
}

func (bs *bitstream) read(width int) (uint32, error) {
	if width < 0 || width >= 32 {
		return 0, ErrBadParam
	}

	for bs.buffered < width {
		if bs.readPos >= len(bs.source) {
			return 0, ErrEndOfStream
		}
		bs.buffer |= uint64(bs.source[bs.readPos]) << bs.buffered
		bs.readPos++
		bs.buffered += 8
	}

	result := uint32(bs.buffer & ((1 << width) - 1))
	bs.buffer >>= width
	bs.buffered -= width

	return result, nil
}
