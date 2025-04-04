package postgres_store

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/catalystcommunity/app-utils-go/env"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/config"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type txError struct {
	cause error
}

// Error implements the error interface.
func (e *txError) Error() string { return e.cause.Error() }

// Cause implements the pkg/errors causer interface.
func (e *txError) Cause() error { return e.cause }

// Unwrap implements the go error causer interface.
func (e *txError) Unwrap() error { return e.cause }

// AmbiguousCommitError represents an error that left a transaction in an
// ambiguous state: unclear if it committed or not.
type AmbiguousCommitError struct {
	txError
}

func newAmbiguousCommitError(err error) *AmbiguousCommitError {
	return &AmbiguousCommitError{txError{cause: err}}
}

var (
	PostgresStore = PostgresDbStore{}
	db            *gorm.DB
	pgxPool       *pgxpool.Pool
)

type PostgresDbStore struct{}

// GetDB returns the underlying gorm.DB connection
func (s PostgresDbStore) GetDB() *gorm.DB {
	return db
}

// Transaction context key type
type txContextKey struct{}

// GetTxContextKey returns the transaction context key for use in middleware
func GetTxContextKey() interface{} {
	return txContextKey{}
}

// GetDBFromContext returns either the transaction from the context or the global DB
func GetDBFromContext(ctx context.Context) *gorm.DB {
	// Check if there's a transaction in the context
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	
	// Otherwise return the global DB
	return db
}

func (s PostgresDbStore) Initialize() (func(), error) {
	var err error
	uri := config.DbUri
	pgxpoolConfig, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}
	logrusLogger := &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(logrus.JSONFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.ErrorLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
	pgxpoolConfig.ConnConfig.Logger = logrusadapter.NewLogger(logrusLogger)
	pgxPool, err = pgxpool.ConnectConfig(context.Background(), pgxpoolConfig)
	if err != nil {
		return nil, err
	}
	logger := getLogger()
	nowFunc := func() time.Time {
		return time.Now().UTC()
	}
	db, err = gorm.Open(postgres.Open(uri), &gorm.Config{Logger: logger, NowFunc: nowFunc})
	if err != nil {
		return nil, err
	}
	return func() {
		pgxPool.Close()
	}, nil
}

func getLogger() logger.Interface {
	slowThresholdSeconds := env.GetEnvAsIntOrDefault("SQL_LOGGER_SLOW_SQL_SECONDS", "1")
	logLevel := env.GetEnvOrDefault("SQL_LOGGER_LEVEL", "error")
	ignoreRecordNotFound := env.GetEnvAsBoolOrDefault("SQL_LOGGER_IGNORE_RECORD_NOT_FOUND", "true")
	colorful := env.GetEnvAsBoolOrDefault("SQL_LOGGER_COLORFUL_LOGS", "true")
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Duration(slowThresholdSeconds) * time.Second, // Slow SQL threshold
			LogLevel:                  getLogLevel(logLevel),                             // Log level
			IgnoreRecordNotFoundError: ignoreRecordNotFound,                              // Ignore ErrRecordNotFound error for logger
			Colorful:                  colorful,                                          // Disable color
		},
	)
}

func getLogLevel(loglevel string) logger.LogLevel {
	switch strings.ToLower(loglevel) {
	case "info":
		return logger.Info
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	case "silent":
		return logger.Silent
	default:
		return logger.Error
	}
}
