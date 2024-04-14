package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/realPointer/banners/internal/entity"
	postgresrepo "github.com/realPointer/banners/internal/repository/postgres"
	redisrepo "github.com/realPointer/banners/internal/repository/redis"
	"github.com/realPointer/banners/pkg/postgres"
	"github.com/realPointer/banners/pkg/redis"
	"github.com/rs/zerolog"
)

type Banner interface {
	GetBanner(ctx context.Context, tagID, featureID int) (string, error)
	GetBanners(ctx context.Context, featureID, tagID, limit, offset *int) ([]entity.BannerInfo, error)
	CreateBanner(ctx context.Context, tagIDs []int, featureID int, content json.RawMessage, isActive bool) (int, error)
	UpdateBanner(ctx context.Context, bannerID int, tagIDs []int, featureID *int, content json.RawMessage, isActive *bool) error
	DeleteBanner(ctx context.Context, bannerID int) error
	DeleteBannersByFeatureID(ctx context.Context, featureID int) error
	DeleteBannersByTagID(ctx context.Context, tagID int) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

type User interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
}

var (
	ErrBannerNotFound = errors.New("banner not found")
)

type Repositories struct {
	Banner
	Cache
	User
}

func NewRepositories(l *zerolog.Logger, pg *postgres.Postgres, rdb *redis.Redis) *Repositories {
	return &Repositories{
		Banner: postgresrepo.NewBannerRepo(pg, l),
		Cache:  redisrepo.NewCacheRepo(rdb),
		User:   postgresrepo.NewUserRepo(pg, l),
	}
}
