package controllers

import (
	"github.com/prest/prest/adapters"
	"github.com/prest/prest/config"
)

type DBHandlers struct {
	Get map[string]DB
}

type DB struct {
	Adapter adapters.Adapter
	Config  config.Prest
}

func New() DBHandlers {
	db := DBHandlers{
		Get: make(map[string]DB),
	}
	// create pg load with custom envs
	// ...
	// load default db
	db.Get[config.PrestConf.PGDatabase] = DB{
		Adapter: config.PrestConf.Adapter,
		Config:  *config.PrestConf,
	}
	return db
}
