// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package itmod

import (
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mukunda.com/modlib/common"
)

func notemapWithSample(sample int) [120]common.NotemapEntry {
	var mapping [120]common.NotemapEntry
	for i := 0; i < 120; i++ {
		mapping[i] = common.NotemapEntry{Note: int16(i), Sample: int16(sample)}
	}
	return mapping
}

var itFixture1 = common.Module{
	Source:                   common.ItSource,
	Title:                    "reflection",
	GlobalVolume:             45,
	MixingVolume:             48,
	InitialSpeed:             6,
	InitialTempo:             135,
	PanSeparation:            128,
	PitchWheelDepth:          2,
	StereoMixing:             true,
	UseInstruments:           true,
	LinearSlides:             true,
	OldEffects:               false,
	LinkEFG:                  false,
	Channels:                 2,
	Message:                  "a test module\rline 2",
	PatternHighlight_Beat:    4,
	PatternHighlight_Measure: 16,
	ChannelSettings: []common.ChannelSetting{
		{Name: "", InitialVolume: 64, InitialPan: 32},
		{Name: "", InitialVolume: 64, InitialPan: 32},
	},
	Order: []int16{0, 254, 255, 0, 255},
	Instruments: []common.Instrument{
		{
			Name:                  "bass",
			DosFilename:           "bass.iti",
			Fadeout:               7,
			PitchPanCenter:        60, // c-5
			GlobalVolume:          126,
			DefaultPan:            33,
			DefaultPanEnabled:     false,
			RandomVolumeVariation: 3,
			RandomPanVariation:    2,
			FilterCutoff:          3,
			FilterResonance:       2,
			MidiProgram:           0xff,
			MidiBank:              0xffff,
			Notemap:               notemapWithSample(1),

			Envelopes: []common.Envelope{
				{
					Type:    common.EnvelopeTypeVolume,
					Enabled: true,
					Loop:    true,
					Sustain: false,

					LoopStart: 0,
					LoopEnd:   2,

					Nodes: []common.EnvelopeNode{
						{X: 0, Y: 32},
						{X: 9, Y: 51},
						{X: 11, Y: 4},
						{X: 53, Y: 0},
					},
				},
				{
					Type:    common.EnvelopeTypePanning,
					Enabled: true,
					Loop:    true,
					Sustain: false,

					LoopStart: 0,
					LoopEnd:   3,

					Nodes: []common.EnvelopeNode{
						{X: 0, Y: 0},
						{X: 31, Y: 2},
						{X: 89, Y: -2},
						{X: 125, Y: 0},
					},
				},
				{
					Type:         common.EnvelopeTypePitch,
					Enabled:      true,
					Loop:         false,
					Sustain:      true,
					SustainStart: 1,
					SustainEnd:   1,
					Nodes: []common.EnvelopeNode{
						{X: 0, Y: 0},
						{X: 286, Y: +1},
						{X: 310, Y: -1},
					},
				},
			},
		},

		{
			Name:               "stab",
			DosFilename:        "stab.iti",
			Fadeout:            8,
			NewNoteAction:      common.NnaContinue,
			DuplicateCheckType: common.DctOff,
			PitchPanCenter:     60, // c-5
			GlobalVolume:       128,
			DefaultPan:         32,
			DefaultPanEnabled:  false,
			MidiProgram:        0xff,
			MidiBank:           0xffff,
			Notemap:            notemapWithSample(1),

			Envelopes: []common.Envelope{
				{
					Type:      common.EnvelopeTypeVolume,
					Enabled:   true,
					Loop:      false,
					Sustain:   false,
					LoopStart: 0,
					LoopEnd:   3,
					Nodes: []common.EnvelopeNode{
						{X: 0, Y: 27},
						{X: 6, Y: 48},
						{X: 8, Y: 3},
						{X: 57, Y: 0},
					},
				},
				{
					Type:      common.EnvelopeTypePanning,
					Enabled:   false,
					Loop:      false,
					Sustain:   false,
					LoopStart: 0,
					LoopEnd:   0,
					Nodes: []common.EnvelopeNode{
						{X: 0, Y: 0},
						{X: 10, Y: 0},
					},
				},
				{
					Type:         common.EnvelopeTypePitch,
					Enabled:      false,
					Loop:         false,
					Sustain:      false,
					SustainStart: 0,
					SustainEnd:   0,
					Nodes: []common.EnvelopeNode{
						{X: 0, Y: 0},
						{X: 10, Y: 0},
					},
				},
			},
		},
	},
	Samples: []common.Sample{
		{
			Name:        "doodle",
			DosFilename: "doodle.raw",

			GlobalVolume:   64,
			DefaultVolume:  64,
			DefaultPanning: 32,

			S16:             false,
			Stereo:          false,
			Loop:            true,
			Sustain:         false,
			PingPong:        false,
			PingPongSustain: false,

			LoopStart:        0,
			LoopEnd:          64,
			SustainLoopStart: 3,
			SustainLoopEnd:   7,

			C5: 8363,

			VibratoSpeed:    2,
			VibratoDepth:    3,
			VibratoSweep:    4,
			VibratoWaveform: common.SampleVibratoWaveformSquare,

			// This will be int16 if S16 is set, int8 otherwise
			// Stereo samples have left,right interleaved
			Data: common.SampleData{
				Channels: 1,
				Bits:     8,
				Data: []any{
					readBinaryPcm8("test/doodle.raw"),
				},
			},
		},
	},
}

