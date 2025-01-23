// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package itmod

import (
	"strings"

	"go.mukunda.com/modlib/common"
)

func iif[T any](cond bool, a, b T) T {
	if cond {
		return a
	} else {
		return b
	}
}

func (itm *ItModule) ToCommon() *common.Module {
	m := new(common.Module)
	m.Source = common.ItSource

	m.Title = strings.TrimRight(string(itm.Header.Title[:]), "\000")

	m.StereoMixing = (itm.Header.Flags & ItFlagStereo) != 0
	m.UseInstruments = (itm.Header.Flags & ItFlagInstruments) != 0
	m.LinearSlides = (itm.Header.Flags & ItFlagLinearSlides) != 0
	m.OldEffects = (itm.Header.Flags & ItFlagOldEffects) != 0
	m.LinkEFG = (itm.Header.Flags & ItFlagLinkEFG) != 0

	m.PatternHighlight_Beat = int16(itm.Header.PatternHighlightBeat)
	m.PatternHighlight_Measure = int16(itm.Header.PatternHighlightMeasure)

	m.GlobalVolume = int16(itm.Header.GlobalVolume)
	m.MixingVolume = int16(itm.Header.MixingVolume)
	m.InitialSpeed = int16(itm.Header.InitialSpeed)
	m.InitialTempo = int16(itm.Header.InitialTempo)
	m.PanSeparation = int16(itm.Header.Sep)
	m.PitchWheelDepth = int16(itm.Header.PWD)

	m.ChannelSettings = make([]common.ChannelSetting, 64)

	for i := 0; i < 64; i++ {
		m.ChannelSettings[i].InitialPan = int16(itm.Header.ChannelPan[i])
	}

	for i := 0; i < 64; i++ {
		m.ChannelSettings[i].InitialVolume = int16(itm.Header.ChannelVolume[i])
	}

	for _, order := range itm.Orders {
		m.Order = append(m.Order, int16(order))
	}

	for _, instrument := range itm.Instruments {
		m.Instruments = append(m.Instruments, instrument.ToCommon())
	}

	for _, sample := range itm.Samples {
		m.Samples = append(m.Samples, sample.ToCommon())
	}

	// Compute number of channels.
	channels := int16(0)

	for _, pattern := range itm.Patterns {
		p := pattern.ToCommon()
		m.Patterns = append(m.Patterns, p)
		channels = max(channels, int16(p.Channels))
	}

	m.Channels = channels
	m.ChannelSettings = m.ChannelSettings[:channels]

	m.Message = strings.TrimRight(string(itm.Message), "\000")

	return m
}

func (iti *ItInstrument) ToCommon() common.Instrument {
	var ins common.Instrument

	ins.Name = strings.TrimRight(string(iti.Name[:]), "\000")
	ins.DosFilename = strings.TrimRight(string(iti.DosFilename[:]), "\000")
	ins.NewNoteAction = int16(iti.NewNoteAction)
	ins.DuplicateCheckType = int16(iti.DuplicateCheckType)
	ins.DuplicateCheckAction = int16(iti.DuplicateCheckAction)
	ins.Fadeout = int16(iti.Fadeout)

	ins.PitchPanSeparation = int16(iti.PPS)
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
		ins.Envelopes = append(ins.Envelopes, translateEnvelope(&iti.Envelopes[i], i))
	}

	return ins
}

func (its *ItSample) ToCommon() common.Sample {
	var s common.Sample
	s.Name = strings.TrimRight(string(its.Header.Name[:]), "\000")
	s.DosFilename = strings.TrimRight(string(its.Header.DosFilename[:]), "\000")

	s.GlobalVolume = int16(its.Header.GlobalVolume)
	s.DefaultVolume = int16(its.Header.DefaultVolume)
	s.DefaultPanning = int16(its.Header.DefaultPanning)

	s.S16 = (its.Header.Flags & SampFlag16bit) != 0
	s.Stereo = (its.Header.Flags & SampFlagStereo) != 0
	s.Loop = (its.Header.Flags & SampFlagLoop) != 0
	s.Sustain = (its.Header.Flags & SampFlagSustain) != 0
	s.PingPong = (its.Header.Flags & SampFlagPingPong) != 0
	s.PingPongSustain = (its.Header.Flags & SampFlagPingPongSustain) != 0

	s.LoopStart = int(its.Header.LoopStart)
	s.LoopEnd = int(its.Header.LoopEnd)
	s.SustainLoopStart = int(its.Header.SustainLoopStart)
	s.SustainLoopEnd = int(its.Header.SustainLoopEnd)

	s.C5 = int(its.Header.C5)

	s.VibratoSpeed = int16(its.Header.VibratoSpeed)
	s.VibratoDepth = int16(its.Header.VibratoDepth)
	s.VibratoSweep = int16(its.Header.VibratoSweep)
	s.VibratoWaveform = int16(its.Header.VibratoWaveform)

	s.Data = common.SampleData{
		Channels: int8(iif(s.Stereo, 2, 1)),
		Bits:     int8(iif(s.S16, 16, 8)),
		Data:     its.Data,
	}

	return s
}

func translateEnvelope(itenv *ItEnvelope, index int) common.Envelope {
	var env common.Envelope

	env.Enabled = (itenv.Flags & EnvFlagEnabled) != 0
	env.Loop = (itenv.Flags & EnvFlagLoop) != 0
	env.Sustain = (itenv.Flags & EnvFlagSustain) != 0

	env.LoopStart = int16(itenv.LoopStart)
	env.LoopEnd = int16(itenv.LoopEnd)
	env.SustainStart = int16(itenv.SustainStart)
	env.SustainEnd = int16(itenv.SustainEnd)

	if index == 0 {
		env.Type = common.EnvelopeTypeVolume
	} else if index == 1 {
		env.Type = common.EnvelopeTypePanning
	} else if index == 2 {
		env.Type = common.EnvelopeTypePitch
		if itenv.Flags&EnvFlagFilter != 0 {
			env.Type = common.EnvelopeTypeFilter
		}
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

	return env
}

func translateNote(note uint8) uint8 {
	if note <= 120 {
		return note + 1 // Normal note, map to +1 so zero is "empty".
	} else if note == 254 || note == 255 {
		return note // Note Cut, Note Off
	} else if note >= 120 {
		return 253 // Fade out
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

func (itp *ItPattern) ToCommon() common.Pattern {
	var p common.Pattern

	// Unpack data
	dataRead := 0
	data := itp.Data

	nextByte := func() byte {
		if dataRead >= len(data) {
			// Should we throw an error if the pattern data is corrupted?
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

	channels := 0

	for row := 0; row < int(itp.Header.Rows); row++ {
		patternRow := common.PatternRow{}
		for {
			channelSelect := nextByte()
			if channelSelect == 0 {
				break
			}

			entry := common.PatternEntry{}

			channel := int((channelSelect - 1) & 63)
			entry.Channel = uint8(channel)
			if channel >= channels {
				channels = channel + 1
			}

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

			patternRow.Entries = append(patternRow.Entries, entry)
		}

		p.Rows = append(p.Rows, patternRow)
	}

	p.Channels = int16(channels)

	return p
}
