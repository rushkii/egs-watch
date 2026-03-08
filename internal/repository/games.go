package repository

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/rushkii/egs-watch/internal/epic"
	"github.com/rushkii/egs-watch/pkg"
)

type gamesRepository struct {
	db *sql.DB
}

func NewGamesRepository(db *sql.DB) *gamesRepository {
	return &gamesRepository{db}
}

func (r *gamesRepository) InsertFreeGames(games []epic.FGElement) error {
	sellerMap := make(map[string]int64)
	developerMap := make(map[string]int64)

	var setupWg sync.WaitGroup
	setupErrChan := make(chan error, 2)

	setupWg.Add(2)

	go func() {
		defer setupWg.Done()
		for _, game := range games {
			if _, exists := sellerMap[game.Seller.ID]; exists {
				continue
			}

			log.Println("Inserting publisher:", game.Seller.Name)

			var internalSellerID int64
			sellerQuery := `
				INSERT INTO sellers (epic_seller_id, name)
				VALUES ($1, $2)
				ON CONFLICT (epic_seller_id) DO UPDATE
				SET name = EXCLUDED.name
				RETURNING id;`

			err := r.db.QueryRow(sellerQuery, game.Seller.ID, game.Seller.Name).Scan(&internalSellerID)
			if err != nil {
				setupErrChan <- fmt.Errorf("Failed to insert seller %s: %w", game.Seller.ID, err)
				return
			}
			sellerMap[game.Seller.ID] = internalSellerID
		}
	}()

	go func() {
		defer setupWg.Done()
		for _, game := range games {
			attrs := make([]pkg.KeyValue, len(game.CustomAttributes))
			for i, a := range game.CustomAttributes {
				attrs[i] = pkg.KeyValue{Key: a.Key, Value: a.Value}
			}

			developer := pkg.GetKVFromArray("developerName", attrs)
			if developer == "" {
				developer = game.Seller.Name
			}

			if _, exists := developerMap[developer]; exists {
				continue
			}

			log.Println("Inserting developer:", developer)

			var internalDevID int64
			developerQuery := `
				INSERT INTO developers (name)
				VALUES ($1)
				ON CONFLICT (name) DO UPDATE
				SET name = EXCLUDED.name
				RETURNING id;`

			err := r.db.QueryRow(developerQuery, developer).Scan(&internalDevID)
			if err != nil {
				setupErrChan <- fmt.Errorf("Failed to insert developer %s: %w", developer, err)
				return
			}
			developerMap[developer] = internalDevID
		}
	}()

	setupWg.Wait()
	close(setupErrChan)

	for err := range setupErrChan {
		return err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(games))

	workers := 10
	sem := make(chan struct{}, workers)

	for _, game := range games {
		wg.Add(1)

		go func(g epic.FGElement) {
			log.Printf("Inserting game %s...\n", g.Title)

			defer wg.Done()

			// acquire lock
			sem <- struct{}{}

			// release lock
			defer func() { <-sem }()

			tx, err := r.db.Begin()
			if err != nil {
				errChan <- fmt.Errorf("Failed to begin tx for game %s: %w", g.ID, err)
				return
			}

			defer tx.Rollback()

			attrs := make([]pkg.KeyValue, len(game.CustomAttributes))

			for i, a := range game.CustomAttributes {
				attrs[i] = pkg.KeyValue{Key: a.Key, Value: a.Value}
			}

			developer := pkg.GetKVFromArray("developerName", attrs)

			if developer == "" {
				developer = game.Seller.Name
			}

			var internalGameID int64
			gameQuery := `
                INSERT INTO free_games (
                    epic_game_id, seller_id, developer_id, namespace, title, description,
                    offer_type, status, requires_redemption_code, product_slug, url_slug,
                    discount_price, original_price, voucher, discount, currency_code, decimals,
                    fmt_original_price, fmt_discount_price, fmt_intermediate_price
                ) VALUES (
                    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
                    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
                )
                ON CONFLICT (epic_game_id) DO UPDATE SET
                    seller_id = EXCLUDED.seller_id,
                    developer_id = EXCLUDED.developer_id, -- <-- The missing comma is added here!
                    namespace = EXCLUDED.namespace,
                    title = EXCLUDED.title,
                    description = EXCLUDED.description,
                    offer_type = EXCLUDED.offer_type,
                    status = EXCLUDED.status,
                    requires_redemption_code = EXCLUDED.requires_redemption_code,
                    product_slug = EXCLUDED.product_slug,
                    url_slug = EXCLUDED.url_slug,
                    discount_price = EXCLUDED.discount_price,
                    original_price = EXCLUDED.original_price,
                    voucher = EXCLUDED.voucher,
                    discount = EXCLUDED.discount,
                    currency_code = EXCLUDED.currency_code,
                    decimals = EXCLUDED.decimals,
                    fmt_original_price = EXCLUDED.fmt_original_price,
                    fmt_discount_price = EXCLUDED.fmt_discount_price,
                    fmt_intermediate_price = EXCLUDED.fmt_intermediate_price,
                    created_at = CURRENT_TIMESTAMP
                RETURNING id;`

			tp := g.Price.TotalPrice
			err = tx.QueryRow(gameQuery,
				g.ID, sellerMap[g.Seller.ID], developerMap[developer], g.Namespace, g.Title, g.Description,
				g.OfferType, g.Status, g.IsCodeRedemptionOnly, g.ProductSlug, g.URLSlug,
				tp.DiscountPrice, tp.OriginalPrice, tp.VoucherDiscount, tp.Discount, tp.CurrencyCode, tp.CurrencyInfo.Decimals,
				tp.FmtPrice.OriginalPrice, tp.FmtPrice.DiscountPrice, tp.FmtPrice.IntermediatePrice,
			).Scan(&internalGameID)

			if err != nil {
				errChan <- fmt.Errorf("Failed to insert game %s: %w", g.ID, err)
				return
			}

			// clean up old data
			tablesToClean := []string{"free_game_images", "free_game_mappings", "free_game_promotions"}
			for _, table := range tablesToClean {
				_, err = tx.Exec(fmt.Sprintf(`DELETE FROM %s WHERE free_game_id = $1`, table), internalGameID)
				if err != nil {
					errChan <- fmt.Errorf("Failed to clean %s for game %s: %w", table, g.ID, err)
					return
				}
			}

			// and then below code, replace with the fresh data

			for _, img := range g.KeyImages {
				_, err = tx.Exec(`
					INSERT INTO free_game_images (
						free_game_id, image_type, url) VALUES
					 ($1, $2, $3)`,
					internalGameID, img.Type, img.URL,
				)

				if err != nil {
					errChan <- err
					return
				}
			}

			for _, cat := range g.CatalogNs.Mappings {
				_, err = tx.Exec(`
					INSERT INTO free_game_mappings (
						free_game_id, mapping_category, slug, mapping_type
					) VALUES ($1, 'catalog', $2, $3)`,
					internalGameID, cat.PageSlug, cat.PageType,
				)

				if err != nil {
					errChan <- err
					return
				}
			}

			for _, off := range g.OfferMappings {
				_, err = tx.Exec(`
					INSERT INTO free_game_mappings (
						free_game_id, mapping_category, slug, mapping_type
					) VALUES ($1, 'offer', $2, $3)`,
					internalGameID, off.PageSlug, off.PageType,
				)

				if err != nil {
					errChan <- err
					return
				}
			}

			if g.Promotions != nil {
				for _, offers := range g.Promotions.PromotionalOffers {
					for _, off := range offers.PromotionalOffers {
						_, err = tx.Exec(`
							INSERT INTO free_game_promotions (
								free_game_id, promo_tier, start_date,
								end_date, discount_type, discount_percentage
							) VALUES ($1, 'current', $2, $3, $4, $5)`,
							internalGameID, off.StartDate, off.EndDate, off.DiscountSetting.DiscountType,
							off.DiscountSetting.DiscountPercentage,
						)

						if err != nil {
							errChan <- err
							return
						}
					}
				}
				for _, offers := range g.Promotions.UpcomingPromotionalOffers {
					for _, off := range offers.PromotionalOffers {
						_, err = tx.Exec(`
							INSERT INTO free_game_promotions (
								free_game_id, promo_tier, start_date,
								end_date, discount_type, discount_percentage
							) VALUES ($1, 'upcoming', $2, $3, $4, $5)`,
							internalGameID, off.StartDate, off.EndDate, off.DiscountSetting.DiscountType,
							off.DiscountSetting.DiscountPercentage,
						)

						if err != nil {
							errChan <- err
							return
						}
					}
				}
			}

			if err := tx.Commit(); err != nil {
				errChan <- fmt.Errorf("Failed to commit game %s: %w", g.Title, err)
			}

			log.Printf("%s has been inserted...\n", g.Title)
		}(game)
	}

	wg.Wait()
	close(errChan)

	var finalErr error
	for err := range errChan {
		log.Printf("Worker error: %v\n", err)
		if finalErr == nil {
			finalErr = err
		}
	}

	return finalErr
}