func assertEqualFields(t *testing.T, mod *common.Module, expected *common.Module, exclude []string) {
	excludeMap := make(map[string]bool)
	for _, name := range exclude {
		excludeMap[name] = true
	}

	v := reflect.ValueOf(*mod)
	ev := reflect.ValueOf(*expected)
	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Name
		if excludeMap[fieldName] {
			continue
		}
		assert.Equal(t, ev.Field(i).Interface(), v.Field(i).Interface(), fieldName+" should match")
	}
}

func readBinaryPcm8(filename string) []int8 {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	copied := make([]int8, len(data))
	for i := 0; i < len(data); i++ {
		copied[i] = int8(data[i])
	}

	return copied
}

func TestLoading(t *testing.T) {

	itmod, err := LoadITFile("test/reflection.it")
	assert.NoError(t, err)
	mod := itmod.ToCommon()

	assertEqualFields(t, mod, &itFixture1, []string{"Patterns"})

	rowsSnippet := []common.PatternRow{
		{
			Entries: []common.PatternEntry{
				{
					Channel:       1,
					VolumeCommand: 1,
					VolumeParam:   15,
					Effect:        8,
					EffectParam:   0x32,
				},
			},
		},
		{
			Entries: []common.PatternEntry{
				{
					// The encoding scheme has a way to eliminate repeated bytes, we're checking the volume one here.
					Channel:       1,
					VolumeCommand: 1,
					VolumeParam:   15,
					Effect:        8,
					EffectParam:   0x13,
				},
			},
		},
		{
			//Entries: []common.PatternEntry{}, nil slice
		},
		{
			Entries: []common.PatternEntry{
				{
					Channel:       0,
					Note:          1 + 12*3 + 10, // a#3
					Instrument:    1,
					VolumeCommand: 1,
					VolumeParam:   33,
				},
				{
					Channel:     1,
					Note:        1 + 12*7 + 10, // a#7
					Instrument:  2,
					Effect:      8, // Hxx
					EffectParam: 0x33,
				},
			},
		},
		{
			Entries: []common.PatternEntry{
				{
					Channel:     1,
					Effect:      8, // Hxx
					EffectParam: 0x00,
				},
			},
		},
	}

	assert.Equal(t, rowsSnippet, mod.Patterns[0].Rows[13:18])
}
