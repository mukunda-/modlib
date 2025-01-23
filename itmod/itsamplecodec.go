// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package itmod

import (
	"encoding/binary"
	"errors"
	"io"
)

/*
For encoding/decoding compressed samples stored in IT files.

Ported from OpenMPT and GreaseMonkey's code
https://github.com/OpenMPT/openmpt/blob/master/soundlib/ITCompression.cpp
https://github.com/iamgreaser/it2everything/blob/master/munch.py#L820
*/
type ItSampleCodec struct {
	// Set this if working with sources that are created with IT 2.15 or later.
	It215 bool

	// Decode/encode 16-bit samples.
	Is16 bool
}

var ErrDecodingError = errors.New("decoding error")

type itSampleCodecParams struct {
	lowerTab []int16
	upperTab []int16
	fetchA   int
	lowerB   int
	upperB   int
	defWidth int
	mask     int
}

// Algorithm parameters for 16-Bit samples
var itSampleCodecParams16 = itSampleCodecParams{
	lowerTab: []int16{0, -1, -3, -7, -15, -31, -56, -120, -248, -504, -1016, -2040, -4088, -8184, -16376, -32760, -32768},
	upperTab: []int16{0, 1, 3, 7, 15, 31, 55, 119, 247, 503, 1015, 2039, 4087, 8183, 16375, 32759, 32767},
	fetchA:   4,
	lowerB:   -8,
	upperB:   7,
	defWidth: 17,
	mask:     0xFFFF,
}

// Algorithm parameters for 8-Bit samples
var itSampleCodecParams8 = itSampleCodecParams{
	lowerTab: []int16{0, -1, -3, -7, -15, -31, -60, -124, -128},
	upperTab: []int16{0, 1, 3, 7, 15, 31, 59, 123, 127},
	fetchA:   3,
	lowerB:   -4,
	upperB:   3,
	defWidth: 9,
	mask:     0xFF,
}

// Decodes a sample from the stream into memory. totalLength is measured in samples.
// For 8-bit samples, the result needs to be converted. Each int16 contains only one 8-bit
// sample.
func (self *ItSampleCodec) Decode(r io.Reader, sampleLength int) ([]int16, error) {
	totalData := []int16{}

	remainingLength := sampleLength
	for remainingLength > 0 {
		chunk, err := self.decodeChunk(r, remainingLength)
		if err != nil {
			return nil, err
		}
		totalData = append(totalData, chunk...)
		remainingLength -= len(chunk)
	}

	return totalData, nil
}

func (*ItSampleCodec) getChunk(r io.Reader) (bitstream, error) {
	// Read in a chunk.
	var byteLength uint16
	err := binary.Read(r, binary.LittleEndian, &byteLength)
	if err != nil {
		return bitstream{}, err
	}

	bytes := make([]byte, byteLength)
	err = binary.Read(r, binary.LittleEndian, &bytes)
	if err != nil {
		return bitstream{}, err
	}

	return createBitstream(bytes), nil
}

func (c *ItSampleCodec) decodeChunk(r io.Reader, remainingLength int) ([]int16, error) {

	var decoded []int16

	dataSource, err := c.getChunk(r)
	if err != nil {
		return nil, err
	}

	// 32kb block
	maxBlockLength := 32 * 1024
	if c.Is16 {
		maxBlockLength /= 2
	}

	curLength := min(remainingLength, maxBlockLength)

	props := &itSampleCodecParams8
	if c.Is16 {
		props = &itSampleCodecParams16
	}
	width := props.defWidth

	changeWidth := func(toWidth int) {
		toWidth++
		if toWidth >= width {
			toWidth++
		}
		width = toWidth
	}

	mem1 := 0
	mem2 := 0

	write := func(v int, topBit int) {
		if v&topBit != 0 {
			v -= (topBit << 1)
		}
		mem1 += v
		mem2 += mem1
		if c.It215 {
			decoded = append(decoded, int16(mem2))
		} else {
			decoded = append(decoded, int16(mem1))
		}
		//writtenSamples++;
		//writePos += mptSample.GetNumChannels();
		curLength--
	}

	for curLength > 0 {
		if width > props.defWidth {
			// Error!
			return nil, ErrDecodingError
		}

		vu, err := dataSource.read(width)
		if err != nil {
			return nil, err
		}
		v := int(vu)

		topBit := (1 << (width - 1))
		if width <= 6 {
			// Mode A: 1 to 6 bits
			if v == topBit {
				toWidth, err := dataSource.read(props.fetchA)
				if err != nil {
					return nil, err
				}
				changeWidth(int(toWidth))
			} else {
				write(int(v), topBit)
			}
		} else if width < props.defWidth {
			// Mode B: 7 to 8 / 16 bits
			if v >= topBit+props.lowerB && v <= topBit+props.upperB {
				changeWidth(v - (topBit + props.lowerB))
			} else {
				write(v, topBit)
			}
		} else {
			// Mode C: 9 / 17 bits
			if v&topBit != 0 {
				width = (v & ^topBit) + 1
			} else {
				write((v & ^topBit), 0)
			}
		}
	}

	return decoded, nil
}

// Todo: encoding.
func (*ItSampleCodec) Encode(r io.Reader, sampleLength int) ([]byte, error) {
	return nil, nil
}
