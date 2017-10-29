package configuration

import (
	"github.com/pelletier/go-toml"
	"log"
)

type AppConfig interface {
	GetOrDie(path string) interface{}
}

type appConfig struct {
	db *toml.Tree
}

func NewAppConfig(filenameFull string) (AppConfig, error) {
	db, err := toml.LoadFile(filenameFull)
	if err!=nil {
		return nil, err
	}

	return &appConfig{
		db: db,
	}, nil
}

func (c* appConfig) GetOrDie(path string) interface{} {
	rv := c.db.Get(path)
	if rv == nil {
		log.Fatalf("%s not set in configuration", path)
	}
	return rv
}