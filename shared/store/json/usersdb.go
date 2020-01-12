package json

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store/entitites"
	scribble "github.com/nanobox-io/golang-scribble"
	"hash"
	"time"
)

type usersDbJson struct {
	dbDir string
	db *scribble.Driver
}

type userRow struct {
	Login     string
	Tries     []int64
	TimeStep  int
	Digits    uint8
	SecretKey string
	Hash      int

	LockStrategy entitites.LockStrategyType
	LockUntil    string
	LockTimeout  int

	FailCountBeforeLock int
	FailCount           int
}

func (u *usersDbJson) Open() (err error) {
	u.db, err = scribble.New(u.dbDir, nil)
	return
}

func (u *usersDbJson) Close() error {
	u.db = nil
	return nil
}

func (u *usersDbJson) Flush() error {
	return nil
}

func (u *usersDbJson) Find(login string) (*entitites.TOTPUserOptions, error) {
	if u.db == nil {
		return nil, fmt.Errorf("db not opened")
	}

	row, hash, secret, err := findStored(u, login)
	if err != nil {
		return nil, err
	}

	var lockUntil time.Time
	if row.LockUntil == "" {
		lockUntil = time.Unix(0, 0)
	} else {
		lockUntil, err = time.Parse(time.RFC3339, row.LockUntil)
		if err != nil {
			return nil, err
		}
	}

	return &entitites.TOTPUserOptions{
		Login:               login,
		Digits:              row.Digits,
		TimeStep:            time.Duration(row.TimeStep) * time.Second,
		Time:                time.Now,
		Hash:                hash,
		Tries:               row.Tries,
		Secret:              secret,
		LockStrategy:        row.LockStrategy,
		LockUntil:           lockUntil,
		LockTimeout:         time.Duration(row.LockTimeout) * time.Second,
		FailCount:           row.FailCount,
		FailCountBeforeLock: row.FailCountBeforeLock,
	}, nil
}

func (u *usersDbJson) Update(login string, lockUntil time.Time, failCount int) error {
	if u.db == nil {
		return fmt.Errorf("db not opened")
	}

	row, _, _, err := findStored(u, login)
	if err != nil {
		return err
	}

	if lockUntil.IsZero() {
		row.LockUntil = ""
	} else {
		row.LockUntil = lockUntil.Format(time.RFC3339)
	}
	row.FailCount = failCount

	err = u.db.Write("user", login, row)
	return err
}
func NewUsersDb(dbDir string, resolver shared.PathResolver) (*usersDbJson, error) {
	return &usersDbJson{
		dbDir: resolver.PathToAbs(dbDir),
	}, nil
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


func findStored(u *usersDbJson, login string) (userRow, func() hash.Hash, []byte, error) {
	var row userRow
	err := u.db.Read("user", login, &row)
	if err != nil {
		return userRow{}, nil, nil, err
	}

	hash := hashFactory(row.Hash)
	if hash == nil {
		return userRow{}, nil, nil, fmt.Errorf("unable to parse hash for algo %d: unknown algo for login %s", row.Hash, login)
	}

	secret, err := hex.DecodeString(row.SecretKey)
	if err != nil {
		return userRow{}, nil, nil, fmt.Errorf("unable to parse secretkey for login %s", login)
	}
	return row, hash, secret, nil
}
