package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/realPointer/banners/internal/entity"
	"github.com/realPointer/banners/pkg/postgres"
	"github.com/rs/zerolog"
)

type BannerRepo struct {
	*postgres.Postgres
	l *zerolog.Logger
}

func NewBannerRepo(pg *postgres.Postgres, l *zerolog.Logger) *BannerRepo {
	return &BannerRepo{
		Postgres: pg,
		l:        l,
	}
}

var ErrBannerNotFound = errors.New("banner not found")

func (r *BannerRepo) GetBanner(ctx context.Context, tagID, featureID int) (string, error) {
	sql, args, _ := r.Builder.
		Select("bv.content").
		From("banners b").
		Join("feature_tag_banners ftb ON b.id = ftb.banner_id").
		Join("banner_versions bv ON b.id = bv.banner_id").
		Where(squirrel.Eq{
			"ftb.tag_id":     tagID,
			"ftb.feature_id": featureID,
			"b.deleted":      false,
			"b.is_active":    true,
		}).
		Where("bv.version = b.last_version").
		ToSql()

	var content string
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(&content)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrBannerNotFound
		}
		return "", fmt.Errorf("BannerRepo.GetBanner - r.Pool.QueryRow: %w", err)
	}

	return content, nil
}

func (r *BannerRepo) GetBanners(ctx context.Context, featureID, tagID, limit, offset *int) ([]entity.BannerInfo, error) {
	query := r.Builder.
		Select("b.id as banner_id", "b.is_active", "b.created_at", "b.updated_at", "ftb.feature_id", "array_agg(ftb.tag_id ORDER BY ftb.tag_id ASC) as tag_ids", "bv.content").
		From("banners b").
		Join("feature_tag_banners ftb ON b.id = ftb.banner_id").
		Join("banner_versions bv ON b.id = bv.banner_id").
		Where("bv.version = b.last_version").
		GroupBy("b.id", "b.is_active", "b.created_at", "b.updated_at", "ftb.feature_id", "bv.content")

	if featureID != nil {
		query = query.Where(squirrel.Eq{"ftb.feature_id": *featureID})
	}

	if tagID != nil {
		query = query.Where(squirrel.Eq{"ftb.tag_id": *tagID})
	}

	if limit != nil {
		query = query.Limit(uint64(*limit))
	}

	if offset != nil {
		query = query.Offset(uint64(*offset))
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("BannerRepo.GetBanners - query.ToSql: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("BannerRepo.GetBanners - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	var banners []entity.BannerInfo
	for rows.Next() {
		var banner entity.BannerInfo

		err := rows.Scan(&banner.BannerID, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt, &banner.FeatureID, &banner.TagIDs, &banner.Content)
		if err != nil {
			return nil, fmt.Errorf("BannerRepo.GetBanners - rows.Scan: %w", err)
		}

		banners = append(banners, banner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("BannerRepo.GetBanners - rows.Err: %w", err)
	}

	return banners, nil
}

func (r *BannerRepo) CreateBanner(ctx context.Context, tagIDs []int, featureID int, content json.RawMessage, isActive bool) (int, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("BannerRepo.CreateBanner - r.Pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	sqlInsertBanner, argsInsertBanner, _ := r.Builder.
		Insert("banners").
		Columns("is_active").
		Values(isActive).
		Suffix("RETURNING id").
		ToSql()

	var bannerID int
	err = tx.QueryRow(ctx, sqlInsertBanner, argsInsertBanner...).Scan(&bannerID)
	if err != nil {
		return 0, fmt.Errorf("BannerRepo.CreateBanner - tx.QueryRow: %w", err)
	}

	sqlInsertVersion, argsInsertVersion, _ := r.Builder.
		Insert("banner_versions").
		Columns("banner_id", "content").
		Values(bannerID, content).
		ToSql()

	_, err = tx.Exec(ctx, sqlInsertVersion, argsInsertVersion...)
	if err != nil {
		return 0, fmt.Errorf("BannerRepo.CreateBanner - tx.Exec: %w", err)
	}

	for _, tagID := range tagIDs {
		sqlInsertFeatureTag, argsInsertFeatureTag, _ := r.Builder.
			Insert("feature_tag_banners").
			Columns("banner_id", "feature_id", "tag_id").
			Values(bannerID, featureID, tagID).
			ToSql()

		_, err = tx.Exec(ctx, sqlInsertFeatureTag, argsInsertFeatureTag...)
		if err != nil {
			return 0, fmt.Errorf("BannerRepo.CreateBanner - tx.Exec: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("BannerRepo.CreateBanner - tx.Commit: %w", err)
	}

	return bannerID, nil
}

func (r *BannerRepo) UpdateBanner(ctx context.Context, bannerID int, tagIDs []int, featureID *int, content json.RawMessage, isActive *bool) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("BannerRepo.UpdateBanner - begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var lastVersion int
	var newVersion int

	if content != nil {
		sqlGetLastVersion, argsGetLastVersion, _ := r.Builder.Select("last_version").From("banners").Where(squirrel.Eq{"id": bannerID}).ToSql()
		err = tx.QueryRow(ctx, sqlGetLastVersion, argsGetLastVersion...).Scan(&lastVersion)
		if err != nil {
			return fmt.Errorf("BannerRepo.UpdateBanner - get last_version: %w", err)
		}
		newVersion = lastVersion + 1

		sqlInsertBannerVersion, argsInsertBannerVersion, _ := r.Builder.Insert("banner_versions").
			Columns("banner_id", "version", "content").
			Values(bannerID, newVersion, content).
			ToSql()
		_, err = tx.Exec(ctx, sqlInsertBannerVersion, argsInsertBannerVersion...)
		if err != nil {
			return fmt.Errorf("BannerRepo.UpdateBanner - insert banner_versions: %w", err)
		}
	}

	var currentFeatureID int
	if len(tagIDs) > 0 {
		sqlGetFeatureID, argsGetFeatureID, _ := r.Builder.Select("feature_id").From("feature_tag_banners").Where(squirrel.Eq{"banner_id": bannerID}).Limit(1).ToSql()
		err = tx.QueryRow(ctx, sqlGetFeatureID, argsGetFeatureID...).Scan(&currentFeatureID)
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("BannerRepo.UpdateBanner - get current feature_id: %w", err)
		}

		sqlDeleteFeatureTags, argsDeleteFeatureTags, _ := r.Builder.Delete("feature_tag_banners").Where(squirrel.Eq{"banner_id": bannerID}).ToSql()
		_, err = tx.Exec(ctx, sqlDeleteFeatureTags, argsDeleteFeatureTags...)
		if err != nil {
			return fmt.Errorf("BannerRepo.UpdateBanner - delete feature_tag_banners: %w", err)
		}

		for _, tagID := range tagIDs {
			sqlInsertFeatureTag, argsInsertFeatureTag, _ := r.Builder.Insert("feature_tag_banners").
				Columns("banner_id", "feature_id", "tag_id").
				Values(bannerID, currentFeatureID, tagID).
				ToSql()
			_, err = tx.Exec(ctx, sqlInsertFeatureTag, argsInsertFeatureTag...)
			if err != nil {
				return fmt.Errorf("BannerRepo.UpdateBanner - insert feature_tag_banners: %w", err)
			}
		}
	}

	if featureID != nil {
		sqlUpdateFeatureTags, argsUpdateFeatureTags, _ := r.Builder.Update("feature_tag_banners").
			Set("feature_id", *featureID).
			Where(squirrel.Eq{"banner_id": bannerID}).
			ToSql()
		_, err = tx.Exec(ctx, sqlUpdateFeatureTags, argsUpdateFeatureTags...)
		if err != nil {
			return fmt.Errorf("BannerRepo.UpdateBanner - update feature_tag_banners: %w", err)
		}
	}

	updateBannerQuery := r.Builder.Update("banners").Where(squirrel.Eq{"id": bannerID})

	if isActive != nil {
		updateBannerQuery = updateBannerQuery.Set("is_active", *isActive)
	}

	if content != nil {
		updateBannerQuery = updateBannerQuery.Set("last_version", newVersion)
	}

	updateBannerQuery = updateBannerQuery.Set("updated_at", squirrel.Expr("NOW()"))

	sqlUpdateBanner, argsUpdateBanner, _ := updateBannerQuery.ToSql()
	_, err = tx.Exec(ctx, sqlUpdateBanner, argsUpdateBanner...)
	if err != nil {
		return fmt.Errorf("BannerRepo.UpdateBanner - update banners: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("BannerRepo.UpdateBanner - commit transaction: %w", err)
	}

	return nil
}

func (r *BannerRepo) DeleteBanner(ctx context.Context, bannerID int) error {
	sql, args, _ := r.Builder.
		Delete("banners").
		Where(squirrel.Eq{"id": bannerID}).
		ToSql()

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("BannerRepo.DeleteBanner - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *BannerRepo) DeleteBannersByFeatureID(ctx context.Context, featureID int) error {
	sql, args, _ := r.Builder.
		Update("banners b").
		Set("deleted", true).
		Where("b.id IN (SELECT banner_id FROM feature_tag_banners WHERE feature_id = $2)", featureID).
		ToSql()

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("BannerRepo.DeleteBannersByFeatureID - update banners: %w", err)
	}

	return nil
}

func (r *BannerRepo) DeleteBannersByTagID(ctx context.Context, tagID int) error {
	sql, args, _ := r.Builder.
		Update("banners b").
		Set("deleted", true).
		Where("b.id IN (SELECT banner_id FROM feature_tag_banners WHERE tag_id = $2)", tagID).
		ToSql()

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("BannerRepo.DeleteBannersByTagID - update banners: %w", err)
	}

	return nil
}
