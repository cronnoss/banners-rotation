package multiarmedbandit

import (
	"math"
)

type Banner interface {
	GetID() int
	GetImpressions() float64
	GetClicks() float64
}

func PickBanner(banners []Banner) int {
	var (
		totalImpressions float64
		maximumRating    float64 = -1
		selectedBannerID         = 0
	)

	// Find the sum of all impressions for subsequent calculation
	for _, b := range banners {
		imp := b.GetImpressions()
		if imp == 0 {
			imp = 1
		}
		totalImpressions += imp
	}

	// Select the banner with the highest rating
	for _, b := range banners {
		rating := CalculateRating(b.GetClicks(), b.GetImpressions(), totalImpressions)
		if rating > maximumRating {
			maximumRating = rating
			selectedBannerID = b.GetID()
		}
	}

	return selectedBannerID
}

// CalculateRating calculates the banner rating.
func CalculateRating(clicks, impressions, totalImpressions float64) float64 {
	if impressions == 0 {
		impressions = 1
	}
	return clicks/impressions + math.Sqrt(2*math.Log(totalImpressions)/impressions)
}
