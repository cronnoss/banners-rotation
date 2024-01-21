//go:build integration
// +build integration

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/cronnoss/banners-rotation/internal/rmq"
	"github.com/cronnoss/banners-rotation/internal/server/pb"
	"github.com/cronnoss/banners-rotation/internal/storage"
	_ "github.com/jackc/pgx/stdlib" // Blank import for side effects
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type BannerSuite struct {
	suite.Suite
	ctx        context.Context
	bannerConn *grpc.ClientConn
	client     pb.BannerServiceClient
	db         *sqlx.DB
	msgs       <-chan amqp.Delivery
}

func (s *BannerSuite) SetupSuite() {
	s.ctx = context.TODO()

	host := os.Getenv("GRPC_HOST")
	port := os.Getenv("GRPC_PORT")
	bannerHost := host + ":" + port
	if bannerHost == "" {
		bannerHost = "127.0.0.1:8082"
	}

	var err error
	s.bannerConn, err = grpc.Dial(bannerHost, grpc.WithInsecure())
	s.Require().NoError(err)
	s.client = pb.NewBannerServiceClient(s.bannerConn)

	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		// "postgres", "postgres", "localhost", 5432, "postgres")
		"postgres", "postgres", os.Getenv("POSTGRES_HOST"), 5432, "postgres")
	s.db, err = sqlx.Open("pgx", connectionString)
	s.Require().NoError(err)

	RabbitmqHost := os.Getenv("RABBITMQ_HOST")
	URI := fmt.Sprintf("%s://%s:%s@%s:%d/", // "amqp://guest:guest@localhost:5672/"
		"amqp", "guest", "guest", RabbitmqHost, 5672,
	)

	eventsConsMq, err := rmq.New(
		// "amqp://guest:guest@localhost:5672/",
		URI,
		"events",
		"fanout",
		"notifications",
		"",
		"1m",
		"1s",
		2,
		"15s",
	)
	s.Require().NoError(err)

	err = eventsConsMq.Init(s.ctx)
	s.Require().NoError(err)

	s.msgs, err = eventsConsMq.Consume("banner_notifications")
	s.Require().NoError(err)
}

func (s *BannerSuite) TearDownTest() {
	query := `DELETE FROM clicks`
	_, err := s.db.Exec(query)
	s.Require().NoError(err)

	query = `DELETE FROM impressions`
	_, err = s.db.Exec(query)
	s.Require().NoError(err)
}

func (s *BannerSuite) TearDownSuite() {
	defer s.db.Close()
}

func TestBannerPost(t *testing.T) {
	suite.Run(t, new(BannerSuite))
}

func (s *BannerSuite) TestBanner_AddBanner() {
	tests := []struct {
		name          string
		slotID        int32
		bannerID      int32
		expectedError string
	}{
		{"Success", 1, 8, ""},
		{"Already assigned to the slot", 3, 6, "banner is already assigned to the slot"},
		{"Slot does not exist", 1000, 6, "specified slot does not exist"},
		{"Banner does not exist", 1, 60, "specified banner does not exist"},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			req := &pb.AddBannerRequest{
				SlotId:   test.slotID,
				BannerId: test.bannerID,
			}

			resp, err := s.client.AddBanner(s.ctx, req)

			switch test.name {
			case "Success":
				s.Require().NoError(err)
				s.NotNil(resp)
				s.Equal("Banner added successfully", resp.Message)
				// Checking the presence of a record in the database
				s.checkingRecordInRotationsTable(test.slotID, test.bannerID)
				s.removeRecord(test.slotID, test.bannerID)
			case "Already assigned to the slot":
				s.Require().Error(err)
				s.Nil(resp)
				s.Equal(test.expectedError, status.Convert(err).Message())
				// Checking the presence of a record in the database
				s.checkingRecordInRotationsTable(test.slotID, test.bannerID)
			default:
				s.Require().Error(err)
				s.Nil(resp)
				s.Equal(test.expectedError, status.Convert(err).Message())
				// Checking the absence of a record in the database
				s.checkingNoRecordInRotationsTable(test.slotID, test.bannerID)
			}
		})
	}
}

