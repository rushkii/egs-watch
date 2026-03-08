package repository

import "time"

type ImageArray []FGDBKeyImage

type FreeGamesFilteredFromDB struct {
	All      []FreeGamesFromDB
	Now      []FreeGamesFromDB
	Upcoming []FreeGamesFromDB
}

type FreeGamesFromDB struct {
	ID                     string     `json:"id"`
	GameID                 string     `json:"game_id"`
	Namespace              string     `json:"namespace"`
	Title                  string     `json:"title"`
	Description            string     `json:"description"`
	OfferType              string     `json:"offer_type"`
	Status                 string     `json:"status"`
	RequiresRedemptionCode bool       `json:"requires_redemption_code"`
	Publisher              string     `json:"seller"`
	Developer              string     `json:"developer"`
	Slug                   string     `json:"slug"`
	Images                 ImageArray `json:"images"`
	Period                 string     `json:"period"`
	FmtOriginalPrice       string     `json:"fmt_original_price"`
	FmtDiscountPrice       string     `json:"fmt_discount_price"`
	FmtIntermediatePrice   string     `json:"fmt_intermediate_price"`
	StartDate              time.Time  `json:"start_date"`
	EndDate                time.Time  `json:"end_date"`
}

type FGDBKeyImage struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// type FGDBSeller struct {
// 	ID   string `json:"epic_game_id"`
// 	Name string `json:"name"`
// }

// type FGDBMapping struct {
// 	Slug string `json:"slug"`
// 	Type string `json:"type"`
// }

// type FGDBFmtPrice struct {
// 	OriginalPrice     string `json:"original_price"`
// 	DiscountPrice     string `json:"discount_price"`
// 	IntermediatePrice string `json:"intermediate_price"`
// }

// type FGDBTotalPrice struct {
// 	DiscountPrice int          `json:"discount_price"`
// 	OriginalPrice int          `json:"original_price"`
// 	Vouche        int          `json:"voucher"`
// 	Discount      int          `json:"discount"`
// 	CurrencyCode  string       `json:"currency_code"`
// 	CurrencyInfo  int          `json:"decimals"`
// 	FmtPrice      FGDBFmtPrice `json:"formatted_price"`
// }

// type FGDBPromotions struct {
// 	Current  []FGDBPromoOffer `json:"current"`
// 	Upcoming []FGDBPromoOffer `json:"upcoming"`
// }

// type FGDBPromoOffer struct {
// 	StartDate time.Time           `json:"start_date"`
// 	EndDate   time.Time           `json:"end_date"`
// 	Setting   FGDBDiscountSetting `json:"setting"`
// }
// type FGDBDiscountSetting struct {
// 	Type       string `json:"type"`
// 	Percentage int    `json:"percentage"`
// }
