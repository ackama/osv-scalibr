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

package pypi

import (
	"encoding/json"
	"testing"
)

func TestYankedUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Yanked
		wantErr bool
	}{
		{
			name: "boolean false",
			json: `false`,
			want: Yanked{Value: false},
		},
		{
			name: "boolean true",
			json: `true`,
			want: Yanked{Value: true},
		},
		{
			name: "string reason",
			json: `"security issue"`,
			want: Yanked{Value: true},
		},
		{
			name:    "invalid type",
			json:    `123`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Yanked
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("UnmarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}
