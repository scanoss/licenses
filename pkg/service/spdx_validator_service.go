package service

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"scanoss.com/licenses/internal/assets"
	"strings"
	"sync"
)

// SPDXLicenseList represents the complete SPDX license list
type SPDXLicenseList struct {
	LicenseListVersion string        `json:"licenseListVersion"`
	Licenses           []SPDXLicense `json:"licenses"`
}

// SPDXLicense represents a single license in the SPDX license list
type SPDXLicense struct {
	Reference               string   `json:"reference"`
	IsDeprecatedLicenseID   bool     `json:"isDeprecatedLicenseId"`
	DetailsURL              string   `json:"detailsUrl"`
	ReferenceNumber         int      `json:"referenceNumber"`
	Name                    string   `json:"name"`
	LicenseID               string   `json:"licenseId"`
	SeeAlso                 []string `json:"seeAlso"`
	IsOSIApproved           bool     `json:"isOsiApproved"`
	IsFSFLibre              *bool    `json:"isFsfLibre,omitempty"`
	IsDeprecated            *bool    `json:"isDeprecated,omitempty"`
	LicenseText             *string  `json:"licenseText,omitempty"`
	StandardLicenseHeader   *string  `json:"standardLicenseHeader,omitempty"`
	StandardLicenseTemplate *string  `json:"standardLicenseTemplate,omitempty"`
}

// SPDXLicenseValidator provides validation using the license list
type SPDXLicenseValidator struct {
	licenseList *SPDXLicenseList
	licenseMap  map[string]SPDXLicense
	mu          sync.RWMutex // Protects concurrent access
}

var (
	instance *SPDXLicenseValidator
	once     sync.Once
)

// GetSPDXValidator returns the singleton instance of SPDXLicenseValidator
func GetSPDXValidator() *SPDXLicenseValidator {
	once.Do(func() {
		instance = &SPDXLicenseValidator{
			licenseMap: make(map[string]SPDXLicense),
		}
		// Initialize with embedded assets
		if err := instance.LoadFromEmbeddedAssets(); err != nil {
			// Handle initialization error - you might want to panic or log
			panic(fmt.Sprintf("Failed to initialize SPDX license validator: %v", err))
		}
	})
	return instance
}

// buildLicenseMap builds a map for quick license lookups
func (v *SPDXLicenseValidator) buildLicenseMap() {
	v.licenseMap = make(map[string]SPDXLicense)
	for _, license := range v.licenseList.Licenses {
		v.licenseMap[license.LicenseID] = license
	}
}

// LoadFromJSON loads the license list from JSON data
func (v *SPDXLicenseValidator) LoadFromJSON(data []byte) error {
	var licenseList SPDXLicenseList
	if err := json.Unmarshal(data, &licenseList); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	v.licenseList = &licenseList
	v.buildLicenseMap()
	return nil
}

// LoadFromEmbeddedAssets loads the license list from embedded assets
func (v *SPDXLicenseValidator) LoadFromEmbeddedAssets() error {
	return v.LoadFromJSON(assets.LicensesJSON)
}

// IsValidLicenseID checks if a license ID is valid (thread-safe)
func (v *SPDXLicenseValidator) IsValidLicenseID(spdxid string) bool {
	if spdxid == "" {
		return false
	}
	v.mu.RLock()
	defer v.mu.RUnlock()

	_, exists := v.licenseMap[spdxid]
	return exists || strings.HasPrefix(spdxid, "LicenseRef-")
}
