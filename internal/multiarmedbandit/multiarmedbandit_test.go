package multiarmedbandit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type mockBanner struct {
	ID          int
	impressions int
	clicks      int
}

func (mb *mockBanner) GetID() int {
	return mb.ID
}

func (mb *mockBanner) GetImpressions() float64 {
	return float64(mb.impressions)
}

func (mb *mockBanner) GetClicks() float64 {
	return float64(mb.clicks)
}

func TestCalculateRating(t *testing.T) {
	tests := []struct {
		name             string
		clicks           float64
		impressions      float64
		totalImpressions float64
		expectedRating   float64
	}{
		{
			name:             "clicks and impressions are zero",
			clicks:           0,
			impressions:      0,
			totalImpressions: 100,
			expectedRating:   3.034854258770293, // 0 / 1 + sqrt(2 * ln(100) / 1)
		},
		{
			name:             "clicks are zero",
			clicks:           0,
			impressions:      5,
			totalImpressions: 15,
			expectedRating:   1.040778593381361, // 0 / 5 + sqrt(2 * ln(15) / 5)
		},
		{
			name:             "impressions are zero",
			clicks:           100,
			impressions:      0,
			totalImpressions: 100,
			expectedRating:   103.0348542587703, // 100 / 1 + sqrt(2 * ln(100) / 1)
		},
		{
			name:             "all parameters are equal",
			clicks:           100,
			impressions:      100,
			totalImpressions: 100,
			expectedRating:   1.3034854258770292, // 100 / 100 + sqrt(2 * ln(100) / 100)
		},
		{
			name:             "all parameters are different",
			clicks:           100,
			impressions:      200,
			totalImpressions: 100,
			expectedRating:   0.7145966026289348, // 100 / 200 + sqrt(2 * ln(100) / 200)
		},
	}

	t.Parallel()
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			calculatedRating := CalculateRating(test.clicks, test.impressions, test.totalImpressions)
			require.Equal(t, test.expectedRating, calculatedRating)
		})
	}
}

func TestPickBanner(t *testing.T) {
	tests := []struct {
		name    string
		banners []Banner
		want    int
	}{
		{
			name: "all banners have zero clicks and impressions, pick first banner",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 0, clicks: 0},
				&mockBanner{ID: 2, impressions: 0, clicks: 0},
				&mockBanner{ID: 3, impressions: 0, clicks: 0},
			},
			want: 1,
		},
		{
			name: "one banner has impressions, but not clicks, pick second banner",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 2, clicks: 0},
				&mockBanner{ID: 2, impressions: 0, clicks: 0},
				&mockBanner{ID: 3, impressions: 0, clicks: 0},
			},
			want: 2,
		},
		{
			name: "all banners have equal amount of impressions, one banner has clicks, pick it",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 5, clicks: 0},
				&mockBanner{ID: 2, impressions: 5, clicks: 3},
				&mockBanner{ID: 3, impressions: 5, clicks: 0},
			},
			want: 2,
		},
		{
			name: "two banners have impressions, but not clicks, pick third banner",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 2, clicks: 0},
				&mockBanner{ID: 2, impressions: 2, clicks: 0},
				&mockBanner{ID: 3, impressions: 0, clicks: 0},
			},
			want: 3,
		},
		{
			name: "all banners have different amount of impressions, but not clicks, pick third banner",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 3, clicks: 0},
				&mockBanner{ID: 2, impressions: 4, clicks: 0},
				&mockBanner{ID: 3, impressions: 2, clicks: 0},
			},
			want: 3,
		},
		{
			name: "one banner has clicks, pick first banner",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 5, clicks: 1},
				&mockBanner{ID: 2, impressions: 4, clicks: 0},
				&mockBanner{ID: 3, impressions: 4, clicks: 0},
			},
			want: 1,
		},
		{
			name: "one banner has clicks, but too many impressions, pick second banner",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 6, clicks: 1},
				&mockBanner{ID: 2, impressions: 4, clicks: 0},
				&mockBanner{ID: 3, impressions: 4, clicks: 0},
			},
			want: 2,
		},
		{
			name: "all banners have clicks, pick second banner",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 7, clicks: 2},
				&mockBanner{ID: 2, impressions: 4, clicks: 1},
				&mockBanner{ID: 3, impressions: 4, clicks: 0},
			},
			want: 2,
		},
		{
			name: "all banners have clicks, pick third banner",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 7, clicks: 2},
				&mockBanner{ID: 2, impressions: 5, clicks: 1},
				&mockBanner{ID: 3, impressions: 4, clicks: 1},
			},
			want: 3,
		},
		{
			name: "one banner has many impressions and clicks",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 16000, clicks: 799},
				&mockBanner{ID: 2, impressions: 9000, clicks: 59},
				&mockBanner{ID: 3, impressions: 3000, clicks: 9},
			},
			want: 1,
		},
		{
			name: "multiple banners have the same number of impressions and clicks, pick the first one",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 50, clicks: 10},
				&mockBanner{ID: 2, impressions: 50, clicks: 10},
				&mockBanner{ID: 3, impressions: 50, clicks: 10},
			},
			want: 1,
		},
		{
			name: "one banner has more impressions than others but fewer clicks, prioritize clicks",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 1000, clicks: 50},
				&mockBanner{ID: 2, impressions: 1500, clicks: 60},
				&mockBanner{ID: 3, impressions: 800, clicks: 100},
			},
			want: 3,
		},
		{
			name: "one banner has more clicks than others but fewer impressions, prioritize impressions",
			banners: []Banner{
				&mockBanner{ID: 1, impressions: 300, clicks: 100},
				&mockBanner{ID: 2, impressions: 500, clicks: 200},
				&mockBanner{ID: 3, impressions: 200, clicks: 50},
			},
			want: 2,
		},
	}

	t.Parallel()
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bannerID := PickBanner(test.banners)
			require.Equal(t, test.want, bannerID)
		})
	}
}
