package store

import (
	"fmt"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store/entitites"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store/json"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store/sqlite"
	"time"
)

type UsersDb interface{
	Open() error
	Close() error
	Flush() error

	Find(login string) (*entitites.TOTPUserOptions, error)

	Update(login string, lockUntil time.Time, failCount int) error
}


func NewUsersDb(dbtype string, connStr string, resolver shared.PathResolver) (UsersDb, error) {
	switch dbtype {
	case "sqlite":
		return sqlite.NewUsersDb(connStr, resolver)
	case "json":
		return json.NewUsersDb(connStr, resolver)
	default:
		return nil, fmt.Errorf("db with type [%s] not supported", dbtype)
	}
}