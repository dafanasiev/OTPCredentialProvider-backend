package entitites

import (
	"time"
	"hash"
)

type LockStrategyType int

const (
	LockStrategyType_None LockStrategyType = 0
	LockStrategyType_Simple  = 1
)

type TOTPUserOptions struct{
	UserId int
	Login string

	Time     func() time.Time
	Tries    []int64
	TimeStep time.Duration
	Digits   uint8
	Hash     func() hash.Hash
	Secret []byte

	LockStrategy LockStrategyType
	LockUntil time.Time
	LockTimeout time.Duration

	FailCount int
	FailCountBeforeLock int
}
