package interfaces

import (
	"context"

	"github.com/cronnoss/banners-rotation/internal/storage"
)

type Storage interface {
	Connect(ctx context.Context, dbPort int, dbHost, dbUser, dbPassword, dbName string) error
	Close(ctx context.Context) error
	Migrate(ctx context.Context, migrate string) error
	AddBanner(ctx context.Context, bannerID, slotID int) error
	RemoveBanner(ctx context.Context, bannerID, slotID int) error
	ClickBanner(ctx context.Context, bannerID, slotID, userGroupID int) (*storage.Click, error)
	PickBanner(ctx context.Context, slotID, usergroupID int) (*storage.Impress, int, error)
	IsBannerAssignedToSlot(ctx context.Context, bannerID, slotID int) (bool, error)
	BannerExists(ctx context.Context, bannerID int) bool
	SlotExists(ctx context.Context, slotID int) bool
	UserGroupExists(ctx context.Context, userGroupID int) bool
}
