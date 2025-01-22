/* itmod
 * Copyright 2025 Mukunda Johnson (mukunda.com)
 * Licensed under MIT
 */

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
)

type ITModule struct {
	Filename         string
	Title            string
	PatternHighlight uint16
	Length           uint16
	InstrumentCount  uint16
	SampleCount      uint16
	PatternCount     uint16
	Cwtv             uint16
	Cmwt             uint16
	Flags            uint16 // TODO: why was this int in itloader.cpp?
	Special          uint16
	GlobalVolume     uint8
	MixingVolume     uint8
	InitialSpeed     uint8
	InitialTempo     uint8
	Sep              uint8
	PWD              uint8

	Message string

	ChannelPan    [64]uint8
	ChannelVolume [64]uint8

	Orders [256]uint8

	Instruments []Instrument
	Samples     []Sample
	Patterns    []Pattern
}

type Instrument struct {
	Name        string
	DOSFilename string

	NewNoteAction        uint8
	DuplicateCheckType   uint8
	DuplicateCheckAction uint8

	Fadeout                uint16
	PPS                    uint8
	PPC                    uint8
	GlobalVolume           uint8
	DefaultPan             uint8
	RandomVolume           uint8
	RandomPanning          uint8
	TrackerVersion         uint16
	NumberOfSamples        uint8
	InitialFilterCutoff    uint8
	InitialFilterResonance uint8

	MidiChannel uint8
	MidiProgram uint8
	MidiBank    uint16

	Notemap [120]NotemapEntry

	VolumeEnvelope  Envelope
	PanningEnvelope Envelope
	PitchEnvelope   Envelope
}

type NotemapEntry struct {
	Note   uint8
	Sample uint8
}

type EnvelopeEntry struct {
	y uint8
	x uint16
}

type Envelope struct {
	Enabled  bool
	Loop     bool
	Sustain  bool
	IsFilter bool

	Length       uint8
	LoopStart    uint8
	LoopEnd      uint8
	SustainStart uint8
	SustainEnd   uint8

	Nodes []EnvelopeEntry
}

type Sample struct {
	Name        string
	DOSFilename string

	GlobalVolume uint8

	HasSample  bool
	Stereo     bool
	Compressed bool

	DefaultVolume  uint8
	DefaultPanning uint8
	Convert        uint8

	VibratoSpeed uint8
	VibratoDepth uint8
	VibratoForm  uint8
	VibratoRate  uint8

	Data SampleData
}

// TODO I don't think this needs to be separated
type SampleData struct {
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

type Pattern struct {
	DataLength uint16
	Rows       uint16
	Data       []byte
}

var ErrInvalidSource = errors.New("invalid source")

func (m *ITModule) LoadFromFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	m.Filename = filename
	m.Load(f)
	return nil
}

type unpacker struct {
	r io.Reader
}

func (up *unpacker) read8() byte {
	var b [1]byte
	_, err := up.r.Read(b[:])
	if err != nil {
		panic(err)
	}
	return b[0]
}

func (up *unpacker) read16() uint16 {
	var b uint16
	err := binary.Read(up.r, binary.LittleEndian, &b)
	if err != nil {
		panic(err)
	}
	return b
}

func (up *unpacker) read32() uint32 {
	var res uint32
	err := binary.Read(up.r, binary.LittleEndian, &res)
	if err != nil {
		panic(err)
	}
	return res
}

func (up *unpacker) readString(length int) string {
	b := make([]byte, length)
	_, err := up.r.Read(b)
	if err != nil {
		panic(err)
	}

	return strings.TrimRight(string(b), "\000")
}