func (s *BannerSuite) TestBanner_RemoveBanner() {
	tests := []struct {
		name          string
		slotID        int32
		bannerID      int32
		expectedError bool
	}{
		{"An existing banner", 1, 1, false},
		{"Slot does not exist", 100, 1, false},
		{"Banner does not exist", 1, 100, false},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			req := &pb.RemoveBannerRequest{
				SlotId:   test.slotID,
				BannerId: test.bannerID,
			}

			resp, err := s.client.RemoveBanner(s.ctx, req)

			if test.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.NotNil(resp)
				s.Equal("Banner removed successfully", resp.Message)
				// Checking the absence of a record in the database
				s.checkingNoRecordInRotationsTable(test.slotID, test.bannerID)
				s.addRecord()
			}
		})
	}
}

func (s *BannerSuite) TestBanner_ClickBanner_NotExist() {
	tests := []struct {
		name          string
		slotID        int32
		bannerID      int32
		UsergroupID   int32
		expectedError string
	}{
		{"Slot does not exist", 1000, 1, 1, "specified slot does not exist"},
		{"Banner does not exist", 1, 60, 1, "specified banner does not exist"},
		{"UserGroupID does not exist", 1, 1, 70, "specified userGroup does not exist"},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			req := &pb.ClickBannerRequest{
				SlotId:      test.slotID,
				BannerId:    test.bannerID,
				UsergroupId: test.UsergroupID,
			}
			beforeCount := s.getCountRecordInClicksTable(test.slotID, test.bannerID, test.UsergroupID)
			resp, err := s.client.ClickBanner(s.ctx, req)

			afterCount := s.getCountRecordInClicksTable(test.slotID, test.bannerID, test.UsergroupID)
			s.Equal(beforeCount, afterCount)
			s.Require().Error(err)
			s.Nil(resp)
			s.Equal(test.expectedError, status.Convert(err).Message())
		})
	}
}

func (s *BannerSuite) TestBanner_ClickBanner_Success() {
	req := &pb.ClickBannerRequest{
		SlotId:      1,
		BannerId:    1,
		UsergroupId: 1,
	}
	beforeCount := s.getCountRecordInClicksTable(req.SlotId, req.BannerId, req.UsergroupId)
	resp, err := s.client.ClickBanner(s.ctx, req)
	afterCount := s.getCountRecordInClicksTable(req.SlotId, req.BannerId, req.UsergroupId)
	s.Equal(beforeCount+1, afterCount)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Equal("Banner clicked successfully", resp.Message)

	// Getting a record from the database
	click, err := s.getRecordInClicksTable(req.SlotId, req.BannerId, req.UsergroupId)
	s.Require().NoError(err)

	// Getting a json string for comparison
	notification := createClickNotification(click)
	notificationJSON, err := serializeNotification(notification)
	s.Require().NoError(err)
	jsonString := string(notificationJSON)

	var body string
	if msg, ok := <-s.msgs; ok {
		body = string(msg.Body)
		msg.Ack(true)
	}

	// Compare the string and the received notification from RMQ
	s.Equal(jsonString, body)
}

