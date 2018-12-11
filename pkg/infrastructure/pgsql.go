package infrastructure

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/loggers"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
)

type PgsqlHandler struct {
	Conn *sql.DB
}

func (handler *PgsqlHandler) Healthcheck() bool {
	return handler.Conn.Ping() == nil
}

func (handler *PgsqlHandler) Close() error {
	return handler.Conn.Close()
}

func (handler *PgsqlHandler) Insert(statement string) error {
	_, err := handler.Conn.Exec(statement)
	return err
}

func (handler *PgsqlHandler) Update(statement string) error {
	_, err := handler.Conn.Exec(statement)
	return err
}

func (handler *PgsqlHandler) Query(statement string) (repository.DbResult, error) {
	rows, err := handler.Conn.Query(statement)
	if err != nil {
		fmt.Println(err)
		return new(PgsqlRow), err
	}
	return PgsqlRow{
		Rows: rows,
	}, nil
}

type PgsqlRow struct {
	Rows *sql.Rows
}

func (r PgsqlRow) Scan(dest ...interface{}) {
	r.Rows.Scan(dest...)
}

func (r PgsqlRow) Next() bool {
	return r.Rows.Next()
}

func (r PgsqlRow) Close() error {
	return r.Rows.Close()
}

func NewPgsqlHandler(conf DatabaseConfig, logger loggers.Logger) (*PgsqlHandler, error) {
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
				Conn: poolDb,
			}, nil
		}
	}

	return nil, fmt.Errorf("Max connection attemtps reached")
}
