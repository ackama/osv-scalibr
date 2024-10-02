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

// Package filebrowser implements a detector for weak/guessable passwords
// on a filebrowser instance.
// To test and install filebrowser, simply follow the instructions in
// https://filebrowser.org/installation
package filebrowser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/osv-scalibr/detector"
	scalibrfs "github.com/google/osv-scalibr/fs"
	"github.com/google/osv-scalibr/inventoryindex"
	"github.com/google/osv-scalibr/log"
	"github.com/google/osv-scalibr/plugin"
)

const (
	fileBrowserIP  = "127.0.0.1"
	requestTimeout = 2 * time.Second
	title          = "Filebrowser is vulnerable to weak credentials using admin/admin"
	description    = "With access to FileBrowser with weak credentials" +
		"an attacker can execute arbitrary code."
	recommendation = "Stop the filebrowser service and uninstall it or change the credentials."
)

var (
	fileBrowserPorts = []int{
		5080,
		8080,
		80,
	}
)

// Detector is a SCALIBR Detector for weak/guessable passwords from /etc/shadow.
type Detector struct{}

// Name of the detector.
func (Detector) Name() string { return "weakcredentials/filebrowser" }

// Version of the detector.
func (Detector) Version() int { return 0 }

// Requirements of the detector.
func (Detector) Requirements() *plugin.Capabilities {
	return &plugin.Capabilities{OS: plugin.OSLinux, RunningSystem: true}
}

// RequiredExtractors returns an empty list as there are no dependencies.
func (Detector) RequiredExtractors() []string { return []string{} }

// Scan starts the scan.
func (d Detector) Scan(ctx context.Context, scanRoot *scalibrfs.ScanRoot, ix *inventoryindex.InventoryIndex) ([]*detector.Finding, error) {
	for _, fileBrowserPort := range fileBrowserPorts {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if !isVulnerable(ctx, fileBrowserIP, fileBrowserPort) {
			continue
		}
		return []*detector.Finding{&detector.Finding{
			Adv: &detector.Advisory{
				ID: &detector.AdvisoryID{
					Publisher: "SCALIBR",
					Reference: "file-browser-weakcredentials",
				},
				Type:           detector.TypeVulnerability,
				Title:          title,
				Description:    description,
				Recommendation: recommendation,
				Sev:            &detector.Severity{Severity: detector.SeverityCritical},
			},
		}}, nil
	}

	return nil, nil

}

func isVulnerable(ctx context.Context, fileBrowserIP string, fileBrowserPort int) bool {
	if !checkAccessibility(ctx, fileBrowserIP, fileBrowserPort) {
		return false
	}
	if !checkLogin(ctx, fileBrowserIP, fileBrowserPort) {
		return false
	}
	return true
}

// checkAccessibility checks if the filebrowser instance is accessible given an IP and port.
func checkAccessibility(ctx context.Context, fileBrowserIP string, fileBrowserPort int) bool {

	client := &http.Client{Timeout: requestTimeout}
	targetURL := fmt.Sprintf("http://%s:%d/", fileBrowserIP, fileBrowserPort)

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		log.Infof(fmt.Sprintf("Error while constructing request %s to the server: %v", targetURL, err))
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Infof(fmt.Sprintf("Timeout exceeded when accessing %s", targetURL))
		} else {
			log.Infof(fmt.Sprintf("Error when sending request %s to the server: %v", targetURL, err))
		}
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	// Expected size for the response is around 6 kilobytes.
	if resp.ContentLength > 20*1024 {
		log.Infof(fmt.Sprintf("Filesize is too large: %d bytes", resp.ContentLength))
		return false
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Infof(fmt.Sprintf("Error reading response body: %v", err))
		return false
	}

	bodyString := string(bodyBytes)
	if !strings.Contains(bodyString, "File Browser") {
		log.Infof("Response body does not contain 'File Browser'")
		return false
	}

	return true
}

// checkLogin checks if the login with default credentials is successful.
func checkLogin(ctx context.Context, fileBrowserIP string, fileBrowserPort int) bool {

	client := &http.Client{Timeout: requestTimeout}
	targetURL := fmt.Sprintf("http://%s:%d/api/login", fileBrowserIP, fileBrowserPort)

	requestBody, _ := json.Marshal(map[string]string{
		"username":  "admin",
		"password":  "admin",
		"recaptcha": "",
	})

	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, io.NopCloser(bytes.NewBuffer(requestBody)))
	if err != nil {
		log.Infof(fmt.Sprintf("Error while constructing request %s to the server: %v", targetURL, err))
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Infof(fmt.Sprintf("Timeout exceeded when accessing %s", targetURL))
		} else {
			log.Infof(fmt.Sprintf("Error when sending request %s to the server: %v", targetURL, err))
		}
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}