package model

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/mickael-kerjean/filestash/server/common"
	"os"
	"path/filepath"
	"time"
)

var DB *sql.DB

func init() {
	cachePath := filepath.Join(GetCurrentDir(), DB_PATH)
	os.MkdirAll(cachePath, os.ModePerm)

	DB, err := sql.Open(Config.Get("database.driver_name").String(), Config.Get("database.data_source_name").String())
	if err != nil {
		panic(err)
	}

	stmt, err := DB.Prepare("CREATE TABLE IF NOT EXISTS Location(backend VARCHAR(16), path VARCHAR(512), CONSTRAINT pk_location PRIMARY KEY(backend, path))")
	if err != nil {
		panic(err)
	}

	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}

	stmt, err = DB.Prepare("CREATE TABLE IF NOT EXISTS Share(id VARCHAR(64) PRIMARY KEY, related_backend VARCHAR(16), related_path VARCHAR(512), params JSON, auth VARCHAR(4093) NOT NULL, FOREIGN KEY (related_backend, related_path) REFERENCES Location(backend, path) ON UPDATE CASCADE ON DELETE CASCADE)")
	if err != nil {
		panic(err)
	}

	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}

	stmt, err = DB.Prepare("CREATE TABLE IF NOT EXISTS Verification(`key` VARCHAR(512), code VARCHAR(4), expire DATETIME NOT NULL)")
	if err != nil {
		panic(err)
	}

	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}

	stmt, err = DB.Prepare("CREATE INDEX idx_verification ON Verification(code, expire)")
	if err != nil {
		panic(err)
	}

	_, err = stmt.Exec()
	if err != nil && err.(*mysql.MySQLError).Number != 1061 {
		panic(err)
	}

	go func() {
		autovacuum()
	}()
}

func autovacuum() {
	if stmt, err := DB.Prepare("DELETE FROM Verification WHERE expire < datetime('now')"); err == nil {
		stmt.Exec()
	}
	time.Sleep(6 * time.Hour)
}