func (m *ITModule) Load(r io.ReadSeeker) error {

	up := unpacker{r}

	if up.read8() != 'I' || up.read8() != 'M' || up.read8() != 'P' || up.read8() != 'M' {
		return fmt.Errorf("%w: expected 'IMPM' header", ErrInvalidSource)
	}

	title := make([]byte, 26)
	_, err := r.Read(title)
	if err != nil {
		return err
	}
	m.Title = strings.TrimRight(string(title), "\000")

	m.PatternHighlight = up.read16()

	m.Length = up.read16()
	m.InstrumentCount = up.read16()
	m.SampleCount = up.read16()
	m.PatternCount = up.read16()
	m.Cwtv = up.read16()
	m.Cmwt = up.read16()
	m.Flags = up.read16()
	m.Special = up.read16()
	m.GlobalVolume = up.read8()
	m.MixingVolume = up.read8()
	m.InitialSpeed = up.read8()
	m.InitialTempo = up.read8()

	m.Sep = up.read8()
	m.PWD = up.read8()

	messageLength := up.read16()
	messageOffset := up.read32()

	up.read32() // skip 4 bytes (reserved)

	for i := 0; i < 64; i++ {
		m.ChannelPan[i] = up.read8()
	}

	for i := 0; i < 64; i++ {
		m.ChannelVolume[i] = up.read8()
	}

	{
		foundend := false
		ActualLength := m.Length
		for i := 0; i < 256; i++ {
			if i < int(m.Length) {
				m.Orders[i] = up.read8()
			} else {
				m.Orders[i] = 255
			}

			if m.Orders[i] == 255 && !foundend {
				foundend = true
				ActualLength = uint16(i + 1)
			}
		}

		m.Length = ActualLength
	}

	instrTable := make([]uint32, m.InstrumentCount)
	sampleTable := make([]uint32, m.SampleCount)
	patternTable := make([]uint32, m.PatternCount)

	for i := 0; i < int(m.InstrumentCount); i++ {
		instrTable[i] = up.read32()
	}
	for i := 0; i < int(m.SampleCount); i++ {
		sampleTable[i] = up.read32()
	}
	for i := 0; i < int(m.PatternCount); i++ {
		patternTable[i] = up.read32()
	}

	for i := 0; i < int(m.InstrumentCount); i++ {
		r.Seek(int64(instrTable[i]), io.SeekStart)
		var ins Instrument
		if err := ins.Load(r); err != nil {
			return err
		}
		m.Instruments = append(m.Instruments, ins)
	}

	for i := 0; i < int(m.SampleCount); i++ {
		r.Seek(int64(sampleTable[i]), io.SeekStart)
		var sample Sample
		if err := sample.Load(r); err != nil {
			return err
		}
		m.Samples = append(m.Samples, sample)
	}

	for i := 0; i < int(m.PatternCount); i++ {
		var pattern Pattern

		// TODO: unsure why the others don't have null checks.
		if patternTable[i] != 0 {
			r.Seek(int64(patternTable[i]), io.SeekStart)
			if err := pattern.Load(r); err != nil {
				return err
			}
			m.Patterns = append(m.Patterns, pattern)
		} else {
			pattern.LoadDefault()
			m.Patterns = append(m.Patterns, pattern)
		}
	}

	if messageLength != 0 {
		r.Seek(int64(messageOffset), io.SeekStart)
		msg := make([]byte, messageLength)
		if _, err := r.Read(msg); err != nil {
			return err
		}

		m.Message = strings.Trim(string(msg), "\000")
	}

	return nil
}

func (ins *Instrument) Load(r io.ReadSeeker) error {
	up := unpacker{r}

	up.read32() // "IMPI"
	ins.DOSFilename = up.readString(12)
	up.read8() // 00h reserved

	ins.NewNoteAction = up.read8()
	ins.DuplicateCheckType = up.read8()
	ins.DuplicateCheckAction = up.read8()
	ins.Fadeout = up.read16()
	ins.PPS = up.read8()
	ins.PPC = up.read8()
	ins.GlobalVolume = up.read8()
	ins.DefaultPan = up.read8()
	ins.RandomVolume = up.read8()
	ins.RandomPanning = up.read8()
	ins.TrackerVersion = up.read16()
	ins.NumberOfSamples = up.read8()
	up.read8()

	ins.Name = up.readString(26)

	ins.InitialFilterCutoff = up.read8()
	ins.InitialFilterResonance = up.read8()

	ins.MidiChannel = up.read8()
	ins.MidiProgram = up.read8()
	ins.MidiBank = up.read16()

	for i := 0; i < 120; i++ {
		ins.Notemap[i].Note = up.read8()
		ins.Notemap[i].Sample = up.read8()
	}

	ins.VolumeEnvelope = Envelope{}
	ins.VolumeEnvelope.Load(r)
	ins.PanningEnvelope = Envelope{}
	ins.PanningEnvelope.Load(r)
	ins.PitchEnvelope = Envelope{}
	ins.PitchEnvelope.Load(r)

	return nil
}

