//+build !windows

package sqlite

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"hash"

	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store/entitites"

	"encoding/hex"
	"fmt"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	"strings"
	"sync"
	"time"
)

type usersDbSqlite struct {
	fileName string
	db       *sql.DB
	lock     sync.RWMutex
}

func NewUsersDb(fileName string, resolver shared.PathResolver) (*usersDbSqlite, error) {
	return &usersDbSqlite{
		fileName: resolver.PathToAbs(fileName),
		lock:     sync.RWMutex{},
	}, nil
}

func (d *usersDbSqlite) Open() error {
	d.lock.Lock()

	var err error
	d.db, err = sql.Open("sqlite3", d.fileName)
	if err != nil {
		return err
	}

	sql := `CREATE TABLE IF NOT EXISTS user(
				login TEXT primary key NOT NULL,
				Tries TEXT NOT NULL,
				TimeStep integer NOT NULL,
				Digits integer NOT NULL,
				SecretKey TEXT NOT NULL,
				Hash integer NOT NULL,	--0:sha1,1:sha256,2:sha512


				LockStrategy TINYINT NOT NULL,	--0:none,1:now()+LockTimeout if FailCount>=FailCountBeforeLock
				LockUntil BIGINT NOT NULL,		--lock until this date (unix timestamp), or 0 of not locked
				LockTimeout	INTEGER NOT NULL,	--in seconds

				FailCountBeforeLock INTEGER NOT NULL,
				FailCount INTEGER NOT NULL
			)`

	if _, err = d.db.Exec(sql); err != nil {
		d.lock.Unlock()
		d.Close()
		return err
	}

	d.lock.Unlock()
	return nil
}

func (d *usersDbSqlite) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.db.Close()
}

func (d *usersDbSqlite) Flush() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	err := d.Close()
	if err != nil {
		return err
	}

	return d.Open()
}

func (d *usersDbSqlite) Find(login string) (*entitites.TOTPUserOptions, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	type userRow struct {
		Login     string
		Tries     string
		TimeStep  int
		Digits    uint8
		SecretKey string
		Hash      int

		LockStrategy entitites.LockStrategyType
		LockUntil    int64
		LockTimeout  int

		FailCountBeforeLock int
		FailCount           int
	}

	sql := `SELECT login,tries,timestep,digits,secretkey,hash,LockStrategy,LockUntil,LockTimeout,FailCount,FailCountBeforeLock  FROM user WHERE login=?`
	stmt, err := d.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	row := userRow{}
	err = stmt.QueryRow(login).Scan(&row.Login,
		&row.Tries,
		&row.TimeStep,
		&row.Digits,
		&row.SecretKey,
		&row.Hash,
		&row.LockStrategy,
		&row.LockUntil,
		&row.LockTimeout,
		&row.FailCount,
		&row.FailCountBeforeLock)
	if err != nil {
		return nil, err
	}

	hash := hashFactory(row.Hash)
	if hash == nil {
		return nil, fmt.Errorf("Unable to parse hash for algo %d: unknown algo for login %s", row.Hash, login)
	}

	tries, err := stringToInt64Array(row.Tries)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse tries for login %s", login)
	}

	secret, err := hex.DecodeString(row.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse secretkey for login %s", login)
	}

	return &entitites.TOTPUserOptions{
		Login:               login,
		Digits:              row.Digits,
		TimeStep:            time.Duration(row.TimeStep) * time.Second,
		Time:                time.Now,
		Hash:                hash,
		Tries:               tries,
		Secret:              secret,
		LockStrategy:        row.LockStrategy,
		LockUntil:           time.Unix(row.LockUntil, 0),
		LockTimeout:         time.Duration(row.LockTimeout) * time.Second,
		FailCount:           row.FailCount,
		FailCountBeforeLock: row.FailCountBeforeLock,
	}, nil

}

func (d *usersDbSqlite) Update(login string, lockUntil time.Time, failCount int) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	sql := `UPDATE user SET LockUntil=?, FailCount=? WHERE Login=?`

	stmt, err := d.db.Prepare(sql)
	defer stmt.Close()

	if err != nil {
		return err
	}

	_, err = stmt.Exec(lockUntil.Unix(), failCount, login)

	return err
}

func stringToInt64Array(s string) ([]int64, error) {
	rv := make([]int64, 0)
	if s == "" {
		return rv, nil
	}

	sArr := strings.Split(s, ",")
	for _, sItem := range sArr {
		i, err := strconv.Atoi(sItem)
		if err != nil {
			return nil, err
		}

		rv = append(rv, int64(i))
	}

	return rv, nil
}

func hashFactory(algo int) func() hash.Hash {
	switch algo {
	case 0:
		return sha1.New
	case 1:
		return sha256.New
	case 2:
		return sha512.New
	default:
		return nil
	}
}
