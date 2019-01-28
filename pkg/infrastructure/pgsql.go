package infrastructure

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jinzhu/gorm/dialects/postgres" // nolint
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/loggers"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
)

// PgsqlHandler holds the connection instance to the DB
type PgsqlHandler struct {
	Conn   *sql.DB
	Logger loggers.Logger
}

// Healthcheck implements a connection validator to check if the host is reachable
func (handler *PgsqlHandler) Healthcheck() bool {
	return handler.Conn.Ping() == nil
}

// Close closes an open connection
func (handler *PgsqlHandler) Close() error {
	if err := handler.Conn.Close(); err != nil {
		handler.Logger.Error("Error closing connection: %+v", err)
		return err
	}
	return nil
}

// Insert implements the incorporation of new data into an specific table of DB
func (handler *PgsqlHandler) Insert(statement string, params ...interface{}) error {
	_, err := handler.Conn.Exec(statement, params...)
	return err
}

// Update implements the actualization of a register from an specific table of DB
func (handler *PgsqlHandler) Update(statement string, params ...interface{}) error {
	_, err := handler.Conn.Exec(statement, params...)
	return err
}

// Query send statement to db returning rows
func (handler *PgsqlHandler) Query(statement string, params ...interface{}) (repository.DbResult, error) {
	rows, err := handler.Conn.Query(statement, params...)
	if err != nil {
		handler.Logger.Error("Query error: %+v", err)
		return new(PgsqlRow), err
	}
	return PgsqlRow{
		Rows:   rows,
		Logger: handler.Logger,
	}, nil
}

// PgsqlRow stores a group of DB registers
type PgsqlRow struct {
	Rows   *sql.Rows
	Logger loggers.Logger
}

// Scan copy the information from the DB registers to the destinations specified
func (r PgsqlRow) Scan(dest ...interface{}) error {
	err := r.Rows.Scan(dest...)
	if err != nil {
		r.Logger.Error("Error Scan DB: %+v", err)
		return err
	}
	return nil
}

// Next prepare the next result row for reading
func (r PgsqlRow) Next() bool {
	return r.Rows.Next()
}

// Close ends read process from db
func (r PgsqlRow) Close() error {
	if err := r.Rows.Close(); err != nil {
		r.Logger.Error("Error Close DB: %+v", err)
		return err
	}
	return nil
}

// NewPgsqlHandler Creates connection handler
func NewPgsqlHandler(conf DatabaseConf, logger loggers.Logger) (*PgsqlHandler, error) {
	poolDb, err := sql.Open("postgres",
		fmt.Sprintf("host=%s dbname=%s port=%d sslmode=%s user=%s password=%s",
			conf.Host, conf.Dbname, conf.Port, conf.Sslmode, conf.DbUser, conf.DbPasswd),
	)

	if err != nil || poolDb == nil {
		logger.Error("Error on pool DB definition %+v\n", err)
		return nil, err
	}

	for i := 0; i < conf.ConnRetries; i++ {
		if err := poolDb.Ping(); err != nil {
			logger.Info("Connection attempt number %d failed\n", i+1)
			time.Sleep(1 * time.Second)
		} else {
			poolDb.SetMaxIdleConns(conf.MaxIdle)
			poolDb.SetMaxOpenConns(conf.MaxOpen)

			return &PgsqlHandler{
				Conn:   poolDb,
				Logger: logger,
			}, nil
		}
	}

	return nil, fmt.Errorf("Max connection attemtps reached")
}
