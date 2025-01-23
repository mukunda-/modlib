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
	"strings"

	"go.mukunda.com/modlib/common"
)

type ItModuleHeader struct {
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

	// Orders [256]uint8

	// Instruments []Instrument
	// Samples     []Sample
	// Patterns    []Pattern
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

	VolumeEnvelope  ItEnvelope
	PanningEnvelope ItEnvelope
	PitchEnvelope   ItEnvelope
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

	Nodes [25]EnvelopeEntry

	_ byte
}

type EnvelopeEntry struct {
	Y uint8
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

type ItSample struct {
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

// TODO I don't think this needs to be separated
type ItSampleData struct {
	Bits16       bool
	Length       int
	LoopStart    int
	LoopEnd      int
	C5Speed      int
	SustainStart int
	SustainEnd   int
	Loop         bool
	Sustain      bool
	BidiLoop     bool
	BidiSustain  bool

	// Can be []int16 or []int8
	Data any
}

type ItPattern struct {
	DataLength uint16
	Rows       uint16
	Data       []byte
}

var ErrInvalidSource = errors.New("invalid/corrupted source")
var ErrUnsupportedSource = errors.New("unsupported source")

func LoadITFile(filename string) (*common.Module, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return LoadITData(f)
}

// func (m *ITModule) LoadFromFile(filename string) error {
// 	f, err := os.Open(filename)
// 	if err != nil {
// 		return err
// 	}

// 	defer f.Close()

// 	m.Load(f)
// 	return nil
// }

// type unpacker struct {
// 	r io.Reader
// }

// func (up *unpacker) read8() byte {
// 	var b [1]byte
// 	_, err := up.r.Read(b[:])
// 	if err != nil {
// 		panic(err)
// 	}
// 	return b[0]
// }

// func (up *unpacker) read16() uint16 {
// 	var b uint16
// 	err := binary.Read(up.r, binary.LittleEndian, &b)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return b
// }

// func (up *unpacker) read32() uint32 {
// 	var res uint32
// 	err := binary.Read(up.r, binary.LittleEndian, &res)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return res
// }

// func (up *unpacker) readString(length int) string {
// 	b := make([]byte, length)
// 	_, err := up.r.Read(b)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return strings.TrimRight(string(b), "\000")
// }

const (
	ItFlagStereo              = 1
	ItFlagMixing              = 2
	ItFlagInstruments         = 4
	ItFlagLinearSlides        = 8
	ItFlagOldEffects          = 16
	ItFlagLinkEFG             = 32
	ItFlagMidiPitchControl    = 64
	ItFlagRequestMidiMacros   = 128
	ItFlagExtendedFilterRange = (1 << 15)
)

func LoadITData(r io.ReadSeeker) (*common.Module, error) {
	var m = new(common.Module)
	m.Source = common.ItSource

	var code [4]byte

	if err := binary.Read(r, binary.LittleEndian, &code); err != nil {
		return m, err
	}

	if string(code[:]) != "IMPM" {
		return m, fmt.Errorf("%w: expected 'IMPM' header", ErrInvalidSource)
	}

	var header ItModuleHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return m, err
	}

	if header.Cwtv < 0x0217 {
		return m, fmt.Errorf("%w: cwtv < 0x0217 (too old!)", ErrUnsupportedSource)
	}

	m.Title = strings.TrimRight(string(header.Title[:]), "\000")
	m.Other = map[string]any{}
	m.Other["cwtv"] = int(header.Cwtv)
	m.Other["cmwt"] = int(header.Cmwt)
	m.Other["itflags"] = int(header.Flags)
	m.Other["itspecial"] = int(header.Special)

	m.StereoMixing = (header.Flags & ItFlagStereo) != 0
	m.UseInstruments = (header.Flags & ItFlagInstruments) != 0
	m.LinearSlides = (header.Flags & ItFlagLinearSlides) != 0
	m.OldEffects = (header.Flags & ItFlagOldEffects) != 0
	m.LinkEFG = (header.Flags & ItFlagLinkEFG) != 0

	m.PatternHighlight_Beat = int16(header.PatternHighlightBeat)
	m.PatternHighlight_Measure = int16(header.PatternHighlightMeasure)

	m.GlobalVolume = int16(header.GlobalVolume)
	m.MixingVolume = int16(header.MixingVolume)
	m.InitialSpeed = int16(header.InitialSpeed)
	m.InitialTempo = int16(header.InitialTempo)
	m.PanSeparation = int16(header.Sep)
	m.PitchWheelDepth = int16(header.PWD)

	m.ChannelSettings = make([]common.ChannelSetting, 64)

	for i := 0; i < 64; i++ {
		m.ChannelSettings[i].InitialPan = int16(header.ChannelPan[i]) * 2
	}

	for i := 0; i < 64; i++ {
		m.ChannelSettings[i].InitialVolume = int16(header.ChannelVolume[i])
	}

	{
		orders := make([]uint8, header.OrderCount)
		if err := binary.Read(r, binary.LittleEndian, &orders); err != nil {
			return m, err
		}

		for i := 0; i < int(header.OrderCount); i++ {
			if orders[i] == 255 {
				break
			}
			m.Order = append(m.Order, int16(orders[i]))
		}
	}

	instrTable := make([]uint32, header.InstrumentCount)
	sampleTable := make([]uint32, header.SampleCount)
	patternTable := make([]uint32, header.PatternCount)

	if err := binary.Read(r, binary.LittleEndian, &instrTable); err != nil {
		return m, err
	}

	if err := binary.Read(r, binary.LittleEndian, &sampleTable); err != nil {
		return m, err
	}

	if err := binary.Read(r, binary.LittleEndian, &patternTable); err != nil {
		return m, err
	}

	for i := 0; i < int(header.InstrumentCount); i++ {
		if instrTable[i] == 0 {
			// unknown behavior
			m.Instruments = append(m.Instruments, common.Instrument{})
			continue
		}

		r.Seek(int64(instrTable[i]), io.SeekStart)
		if ins, err := loadInstrumentData(r); err != nil {
			return m, err
		} else {
			m.Instruments = append(m.Instruments, ins)
		}
	}

	for i := 0; i < int(header.SampleCount); i++ {
		if sampleTable[i] == 0 {
			// unknown behavior
			m.Samples = append(m.Samples, common.Sample{})
			continue
		}

		r.Seek(int64(sampleTable[i]), io.SeekStart)
		if sample, err := loadSampleData(r, header.Cwtv >= 0x215); err != nil {
			return m, err
		} else {
			m.Samples = append(m.Samples, sample)
		}
	}

	for i := 0; i < int(header.PatternCount); i++ {
		if patternTable[i] == 0 {
			// unknown behavior
			m.Patterns = append(m.Patterns, common.Pattern{})
			continue
		}

		r.Seek(int64(patternTable[i]), io.SeekStart)
		if pattern, err := loadPattern(r); err != nil {
			return m, err
		} else {
			m.Patterns = append(m.Patterns, pattern)
		}
	}

	if header.MessageLength != 0 {
		r.Seek(int64(header.MessageOffset), io.SeekStart)
		msg := make([]byte, header.MessageLength)

		if err := binary.Read(r, binary.LittleEndian, msg); err != nil {
			return m, err
		}

		m.Message = strings.Trim(string(msg), "\000")
	}

	return m, nil
}

func loadInstrumentData(r io.ReadSeeker) (common.Instrument, error) {
	var ins common.Instrument

	var iti ItInstrument
	if err := binary.Read(r, binary.LittleEndian, &iti); err != nil {
		return ins, err
	}

	if string(iti.FileCode[:]) != "IMPI" {
		// Ignore if the code is incorrect, maybe propagate a warning somewhere?
	}

	ins.Name = strings.TrimRight(string(iti.Name[:]), "\000")
	ins.DosFilename = strings.TrimRight(string(iti.DosFilename[:]), "\000")
	ins.NewNoteAction = int16(iti.NewNoteAction)
	ins.DuplicateCheckType = int16(iti.DuplicateCheckType)
	ins.DuplicateCheckAction = int16(iti.DuplicateCheckAction)
	ins.Fadeout = int16(iti.Fadeout)

	ins.PitchPanSeparation = int16(iti.PPC)
	ins.PitchPanCenter = int16(iti.PPC)

	ins.GlobalVolume = int16(iti.GlobalVolume)

	ins.DefaultPan = int16(iti.DefaultPan & 0x7F)
	ins.DefaultPanEnabled = iti.DefaultPan&128 == 0

	ins.RandomVolumeVariation = int16(iti.RandomVolume)
	ins.RandomPanVariation = int16(iti.RandomPanning)

	ins.FilterCutoff = int16(iti.InitialFilterCutoff)
	ins.FilterResonance = int16(iti.InitialFilterResonance)

	ins.MidiChannel = int16(iti.MidiChannel)
	ins.MidiProgram = int16(iti.MidiProgram)
	ins.MidiBank = uint16(iti.MidiBank)

	for i := 0; i < 120; i++ {
		ins.Notemap[i].Note = int16(iti.Notemap[i].Note)
		ins.Notemap[i].Sample = int16(iti.Notemap[i].Sample)
	}

	for i := 0; i < 3; i++ {
		if env, err := loadEnvelopeData(r, i); err != nil {
			return ins, err
		} else {
			ins.Envelopes = append(ins.Envelopes, env)
		}
	}

	return ins, nil
}

func loadEnvelopeData(r io.ReadSeeker, index int) (common.Envelope, error) {
	var env common.Envelope

	var itenv ItEnvelope
	if err := binary.Read(r, binary.LittleEndian, &itenv); err != nil {
		return env, err
	}

	env.Enabled = (itenv.Flags & EnvFlagEnabled) != 0
	env.Loop = (itenv.Flags & EnvFlagLoop) != 0
	env.Sustain = (itenv.Flags & EnvFlagSustain) != 0

	if index == 0 {
		env.Type = common.EnvelopeTypeVolume
	} else if index == 1 {
		env.Type = common.EnvelopeTypePanning
	} else if index == 2 {
		env.Type = common.EnvelopeTypePitch
		if itenv.Flags&EnvFlagFilter != 0 {
			env.Type = common.EnvelopeTypeFilter
		}
	} else {
		return env, fmt.Errorf("%w: invalid envelope index", ErrInvalidSource)
	}

	for i := 0; i < 25; i++ {
		if i >= int(itenv.NodeCount) {
			break
		}
		env.Nodes = append(env.Nodes, common.EnvelopeNode{
			Y: int16(itenv.Nodes[i].Y),
			X: int16(itenv.Nodes[i].X),
		})
	}

	return env, nil
}

func loadSampleData(r io.ReadSeeker, it215 bool) (common.Sample, error) {
	var s common.Sample
	var its ItSample
	if err := binary.Read(r, binary.LittleEndian, &its); err != nil {
		return s, err
	}

	if string(its.FileCode[:]) != "IMPS" {
		// Ignore if the code is incorrect, maybe propagate a warning somewhere?
	}

	s.Name = strings.TrimRight(string(its.Name[:]), "\000")
	s.DosFilename = strings.TrimRight(string(its.DosFilename[:]), "\000")

	s.GlobalVolume = int16(its.GlobalVolume)
	s.DefaultVolume = int16(its.DefaultVolume)
	s.DefaultPanning = int16(its.DefaultPanning)

	s.S16 = (its.Flags & SampFlag16bit) != 0
	s.Stereo = (its.Flags & SampFlagStereo) != 0
	s.Loop = (its.Flags & SampFlagLoop) != 0
	s.Sustain = (its.Flags & SampFlagSustain) != 0
	s.PingPong = (its.Flags & SampFlagPingPong) != 0
	s.PingPongSustain = (its.Flags & SampFlagPingPongSustain) != 0

	s.LoopStart = int(its.LoopStart)
	s.LoopEnd = int(its.LoopEnd)

	s.C5 = int(its.C5)

	s.VibratoSpeed = int16(its.VibratoSpeed)
	s.VibratoDepth = int16(its.VibratoDepth)
	s.VibratoSweep = int16(its.VibratoSweep)
	s.VibratoWaveform = int16(its.VibratoWaveform)

	r.Seek(int64(its.SamplePointer), io.SeekStart)
	if data, err := its.loadSampleData(r, it215); err != nil {
		return s, err
	} else {
		s.Data = data
	}

	return s, nil
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

func (s *ItSample) loadSampleData(r io.ReadSeeker, it215 bool) (common.SampleData, error) {

	if s.Convert&SampConvDelta != 0 {
		return common.SampleData{}, fmt.Errorf("%w: delta-encoded samples not supported", ErrUnsupportedSource)
	}

	data := common.SampleData{}

	compressed := s.Flags&SampFlagCompressed != 0
	signed := s.Convert&SampConvSigned != 0
	bits16 := s.Flags&SampFlag16bit != 0
	stereo := s.Flags&SampFlagStereo != 0
	length := int(s.Length)

	data.Channels = 1
	if stereo {
		data.Channels = 2
		length >>= 1
	}

	data.Bits = 8
	if bits16 {
		data.Bits = 16
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

	for ch := 0; ch < int(data.Channels); ch++ {
		if !compressed {

			if bits16 {
				d, err := readPcm[int16](r, length, offset)
				if err != nil {
					return common.SampleData{}, err
				}

				data.Data = append(data.Data, d)
			} else {
				d, err := readPcm[int8](r, length, offset)
				if err != nil {
					return common.SampleData{}, err
				}

				data.Data = append(data.Data, d)
			}
		} else {
			decoder := ItSampleCodec{
				Is16:  bits16,
				It215: it215,
			}

			decoded, err := decoder.Decode(r, length)
			if err != nil {
				return common.SampleData{}, err
			}

			if bits16 {
				data.Data = append(data.Data, decoded)
			} else {
				data8 := make([]int8, len(decoded))
				for i := 0; i < len(decoded); i++ {
					data8[i] = int8(decoded[i])
				}
				data.Data = append(data.Data, data8)
			}

			/*
				totalData := []int16{}
				remainingLength := length
				for remainingLength > 0 {
					d, err := s.decompressItSampleChunk(r, remainingLength, bits16, it215)
					if err != nil {
						return common.SampleData{}, err
					}

					totalData = append(totalData, d...)
					remainingLength -= len(d)
				}

				data.Data = append(data.Data, totalData)*/

			//return nil, fmt.Errorf("%w: compressed samples not supported", ErrUnsupportedSource)
		}
	}

	return data, nil
}

/*
	func (s *ItSample) decompressItSampleChunk(r io.ReadSeeker, remainingLength int, bits16 bool, it215 bool) ([]int16, error) {
		var chunkSize uint16
		if err := binary.Read(r, binary.LittleEndian, &chunkSize); err != nil {
			return nil, err
		}

		chunk := make([]byte, chunkSize)
		if err := binary.Read(r, binary.LittleEndian, &chunk); err != nil {
			return nil, err
		}

		return nil, nil
	}
*/
func translateNote(note uint8) uint8 {
	if note <= 120 {
		return note + 1
	} else if note == 253 {
		return 200
	} else if note == 254 {
		return 201
	} else if note == 255 {
		return 202
	} else {
		return 0
	}
}

func translatePatternVolume(vol uint8) (uint8, uint8) {
	if vol <= 64 {
		return 1, vol
	} else if vol <= 74 {
		return 2, vol - 65
	} else if vol <= 84 {
		return 3, vol - 75
	} else if vol <= 94 {
		return 4, vol - 85
	} else if vol <= 104 {
		return 5, vol - 95
	} else if vol <= 114 {
		return 6, vol - 105
	} else if vol <= 124 {
		return 7, vol - 125
	} else if vol <= 127 {
		return 0, 0
	} else if vol <= 128 {
		return 8, vol - 128
	} else if vol <= 202 {
		return 9, vol - 129
	} else if vol <= 212 {
		return 10, vol - 203
	}
	return 0, 0
}

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

func loadPattern(r io.ReadSeeker) (common.Pattern, error) {
	var p common.Pattern
	var itp ItPattern
	if err := binary.Read(r, binary.LittleEndian, &itp); err != nil {
		return p, err
	}

	data := make([]byte, itp.DataLength)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return p, err
	}

	// Unpack data
	dataRead := 0
	failure := false

	nextByte := func() byte {
		if dataRead >= len(data) {
			failure = true
			return 0
		}

		byt := data[dataRead]
		dataRead++
		return byt
	}

	var lastMask [64]byte
	var lastNote [64]byte
	var lastIns [64]byte
	var lastVol [64]byte
	var lastEffect [64]byte
	var lastEffectParam [64]byte

	for row := 0; row < int(itp.Rows); row++ {
		for {
			channelSelect := nextByte()
			if channelSelect == 0 {
				break
			}

			entry := common.PatternChannelEntry{}

			channel := int((channelSelect - 1) & 63)
			if channelSelect&0x80 != 0 {
				lastMask[channel] = nextByte()
			}
			mask := lastMask[channel]

			if mask&PmaskNote != 0 {
				lastNote[channel] = nextByte()
			}

			if mask&(PmaskNote|PmaskLastNote) != 0 {
				entry.Note = translateNote(lastNote[channel])
			}

			if mask&PmaskIns != 0 {
				lastIns[channel] = nextByte()
			}

			if mask&(PmaskIns|PmaskLastIns) != 0 {
				entry.Instrument = int16(lastIns[channel]) // add one here? or is 1 lowest?
			}

			if mask&PmaskVol != 0 {
				lastVol[channel] = nextByte()
			}

			if mask&(PmaskVol|PmaskLastVol) != 0 {
				entry.VolumeCommand, entry.VolumeParam = translatePatternVolume(lastVol[channel])
			}

			if mask&PmaskEffect != 0 {
				lastEffect[channel] = nextByte()
				lastEffectParam[channel] = nextByte()
			}

			if mask&(PmaskEffect|PmaskLastEffect) != 0 {
				entry.Effect = lastEffect[channel]
				entry.EffectParam = lastEffectParam[channel]
			}
		}

		if failure {
			return p, fmt.Errorf("%w: unexpected end of pattern data", ErrInvalidSource)
		}
	}

	return p, nil
}
