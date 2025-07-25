// Copyright 2025 Google LLC
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

package packagelockjson_test

import (
	"context"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem"
	"github.com/google/osv-scalibr/extractor/filesystem/internal/units"
	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/packagelockjson"
	"github.com/google/osv-scalibr/extractor/filesystem/osv"
	"github.com/google/osv-scalibr/extractor/filesystem/simplefileapi"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/purl"
	"github.com/google/osv-scalibr/stats"
	"github.com/google/osv-scalibr/testing/extracttest"
	"github.com/google/osv-scalibr/testing/fakefs"
	"github.com/google/osv-scalibr/testing/testcollector"
)

func TestExtractor_FileRequired(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		fileSizeBytes    int64
		maxFileSizeBytes int64
		wantRequired     bool
		wantResultMetric stats.FileRequiredResult
	}{
		{
			name:         "Empty path",
			path:         filepath.FromSlash(""),
			wantRequired: false,
		},
		{
			name:             "package-lock.json",
			path:             filepath.FromSlash("package-lock.json"),
			wantRequired:     true,
			wantResultMetric: stats.FileRequiredResultOK,
		},
		{
			name:             "package-lock.json at the end of a path",
			path:             filepath.FromSlash("path/to/my/package-lock.json"),
			wantRequired:     true,
			wantResultMetric: stats.FileRequiredResultOK,
		},
		{
			name:         "package-lock.json as path segment",
			path:         filepath.FromSlash("path/to/my/package-lock.json/file"),
			wantRequired: false,
		},
		{
			name:         "package-lock.json.file (wrong extension)",
			path:         filepath.FromSlash("path/to/my/package-lock.json.file"),
			wantRequired: false,
		},
		{
			name:         "path.to.my.package.lock.json",
			path:         filepath.FromSlash("path.to.my.package.lock.json"),
			wantRequired: false,
		},
		{
			name:         "skip from inside node_modules dir",
			path:         filepath.FromSlash("foo/node_modules/bar/package-lock.json"),
			wantRequired: false,
		},
		{
			name:             "package-lock.json required if file size < max file size",
			path:             "foo/package-lock.json",
			fileSizeBytes:    100 * units.KiB,
			maxFileSizeBytes: 1 * units.MiB,
			wantRequired:     true,
			wantResultMetric: stats.FileRequiredResultOK,
		},
		{
			name:             "package-lock.json required if file size == max file size",
			path:             "foo/package-lock.json",
			fileSizeBytes:    1 * units.MiB,
			maxFileSizeBytes: 1 * units.MiB,
			wantRequired:     true,
			wantResultMetric: stats.FileRequiredResultOK,
		},
		{
			name:             "package-lock.json not required if file size > max file size",
			path:             "foo/package-lock.json",
			fileSizeBytes:    1 * units.MiB,
			maxFileSizeBytes: 100 * units.KiB,
			wantRequired:     false,
			wantResultMetric: stats.FileRequiredResultSizeLimitExceeded,
		},
		{
			name:             "package-lock.json required if max file size set to 0",
			path:             "foo/package-lock.json",
			fileSizeBytes:    1 * units.MiB,
			maxFileSizeBytes: 0,
			wantRequired:     true,
			wantResultMetric: stats.FileRequiredResultOK,
		},
		{
			name:             "npm-shrinkwrap.json",
			path:             filepath.FromSlash("npm-shrinkwrap.json"),
			wantRequired:     true,
			wantResultMetric: stats.FileRequiredResultOK,
		},
		{
			name:             "npm-shrinkwrap.json at the end of a path",
			path:             filepath.FromSlash("path/to/my/npm-shrinkwrap.json"),
			wantRequired:     true,
			wantResultMetric: stats.FileRequiredResultOK,
		},
		{
			name:         "npm-shrinkwrap.json as path segment",
			path:         filepath.FromSlash("path/to/my/npm-shrinkwrap.json/file"),
			wantRequired: false,
		},
		{
			name:         "npm-shrinkwrap.json.file (wrong extension)",
			path:         filepath.FromSlash("path/to/my/npm-shrinkwrap.json.file"),
			wantRequired: false,
		},
		{
			name:         "path.to.my.npm-shrinkwrap.json",
			path:         filepath.FromSlash("path.to.my.npm-shrinkwrap.json"),
			wantRequired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := testcollector.New()
			var e filesystem.Extractor = packagelockjson.New(
				packagelockjson.Config{
					Stats:            collector,
					MaxFileSizeBytes: tt.maxFileSizeBytes,
				},
			)

			// Set default size if not provided.
			fileSizeBytes := tt.fileSizeBytes
			if fileSizeBytes == 0 {
				fileSizeBytes = 100 * units.KiB
			}

			isRequired := e.FileRequired(simplefileapi.New(tt.path, fakefs.FakeFileInfo{
				FileName: filepath.Base(tt.path),
				FileMode: fs.ModePerm,
				FileSize: fileSizeBytes,
			}))
			if isRequired != tt.wantRequired {
				t.Fatalf("FileRequired(%s): got %v, want %v", tt.path, isRequired, tt.wantRequired)
			}

			gotResultMetric := collector.FileRequiredResult(tt.path)
			if gotResultMetric != tt.wantResultMetric {
				t.Errorf("FileRequired(%s) recorded result metric %v, want result metric %v", tt.path, gotResultMetric, tt.wantResultMetric)
			}
		})
	}
}

