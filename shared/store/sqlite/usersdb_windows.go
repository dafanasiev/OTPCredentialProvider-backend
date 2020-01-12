//+build windows

package sqlite

import (
	"errors"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store/entitites"
	"time"
)

type fake struct{

}

func (f *fake) Open() error {
	panic("implement me")
}

func (f *fake) Close() error {
	panic("implement me")
}

func (f *fake) Flush() error {
	panic("implement me")
}

func (f *fake) Find(login string) (*entitites.TOTPUserOptions, error) {
	panic("implement me")
}

func (f *fake) Update(login string, lockUntil time.Time, failCount int) error {
	panic("implement me")
}

func NewUsersDb(fileName string, resolver shared.PathResolver) (*fake, error) {
	return nil, errors.New("windows not supported by sqlite provider")
}

