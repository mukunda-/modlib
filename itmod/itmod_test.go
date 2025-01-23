// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package itmod

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mukunda.com/modlib/common"
)

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
	Order: []int16{0, 0},
	// Other: map[string]any{

	// 	"cwtv":      0x5131, // OpenMPT
	// 	"cmwt":      0x214,  // Compatible with IT2.14
	// 	"itflags":   0x4D,
	// 	"itspecial": 0x7,
	// },
}

func validateHeader(t *testing.T, mod *common.Module, expected *common.Module) {
	excludedFields := map[string]bool{
		//"Other": true,
	}
	v := reflect.ValueOf(*mod)
	ev := reflect.ValueOf(*expected)
	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Name
		if excludedFields[fieldName] {
			continue
		}
		assert.Equal(t, ev.Field(i).Interface(), v.Field(i).Interface(), fieldName)
	}
	// assert.Equal(t, expected.FilePath, mod.FilePath)
	// assert.Equal(t, expected.Title, mod.Title)
	// assert.Equal(t, expected.GlobalVolume, mod.GlobalVolume)
	// assert.Equal(t, expected.MixingVolume, mod.MixingVolume)
	// assert.Equal(t, expected.InitialSpeed, mod.InitialSpeed)
	// assert.Equal(t, expected.InitialTempo, mod.InitialTempo)
	// assert.Equal(t, expected.PanSeparation, mod.PanSeparation)
	// assert.Equal(t, expected.PitchWheelDepth, mod.PitchWheelDepth)
	// assert.Equal(t, expected.StereoMixing, mod.StereoMixing)
	// assert.Equal(t, expected.UseInstruments, mod.UseInstruments)
	// assert.Equal(t, expected.LinearSlides, mod.LinearSlides)
	// assert.Equal(t, expected.OldEffects, mod.OldEffects)
	// assert.Equal(t, expected.LinkEFG, mod.LinkEFG)
	// assert.Equal(t, expected.Channels, mod.Channels)
	// assert.Equal(t, expected.Message, mod.Message)
	// assert.Equal(t, expected.PatternHighlight_Beat, mod.PatternHighlight_Beat)
	// assert.Equal(t, expected.PatternHighlight_Measure, mod.PatternHighlight_Measure)
	// assert.Equal(t, expected.ChannelSettings, mod.ChannelSettings)
	// assert.Equal(t, expected.Order, mod.Order)
}

func TestLoading(t *testing.T) {

	mod, err := LoadITFile("test/reflection.it")
	assert.NoError(t, err)

	validateHeader(t, mod, &itFixture1)
	//validateInstruments(t, mod)

	// i0 := &mod.Instruments[0]
	// assert.EqualValues(t, "bass", i0.Name)
	// assert.EqualValues(t, "bass.iti", i0.DosFilename)
	// assert.EqualValues(t, 0, i0.NewNoteAction)
	// assert.EqualValues(t, 0, i0.DuplicateCheckType)
	// assert.EqualValues(t, 0, i0.DuplicateCheckAction)
	// assert.EqualValues(t, 8, i0.Fadeout) //256/32
	// assert.EqualValues(t, 0, i0.PitchPanSeparation)
	// assert.EqualValues(t, 60, i0.PitchPanCenter) // C-5

	// assert.EqualValues(t, 126, i0.GlobalVolume)
	// assert.EqualValues(t, 32, i0.DefaultPan)
	// assert.False(t, i0.DefaultPanEnabled)

	// assert.EqualValues(t, 0, i0.RandomVolumeVariation)
	// assert.EqualValues(t, 0, i0.RandomPanVariation)
	// //assert.Equal( t, uint16(), TrackerVersion         )

	// assert.EqualValues(t, 0, i0.FilterCutoff)
	// assert.EqualValues(t, 0, i0.FilterResonance)

	// assert.EqualValues(t, 0, i0.MidiChannel)
	// assert.EqualValues(t, 0xff, i0.MidiProgram)
	// assert.EqualValues(t, 0xffff, i0.MidiBank)

	// for i := 0; i < 120; i++ {
	// 	assert.EqualValues(t, i, i0.Notemap[i].Note)
	// 	assert.EqualValues(t, 1, i0.Notemap[i].Sample)

	// 	// Second instrument uses same sample (1)
	// 	assert.EqualValues(t, i, mod.Instruments[1].Notemap[i].Note)
	// 	assert.EqualValues(t, 1, mod.Instruments[1].Notemap[i].Sample)
	// }

	// e := &mod.Instruments[0].Envelopes[0]
	// assert.Equal(t, common.EnvelopeTypeVolume, e.Type)
	// assert.True(t, e.Enabled)
	// assert.True(t, e.Loop)
	// assert.False(t, e.Sustain)

	// assert.Len(t, e.Nodes, 4)
	// assert.EqualValues(t, 0, e.LoopStart)
	// assert.EqualValues(t, 2, e.LoopEnd)

	// e = &mod.Instruments[0].Envelopes[1]
	// assert.EqualValues(t, common.EnvelopeTypePanning, e.Type)
	// assert.Len(t, e.Nodes, 4)
	// assert.EqualValues(t, 0, e.LoopStart)
	// assert.EqualValues(t, 3, e.LoopEnd)
	// assert.True(t, e.Enabled)

	// e = &mod.Instruments[0].Envelopes[2]
	// assert.EqualValues(t, common.EnvelopeTypePitch, e.Type)
	// assert.Len(t, e.Nodes, 3)
	// assert.True(t, e.Enabled)
	// assert.False(t, e.Loop)
	// assert.True(t, e.Sustain)
	// assert.EqualValues(t, 1, e.SustainStart)
	// assert.EqualValues(t, 1, e.SustainEnd)
}
