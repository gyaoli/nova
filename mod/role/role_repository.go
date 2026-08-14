package role

import (
	"context"
	"errors"
	"strings"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type roleModel struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Platform       string    `gorm:"column:platform;primaryKey"`
	ZoneID         int32     `gorm:"column:zone_id;primaryKey"`
	Account        string    `gorm:"column:account"`
	Name           string    `gorm:"column:name"`
	Level          int8      `gorm:"column:lev"`
	RegIP          *string   `gorm:"column:reg_ip"`
	RegTime        int64     `gorm:"column:reg_time"`
	Data           []byte    `gorm:"column:data"`
	LastUpdateTime time.Time `gorm:"column:last_update_time;autoCreateTime;autoUpdateTime"`
	LastLoginTime  int64     `gorm:"column:last_login_time"`
	LastLogoutTime int64     `gorm:"column:last_logout_time"`
}

func (roleModel) TableName() string { return "role" }

// mysqlRepository persists role snapshots for the role module.
type mysqlRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &mysqlRepository{db: db} }

func (r *mysqlRepository) Create(ctx context.Context, value Role) (Role, error) {
	if len(value.Data) == 0 {
		return Role{}, ErrInvalidArgument
	}
	model := modelFromRole(value)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return Role{}, mapCreateError(err)
	}
	return roleFromModel(model), nil
}

func (r *mysqlRepository) FindByAccount(ctx context.Context, account, platform string, zoneID int32) (Role, error) {
	var model roleModel
	err := r.db.WithContext(ctx).
		Where("account = ? AND platform = ? AND zone_id = ?", account, platform, zoneID).
		Take(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Role{}, ErrNotFound
	}
	if err != nil {
		return Role{}, err
	}
	return roleFromModel(model), nil
}

func (r *mysqlRepository) UpdateLastLogin(ctx context.Context, id int64, platform string, zoneID int32, loginTime int64) error {
	result := r.db.WithContext(ctx).Model(&roleModel{}).
		Where("id = ? AND platform = ? AND zone_id = ?", id, platform, zoneID).
		Update("last_login_time", loginTime)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func modelFromRole(value Role) roleModel {
	var regIP *string
	if value.RegIP != "" {
		regIP = &value.RegIP
	}
	return roleModel{
		ID: value.ID, Platform: value.Platform, ZoneID: value.ZoneID,
		Account: value.Account, Name: value.Name, Level: value.Level,
		RegIP: regIP, RegTime: value.RegTime, Data: append([]byte(nil), value.Data...),
		LastUpdateTime: value.LastUpdateTime, LastLoginTime: value.LastLoginTime,
		LastLogoutTime: value.LastLogoutTime,
	}
}

func roleFromModel(model roleModel) Role {
	value := Role{
		ID: model.ID, Platform: model.Platform, ZoneID: model.ZoneID,
		Account: model.Account, Name: model.Name, Level: model.Level,
		RegTime: model.RegTime, Data: append([]byte(nil), model.Data...),
		LastUpdateTime: model.LastUpdateTime, LastLoginTime: model.LastLoginTime,
		LastLogoutTime: model.LastLogoutTime,
	}
	if model.RegIP != nil {
		value.RegIP = *model.RegIP
	}
	return value
}

func mapCreateError(err error) error {
	var mysqlErr *drivermysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		switch duplicateKey(mysqlErr.Message) {
		case "acc":
			return ErrAlreadyExists
		case "name":
			return ErrNameConflict
		}
	}
	return err
}

func duplicateKey(message string) string {
	message = strings.ToLower(message)
	index := strings.LastIndex(message, "for key ")
	if index < 0 {
		return ""
	}
	key := strings.Trim(strings.TrimSpace(message[index+len("for key "):]), "'`\"")
	if separator := strings.LastIndexByte(key, '.'); separator >= 0 {
		key = key[separator+1:]
	}
	return key
}
