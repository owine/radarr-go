package services_test

import (
	"fmt"

	"github.com/radarr/radarr-go/internal/models"
)

// ExampleQualityProfile_GetAllowedQualities demonstrates finding allowed qualities in a profile
func ExampleQualityProfile_GetAllowedQualities() {
	profile := &models.QualityProfile{
		Name:   "HD-1080p",
		Cutoff: 7, // Bluray-1080p
		Items: []*models.QualityProfileItem{
			{
				Quality: &models.QualityLevel{
					ID:    3,
					Title: "WEBDL-1080p",
				},
				Allowed: true,
			},
			{
				Quality: &models.QualityLevel{
					ID:    7,
					Title: "Bluray-1080p",
				},
				Allowed: true,
			},
			{
				Quality: &models.QualityLevel{
					ID:    19,
					Title: "Bluray-2160p",
				},
				Allowed: false, // Not allowed in this profile
			},
		},
	}

	allowedQualities := profile.GetAllowedQualities()
	for _, quality := range allowedQualities {
		fmt.Printf("Allowed: %s\n", quality.Title)
	}
	// Output:
	// Allowed: WEBDL-1080p
	// Allowed: Bluray-1080p
}

// ExampleQualityProfile_IsUpgradeAllowed demonstrates upgrade checking
func ExampleQualityProfile_IsUpgradeAllowed() {
	profile := &models.QualityProfile{
		Name:           "Standard",
		UpgradeAllowed: true,
	}

	fmt.Printf("Upgrades allowed: %t", profile.IsUpgradeAllowed())
	// Output: Upgrades allowed: true
}

// ExampleDefaultQualityDefinitions shows how to get standard quality definitions
func ExampleDefaultQualityDefinitions() {
	qualities := models.DefaultQualityDefinitions()

	// Show first few quality definitions
	for i, quality := range qualities[:3] {
		fmt.Printf("%d. %s (Weight: %d)\n", i+1, quality.Title, quality.Weight)
	}
	// Output:
	// 1. Unknown (Weight: 1)
	// 2. WORKPRINT (Weight: 2)
	// 3. CAM (Weight: 3)
}
