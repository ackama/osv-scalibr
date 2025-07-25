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

package gcpexpressmode_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/osv-scalibr/veles"
	"github.com/google/osv-scalibr/veles/secrets/gcpexpressmode"
)

const (
	testKey = "AQ.Ab8Rthat-is-1-very-nice-KeyYouGotThere_ShameIfLost"
)

func TestDetector_truePositives(t *testing.T) {
	engine, err := veles.NewDetectionEngine([]veles.Detector{gcpexpressmode.NewDetector()})
	if err != nil {
		t.Fatalf("veles.NewDetectionEngine() error: %v", err)
	}
	cases := []struct {
		name  string
		input string
		want  []veles.Secret
	}{
		{
			name:  "simple matching string",
			input: testKey,
			want: []veles.Secret{
				gcpexpressmode.APIKey{Key: testKey},
			},
		},
		{
			name:  "match at end of string",
			input: "API_KEY=" + testKey,
			want: []veles.Secret{
				gcpexpressmode.APIKey{Key: testKey},
			},
		},
		{
			name:  "match in middle of string",
			input: `API_KEY="` + testKey + `"`,
			want: []veles.Secret{
				gcpexpressmode.APIKey{Key: testKey},
			},
		},
		{
			name:  "multiple matches",
			input: testKey + "\n" + testKey + testKey,
			want: []veles.Secret{
				gcpexpressmode.APIKey{Key: testKey},
				gcpexpressmode.APIKey{Key: testKey},
				gcpexpressmode.APIKey{Key: testKey},
			},
		},
		{
			name:  "multiple distinct matches",
			input: testKey + "\n" + testKey[:len(testKey)-1] + "1\n",
			want: []veles.Secret{
				gcpexpressmode.APIKey{Key: testKey},
				gcpexpressmode.APIKey{Key: testKey[:len(testKey)-1] + "1"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := engine.Detect(t.Context(), strings.NewReader(tc.input))
			if err != nil {
				t.Errorf("Detect() error: %v, want nil", err)
			}
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Detect() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDetector_trueNegatives(t *testing.T) {
	engine, err := veles.NewDetectionEngine([]veles.Detector{gcpexpressmode.NewDetector()})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name  string
		input string
		want  []veles.Secret
	}{
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "short key should not match",
			input: testKey[:len(testKey)-1],
		},
		{
			name:  "incorrect casing of prefix should not match",
			input: strings.ToLower(testKey),
		},
		{
			name:  "special character in key should not match",
			input: "AQ.Ab8Rthat-is-1-very-n.ce-KeyYouGotThere_ShameIfLost",
		},
		{
			name:  "wrong prefix should not match",
			input: "AQ-Ab8Rthat-is-1-very-n.ce-KeyYouGotThere_ShameIfLost",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := engine.Detect(t.Context(), strings.NewReader(tc.input))
			if err != nil {
				t.Errorf("Detect() error: %v, want nil", err)
			}
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Detect() diff (-want +got):\n%s", diff)
			}
		})
	}
}
