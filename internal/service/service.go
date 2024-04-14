package service

import (
	"context"
	"time"

	"github.com/realPointer/banners/internal/entity"
	"github.com/realPointer/banners/internal/repository"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Banner interface {
	GetBanner(ctx context.Context, tagID, featureID int, useLastRevision bool) (string, error)
	GetBanners(ctx context.Context, filter *entity.BannerFilter) ([]entity.BannerInfo, error)
	CreateBanner(ctx context.Context, banner *entity.BannerCreate) (int, error)
	UpdateBanner(ctx context.Context, bannerID int, update *entity.BannerUpdate) error
	DeleteBanner(ctx context.Context, bannerID int) error
	DeleteBannersByFeatureID(ctx context.Context, featureID int) error
	DeleteBannersByTagID(ctx context.Context, tagID int) error
}

type Auth interface {
	Register(ctx context.Context, username, password, role string) error
	Login(ctx context.Context, username, password string) (string, error)
	ParseToken(tokenString string) (*TokenClaims, error)
}

type Services struct {
	Banner
	Auth
}

type ServicesDependencies struct {
	Repositories *repository.Repositories
	SignKey      string
	TokenTTL     time.Duration
	Salt         string
}

func NewServices(l *zerolog.Logger, deps ServicesDependencies) *Services {
	return &Services{
		Banner: NewBannerService(l, deps.Repositories.Banner, deps.Repositories.Cache),
		Auth:   NewAuthService(l, deps.Repositories.User, deps.SignKey, deps.TokenTTL, deps.Salt),
	}
}
