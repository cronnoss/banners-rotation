package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/cronnoss/banners-rotation/internal/multiarmedbandit"
	"github.com/cronnoss/banners-rotation/internal/storage"
	_ "github.com/jackc/pgx/stdlib" // The underscore _ before the import path is used when the package is imported for
	// its side effects, in this case, the initialization of the pgx driver.
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

var errNoBannersForGivenSlot = errors.New("no banners for a given slot")

type Storage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Migrate(ctx context.Context, migrate string) (err error) {
	_ = ctx
	if err := goose.SetDialect("pgx"); err != nil {
		return fmt.Errorf("cannot set dialect: %w", err)
	}

	if err := goose.Up(s.db.DB, migrate); err != nil {
		return fmt.Errorf("cannot migrate: %w", err)
	}

	return nil
}

func (s *Storage) Connect(ctx context.Context, dbPort int, dbHost, dbUser, dbPassword, dbName string) (err error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)
	s.db, err = sqlx.Open("pgx", connStr)

	if err != nil {
		return fmt.Errorf("cannot open pgx driver: %w", err)
	}

	return s.db.PingContext(ctx)
}

func (s *Storage) Close(ctx context.Context) error {
	_ = ctx
	return s.db.Close()
}

func (s *Storage) AddBanner(ctx context.Context, bannerID, slotID int) error {
	const query = `
		INSERT INTO rotations (slot_id, banner_id, created_at)
		VALUES ($1, $2, NOW());
	`
	_, err := s.db.ExecContext(ctx, query, slotID, bannerID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) RemoveBanner(ctx context.Context, bannerID, slotID int) error {
	const query = `DELETE FROM rotations WHERE slot_id = $1 and banner_id = $2;`

	_, err := s.db.ExecContext(ctx, query, slotID, bannerID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) ClickBanner(ctx context.Context, bannerID, slotID, userGroupID int) (*storage.Click, error) {
	const query = `
		INSERT INTO clicks (slot_id, banner_id, usergroup_id, created_at) 
		VALUES ($1, $2, $3, NOW())
		RETURNING id, slot_id, banner_id, usergroup_id, created_at;`

	click := &storage.Click{}
	err := s.db.QueryRowContext(ctx, query, slotID, bannerID, userGroupID).
		Scan(&click.ID, &click.SlotID, &click.BannerID, &click.UserGroupID, &click.CreatedAt)
	if err != nil {
		return nil, err
	}

	return click, nil
}

func (s *Storage) PickBanner(ctx context.Context, slotID, usergroupID int) (*storage.Impress, int, error) {
	const query = `
		SELECT
			r.banner_id,
			(SELECT COUNT(*) FROM impressions i WHERE i.banner_id = r.banner_id AND i.usergroup_id = $1) AS impressions,
			(SELECT COUNT(*) FROM clicks c WHERE c.banner_id = r.banner_id AND c.usergroup_id = $1) AS clicks
		FROM rotations r
		WHERE r.slot_id = $2;`

	rows, err := s.db.QueryContext(ctx, query, usergroupID, slotID)
	if err != nil {
		return nil, 0, err
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	banners := make([]multiarmedbandit.Banner, 0)
	for rows.Next() {
		var bnr storage.BannerStatistics
		if err := rows.Scan(&bnr.BannerID, &bnr.Impressions, &bnr.Clicks); err != nil {
			return nil, 0, err
		}
		banners = append(banners, &bnr)
	}

	if len(banners) == 0 {
		return nil, 0, errNoBannersForGivenSlot
	}

	bannerID := multiarmedbandit.PickBanner(banners)

	impress, err := s.ImpressBanner(ctx, bannerID, slotID, usergroupID)
	if err != nil {
		return nil, 0, err
	}

	return impress, bannerID, nil
}

func (s *Storage) ImpressBanner(ctx context.Context, bannerID, slotID, userGroupID int) (*storage.Impress, error) {
	const query = `
		INSERT INTO impressions
		(slot_id, banner_id, usergroup_id, created_at) VALUES
		($1, $2, $3, NOW())
		RETURNING id, slot_id, banner_id, usergroup_id, created_at;`

	impress := &storage.Impress{}
	err := s.db.QueryRowContext(ctx, query, slotID, bannerID, userGroupID).
		Scan(&impress.ID, &impress.SlotID, &impress.BannerID, &impress.UserGroupID, &impress.CreatedAt)
	if err != nil {
		return nil, err
	}

	return impress, err
}

func (s *Storage) IsBannerAssignedToSlot(ctx context.Context, bannerID, slotID int) (bool, error) {
	const query = `
        SELECT COUNT(*)
        FROM rotations
        WHERE banner_id = $1 AND slot_id = $2;`

	var count int
	err := s.db.QueryRowContext(ctx, query, bannerID, slotID).Scan(&count)
	if err != nil {
		return false, err
	}

	// Если count > 0, значит баннер уже присвоен слоту
	return count > 0, nil
}

func (s *Storage) BannerExists(ctx context.Context, bannerID int) bool {
	const query = `
      SELECT COUNT(*)
      FROM banners
      WHERE id = $1;`

	var count int
	err := s.db.QueryRowContext(ctx, query, bannerID).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

func (s *Storage) SlotExists(ctx context.Context, slotID int) bool {
	const query = `
      SELECT COUNT(*)
      FROM slots
      WHERE id = $1;`

	var count int
	err := s.db.QueryRowContext(ctx, query, slotID).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

func (s *Storage) UserGroupExists(ctx context.Context, userGroupID int) bool {
	const query = `
      SELECT COUNT(*)
      FROM usergroups
      WHERE id = $1;`

	var count int
	err := s.db.QueryRowContext(ctx, query, userGroupID).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}
