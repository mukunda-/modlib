// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

/*
Package modlib is for working with tracker (music) module files.
*/
package modlib

import (
	"errors"
	"io"
	"os"

	"go.mukunda.com/modlib/common"
	"go.mukunda.com/modlib/itmod"
)

// Returned when the module format could not be detected.
var ErrUnknownModuleFormat = errors.New("unknown or unsupported module format")

type Module = common.Module

// Load a module by filename.
func LoadModule(filename string) (*Module, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return LoadModuleFromStream(file)

}

// Load a module from an open stream. Seeking is required for module loading.
func LoadModuleFromStream(r io.ReadSeeker) (*Module, error) {
	signature := make([]byte, 4)
	if _, err := io.ReadFull(r, signature); err != nil {
		return nil, err
	}

	if string(signature) == "IMPM" {
		r.Seek(0, io.SeekStart)
		reader := itmod.ItReader{}

		mod, err := reader.ReadItModule(r)
		if err != nil {
			return nil, err
		}

		return mod.ToCommon(), nil
	}

	return nil, ErrUnknownModuleFormat
}
