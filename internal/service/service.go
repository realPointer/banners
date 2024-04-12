package service

import "github.com/realPointer/banners/internal/repository"

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Services struct {
}

type ServicesDependencies struct {
	Repositories *repository.Repositories
}

func NewServices(deps ServicesDependencies) *Services {
	return &Services{}
}
