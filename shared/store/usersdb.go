package store

import (
	"fmt"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store/sqlite"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store/entitites"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared"
	"time"
)

type UsersDb interface{
	Open() error
	Close() error
	Flush() error

	FindTOTPUserOptions(login string) (*entitites.TOTPUserOptions, error)

	UpdateUser(userId int, lockUntil time.Time, failCount int) error
}


func NewUsersDb(dbtype string, connStr string, resolver shared.PathResolver) (UsersDb, error) {
	if dbtype!= "sqlite" {
		return nil, fmt.Errorf("Db with type [%s] not supported", dbtype)
	}

	return sqlite.NewUsersDbSqlite(connStr, resolver)
}