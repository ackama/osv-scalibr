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

package datasource_test

import (
	"maps"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/osv-scalibr/clients/datasource"
)

func TestRequestCache(t *testing.T) {
	// Test that RequestCache calls each function exactly once per key.
	requestCache := datasource.NewRequestCache[int, int]()

	const numKeys = 20
	const requestsPerKey = 50

	var wg sync.WaitGroup
	var fnCalls [numKeys]int32

	for i := range numKeys {
		for range requestsPerKey {
			wg.Add(1)
			go func() {
				t.Helper()
				_, _ = requestCache.Get(i, func() (int, error) {
					// Count how many times this function gets called for this key,
					// then return the key as the value.
					atomic.AddInt32(&fnCalls[i], 1)
					return i, nil
				})
				wg.Done()
			}()
		}
	}

	wg.Wait() // Make sure all the goroutines are finished

	for i, c := range fnCalls {
		if c != 1 {
			t.Errorf("RequestCache Get(%d) function called %d times", i, c)
		}
	}

	cacheMap := requestCache.GetMap()
	if len(cacheMap) != numKeys {
		t.Errorf("RequestCache GetMap length was %d, expected %d", len(cacheMap), numKeys)
	}

	for k, v := range cacheMap {
		if k != v {
			t.Errorf("RequestCache GetMap key %d has unexpected value %d", k, v)
		}
	}
}

func TestRequestCacheSetMap(t *testing.T) {
	requestCache := datasource.NewRequestCache[string, string]()
	requestCache.SetMap(map[string]string{"foo": "foo1", "bar": "bar2"})
	fn := func() (string, error) { return "CACHE MISS", nil }

	want := map[string]string{
		"foo": "foo1",
		"bar": "bar2",
		"baz": "CACHE MISS",
		"FOO": "CACHE MISS",
	}

	for k, v := range want {
		got, err := requestCache.Get(k, fn)
		if err != nil {
			t.Errorf("Get(%v) returned an error: %v", v, err)
		} else if got != v {
			t.Errorf("Get(%v) got: %v, want %v", k, got, v)
		}
	}

	gotMap := requestCache.GetMap()
	if !maps.Equal(want, gotMap) {
		t.Errorf("GetMap() got %v, want %v", gotMap, want)
	}
}
