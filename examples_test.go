// modlib
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package modlib_test

import (
	"fmt"

	"go.mukunda.com/modlib"
)

func ExampleLoadModule() error {

	// Load a module by filename.
	mod, err := modlib.LoadModule("my_module.it")
	if err != nil {
		return err
	}

	fmt.Println("Title:", mod.Title)
	return nil
}
