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

type ItSampleCodecParams struct {
	lowerTab []int16
	upperTab []int16
	fetchA   int
	lowerB   int
	upperB   int
	defWidth int
	mask     int
}

// Algorithm parameters for 16-Bit samples
var ItSampleCodecParams16 = ItSampleCodecParams{
	lowerTab: []int16{0, -1, -3, -7, -15, -31, -56, -120, -248, -504, -1016, -2040, -4088, -8184, -16376, -32760, -32768},
	upperTab: []int16{0, 1, 3, 7, 15, 31, 55, 119, 247, 503, 1015, 2039, 4087, 8183, 16375, 32759, 32767},
	fetchA:   4,
	lowerB:   -8,
	upperB:   7,
	defWidth: 17,
	mask:     0xFFFF,
}

// Algorithm parameters for 8-Bit samples
var ItSampleCodecParams8 = ItSampleCodecParams{
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

	props := &ItSampleCodecParams8
	if c.Is16 {
		props = &ItSampleCodecParams16
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

/*
me trying to make sense of greasemonkey's code until looking at openmpt
func (*ItSampleDecoder) decodeChunk(r io.Reader, is16 bool, remainingLength int) ([]int16, error) {

	bytepos := 0
	bitpos := 0

	//base_length := remainingLength
	grab_length := remainingLength
	running_count := 0

	fetch_a := 3
	//spread_b := 8
	lower_b := -4
	upper_b := 3
	width := 9
	widthtop := 9
	unpack_mask := 0xFF
	maxgrablen := 0x8000
	if is16 {
		fetch_a = 4
		//spread_b = 16
		lower_b = -8
		upper_b = 7
		width = 17
		widthtop = 17
		unpack_mask = 0xFFFF
		maxgrablen = 0x4000
	}

	// Read in a chunk.
	var byteLength uint16
	err := binary.Read(r, binary.LittleEndian, &byteLength)
	if err != nil {
		return nil, err
	}

	bytes := make([]byte, byteLength)
	err = binary.Read(r, binary.LittleEndian, &bytes)
	if err != nil {
		return nil, err
	}

	end_of_block := func() bool {
		return bytepos >= len(bytes)
	}

	change_width := func(w int) {
		w += 1
		if w >= width {
			w += 1
		}
		width = w
	}

	// Read a number of bits from the stream.
	read := func(numBits int) (int, error) {
		result := 0
		valueWritePos := 0

		for numBits > 0 {
			if bytepos > len(bytes) {
				return 0, fmt.Errorf("%w: unexpected end of data", ErrDecodingError)
			}
			remaining := 8 - bitpos
			if numBits >= remaining {
				result |= int(bytes[bytepos]) >> bitpos << valueWritePos
				bytepos++
				bitpos = 0
				numBits -= remaining
				valueWritePos += remaining
			} else {
				result |= ((int(bytes[bytepos]) >> bitpos) & (1<<numBits - 1)) << valueWritePos
				valueWritePos += numBits
				bitpos += numBits
				numBits = 0
			}
		}
		return result, nil
	}
	unpacked_root := 0
	length := min(grab_length, maxgrablen)
	unpacked_data := []int16{}

	write := func(value int, topbit int) {
		running_count += 1
		length -= 1

		v := value
		if v&topbit != 0 {
			v -= topbit * 2
		}
		unpacked_root = (unpacked_root + v) & unpack_mask
		unpacked_data = append(unpacked_data, int16(unpacked_root))
	}

	grab_length -= length
	//print "subchunk length: %i" % length

	for length > 0 && !end_of_block() {
		if width == 0 || width > widthtop {
			return nil, fmt.Errorf("%w: invalid bit width", ErrDecodingError)
		}

		value, err := read(width)
		if err != nil {
			return nil, err
		}

		topbit := int(1 << (width - 1))

		if width <= 6 { // MODE A
			if value == topbit {
				w, err := read(fetch_a)
				if err != nil {
					return nil, err
				}
				change_width(int(w))
				//#print width
			} else {
				write(int(value), topbit)
			}
		} else if width < widthtop { // # MODE B
			if value >= topbit+lower_b && value <= topbit+upper_b {
				qv := value - (topbit + lower_b)
				//#print "MODE B CHANGE",width,value,qv
				change_width(qv)
				//#print width
			} else {
				write(value, topbit)
			}
		} else { //# MODE C
			if value&topbit != 0 {
				width = (value & ^topbit) + 1
				//#print width
			} else {
				write((value & ^topbit), 0)
			}
		}
	}

	//print "bytes remaining in block: %i" % (len(data)-dpos)

	return unpacked_data, nil
}
*/
