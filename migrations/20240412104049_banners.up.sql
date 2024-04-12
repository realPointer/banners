CREATE TABLE IF NOT EXISTS banners (
    id SERIAL PRIMARY KEY,
    last_version INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS feature_tag_banners (
    banner_id INTEGER NOT NULL REFERENCES banners (id) ON DELETE CASCADE,
    feature_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (feature_id, tag_id, banner_id),
    CONSTRAINT unq_banner_feature UNIQUE (banner_id, feature_id)
);

CREATE TABLE IF NOT EXISTS banner_versions (
    banner_id INTEGER NOT NULL REFERENCES banners (id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1,
    content JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (banner_id, version)
);

CREATE INDEX idx_banners_deleted ON banners (deleted);

CREATE INDEX idx_feature_tag_banners_feature_id ON feature_tag_banners (feature_id);

CREATE INDEX idx_feature_tag_banners_tag_id ON feature_tag_banners (tag_id);

CREATE INDEX idx_feature_tag_banners_feature_tag ON feature_tag_banners (feature_id, tag_id);