func (r *gamesRepository) InsertUpdateSent(freeGameID string) error {
	query := `
		INSERT INTO free_games_sent (free_game_id)
		VALUES ($1)
		ON CONFLICT (free_game_id) DO UPDATE
		SET free_game_id = EXCLUDED.free_game_id;`

	_, err := r.db.Exec(query, freeGameID)
	if err != nil {
		return fmt.Errorf("Failed to insert free games update sent %s: %w", freeGameID, err)
	}

	return nil
}

func (r *gamesRepository) SelectFreeGames() ([]FreeGamesFromDB, error) {
	query := `
		SELECT
			fg.id,
			fg.epic_game_id game_id,
			fg.namespace,
			fg.title,
			fg.description,
			fg.offer_type,
			fg.status,
			fg.requires_redemption_code,
			s."name" seller,
			d."name" developer,
			fgm.slug,
			json_agg(
				json_build_object('type', image_type, 'url', url)
			) AS images,
			fgp.promo_tier period,
			fg.fmt_original_price,
			fg.fmt_discount_price,
			fg.fmt_intermediate_price,
			fgp.start_date,
			fgp.end_date
		FROM
			free_games fg
		LEFT JOIN sellers s ON
			s.id = fg.seller_id
		LEFT JOIN developers d ON
			d.id = fg.developer_id
		LEFT JOIN free_game_images fgi ON
			fgi.free_game_id = fg.id
		LEFT JOIN free_game_mappings fgm ON
			fgm.free_game_id = fg.id
		JOIN free_game_promotions fgp ON
			fgp.free_game_id = fg.id
		LEFT JOIN free_games_sent fgs ON
			fgs.free_game_id = fg.id
		WHERE
			fg.offer_type = 'BASE_GAME'
			AND fg.requires_redemption_code = FALSE
			AND fgi.image_type IN (
				'OfferImageWide', 'Thumbnail'
			)
			AND fgs.free_game_id IS NULL
		--    AND fgm.mapping_category = 'offer'
		GROUP BY
			fg.id,
			fg.epic_game_id,
			fg.namespace,
			fg.title,
			fg.description,
			fg.offer_type,
			fg.status,
			fg.requires_redemption_code,
			s."name",
			d."name",
			fgm.slug,
			fgp.promo_tier,
			fg.fmt_original_price,
			fg.fmt_discount_price,
			fg.fmt_intermediate_price,
			fgp.start_date,
			fgp.end_date
		ORDER BY fgp.start_date ASC, fgm.slug ASC;
	`

	results := []FreeGamesFromDB{}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Error querying albums: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		fg := FreeGamesFromDB{}
		err := rows.Scan(
			&fg.ID,
			&fg.GameID,
			&fg.Namespace,
			&fg.Title,
			&fg.Description,
			&fg.OfferType,
			&fg.Status,
			&fg.RequiresRedemptionCode,
			&fg.Publisher,
			&fg.Developer,
			&fg.Slug,
			&fg.Images,
			&fg.Period,
			&fg.FmtOriginalPrice,
			&fg.FmtDiscountPrice,
			&fg.FmtIntermediatePrice,
			&fg.StartDate,
			&fg.EndDate,
		)

		if err != nil {
			return nil, fmt.Errorf("Error scanning album row: %w", err)
		}

		results = append(results, fg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error during rows iteration: %w", err)
	}

	return results, nil
}

func (r *gamesRepository) GetFreeGamesFromDB() (FreeGamesFilteredFromDB, error) {
	filtered := FreeGamesFilteredFromDB{}

	freeGames, err := r.SelectFreeGames()
	if err != nil {
		return filtered, fmt.Errorf("Failed to get free games from db: %w", err)
	}

	filtered.All = freeGames

	for _, fg := range freeGames {
		if fg.Period == "current" {
			filtered.Now = append(filtered.Now, fg)
		} else if fg.Period == "upcoming" {
			filtered.Upcoming = append(filtered.Upcoming, fg)
		}
	}

	return filtered, nil
}

func (r *gamesRepository) CleanupFreeGames() (int64, error) {
	query := `DELETE FROM free_games WHERE created_at < NOW() - INTERVAL '2 weeks';`

	result, err := r.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("Error while clean up free games root")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("Error while getting affected rows of free games root")
	}

	query = `DELETE FROM free_games_sent WHERE created_at < NOW() - INTERVAL '2 weeks';`

	result, err = r.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("Error while clean up free games sent")
	}

	rows, err = result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("Error while getting affected rows of free games sent")
	}

	return rows, nil
}
