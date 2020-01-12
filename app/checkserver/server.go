package main

import (
	"context"
	"github.com/balasanjay/totp"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
	"log"

	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/api"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store/entitites"
	"time"
)

type server struct {
	db     store.UsersDb
	apikey string
}

func (s *server) Check(ctx context.Context, req *api.CheckRequest) (*api.CheckResponse, error) {
	const (
		// CheckError_None =  "0x00"

		CheckError_NoMeta = "0x01"
		CheckError_NoApiKey = "0x02"
		CheckError_EmptyApiKey = "0x03"
		CheckError_IncorrectApiKey = "0x04"

		CheckError_UnsupportedOtpType = "0x50"
		CheckError_LoginNotSet = "0x51"
		CheckError_CodeNotSet = "0x52"

		CheckError_UnknownUser = "0x53"
		CheckError_UserLocked = "0x54"

		CheckError_IncorrectCode = "0x999"
	)

	checkApiKeyFn := func(ctx context.Context) error {
		var meta metadata.MD
		var ok bool

		if meta, ok = metadata.FromIncomingContext(ctx); !ok {
			return errors.New(CheckError_NoMeta)
		}

		if apikeys, ok := meta["apikey"]; !ok {
			return  errors.New(CheckError_NoApiKey)
		} else {
			if apikeys == nil || len(apikeys) != 1 {
				return  errors.New(CheckError_EmptyApiKey)
			}

			if s.apikey != apikeys[0] {
				return  errors.New(CheckError_IncorrectApiKey)
			}

			return nil
		}
	}

	checkRequestFn := func(req *api.CheckRequest) error {
		if req.Type != api.CheckRequest_TOTP {
			return  errors.New(CheckError_UnsupportedOtpType)
		}

		if len(req.Login) == 0 {
			return errors.New(CheckError_LoginNotSet)
		}

		if len(req.Code) == 0 {
			return errors.New(CheckError_CodeNotSet)
		}

		return nil
	}

	err := checkApiKeyFn(ctx)
	if err != nil {
		return nil, err
	}

	log.Printf("INFO: try to Authentificate OTP code. type:%d, login:%s, code:%s", req.Type, req.Login, req.Code)

	err = checkRequestFn(req)
	if err != nil {
		return nil, err
	}

	user, err := s.db.Find(req.Login)
	if err != nil {
		log.Printf("ERROR: unable to fetch user %s from db; error:%s", req.Login, err.Error())
		return nil, errors.New(CheckError_UnknownUser)
	}

	now := time.Now()
	if now.Before(user.LockUntil) {
		log.Printf("INFO: login %s is LOCKED until [%s], now: [%s]", req.Login, user.LockUntil.String(), now.String())
		onAuthenticateLocked(user, s.db)
		return nil, errors.New(CheckError_UserLocked)
	}

	authRes := totp.Authenticate(user.Secret, req.Code, &totp.Options{
		Hash:     user.Hash,
		Time:     user.Time,
		TimeStep: user.TimeStep,
		Digits:   user.Digits,
		Tries:    user.Tries,
	})

	if !authRes {
		onAuthenticateFailed(user, s.db)
		log.Printf("ERROR: code %s for user %s is incorrect", req.Code, req.Login)
		return nil, errors.New(CheckError_IncorrectCode)
	}

	log.Printf("INFO: success login:%s, code:%s", req.Login, req.Code)
	onAuthenticateSuccess(user, s.db)

	resp := &api.CheckResponse{}
	return resp, nil
}

func onAuthenticateLocked(user *entitites.TOTPUserOptions, db store.UsersDb) {
	err := db.Update(user.Login, user.LockUntil, user.FailCount+1)
	if err != nil {
		log.Printf("WARNING: cant update user %s with login %s; error:%s", user.Login, user.Login, err.Error())
	}
}

func onAuthenticateSuccess(user *entitites.TOTPUserOptions, db store.UsersDb) {
	err := db.Update(user.Login, time.Unix(0, 0), 0)
	if err != nil {
		log.Printf("WARNING: cant update user %s with login %s; error:%s", user.Login, user.Login, err.Error())
	}
}

func onAuthenticateFailed(user *entitites.TOTPUserOptions, db store.UsersDb) {
	newLockUntil := user.LockUntil
	if user.FailCount+1 >= user.FailCountBeforeLock {
		newLockUntil = time.Now().Add(user.LockTimeout)
	}

	switch user.LockStrategy {
	case entitites.LockStrategyType_None:
		break
	case entitites.LockStrategyType_Simple:
		err := db.Update(user.Login, newLockUntil, user.FailCount+1)
		if err != nil {
			log.Printf("WARNING: cant update user %s with login %s; error:%s", user.Login, user.Login, err.Error())
		}
		break
	default:
		log.Printf("WARNING: unknown LockStrategy value [%d] for user %s", user.LockStrategy, user.Login)
		break
	}
}
