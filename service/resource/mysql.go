package resource

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	defaultMaxOpenConns    = 32
	defaultMaxIdleConns    = 16
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
	defaultIOTimeout       = 2 * time.Second
)

// Config contains connection-pool settings shared by every MySQL consumer.
type MySQLConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// Client owns one concurrency-safe GORM/database.sql connection pool.
type MySQLClient struct {
	db  *gorm.DB
	sql *sql.DB
}

func openMySQL(ctx context.Context, config MySQLConfig) (*MySQLClient, error) {
	config = withMySQLDefaults(config)
	driverConfig := drivermysql.NewConfig()
	driverConfig.User = config.User
	driverConfig.Passwd = config.Password
	driverConfig.Net = "tcp"
	driverConfig.Addr = netJoinHostPort(config.Host, config.Port)
	driverConfig.DBName = config.Database
	driverConfig.ParseTime = true
	driverConfig.Loc = time.Local
	driverConfig.Collation = "utf8_bin"
	driverConfig.Timeout = defaultIOTimeout
	driverConfig.ReadTimeout = defaultIOTimeout
	driverConfig.WriteTimeout = defaultIOTimeout

	db, err := gorm.Open(mysql.Open(driverConfig.FormatDSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get mysql connection pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return &MySQLClient{db: db, sql: sqlDB}, nil
}

func (c *MySQLClient) DB() *gorm.DB { return c.db }

func (c *MySQLClient) Ping(ctx context.Context) error { return c.sql.PingContext(ctx) }

func (c *MySQLClient) Stats() sql.DBStats { return c.sql.Stats() }

func (c *MySQLClient) Close() error { return c.sql.Close() }

func withMySQLDefaults(config MySQLConfig) MySQLConfig {
	if config.MaxOpenConns <= 0 {
		config.MaxOpenConns = defaultMaxOpenConns
	}
	if config.MaxIdleConns <= 0 || config.MaxIdleConns > config.MaxOpenConns {
		config.MaxIdleConns = defaultMaxIdleConns
	}
	if config.ConnMaxLifetime <= 0 {
		config.ConnMaxLifetime = defaultConnMaxLifetime
	}
	if config.ConnMaxIdleTime <= 0 {
		config.ConnMaxIdleTime = defaultConnMaxIdleTime
	}
	return config
}

func netJoinHostPort(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}
