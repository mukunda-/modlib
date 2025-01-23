// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package itmod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mukunda.com/modlib/common"
)

func TestLoading(t *testing.T) {

	mod, err := LoadITFile("test/reflection.it")
	assert.NoError(t, err)

	assert.Equal(t, "reflection", mod.Title)
	assert.Len(t, mod.Order, 1)

	assert.Len(t, mod.ChannelSettings, 64)
	assert.Len(t, mod.Instruments, 2)
	assert.Len(t, mod.Samples, 1)
	assert.Len(t, mod.Patterns, 1)

	// Flags todo?
	assert.True(t, mod.StereoMixing)
	assert.True(t, mod.UseInstruments)
	//assert.Equal(t, uint16(0x4d), mod.Flags) // stereo, instruments + samples, midi pitch(?)

	//assert.Equal(t, uint16(7), mod.Special) // message + edithistory (??) + highlight (??)

	assert.EqualValues(t, 45, mod.GlobalVolume)
	assert.EqualValues(t, 48, mod.MixingVolume)
	assert.EqualValues(t, 6, mod.InitialSpeed)
	assert.EqualValues(t, 135, mod.InitialTempo)
	assert.EqualValues(t, 128, mod.PanSeparation)
	//assert.Equal(t, uint8(2), mod.PWD) // I don't know how to set this in openmpt

	assert.EqualValues(t, "a test module", mod.Message)

	assert.EqualValues(t, 32, mod.ChannelSettings[0].InitialPan)
	assert.EqualValues(t, 32, mod.ChannelSettings[1].InitialPan)

	assert.EqualValues(t, 64, mod.ChannelSettings[0].InitialVolume)
	assert.EqualValues(t, 64, mod.ChannelSettings[1].InitialVolume)

	assert.EqualValues(t, 0, mod.Order[0])

	i0 := &mod.Instruments[0]
	assert.EqualValues(t, "bass", i0.Name)
	assert.EqualValues(t, "bass.iti", i0.DosFilename)
	assert.EqualValues(t, 0, i0.NewNoteAction)
	assert.EqualValues(t, 0, i0.DuplicateCheckType)
	assert.EqualValues(t, 0, i0.DuplicateCheckAction)
	assert.EqualValues(t, 8, i0.Fadeout) //256/32
	assert.EqualValues(t, 0, i0.PitchPanSeparation)
	assert.EqualValues(t, 60, i0.PitchPanCenter) // C-5

	assert.EqualValues(t, 128, i0.GlobalVolume)
	assert.EqualValues(t, 64, i0.DefaultPan)
	assert.False(t, i0.DefaultPanEnabled)

	assert.EqualValues(t, 0, i0.RandomVolumeVariation)
	assert.EqualValues(t, 0, i0.RandomPanVariation)
	//assert.Equal( t, uint16(), TrackerVersion         )

	assert.EqualValues(t, 0, i0.FilterCutoff)
	assert.EqualValues(t, 0, i0.FilterResonance)

	assert.EqualValues(t, 0, i0.MidiChannel)
	assert.EqualValues(t, 0xff, i0.MidiProgram)
	assert.EqualValues(t, 0xffff, i0.MidiBank)

	for i := 0; i < 120; i++ {
		assert.EqualValues(t, i, i0.Notemap[i].Note)
		assert.EqualValues(t, 1, i0.Notemap[i].Sample)

		// Second instrument uses same sample (1)
		assert.EqualValues(t, i, mod.Instruments[1].Notemap[i].Note)
		assert.EqualValues(t, 1, mod.Instruments[1].Notemap[i].Sample)
	}

	e := &mod.Instruments[0].Envelopes[0]
	assert.Equal(t, common.EnvelopeTypeVolume, e.Type)
	assert.True(t, e.Enabled)
	assert.True(t, e.Loop)
	assert.False(t, e.Sustain)

	assert.Len(t, e.Nodes, 4)
	assert.EqualValues(t, 0, e.LoopStart)
	assert.EqualValues(t, 2, e.LoopEnd)

	e = &mod.Instruments[0].Envelopes[1]
	assert.Len(t, e.Nodes, 4)
	assert.EqualValues(t, 0, e.LoopStart)
	assert.EqualValues(t, 3, e.LoopEnd)
	assert.True(t, e.Enabled)

	e = &mod.Instruments[0].Envelopes[2]
	assert.Len(t, e.Nodes, 2)
	assert.True(t, e.Enabled)

}
