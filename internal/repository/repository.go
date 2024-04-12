package repository

import "github.com/realPointer/banners/pkg/postgres"

type Repositories struct {
}

func NewRepositories(pg *postgres.Postgres) *Repositories {
	return &Repositories{}
}
