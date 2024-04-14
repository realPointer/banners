package service

import (
	"context"
	"fmt"
	"time"

	"github.com/realPointer/banners/internal/entity"
	"github.com/realPointer/banners/internal/repository"
	"github.com/rs/zerolog"
)

type BannerService struct {
	bannerRepo repository.Banner
	cacheRepo  repository.Cache
	l          *zerolog.Logger
}

func NewBannerService(l *zerolog.Logger, bannerRepo repository.Banner, cacheRepo repository.Cache) *BannerService {
	return &BannerService{
		bannerRepo: bannerRepo,
		cacheRepo:  cacheRepo,
		l:          l,
	}
}

func (s *BannerService) GetBanner(ctx context.Context, tagID, featureID int, useLastRevision bool) (string, error) {
	cacheKey := fmt.Sprintf("banner:%d:%d", tagID, featureID)

	if !useLastRevision {
		cachedData, err := s.cacheRepo.Get(ctx, cacheKey)
		if err == nil && cachedData != "" {
			return cachedData, nil
		}
	}

	data, err := s.bannerRepo.GetBanner(ctx, tagID, featureID)
	if err != nil {
		return "", err
	}

	err = s.cacheRepo.Set(ctx, cacheKey, data, 5*time.Minute)
	if err != nil {
		return "", err
	}

	return data, nil
}

func (s *BannerService) GetBanners(ctx context.Context, filter *entity.BannerFilter) ([]entity.BannerInfo, error) {
	return s.bannerRepo.GetBanners(ctx, filter.FeatureID, filter.TagID, filter.Limit, filter.Offset)
}

func (s *BannerService) CreateBanner(ctx context.Context, banner *entity.BannerCreate) (int, error) {
	return s.bannerRepo.CreateBanner(ctx, banner.TagIDs, *banner.FeatureID, banner.Content, banner.IsActive)
}

func (s *BannerService) UpdateBanner(ctx context.Context, bannerID int, update *entity.BannerUpdate) error {
	return s.bannerRepo.UpdateBanner(ctx, bannerID, update.TagIDs, update.FeatureID, update.Content, update.IsActive)
}

func (s *BannerService) DeleteBanner(ctx context.Context, bannerID int) error {
	return s.bannerRepo.DeleteBanner(ctx, bannerID)
}

func (s *BannerService) DeleteBannersByFeatureID(ctx context.Context, featureID int) error {
	return s.bannerRepo.DeleteBannersByFeatureID(ctx, featureID)
}

func (s *BannerService) DeleteBannersByTagID(ctx context.Context, tagID int) error {
	return s.bannerRepo.DeleteBannersByTagID(ctx, tagID)
}