func TestMetricCollector(t *testing.T) {
	tests := []struct {
		name             string
		inputConfig      extracttest.ScanInputMockConfig
		wantResultMetric stats.FileExtractedResult
	}{
		{
			name: "invalid package-lock.json",
			inputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/not-json.txt",
			},
			wantResultMetric: stats.FileExtractedResultErrorUnknown,
		},
		{
			name: "valid package-lock.json",
			inputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/one-package.v1.json",
			},
			wantResultMetric: stats.FileExtractedResultSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := testcollector.New()
			extr := packagelockjson.New(packagelockjson.Config{
				Stats: collector,
			})

			scanInput := extracttest.GenerateScanInputMock(t, tt.inputConfig)
			defer extracttest.CloseTestScanInput(t, scanInput)

			// Results are tested in the other files
			_, _ = extr.Extract(context.Background(), &scanInput)

			gotResultMetric := collector.FileExtractedResult(tt.inputConfig.Path)
			if gotResultMetric != tt.wantResultMetric {
				t.Errorf("Extract(%s) recorded result metric %v, want result metric %v", tt.inputConfig.Path, gotResultMetric, tt.wantResultMetric)
			}

			gotFileSizeMetric := collector.FileExtractedFileSize(tt.inputConfig.Path)
			if gotFileSizeMetric != scanInput.Info.Size() {
				t.Errorf("Extract(%s) recorded file size %v, want file size %v", tt.inputConfig.Path, gotFileSizeMetric, scanInput.Info.Size())
			}
		})
	}
}

func TestExtractor_Extract_Shrinkwrap_JSON(t *testing.T) {
	tests := []extracttest.TestTableEntry{
		{
			Name: "invalid json",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/not-json.txt",
			},
			WantErr: extracttest.ContainsErrStr{Str: "could not extract from"},
		},
		{
			Name: "valid package-lock.json only",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/package-lock-only/package-lock.json",
			},
			WantPackages: []*extractor.Package{
				{
					Name:      "wrappy",
					Version:   "1.0.2",
					PURLType:  purl.TypeNPM,
					Locations: []string{"testdata/package-lock-only/package-lock.json"},
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					Name:      "supports-color",
					Version:   "5.5.0",
					PURLType:  purl.TypeNPM,
					Locations: []string{"testdata/package-lock-only/package-lock.json"},
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "valid npm-shrinkwrap.json only",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/npm-shrinkwrap-only/npm-shrinkwrap.json",
			},
			WantPackages: []*extractor.Package{
				{
					Name:      "wrappy",
					Version:   "1.0.2",
					PURLType:  purl.TypeNPM,
					Locations: []string{"testdata/npm-shrinkwrap-only/npm-shrinkwrap.json"},
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					Name:      "supports-color",
					Version:   "5.5.0",
					PURLType:  purl.TypeNPM,
					Locations: []string{"testdata/npm-shrinkwrap-only/npm-shrinkwrap.json"},
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "valid package-lock.json and npm-shrinkwrap.json and extract package-lock.json",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/both/package-lock.json",
			},
			WantPackages: nil,
		},
		{
			Name: "valid package-lock.json and npm-shrinkwrap.json and extract npm-shrinkwrap.json",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/both/npm-shrinkwrap.json",
			},
			WantPackages: []*extractor.Package{
				{
					Name:      "wrappy",
					Version:   "1.0.2",
					PURLType:  purl.TypeNPM,
					Locations: []string{"testdata/both/npm-shrinkwrap.json"},
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					Name:      "supports-color",
					Version:   "5.5.0",
					PURLType:  purl.TypeNPM,
					Locations: []string{"testdata/both/npm-shrinkwrap.json"},
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			collector := testcollector.New()
			extr := packagelockjson.New(packagelockjson.Config{
				Stats: collector,
			})

			scanInput := extracttest.GenerateScanInputMock(t, tt.InputConfig)
			defer extracttest.CloseTestScanInput(t, scanInput)

			got, err := extr.Extract(context.Background(), &scanInput)

			if diff := cmp.Diff(tt.WantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s.Extract(%q) error diff (-want +got):\n%s", extr.Name(), tt.InputConfig.Path, diff)
				return
			}

			wantInv := inventory.Inventory{Packages: tt.WantPackages}
			if diff := cmp.Diff(wantInv, got, cmpopts.SortSlices(extracttest.PackageCmpLess)); diff != "" {
				t.Errorf("%s.Extract(%q) diff (-want +got):\n%s", extr.Name(), tt.InputConfig.Path, diff)
			}

			gotFileSizeMetric := collector.FileExtractedFileSize(tt.InputConfig.Path)
			if gotFileSizeMetric != scanInput.Info.Size() {
				t.Errorf("Extract(%s) recorded file size %v, want file size %v", tt.InputConfig.Path, gotFileSizeMetric, scanInput.Info.Size())
			}
		})
	}
}
