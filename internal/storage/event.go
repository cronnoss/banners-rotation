package storage

import "time"

type Click struct {
	ID          int       `json:"id"`
	SlotID      int       `json:"slot_id"`      //nolint:tagliatelle
	BannerID    int       `json:"banner_id"`    //nolint:tagliatelle
	UserGroupID int       `json:"usergroup_id"` //nolint:tagliatelle
	CreatedAt   time.Time `json:"created_at"`   //nolint:tagliatelle
}

type Impress struct {
	ID          int       `json:"id"`
	SlotID      int       `json:"slot_id"`      //nolint:tagliatelle
	BannerID    int       `json:"banner_id"`    //nolint:tagliatelle
	UserGroupID int       `json:"usergroup_id"` //nolint:tagliatelle
	CreatedAt   time.Time `json:"created_at"`   //nolint:tagliatelle
}

type Notification struct {
	TypeEvent   string    `json:"type_event"`   //nolint:tagliatelle
	SlotID      int       `json:"slot_id"`      //nolint:tagliatelle
	BannerID    int       `json:"banner_id"`    //nolint:tagliatelle
	UsergroupID int       `json:"usergroup_id"` //nolint:tagliatelle
	DateTime    time.Time `json:"date_time"`    //nolint:tagliatelle
}
