-- Sellers Table
CREATE TABLE IF NOT EXISTS sellers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    epic_seller_id TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL
);

-- Developers Table
CREATE TABLE IF NOT EXISTS developers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Main Free Games Table
CREATE TABLE IF NOT EXISTS free_games (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    epic_game_id TEXT UNIQUE NOT NULL,
    seller_id BIGINT REFERENCES sellers(id),
    developer_id BIGINT REFERENCES developers(id),

    namespace TEXT,
    title TEXT NOT NULL,
    description TEXT,
    offer_type TEXT,
    status TEXT,
    requires_redemption_code BOOLEAN,
    product_slug TEXT,
    url_slug TEXT,

    discount_price INT,
    original_price INT,
    voucher INT,
    discount INT,
    currency_code TEXT,
    decimals INT,

    fmt_original_price TEXT,
    fmt_discount_price TEXT,
    fmt_intermediate_price TEXT,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Images Table (1-to-Many)
CREATE TABLE IF NOT EXISTS free_game_images (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    free_game_id BIGINT NOT NULL REFERENCES free_games(id) ON DELETE CASCADE,
    image_type TEXT NOT NULL,
    url TEXT NOT NULL
);

-- Mappings Table (Handles both 'Catalogs' and 'Offers')
CREATE TABLE IF NOT EXISTS free_game_mappings (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    free_game_id BIGINT NOT NULL REFERENCES free_games(id) ON DELETE CASCADE,
    mapping_category TEXT NOT NULL, -- e.g., 'catalog' or 'offer'
    slug TEXT NOT NULL,
    mapping_type TEXT NOT NULL
);

-- Promotions Table (Handles both 'Current' and 'Upcoming' with flattened settings)
CREATE TABLE IF NOT EXISTS free_game_promotions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    free_game_id BIGINT NOT NULL REFERENCES free_games(id) ON DELETE CASCADE,
    promo_tier TEXT NOT NULL,

    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,

    discount_type TEXT,
    discount_percentage INT
);

-- Free games sent to WhatsApp
CREATE TABLE IF NOT EXISTS free_games_sent (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    free_game_id BIGINT NOT NULL REFERENCES free_games(id) ON DELETE CASCADE
);

-- CREATE INDEX

CREATE INDEX IF NOT EXISTS idx_fgs_free_game_id ON free_games_sent(free_game_id);
CREATE INDEX IF NOT EXISTS idx_fgi_free_game_id ON free_game_images(free_game_id);
CREATE INDEX IF NOT EXISTS idx_fgm_free_game_id ON free_game_mappings(free_game_id);
CREATE INDEX IF NOT EXISTS idx_fgp_free_game_id ON free_game_promotions(free_game_id);
CREATE INDEX IF NOT EXISTS idx_fg_seller_id ON free_games(seller_id);
CREATE INDEX IF NOT EXISTS idx_fg_developer_id ON free_games(developer_id);

--- SELECT QUERY

WITH GameImages AS (
    SELECT
        free_game_id,
        json_agg(json_build_object('type', image_type, 'url', url)) AS images
    FROM
        free_game_images
    WHERE
        image_type IN ('OfferImageWide', 'Thumbnail')
    GROUP BY
        free_game_id
),
UniquePromotions AS (
    SELECT DISTINCT
        free_game_id,
        promo_tier,
        start_date,
        end_date
    FROM
        free_game_promotions
),
UniqueMappings AS (
    SELECT DISTINCT
        free_game_id,
        slug
    FROM
        free_game_mappings
)
SELECT
    fg.id,
    fg.epic_game_id AS game_id,
    fg.namespace,
    fg.title,
    fg.description,
    fg.offer_type,
    fg.status,
    fg.requires_redemption_code,
    s."name" AS seller,
    d."name" AS developer,
    fgm.slug,
    fgi.images,
    fgp.promo_tier AS period,
    fg.fmt_original_price,
    fg.fmt_discount_price,
    fg.fmt_intermediate_price,
    fgp.start_date,
    fgp.end_date
FROM
    free_games fg
JOIN
    GameImages fgi ON fgi.free_game_id = fg.id
JOIN
    UniquePromotions fgp ON fgp.free_game_id = fg.id
LEFT JOIN
    sellers s ON s.id = fg.seller_id
LEFT JOIN
    developers d ON d.id = fg.developer_id
LEFT JOIN
    UniqueMappings fgm ON fgm.free_game_id = fg.id
WHERE
    fg.offer_type = 'BASE_GAME'
    AND fg.requires_redemption_code = FALSE
    AND NOT EXISTS (
        SELECT 1 FROM free_games_sent fgs WHERE fgs.free_game_id = fg.id
    )
--    AND fgm.mapping_category = 'offer'
ORDER BY
    fgp.start_date ASC,
    fgm.slug ASC;
