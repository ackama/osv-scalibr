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

package python

import (
	"testing"

	"deps.dev/util/resolve"
	"github.com/google/go-cmp/cmp"
	"github.com/google/osv-scalibr/fs"
	"github.com/google/osv-scalibr/guidedremediation/internal/manifest"
)

type testManifest struct {
	FilePath     string
	Root         resolve.Version
	System       resolve.System
	Requirements []resolve.RequirementVersion
}

func checkManifest(t *testing.T, name string, got manifest.Manifest, want testManifest) {
	t.Helper()
	if want.FilePath != got.FilePath() {
		t.Errorf("%s.FilePath() = %q, want %q", name, got.FilePath(), want.FilePath)
	}
	if diff := cmp.Diff(want.Root, got.Root()); diff != "" {
		t.Errorf("%s.Root() (-want +got):\n%s", name, diff)
	}
	if want.System != got.System() {
		t.Errorf("%s.System() = %v, want %v", name, got.System(), want.System)
	}
	if diff := cmp.Diff(want.Requirements, got.Requirements()); diff != "" {
		t.Errorf("%s.Requirements() (-want +got):\n%s", name, diff)
	}
}

func TestRead(t *testing.T) {
	fsys := fs.DirFS("./testdata")
	pypiRW := GetReadWriter()
	got, err := pypiRW.Read("requirements.txt", fsys)
	if err != nil {
		t.Fatalf("error reading manifest: %v", err)
	}
	want := testManifest{
		FilePath: "requirements.txt",
		Root: resolve.Version{
			VersionKey: resolve.VersionKey{
				PackageKey: resolve.PackageKey{
					System: resolve.PyPI,
					Name:   "myproject",
				},
				VersionType: resolve.Concrete,
				Version:     "1.0.0",
			},
		},
		System: resolve.PyPI,
		Requirements: []resolve.RequirementVersion{
			{
				VersionKey: resolve.VersionKey{
					PackageKey: resolve.PackageKey{
						System: resolve.PyPI,
						Name:   "pytest",
					},
					VersionType: resolve.Requirement,
				},
			},
			{
				VersionKey: resolve.VersionKey{
					PackageKey: resolve.PackageKey{
						System: resolve.PyPI,
						Name:   "pytest-cov",
					},
					VersionType: resolve.Requirement,
				},
			},
			{
				VersionKey: resolve.VersionKey{
					PackageKey: resolve.PackageKey{
						System: resolve.PyPI,
						Name:   "beautifulsoup4",
					},
					VersionType: resolve.Requirement,
				},
			},

			{
				VersionKey: resolve.VersionKey{
					PackageKey: resolve.PackageKey{
						System: resolve.PyPI,
						Name:   "docopt",
					},
					VersionType: resolve.Requirement,
					Version:     "== 0.6.1",
				},
			},
			{
				VersionKey: resolve.VersionKey{
					PackageKey: resolve.PackageKey{
						System: resolve.PyPI,
						Name:   "requests",
					},
					VersionType: resolve.Requirement,
					Version:     ">= 2.8.1, == 2.8.*",
				},
			},

			{
				VersionKey: resolve.VersionKey{
					PackageKey: resolve.PackageKey{
						System: resolve.PyPI,
						Name:   "keyring",
					},
					VersionType: resolve.Requirement,
					Version:     ">= 4.1.1",
				},
			},
			{
				VersionKey: resolve.VersionKey{
					PackageKey: resolve.PackageKey{
						System: resolve.PyPI,
						Name:   "coverage",
					},
					VersionType: resolve.Requirement,
					Version:     "!= 3.5",
				},
			},

			{
				VersionKey: resolve.VersionKey{
					PackageKey: resolve.PackageKey{
						System: resolve.PyPI,
						Name:   "Mopidy-Dirble",
					},
					VersionType: resolve.Requirement,
					Version:     "~= 1.1",
				},
			},
		},
	}
	checkManifest(t, "Manifest", got, want)
}
