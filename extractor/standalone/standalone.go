// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package standalone provides a way to extract in a standalone mode (e.g. a command).
package standalone

import (
	"context"
	"path/filepath"

	"github.com/google/osv-scalibr/extractor"
	scalibrfs "github.com/google/osv-scalibr/fs"
	"github.com/google/osv-scalibr/plugin"
)

// Extractor is an interface for plugins that extract information independently. For
// example, a plugin that executes a command or retrieves information from only one file.
type Extractor interface {
	extractor.Extractor
	// Extract the information.
	Extract(ctx context.Context, input *ScanInput) ([]*extractor.Inventory, error)
}

// Config for running standalone extractors.
type Config struct {
	Extractors []Extractor
	ScanRoot   *scalibrfs.ScanRoot
}

// ScanInput provides information for the extractor about the scan.
type ScanInput struct {
	// FS for file access. This is rooted at Root.
	FS scalibrfs.FS
	// The root directory to start the extraction from.
	Root string
}

// Run the extractors that are specified in the config.
func Run(ctx context.Context, config *Config) ([]*extractor.Inventory, []*plugin.Status, error) {
	var inventories []*extractor.Inventory
	var statuses []*plugin.Status

	if !config.ScanRoot.IsVirtual() {
		p, err := filepath.Abs(config.ScanRoot.Path)
		if err != nil {
			return nil, nil, err
		}
		config.ScanRoot.Path = p
	}

	scanInput := &ScanInput{
		FS:   config.ScanRoot.FS,
		Root: config.ScanRoot.Path,
	}

	for _, extractor := range config.Extractors {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}

		inv, err := extractor.Extract(ctx, scanInput)
		if err != nil {
			statuses = append(statuses, plugin.StatusFromErr(extractor, false, err))
			continue
		}
		for _, i := range inv {
			i.Extractor = extractor
		}

		inventories = append(inventories, inv...)
		statuses = append(statuses, plugin.StatusFromErr(extractor, false, nil))
	}

	return inventories, statuses, nil
}