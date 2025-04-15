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
	GetLatestDeviceData(ctx context.Context, username string) (*model.DeviceData, error)
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

func (q *query) GetLatestDeviceData(ctx context.Context, username string) (*model.DeviceData, error) {
	var deviceData model.DeviceData
	if err := q.db.Joins("JOIN users ON users.id = device_data.user_id").
		Where("users.user_name = ?", username).
		Order("device_data.timestamp DESC").
		First(&deviceData).Error; err != nil {
		return nil, err
	}
	return &deviceData, nil
}

func (q *query) AddDeviceData(ctx context.Context, deviceData *model.DeviceData) error {
	return q.db.Create(deviceData).Error
}
