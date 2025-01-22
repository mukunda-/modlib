/* itmod
 * Copyright 2025 Mukunda Johnson (mukunda.com)
 * Licensed under MIT
 */

package itmod

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoading(t *testing.T) {
	var mod ITModule
	mod.LoadFromFile("test/reflection.it")

	assert.Equal(t, "reflection", mod.Title)
	assert.Equal(t, uint16(2), mod.Length)

	assert.Equal(t, uint16(2), mod.InstrumentCount)
	assert.Equal(t, uint16(1), mod.SampleCount)
	assert.Equal(t, uint16(1), mod.PatternCount)
	// Flags todo?
	assert.Equal(t, uint16(0x4d), mod.Flags) // stereo, instruments + samples, midi pitch(?)

	assert.Equal(t, uint16(7), mod.Special) // message + edithistory (??) + highlight (??)

	assert.Equal(t, uint8(45), mod.GlobalVolume)
	assert.Equal(t, uint8(48), mod.MixingVolume)
	assert.Equal(t, uint8(6), mod.InitialSpeed)
	assert.Equal(t, uint8(135), mod.InitialTempo)
	assert.Equal(t, uint8(128), mod.Sep)
	//assert.Equal(t, uint8(2), mod.PWD) // I don't know how to set this in openmpt

	assert.Equal(t, "a test module", mod.Message)

	assert.Equal(t, uint8(32), mod.ChannelPan[0])
	assert.Equal(t, uint8(32), mod.ChannelPan[1])

	assert.Equal(t, uint8(64), mod.ChannelVolume[0])
	assert.Equal(t, uint8(64), mod.ChannelVolume[1])

	assert.Equal(t, uint8(0), mod.Orders[0])
	assert.Equal(t, uint8(255), mod.Orders[1])

	assert.Equal(t, "bass", mod.Instruments[0].Name)
	assert.Equal(t, "bass.iti", mod.Instruments[0].DOSFilename)
	assert.Equal(t, uint8(0), mod.Instruments[0].NewNoteAction)
	assert.Equal(t, uint8(0), mod.Instruments[0].DuplicateCheckType)
	assert.Equal(t, uint8(0), mod.Instruments[0].DuplicateCheckAction)
	assert.Equal(t, uint16(8), mod.Instruments[0].Fadeout) //256/32
	assert.Equal(t, uint8(0), mod.Instruments[0].PPS)
	assert.Equal(t, uint8(0x3c), mod.Instruments[0].PPC) // Unsure what 3c is

	assert.Equal(t, uint8(128), mod.Instruments[0].GlobalVolume)
	assert.Equal(t, uint8(32|128), mod.Instruments[0].DefaultPan)
	assert.Equal(t, uint8(0), mod.Instruments[0].RandomVolume)
	assert.Equal(t, uint8(0), mod.Instruments[0].RandomPanning)
	//assert.Equal( t, uint16(), TrackerVersion         )
	assert.Equal(t, uint8(1), mod.Instruments[0].NumberOfSamples)
	assert.Equal(t, uint8(0), mod.Instruments[0].InitialFilterCutoff)
	assert.Equal(t, uint8(0), mod.Instruments[0].InitialFilterResonance)

	assert.Equal(t, uint8(0), mod.Instruments[0].MidiChannel)
	assert.Equal(t, uint8(0xff), mod.Instruments[0].MidiProgram)
	assert.Equal(t, uint16(0xffff), mod.Instruments[0].MidiBank)

	for i := 0; i < 120; i++ {
		assert.Equal(t, uint8(i), mod.Instruments[0].Notemap[i].Note)
		assert.Equal(t, uint8(1), mod.Instruments[0].Notemap[i].Sample)

		// Second instrument uses same sample.
		assert.Equal(t, uint8(i), mod.Instruments[1].Notemap[i].Note)
		assert.Equal(t, uint8(1), mod.Instruments[1].Notemap[i].Sample)
	}

	assert.True(t, mod.Instruments[0].VolumeEnvelope.Enabled)
	assert.True(t, mod.Instruments[0].VolumeEnvelope.Loop)
	assert.False(t, mod.Instruments[0].VolumeEnvelope.Sustain)
	assert.False(t, mod.Instruments[0].VolumeEnvelope.IsFilter)

	assert.Equal(t, uint8(4), mod.Instruments[0].VolumeEnvelope.Length)
	assert.Equal(t, uint8(0), mod.Instruments[0].VolumeEnvelope.LoopStart)
	assert.Equal(t, uint8(2), mod.Instruments[0].VolumeEnvelope.LoopEnd)

	assert.Equal(t, uint8(4), mod.Instruments[0].PanningEnvelope.Length)
	assert.Equal(t, uint8(0), mod.Instruments[0].PanningEnvelope.LoopStart)
	assert.Equal(t, uint8(3), mod.Instruments[0].PanningEnvelope.LoopEnd)
	assert.True(t, mod.Instruments[0].PanningEnvelope.Enabled)

	assert.Equal(t, uint8(2), mod.Instruments[0].PitchEnvelope.Length)
	assert.True(t, mod.Instruments[0].PitchEnvelope.Enabled)

}
