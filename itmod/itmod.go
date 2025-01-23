// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

/*
Package itmod is for working with Impulse Tracker files.
*/
package itmod

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

type ItReader struct {
	Strict bool
}

type ItModule struct {
	Header ItModuleHeader

	Orders      []uint8
	Instruments []ItInstrument
	Samples     []ItSample
	Patterns    []ItPattern
	Message     []byte
}

type ItModuleHeader struct {
	FileCode                [4]byte
	Title                   [26]byte
	PatternHighlightBeat    uint8
	PatternHighlightMeasure uint8
	OrderCount              uint16
	InstrumentCount         uint16
	SampleCount             uint16
	PatternCount            uint16
	Cwtv                    uint16
	Cmwt                    uint16
	Flags                   uint16
	Special                 uint16
	GlobalVolume            uint8
	MixingVolume            uint8
	InitialSpeed            uint8
	InitialTempo            uint8
	Sep                     uint8
	PWD                     uint8

	MessageLength uint16
	MessageOffset uint32

	Reserved_MPT uint32

	ChannelPan    [64]uint8
	ChannelVolume [64]uint8
}

type ItInstrument struct {
	FileCode    [4]byte
	DosFilename [12]byte

	_ byte

	NewNoteAction        uint8
	DuplicateCheckType   uint8
	DuplicateCheckAction uint8

	Fadeout         uint16
	PPS             uint8
	PPC             uint8
	GlobalVolume    uint8
	DefaultPan      uint8
	RandomVolume    uint8
	RandomPanning   uint8
	TrackerVersion  uint16
	NumberOfSamples uint8

	_ byte

	Name [26]byte

	InitialFilterCutoff    uint8
	InitialFilterResonance uint8

	MidiChannel uint8
	MidiProgram uint8
	MidiBank    uint16

	Notemap [120]NotemapEntry

	Envelopes [3]ItEnvelope
}

type NotemapEntry struct {
	Note   uint8
	Sample uint8
}

const (
	EnvFlagEnabled = 1
	EnvFlagLoop    = 2
	EnvFlagSustain = 4
	EnvFlagFilter  = 128
)

type ItEnvelope struct {
	Flags        uint8
	NodeCount    uint8
	LoopStart    uint8
	LoopEnd      uint8
	SustainStart uint8
	SustainEnd   uint8

	Nodes [25]EnvelopeNode

	_ byte
}

type EnvelopeNode struct {
	Y int8
	X uint16
}

const (
	SampFlagHeader          = 1
	SampFlag16bit           = 2
	SampFlagStereo          = 4
	SampFlagCompressed      = 8
	SampFlagLoop            = 16
	SampFlagSustain         = 32
	SampFlagPingPong        = 64
	SampFlagPingPongSustain = 128
)

const (
	SampConvSigned    = 1
	SampConvBigEndian = 2
	SampConvDelta     = 4
	SampConvByteDelta = 8
	SampConvTxWave    = 16
)

type ItSampleHeader struct {
	FileCode       [4]byte
	DosFilename    [12]byte
	_              byte
	GlobalVolume   uint8
	Flags          uint8
	DefaultVolume  uint8
	Name           [26]byte
	Convert        uint8
	DefaultPanning uint8

	Length uint32

	LoopStart uint32
	LoopEnd   uint32

	C5 uint32

	SustainLoopStart uint32
	SustainLoopEnd   uint32
	SamplePointer    uint32

	VibratoSpeed    uint8
	VibratoDepth    uint8
	VibratoSweep    uint8
	VibratoWaveform uint8
}

type ItSample struct {
	Header   ItSampleHeader
	Channels uint8
	Bits     uint8

	// This is decoded data and not the original bytes.
	// Original is difficult since the samples can be compressed and have no byte-length
	// indicator in that case.
	// Contains [][]int16 or [][]int8 (Data[channel][sample])
	Data []any
}

type ItPatternHeader struct {
	DataLength uint16 // Length of packed data
	Rows       uint16 // Number of rows in the pattern
	_          uint32 // Reserved
	//Data       []byte
}

// Mask constants for the packed pattern data.
const (
	PmaskNote       = 1
	PmaskIns        = 2
	PmaskVol        = 4
	PmaskEffect     = 8
	PmaskLastNote   = 16
	PmaskLastIns    = 32
	PmaskLastVol    = 64
	PmaskLastEffect = 128
)

type ItPattern struct {
	Header ItPatternHeader

	// Packed data
	Data []byte
}

var ErrInvalidSource = errors.New("invalid/corrupted source")
var ErrUnsupportedSource = errors.New("unsupported source")

func LoadITFile(filename string) (*ItModule, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	reader := ItReader{}

	return reader.ReadItModule(f)
}

const (
	ItFlagStereo              = 1
	ItFlagMixing              = 2
	ItFlagInstruments         = 4
	ItFlagLinearSlides        = 8
	ItFlagOldEffects          = 16
	ItFlagLinkEFG             = 32
	ItFlagMidiPitchControl    = 64
	ItFlagRequestMidiMacros   = 128
	ItFlagExtendedFilterRange = 32768
)