func (e *Envelope) Load(r io.ReadSeeker) error {
	up := unpacker{r}

	flags := up.read8()

	e.Enabled = (flags & 1) == 1
	e.Loop = (flags & 2) == 2
	e.Sustain = (flags & 4) == 4
	e.IsFilter = (flags & 128) == 128

	e.Length = up.read8()

	e.LoopStart = up.read8()
	e.LoopEnd = up.read8()

	e.SustainStart = up.read8()
	e.SustainEnd = up.read8()

	for i := 0; i < 25; i++ {
		if i < int(e.Length) {
			y := up.read8()
			x := up.read16()
			e.Nodes = append(e.Nodes, EnvelopeEntry{y, x})
		} else {
			// Discard remaining data.
			up.read8()
			up.read16()
		}
	}

	up.read8() // extra byte
	return nil
}

func (s *Sample) Load(r io.ReadSeeker) error {
	up := unpacker{r}

	up.read32() // IMPS

	s.DOSFilename = up.readString(12)
	up.read8() // 00h
	s.GlobalVolume = up.read8()
	flags := up.read8()

	s.HasSample = (flags & 1) == 1
	s.Data.Bits16 = (flags & 2) == 2
	s.Stereo = (flags & 4) == 4
	s.Compressed = (flags & 8) == 8
	s.Data.Loop = (flags & 16) == 16
	s.Data.Sustain = (flags & 32) == 32
	s.Data.BidiLoop = (flags & 64) == 64
	s.Data.BidiSustain = (flags & 128) == 128

	s.DefaultVolume = up.read8()

	s.Name = up.readString(26)

	s.Convert = up.read8()
	s.DefaultPanning = up.read8()

	s.Data.Length = int(up.read32())
	s.Data.LoopStart = int(up.read32())
	s.Data.LoopEnd = int(up.read32())
	s.Data.C5Speed = int(up.read32())
	s.Data.SustainStart = int(up.read32())
	s.Data.SustainEnd = int(up.read32())

	samplePointer := up.read32()

	s.VibratoSpeed = up.read8()
	s.VibratoDepth = up.read8()
	s.VibratoRate = up.read8()
	s.VibratoForm = up.read8()

	r.Seek(int64(samplePointer), io.SeekStart)
	s.LoadData(r)

	return nil
}

func (s *Sample) LoadData(r io.ReadSeeker) error {
	up := unpacker{r}

	if !s.Compressed {

		// subtract offset for unsigned samples
		offset := 0
		if s.Convert&1 == 0 {
			if s.Data.Bits16 {
				offset = -32768
			} else {
				offset = -128
			}
		}

		// signed samples
		if s.Data.Bits16 {
			data := make([]int16, s.Data.Length)
			for i := 0; i < s.Data.Length; i++ {
				data[i] = int16(int(up.read16()) + offset)
			}
			s.Data.Data = data
		} else {
			data := make([]int8, s.Data.Length)
			for i := 0; i < s.Data.Length; i++ {
				data[i] = int8(int(up.read8()) + offset)
			}
			s.Data.Data = data
		}
	} else {
		// TODO : accept compressed samples.
	}

	return nil
}

func (p *Pattern) Load(r io.ReadSeeker) error {
	up := unpacker{r}

	p.DataLength = up.read16()
	p.Rows = up.read16()

	up.read32() // reserved

	for i := 0; i < int(p.DataLength); i++ {
		p.Data = append(p.Data, up.read8())
	}

	return nil
}

// TODO: this is new/untested
func (p *Pattern) LoadDefault() {
	p.DataLength = 0
	p.Rows = 64
	p.Data = make([]byte, 64)
}
