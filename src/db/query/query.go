package query

import (
	"context"
	"time"

	"github.com/SoumyadipPayra/NightsWatch/src/db/conn"
	"github.com/SoumyadipPayra/NightsWatch/src/db/model"
	"gorm.io/gorm"
)

type Query interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUser(ctx context.Context, username string) (*model.User, error)
	UpdateUserLoginTime(ctx context.Context, username string) error
	GetLatestAppData(ctx context.Context, username string) (*model.AppData, error)
	GetLatestOsInfo(ctx context.Context, username string) (*model.OsInfo, error)
	AddDeviceData(ctx context.Context, deviceData *model.DeviceData) error
}

type query struct {
	db *gorm.DB
}

func NewQuery(ctx context.Context) Query {
	return &query{db: conn.GetDB(ctx)}
}

func (q *query) CreateUser(ctx context.Context, user *model.User) error {
	return q.db.Create(user).Error
}

func (q *query) GetUser(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := q.db.Where("user_name = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (q *query) UpdateUserLoginTime(ctx context.Context, username string) error {
	return q.db.Model(&model.User{}).Where("user_name = ?", username).Update("last_login_time", time.Now()).Error
}

func (q *query) AddDeviceData(ctx context.Context, deviceData *model.DeviceData) error {
	tx := q.db.Begin()
	appData := &model.AppData{
		UserID:        deviceData.UserID,
		InstalledApps: deviceData.InstalledApps,
		Timestamp:     deviceData.Timestamp,
	}
	if err := tx.Create(appData).Error; err != nil {
		tx.Rollback()
		return err
	}
	osInfo := &model.OsInfo{
		UserID:         deviceData.UserID,
		OSQueryVersion: deviceData.OSQueryVersion,
		OSVersion:      deviceData.OSVersion,
		Timestamp:      deviceData.Timestamp,
	}
	if err := tx.Create(osInfo).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (q *query) GetLatestAppData(ctx context.Context, username string) (*model.AppData, error) {
	var appData model.AppData
	if err := q.db.Joins("JOIN users ON users.id = app_data.user_id").
		Where("users.user_name = ?", username).
		Order("app_data.timestamp DESC").
		First(&appData).Error; err != nil {
		return nil, err
	}
	return &appData, nil
}

func (q *query) GetLatestOsInfo(ctx context.Context, username string) (*model.OsInfo, error) {
	var osInfo model.OsInfo
	if err := q.db.Joins("JOIN users ON users.id = os_info.user_id").
		Where("users.user_name = ?", username).
		Order("os_info.timestamp DESC").
		First(&osInfo).Error; err != nil {
		return nil, err
	}
	return &osInfo, nil
}