func (s *BannerSuite) TestBanner_PickBanner() {

	req := &pb.PickBannerRequest{
		SlotId:      1,
		UsergroupId: 1,
	}

	resp, err := s.client.PickBanner(s.ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Equal("Banner picked successfully", resp.Message)

	impression, err := s.getRecordInImpressionsTable(req.SlotId, resp.BannerId, req.UsergroupId)
	s.Require().NoError(err)

	// Getting a json string for comparison
	notification := createImpressNotification(impression)
	notificationJSON, err := serializeNotification(notification)
	s.Require().NoError(err)
	jsonString := string(notificationJSON)

	var body string
	if msg, ok := <-s.msgs; ok {
		body = string(msg.Body)
		msg.Ack(true)
	}

	// Compare the string and the received notification from RMQ
	s.Equal(jsonString, body)
}

func (s *BannerSuite) checkingRecordInRotationsTable(slotID, bannerID int32) {
	query := `SELECT COUNT(*) FROM rotations WHERE slot_id = $1 AND banner_id = $2;`
	var count int
	err := s.db.QueryRow(query, slotID, bannerID).Scan(&count)
	s.Require().NoError(err)
	s.Equal(1, count)
}

func (s *BannerSuite) checkingNoRecordInRotationsTable(slotID, bannerID int32) {
	query := `SELECT COUNT(*) FROM rotations WHERE slot_id = $1 AND banner_id = $2;`

	var count int
	err := s.db.QueryRow(query, slotID, bannerID).Scan(&count)
	s.Require().NoError(err)
	s.Equal(0, count)
}

func (s *BannerSuite) getCountRecordInClicksTable(slotID, bannerID, userGroupID int32) int {
	query := `SELECT COUNT(*) FROM clicks WHERE slot_id = $1 AND banner_id = $2 AND usergroup_id = $3;`
	var count int
	err := s.db.QueryRow(query, slotID, bannerID, userGroupID).Scan(&count)
	s.Require().NoError(err)
	return count
}

func (s *BannerSuite) getRecordInClicksTable(slotID, bannerID, userGroupID int32) (*storage.Click, error) {
	query := `SELECT * FROM clicks WHERE slot_id = $1 AND banner_id = $2 AND usergroup_id = $3;`
	row := s.db.QueryRow(query, slotID, bannerID, userGroupID)

	click := &storage.Click{}
	err := row.Scan(&click.ID, &click.SlotID, &click.BannerID, &click.UserGroupID, &click.CreatedAt)
	if err != nil {
		return nil, err
	}

	return click, nil
}

func (s *BannerSuite) getRecordInImpressionsTable(slotID, bannerID, userGroupID int32) (*storage.Impress, error) {
	query := `
	SELECT * FROM impressions WHERE slot_id = $1 AND banner_id = $2 AND usergroup_id = $3 
	ORDER BY created_at desc 
	LIMIT 1;
	`
	row := s.db.QueryRow(query, slotID, bannerID, userGroupID)

	impress := &storage.Impress{}
	err := row.Scan(&impress.ID, &impress.SlotID, &impress.BannerID, &impress.UserGroupID, &impress.CreatedAt)
	if err != nil {
		return nil, err
	}

	return impress, nil
}

func (s *BannerSuite) removeRecord(slotID, bannerID int32) {
	query := `DELETE FROM rotations WHERE slot_id = $1 and banner_id = $2;`
	_, err := s.db.Exec(query, slotID, bannerID)
	s.Require().NoError(err)
}

func (s *BannerSuite) addRecord() {
	query := `
	INSERT INTO rotations (slot_id, banner_id, created_at)
	SELECT 1, 1, NOW()
	WHERE NOT EXISTS (
		SELECT 1 FROM rotations
		WHERE slot_id = 1 AND banner_id = 1
	)
	`
	_, err := s.db.Exec(query)
	s.Require().NoError(err)
}

func createClickNotification(click *storage.Click) storage.Notification {
	notification := storage.Notification{
		TypeEvent:   "click",
		SlotID:      click.SlotID,
		BannerID:    click.BannerID,
		UsergroupID: click.UserGroupID,
		DateTime:    click.CreatedAt,
	}
	return notification
}

func createImpressNotification(impress *storage.Impress) storage.Notification {
	notification := storage.Notification{
		TypeEvent:   "impress",
		SlotID:      impress.SlotID,
		BannerID:    impress.BannerID,
		UsergroupID: impress.UserGroupID,
		DateTime:    impress.CreatedAt,
	}
	return notification
}

func serializeNotification(notification storage.Notification) ([]byte, error) {
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return nil, err
	}
	return notificationJSON, nil
}
