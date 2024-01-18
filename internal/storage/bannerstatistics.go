package storage

type BannerStatistics struct {
	BannerID    int `db:"banner_id"`
	Impressions int `db:"impressions"`
	Clicks      int `db:"clicks"`
}

func (b *BannerStatistics) GetID() int {
	return b.BannerID
}

func (b *BannerStatistics) GetImpressions() float64 {
	return float64(b.Impressions)
}

func (b *BannerStatistics) GetClicks() float64 {
	return float64(b.Clicks)
}
