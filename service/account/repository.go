package account

import (
	"context"
	"errors"
	"strings"

	drivermysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type accountModel struct {
	AcctName   string  `gorm:"column:acct_name;primaryKey"`
	Password   string  `gorm:"column:password"`
	Account    string  `gorm:"column:account"`
	Platform   string  `gorm:"column:platform"`
	ClientType int     `gorm:"column:client_type"`
	DeviceID   string  `gorm:"column:device_id"`
	DeviceName string  `gorm:"column:device_name"`
	RegIP      *string `gorm:"column:reg_ip"`
	RegTime    int64   `gorm:"column:reg_time"`
}

func (accountModel) TableName() string { return "account" }

type mysqlRepository struct{ db *gorm.DB }

func newRepository(db *gorm.DB) *mysqlRepository { return &mysqlRepository{db: db} }

func (r *mysqlRepository) Create(ctx context.Context, value Account) error {
	var regIP *string
	if value.RegIP != "" {
		regIP = &value.RegIP
	}
	err := r.db.WithContext(ctx).Create(&accountModel{
		AcctName: value.AcctName, Password: value.Password, Account: value.Account,
		Platform: value.Platform, ClientType: value.ClientType,
		DeviceID: value.DeviceID, DeviceName: value.DeviceName,
		RegIP: regIP, RegTime: value.RegTime,
	}).Error
	if err == nil {
		return nil
	}
	if mysqlErr, ok := errors.AsType[*drivermysql.MySQLError](err); ok && mysqlErr.Number == 1062 {
		if duplicateAccountID(mysqlErr.Message) {
			return ErrAccountIDConflict
		}
		return ErrAccountExists
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrAccountExists
	}
	return err
}

func duplicateAccountID(message string) bool {
	message = strings.ToLower(message)
	index := strings.LastIndex(message, "for key ")
	if index < 0 {
		return false
	}
	key := strings.Trim(strings.TrimSpace(message[index+len("for key "):]), "'`\"")
	if separator := strings.LastIndexByte(key, '.'); separator >= 0 {
		key = key[separator+1:]
	}
	return key == "acc"
}

func (r *mysqlRepository) FindByName(ctx context.Context, accountName string) (Account, error) {
	var model accountModel
	err := r.db.WithContext(ctx).Where("acct_name = ?", accountName).Take(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Account{}, ErrAccountNotFound
	}
	if err != nil {
		return Account{}, err
	}
	value := Account{
		AcctName: model.AcctName, Password: model.Password, Account: model.Account,
		Platform: model.Platform, ClientType: model.ClientType,
		DeviceID: model.DeviceID, DeviceName: model.DeviceName, RegTime: model.RegTime,
	}
	if model.RegIP != nil {
		value.RegIP = *model.RegIP
	}
	return value, nil
}
