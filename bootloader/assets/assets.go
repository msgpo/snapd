// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2020 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package assets

import (
	"fmt"
	"sort"

	"github.com/snapcore/snapd/osutil"
)

var registeredAssets = map[string][]byte{}

type forEditions struct {
	// First edition this snippet is used in
	FirstEdition uint
	// Snippet data
	Snippet []byte
}

var registeredEditionSnippets = map[string][]forEditions{}

// registerInternal registers an internal asset under the given name.
func registerInternal(name string, data []byte) {
	if _, ok := registeredAssets[name]; ok {
		panic(fmt.Sprintf("asset %q is already registered", name))
	}
	registeredAssets[name] = data
}

// Internal returns the content of an internal asset registered under the given
// name, or nil when none was found.
func Internal(name string) []byte {
	return registeredAssets[name]
}

type byFirstEdition []forEditions

func (b byFirstEdition) Len() int           { return len(b) }
func (b byFirstEdition) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byFirstEdition) Less(i, j int) bool { return b[i].FirstEdition < b[j].FirstEdition }

// registerSnippetForEditions register a set of snippets, each carrying the
// first edition number it applies to, under a given key.
func registerSnippetForEditions(name string, snippets []forEditions) {
	if _, ok := registeredEditionSnippets[name]; ok {
		panic(fmt.Sprintf("edition snippets %q are already registered", name))
	}

	if !sort.IsSorted(byFirstEdition(snippets)) {
		panic(fmt.Sprintf("edition snippets %q must be sorted in ascending edition number order", name))
	}
	for i := range snippets {
		if i == 0 {
			continue
		}
		if snippets[i-1].FirstEdition == snippets[i].FirstEdition {
			panic(fmt.Sprintf(`first edition %v repeated in edition snippets %q`,
				snippets[i].FirstEdition, name))
		}
	}
	registeredEditionSnippets[name] = snippets
}

// SnippetForEdition returns a snippet registered under given name,
// applicable for the provided edition number.
func SnippetForEdition(name string, edition uint) []byte {
	snippets := registeredEditionSnippets[name]
	if snippets == nil {
		return nil
	}
	var current []byte
	// snippets are sorted by ascending edition number when adding
	for _, snip := range snippets {
		if edition >= snip.FirstEdition {
			current = snip.Snippet
		} else {
			break
		}
	}
	return current
}

// MockInternal mocks the contents of an internal asset for use in testing.
func MockInternal(name string, data []byte) (restore func()) {
	osutil.MustBeTestBinary("mocking can be done only in tests")

	old, ok := registeredAssets[name]
	registeredAssets[name] = data
	return func() {
		if ok {
			registeredAssets[name] = old
		} else {
			delete(registeredAssets, name)
		}
	}
}
