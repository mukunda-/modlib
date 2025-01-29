// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package modlib

import "go.mukunda.com/modlib/common"

// Export all common types into this package.

type Module = common.Module
type ChannelSetting = common.ChannelSetting
type Instrument = common.Instrument
type NotemapEntry = common.NotemapEntry
type EnvelopeType = common.EnvelopeType
type Envelope = common.Envelope
type EnvelopeNode = common.EnvelopeNode
type Sample = common.Sample
type SampleData = common.SampleData
type Pattern = common.Pattern
type PatternRow = common.PatternRow
type PatternEntry = common.PatternEntry

const (
	UnknownSource = common.UnknownSource
	ModSource     = common.ModSource
	S3mSource     = common.S3mSource
	XmSource      = common.XmSource
	ItSource      = common.ItSource
)

const (
	NnaNoteCut  = common.NnaNoteCut
	NnaContinue = common.NnaContinue
	NnaNoteOff  = common.NnaNoteOff
	NnaFade     = common.NnaFade
)

const (
	DctOff        = common.DctOff
	DctNote       = common.DctNote
	DctSample     = common.DctSample
	DctInstrument = common.DctInstrument
	DctPlugin     = common.DctPlugin
)