func (reader *ItReader) ReadItModule(r io.ReadSeeker) (*ItModule, error) {
	itm := new(ItModule)

	var header ItModuleHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	itm.Header = header

	if string(header.FileCode[:]) != "IMPM" {
		return nil, fmt.Errorf("%w: expected 'IMPM' header", ErrInvalidSource)
	}

	if header.Cwtv < 0x0217 {
		// TODO: more support for older versions
		return nil, fmt.Errorf("%w: cwtv < 0x0217 (too old!)", ErrUnsupportedSource)
	}

	orders := make([]uint8, header.OrderCount)
	if err := binary.Read(r, binary.LittleEndian, &orders); err != nil {
		return itm, err
	}

	itm.Orders = orders

	instrTable := make([]uint32, header.InstrumentCount)
	sampleTable := make([]uint32, header.SampleCount)
	patternTable := make([]uint32, header.PatternCount)

	if err := binary.Read(r, binary.LittleEndian, &instrTable); err != nil {
		return itm, err
	}

	if err := binary.Read(r, binary.LittleEndian, &sampleTable); err != nil {
		return itm, err
	}

	if err := binary.Read(r, binary.LittleEndian, &patternTable); err != nil {
		return itm, err
	}

	for i := 0; i < int(header.InstrumentCount); i++ {
		if instrTable[i] == 0 {
			// is this possible?
			itm.Instruments = append(itm.Instruments, ItInstrument{})
			continue
		}

		r.Seek(int64(instrTable[i]), io.SeekStart)
		if ins, err := reader.ReadItInstrument(r); err != nil {
			return itm, err
		} else {
			itm.Instruments = append(itm.Instruments, ins)
		}
	}

	it215 := header.Cmwt >= 0x215

	for i := 0; i < int(header.SampleCount); i++ {
		if sampleTable[i] == 0 {
			// unknown behavior
			itm.Samples = append(itm.Samples, ItSample{})
			continue
		}

		r.Seek(int64(sampleTable[i]), io.SeekStart)
		if sample, err := reader.ReadItSample(r, it215); err != nil {
			return itm, err
		} else {
			itm.Samples = append(itm.Samples, sample)
		}
	}

	for i := 0; i < int(header.PatternCount); i++ {
		if patternTable[i] == 0 {
			// unknown behavior
			itm.Patterns = append(itm.Patterns, ItPattern{})
			continue
		}

		r.Seek(int64(patternTable[i]), io.SeekStart)
		if pattern, err := reader.readItPattern(r); err != nil {
			return itm, err
		} else {
			itm.Patterns = append(itm.Patterns, pattern)
		}
	}

	if header.MessageLength != 0 {
		r.Seek(int64(header.MessageOffset), io.SeekStart)
		msg := make([]byte, header.MessageLength)

		if err := binary.Read(r, binary.LittleEndian, msg); err != nil {
			return itm, err
		}

		itm.Message = msg
	}

	return itm, nil
}

func (reader *ItReader) ReadItInstrument(r io.Reader) (ItInstrument, error) {
	var iti ItInstrument

	if err := binary.Read(r, binary.LittleEndian, &iti); err != nil {
		return iti, err
	}

	if string(iti.FileCode[:]) != "IMPI" {
		if reader.Strict {
			return iti, fmt.Errorf("%w: strict - expected 'IMPI' header", ErrInvalidSource)
		}
	}

	return iti, nil
}

func readPcm[T int8 | int16](r io.ReadSeeker, length int, offset int) ([]T, error) {
	data := make([]T, length)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}
	if offset != 0 {
		for i := 0; i < len(data); i++ {
			data[i] += T(offset)
		}
	}
	return data, nil
}

func (reader *ItReader) ReadItSample(r io.ReadSeeker, it215 bool) (ItSample, error) {
	var header ItSampleHeader
	var its ItSample
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return its, err
	}

	its.Header = header
	if string(header.FileCode[:]) != "IMPS" {
		if reader.Strict {
			return its, fmt.Errorf("%w: strict - expected 'IMPS' header", ErrInvalidSource)
		}
	}

	r.Seek(int64(header.SamplePointer), io.SeekStart)

	if header.Convert&SampConvDelta != 0 {
		// TODO: support this.
		return its, fmt.Errorf("%w: delta-encoded samples not supported", ErrUnsupportedSource)
	}

	//data := common.SampleData{}

	compressed := header.Flags&SampFlagCompressed != 0
	signed := header.Convert&SampConvSigned != 0
	bits16 := header.Flags&SampFlag16bit != 0
	stereo := header.Flags&SampFlagStereo != 0
	length := int(header.Length)

	its.Channels = 1
	if stereo {
		its.Channels = 2
	}

	its.Bits = 8
	if bits16 {
		its.Bits = 16
	}

	// For unsigned samples, use an offset.
	offset := 0
	if !signed {
		if bits16 {
			offset = -32768
		} else {
			offset = -128
		}
	}

	for ch := 0; ch < int(its.Channels); ch++ {
		if !compressed {

			if bits16 {
				d, err := readPcm[int16](r, length, offset)
				if err != nil {
					return its, err
				}

				its.Data = append(its.Data, d)
			} else {
				d, err := readPcm[int8](r, length, offset)
				if err != nil {
					return its, err
				}

				its.Data = append(its.Data, d)
			}
		} else {
			decoder := ItSampleCodec{
				Is16:  bits16,
				It215: it215,
			}

			decoded, err := decoder.Decode(r, length)
			if err != nil {
				return its, err
			}

			if bits16 {
				its.Data = append(its.Data, decoded)
			} else {
				data8 := make([]int8, len(decoded))
				for i := 0; i < len(decoded); i++ {
					data8[i] = int8(decoded[i])
				}
				its.Data = append(its.Data, data8)
			}
		}
	}

	return its, nil
}

func (reader *ItReader) readItPattern(r io.ReadSeeker) (ItPattern, error) {
	var itp ItPattern
	var header ItPatternHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return itp, err
	}

	itp.Header = header

	data := make([]byte, header.DataLength)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return itp, err
	}

	itp.Data = data

	return itp, nil
}
