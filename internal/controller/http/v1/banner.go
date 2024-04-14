package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/realPointer/banners/internal/controller/http/v1/middlewares"
	"github.com/realPointer/banners/internal/entity"
	postgresrepo "github.com/realPointer/banners/internal/repository/postgres"
	"github.com/realPointer/banners/internal/service"
	"github.com/rs/zerolog"
)

type bannerRoutes struct {
	bannerService service.Banner
	l             *zerolog.Logger
}

func NewBannerRouter(bannerService service.Banner, l *zerolog.Logger) http.Handler {
	s := &bannerRoutes{
		bannerService: bannerService,
		l:             l,
	}
	r := chi.NewRouter()

	r.Get("/user_banner", s.getBanner)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.AdminOnly)

		r.Route("/banner", func(r chi.Router) {
			r.Get("/", s.getBanners)
			r.Post("/", s.createBanner)
			r.Patch("/{bannerID:[0-9]+}", s.updateBanner)
			r.Delete("/{bannerID:[0-9]+}", s.deleteBanner)
			r.Delete("/feature/{featureID:[0-9]+}", s.deleteBannersByFeatureId)
			r.Delete("/tag/{tagID:[0-9]+}", s.deleteBannersByTagId)
		})
	})

	return r
}

func (s *bannerRoutes) getBanner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tagIDStr := r.URL.Query().Get("tag_id")
	if tagIDStr == "" {
		http.Error(w, "Missing required query parameter tag_id", http.StatusBadRequest)
		return
	}

	featureIDStr := r.URL.Query().Get("feature_id")
	if featureIDStr == "" {
		http.Error(w, "Missing required query parameter feature_id", http.StatusBadRequest)
		return
	}

	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		http.Error(w, "Invalid tag_id", http.StatusBadRequest)
		return
	}

	featureID, err := strconv.Atoi(featureIDStr)
	if err != nil {
		http.Error(w, "Invalid feature_id", http.StatusBadRequest)
		return
	}

	useLastRevision := false
	if useLastRevisionStr := r.URL.Query().Get("use_last_revision"); useLastRevisionStr != "" {
		useLastRevision, err = strconv.ParseBool(useLastRevisionStr)
		if err != nil {
			http.Error(w, "Invalid value for use_last_revision. Allowed values: true, false", http.StatusBadRequest)
			return
		}
	}

	banner, err := s.bannerService.GetBanner(ctx, tagID, featureID, useLastRevision)
	if err != nil {
		switch {
		case errors.Is(err, postgresrepo.ErrBannerNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	render.JSON(w, r, banner)
}

func (s *bannerRoutes) getBanners(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := &entity.BannerFilter{}

	if featureID := r.URL.Query().Get("feature_id"); featureID != "" {
		id, err := strconv.Atoi(featureID)
		if err != nil {
			http.Error(w, "Invalid feature_id", http.StatusBadRequest)
			return
		}
		filter.FeatureID = &id
	}

	if tagID := r.URL.Query().Get("tag_id"); tagID != "" {
		id, err := strconv.Atoi(tagID)
		if err != nil {
			http.Error(w, "Invalid tag_id", http.StatusBadRequest)
			return
		}
		filter.TagID = &id
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		filter.Limit = &l
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
		filter.Offset = &o
	}

	banners, err := s.bannerService.GetBanners(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, banners)
}

func (s *bannerRoutes) createBanner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	banner := entity.BannerCreate{}
	if err := json.NewDecoder(r.Body).Decode(&banner); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(banner.TagIDs) == 0 || banner.FeatureID == nil || len(banner.Content) == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	bannerID, err := s.bannerService.CreateBanner(ctx, &banner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, map[string]int{"banner_id": bannerID})
}

func (s *bannerRoutes) updateBanner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bannerIDStr := chi.URLParam(r, "bannerID")
	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		http.Error(w, "Invalid banner ID", http.StatusBadRequest)
		return
	}

	update := entity.BannerUpdate{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(update.TagIDs) == 0 && update.FeatureID == nil && len(update.Content) == 0 && update.IsActive == nil {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	err = s.bannerService.UpdateBanner(ctx, bannerID, &update)
	if err != nil {
		switch {
		case errors.Is(err, postgresrepo.ErrBannerNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	render.NoContent(w, r)
}

func (s *bannerRoutes) deleteBanner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bannerIDStr := chi.URLParam(r, "bannerID")
	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		http.Error(w, "Invalid banner ID", http.StatusBadRequest)
		return
	}

	err = s.bannerService.DeleteBanner(ctx, bannerID)
	if err != nil {
		switch {
		case errors.Is(err, postgresrepo.ErrBannerNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	render.NoContent(w, r)
}

func (s *bannerRoutes) deleteBannersByFeatureId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	featureIDStr := chi.URLParam(r, "featureID")
	featureID, err := strconv.Atoi(featureIDStr)
	if err != nil {
		http.Error(w, "Invalid feature ID", http.StatusBadRequest)
		return
	}

	err = s.bannerService.DeleteBannersByFeatureID(ctx, featureID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.NoContent(w, r)
}

func (s *bannerRoutes) deleteBannersByTagId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tagID, err := strconv.Atoi(chi.URLParam(r, "tagID"))
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	err = s.bannerService.DeleteBannersByTagID(ctx, tagID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.NoContent(w, r)
}
