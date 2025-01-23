// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

/*
Package common provides a medium for all supported sources. All submodules load into this
intermediate format, which is based on IT.
*/
package common

type ModuleSourceFormat int16

const (
	UnknownSource ModuleSourceFormat = iota
	ModSource
	S3mSource
	XmSource
	ItSource
)

type Module struct {
	Source          ModuleSourceFormat
	Title           string // The title of the song.
	GlobalVolume    int16  // The initial global volume. 0 = 0%, 128 = 100%
	MixingVolume    int16  // Mixing volume of the song. 0 = 0%, 128 = 100%
	InitialSpeed    int16  // Initial ticks per row (Axx)
	InitialTempo    int16  // Initial BPM.
	PanSeparation   int16  // TODO: how does it work
	PitchWheelDepth int16  // TODO: what is it for
	StereoMixing    bool   // Enable stereo audio mixing.
	UseInstruments  bool   // Enable use of instruments.
	LinearSlides    bool   // Linear slides instead of Amiga slides.
	OldEffects      bool   // Enable old effect behavior (IT)
	LinkEFG         bool   // Share memory between G and EF.
	Channels        int16  // Number of channels.

	// The embedded "song message" text.
	Message string

	// For editing, where to highlight the patterns.
	PatternHighlight_Beat    int16 // Rows per beat
	PatternHighlight_Measure int16 // Rows per measure

	ChannelSettings []ChannelSetting
	Order           []int16
	Instruments     []Instrument
	Samples         []Sample
	Patterns        []Pattern
}

type ChannelSetting struct {
	Name          string
	InitialVolume int16 // 0-64
	InitialPan    int16 // 0-64
	Mute          bool
	Surround      bool
}

const (
	NnaNoteCut  = 0
	NnaContinue = 1
	NnaNoteOff  = 2
	NnaFade     = 3
)

const (
	DctOff        = 0
	DctNote       = 1
	DctSample     = 2
	DctInstrument = 3
	DctPlugin     = 4
)

type Instrument struct {
	Name                 string
	DosFilename          string
	NewNoteAction        int16 // Nna*
	DuplicateCheckType   int16 // Dct*
	DuplicateCheckAction int16 // Dca*
	Fadeout              int16

	// Controls changing pan according to pitch, for example, lower notes coming from one
	// side, and higher notes coming from the other.
	PitchPanSeparation int16
	PitchPanCenter     int16 // (0-119)

	GlobalVolume int16

	DefaultPan        int16 // 0-64
	DefaultPanEnabled bool

	RandomVolumeVariation int16 // percentage (0-100)
	RandomPanVariation    int16 // percentage (0-100)

	FilterCutoff    int16
	FilterResonance int16

	MidiChannel int16
	MidiProgram int16
	MidiBank    uint16

	Notemap [120]NotemapEntry

	Envelopes []Envelope
}

type NotemapEntry struct {
	Note   int16
	Sample int16
}

type EnvelopeType int16

const (
	EnvelopeTypeVolume  EnvelopeType = 0
	EnvelopeTypePanning EnvelopeType = 1
	EnvelopeTypePitch   EnvelopeType = 2
	EnvelopeTypeFilter  EnvelopeType = 3
)

type Envelope struct {
	Enabled bool
	Loop    bool
	Sustain bool
	Type    EnvelopeType

	LoopStart    int16
	LoopEnd      int16
	SustainStart int16
	SustainEnd   int16

	Nodes []EnvelopeNode
}

type EnvelopeNode struct {
	X int16
	Y int16
}

const (
	SampleVibratoWaveformSine   = 0
	SampleVibratoWaveformRamp   = 1
	SampleVibratoWaveformSquare = 2
	SampleVibratoWaveformRandom = 3
)

type Sample struct {
	Name        string
	DosFilename string

	GlobalVolume   int16
	DefaultVolume  int16
	DefaultPanning int16

	S16             bool
	Stereo          bool
	Loop            bool
	Sustain         bool
	PingPong        bool
	PingPongSustain bool

	LoopStart        int
	LoopEnd          int
	SustainLoopStart int
	SustainLoopEnd   int

	C5 int

	VibratoSpeed    int16
	VibratoDepth    int16
	VibratoSweep    int16
	VibratoWaveform int16

	// This will be int16 if S16 is set, int8 otherwise
	// Stereo samples have left,right interleaved
	Data SampleData
}

type SampleData struct {
	Channels int8
	Bits     int8

	// This will contain []int8 or []int16 depending on Bits
	Data []any
}

type Pattern struct {
	Channels int16
	Rows     []PatternRow
}

type PatternRow struct {
	Entries []PatternEntry
}

type PatternEntry struct {
	Channel uint8

	// 0 = Empty, 1 = C-0, 120 = B-9, 253 = NoteFade, 254 = NoteCut, 255 = NoteOff
	Note uint8

	// Set instrument index (0 = empty)
	Instrument int16

	// It's called the volume column, but IT squeezes a lot into it.
	// 0 = empty
	// 1 = Set Volume (p=0-64)
	// 2 = Fine vol up (p=0-9)
	// 3 = Fine vol down (p=0-9)
	// 4 = Vol slide up (p=0-9)
	// 5 = Vol slide down (p=0-9)
	// 6 = Pitch slide down (p=0-9)
	// 7 = Pitch slide up (p=0-9)
	// 8 = Set Pan (p=0-64)
	// 9 = Porta to note (p=0-9)
	// 10 = Vibrato depth (p=0-9)
	VolumeCommand uint8
	VolumeParam   uint8

	// 0 = Empty
	// 1 = Axx, 2 = Bxx, ...
	Effect uint8

	// 00-FF data for the effect
	EffectParam uint8
}
