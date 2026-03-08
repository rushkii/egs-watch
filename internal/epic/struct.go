package epic

import "time"

type FreeGamesFiltered struct {
	All      []FGElement
	Now      []FGElement
	Upcoming []FGElement
}

type FreeGamesResponse struct {
	Data FGData `json:"data"`
}

type FGData struct {
	Catalog FGCatalog `json:"Catalog"`
}

type FGCatalog struct {
	SearchStore FGSearchStore `json:"searchStore"`
}

type FGSearchStore struct {
	Elements []FGElement `json:"elements"`
	Paging   FGPaging    `json:"paging"`
}

type FGPaging struct {
	Count int `json:"count"`
	Total int `json:"total"`
}

type FGElement struct {
	ID                   string              `json:"id"`
	Namespace            string              `json:"namespace"`
	Title                string              `json:"title"`
	Description          string              `json:"description"`
	EffectiveDate        time.Time           `json:"effectiveDate"`
	OfferType            string              `json:"offerType"`
	ExpiryDate           interface{}         `json:"expiryDate"`
	ViewableDate         time.Time           `json:"viewableDate"`
	Status               string              `json:"status"`
	IsCodeRedemptionOnly bool                `json:"isCodeRedemptionOnly"`
	KeyImages            []FGKeyImage        `json:"keyImages"`
	Seller               FGSeller            `json:"seller"`
	ProductSlug          string              `json:"productSlug"`
	URLSlug              string              `json:"urlSlug"`
	Items                []FGItem            `json:"items"`
	CustomAttributes     []FGCustomAttribute `json:"customAttributes"`
	Categories           []FGCategory        `json:"categories"`
	Tags                 []FGTag             `json:"tags"`
	CatalogNs            FGCatalogNs         `json:"catalogNs"`
	OfferMappings        []FGMapping         `json:"offerMappings"`
	Price                FGPrice             `json:"price"`
	Promotions           *FGPromotions       `json:"promotions"`
}

type FGKeyImage struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type FGSeller struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FGItem struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
}

type FGCustomAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type FGCategory struct {
	Path string `json:"path"`
}

type FGTag struct {
	ID string `json:"id"`
}

type FGCatalogNs struct {
	Mappings []FGMapping `json:"mappings"`
}

type FGMapping struct {
	PageSlug string `json:"pageSlug"`
	PageType string `json:"pageType"`
}

type FGPrice struct {
	TotalPrice FGTotalPrice  `json:"totalPrice"`
	LineOffers []FGLineOffer `json:"lineOffers"`
}

type FGTotalPrice struct {
	DiscountPrice   int         `json:"discountPrice"`
	OriginalPrice   int         `json:"originalPrice"`
	VoucherDiscount int         `json:"voucherDiscount"`
	Discount        int         `json:"discount"`
	CurrencyCode    string      `json:"currencyCode"`
	CurrencyInfo    FGCurrency  `json:"currencyInfo"`
	FmtPrice        FGFmtPrice  `json:"fmtPrice"`
	DualPrice       interface{} `json:"dualPrice"`
}

type FGCurrency struct {
	Decimals int `json:"decimals"`
}

type FGFmtPrice struct {
	OriginalPrice     string `json:"originalPrice"`
	DiscountPrice     string `json:"discountPrice"`
	IntermediatePrice string `json:"intermediatePrice"`
}

type FGLineOffer struct {
	AppliedRules []interface{} `json:"appliedRules"`
}

type FGPromotions struct {
	PromotionalOffers         []FGPromoOfferList `json:"promotionalOffers"`
	UpcomingPromotionalOffers []FGPromoOfferList `json:"upcomingPromotionalOffers"`
}

type FGPromoOfferList struct {
	PromotionalOffers []FGPromoOffer `json:"promotionalOffers"`
}

type FGPromoOffer struct {
	StartDate       time.Time         `json:"startDate"`
	EndDate         time.Time         `json:"endDate"`
	DiscountSetting FGDiscountSetting `json:"discountSetting"`
}

type FGDiscountSetting struct {
	DiscountType       string `json:"discountType"`
	DiscountPercentage int    `json:"discountPercentage"`
}
