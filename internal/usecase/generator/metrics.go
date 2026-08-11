package generator

import (
	"sort"
	"time"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
)

// UserMetrics is retained as the small input contract for the existing
// achievement rules.
type UserMetrics struct {
	TotalViews     int
	TotalMessages  int
	TotalFavorites int
	TotalPurchases int
	TotalSales     int
	TopCategories  []domain.CategoryStat
	ActivityDays   int
}

// ProfileMetrics is calculated from the profile payload that the backend
// already stores. Dated buyer/listing/sale/review facts are filtered by Year.
// Values without a date (chats, likes and legacy listings) are treated as a
// profile snapshot and the generated card copy avoids calling them annual.
type ProfileMetrics struct {
	Year                    int
	Buyer                   domain.BuyerRecapSummary
	Seller                  domain.SellerRecapSummary
	Combined                domain.CombinedRecapSummary
	TopCategories           []domain.CategoryStat
	CategoryStats           []domain.CategoryStat
	ActivityDays            int
	Monthly                 [12]MonthlyActivity
	StarListingViews        int
	SellerListingsAreAnnual bool
}

type MonthlyActivity struct {
	Views     int
	Favorites int
	Purchases int
	Listings  int
	Sales     int
}

type categoryMetric struct {
	Name         string
	ViewedAds    int
	Views        int
	Favorites    int
	Purchases    int
	Listings     int
	ListingViews int
	Sales        int
}

func calculateProfileMetrics(user *domain.User, year int) ProfileMetrics {
	metrics := ProfileMetrics{
		Year:                    year,
		TopCategories:           []domain.CategoryStat{},
		SellerListingsAreAnnual: true,
	}
	categories := make(map[string]*categoryMetric)
	activeDays := make(map[string]struct{})
	var largestPurchase *domain.RecapItem
	var starListing *domain.RecapItem
	starListingViews := -1
	ratingTotal := 0

	for _, view := range user.Views {
		category := categoryFor(categories, view.Category)
		activity := viewedActivityForYear(view, year)
		addViewedMonthlyActivity(&metrics.Monthly, view, year)

		if activity.WatchCount > 0 {
			metrics.Buyer.ViewedAdsCount++
			metrics.Buyer.TotalViews += activity.WatchCount
			category.ViewedAds++
			category.Views += activity.WatchCount
		}
		if activity.Liked {
			metrics.Buyer.FavoritesCount++
			category.Favorites++
		}
		if activity.Bought {
			metrics.Buyer.PurchasesCount++
			category.Purchases++
			if activity.UsedAvitoDelivery {
				metrics.Buyer.AvitoDeliveryPurchases++
			}
			item := recapItem(view.Ad)
			if largestPurchase == nil || item.Price > largestPurchase.Price ||
				(item.Price == largestPurchase.Price && item.Title < largestPurchase.Title) {
				selected := item
				largestPurchase = &selected
			}
		}
		for _, activityTime := range activity.Times {
			addActiveDay(activeDays, activityTime)
		}
	}

	metrics.Buyer.ChatsCount = user.ChatsCount
	metrics.Buyer.LargestPurchase = largestPurchase

	for _, ad := range user.OwnAds {
		category := categoryFor(categories, ad.Category)
		listingInYear := inYear(ad.PublishedAt, year)
		if ad.PublishedAt.IsZero() {
			// Existing JSONB rows created before publishedAt was introduced are
			// retained as a snapshot until the profile is updated.
			listingInYear = true
			metrics.SellerListingsAreAnnual = false
		}
		if listingInYear {
			metrics.Seller.ListingsCount++
			metrics.Seller.ListingViews += ad.ViewCount
			metrics.Seller.FavoritesReceived += ad.FavoritesCount
			metrics.Seller.ContactsReceived += ad.ContactsCount
			if !ad.IsArchived && !ad.IsSold {
				metrics.Seller.ActiveListings++
			}
			if ad.IsArchived {
				metrics.Seller.ArchivedListings++
			}
			category.Listings++
			category.ListingViews += ad.ViewCount

			item := recapItem(ad.Ad)
			if starListing == nil || ad.ViewCount > starListingViews ||
				(ad.ViewCount == starListingViews && item.Title < starListing.Title) {
				selected := item
				starListing = &selected
				starListingViews = ad.ViewCount
			}
			if !ad.PublishedAt.IsZero() {
				addActiveDay(activeDays, ad.PublishedAt)
				metrics.Monthly[int(ad.PublishedAt.UTC().Month())-1].Listings++
			}
		}

		if ad.IsSold && ad.SoldAt != nil && inYear(*ad.SoldAt, year) {
			metrics.Seller.SalesCount++
			category.Sales++
			addActiveDay(activeDays, *ad.SoldAt)
			metrics.Monthly[int(ad.SoldAt.UTC().Month())-1].Sales++
		}
		if ad.Review != nil && inYear(ad.Review.CreatedAt, year) {
			metrics.Seller.ReviewsCount++
			ratingTotal += ad.Review.Rating
			addActiveDay(activeDays, ad.Review.CreatedAt)
		}
	}

	metrics.Seller.LikesReceived = user.Likes
	metrics.Seller.StarListing = starListing
	metrics.StarListingViews = starListingViews
	if metrics.Seller.ReviewsCount > 0 {
		average := float64(ratingTotal) / float64(metrics.Seller.ReviewsCount)
		metrics.Seller.AverageRating = &average
	}

	metrics.Buyer.HasData = metrics.Buyer.ViewedAdsCount > 0 ||
		metrics.Buyer.FavoritesCount > 0 || metrics.Buyer.PurchasesCount > 0 ||
		metrics.Buyer.ChatsCount > 0
	metrics.Seller.HasData = metrics.Seller.ListingsCount > 0 || metrics.Seller.SalesCount > 0 ||
		metrics.Seller.LikesReceived > 0 || metrics.Seller.FavoritesReceived > 0 ||
		metrics.Seller.ContactsReceived > 0
	metrics.Buyer.MainCategory = selectCategory(categories, hasBuyerSignals, buyerCategoryLess)
	metrics.Seller.MainCategory = selectCategory(categories, hasSellerSignals, sellerCategoryLess)
	metrics.Combined = domain.CombinedRecapSummary{
		HasBuyerData:  metrics.Buyer.HasData,
		HasSellerData: metrics.Seller.HasData,
		Categories:    countUsedCategories(categories),
		Deals:         metrics.Buyer.PurchasesCount + metrics.Seller.SalesCount,
		MainCategory:  selectCategory(categories, hasAnySignals, combinedCategoryLess),
	}
	metrics.CategoryStats = categoryStats(categories)
	metrics.TopCategories = append([]domain.CategoryStat(nil), metrics.CategoryStats...)
	if len(metrics.TopCategories) > 3 {
		metrics.TopCategories = metrics.TopCategories[:3]
	}
	metrics.ActivityDays = len(activeDays)
	return metrics
}

func (metrics ProfileMetrics) achievementMetrics() *UserMetrics {
	return &UserMetrics{
		TotalViews: metrics.Buyer.TotalViews,
		// chatsCount has no timestamp in the current profile contract, so it
		// must not unlock an achievement that claims to be annual.
		TotalMessages:  0,
		TotalFavorites: metrics.Buyer.FavoritesCount,
		TotalPurchases: metrics.Buyer.PurchasesCount,
		TotalSales:     metrics.Seller.SalesCount,
		TopCategories:  metrics.TopCategories,
		ActivityDays:   metrics.ActivityDays,
	}
}

func categoryFor(categories map[string]*categoryMetric, name string) *categoryMetric {
	metric, exists := categories[name]
	if !exists {
		metric = &categoryMetric{Name: name}
		categories[name] = metric
	}
	return metric
}

func countUsedCategories(categories map[string]*categoryMetric) int {
	count := 0
	for _, category := range categories {
		if categorySignals(category) > 0 {
			count++
		}
	}
	return count
}

type categoryLess func(candidate, selected *categoryMetric) bool
type categoryFilter func(category *categoryMetric) bool

func selectCategory(categories map[string]*categoryMetric, used categoryFilter, less categoryLess) string {
	var selected *categoryMetric
	for _, candidate := range categories {
		if !used(candidate) {
			continue
		}
		if selected == nil || less(candidate, selected) {
			selected = candidate
		}
	}
	if selected == nil {
		return ""
	}
	return selected.Name
}

func hasBuyerSignals(category *categoryMetric) bool {
	return category.ViewedAds+category.Favorites+category.Purchases > 0
}

func hasSellerSignals(category *categoryMetric) bool {
	return category.Listings+category.Sales > 0
}

func hasAnySignals(category *categoryMetric) bool {
	return categorySignals(category) > 0
}

func buyerCategoryLess(candidate, selected *categoryMetric) bool {
	return compareCategory(
		candidate,
		selected,
		[]func(*categoryMetric) int{
			func(value *categoryMetric) int { return value.Purchases },
			func(value *categoryMetric) int { return value.Favorites },
			func(value *categoryMetric) int { return value.Views },
		},
	)
}

func sellerCategoryLess(candidate, selected *categoryMetric) bool {
	return compareCategory(
		candidate,
		selected,
		[]func(*categoryMetric) int{
			func(value *categoryMetric) int { return value.Sales },
			func(value *categoryMetric) int { return value.ListingViews },
			func(value *categoryMetric) int { return value.Listings },
		},
	)
}

func combinedCategoryLess(candidate, selected *categoryMetric) bool {
	return compareCategory(
		candidate,
		selected,
		[]func(*categoryMetric) int{
			func(value *categoryMetric) int { return value.Purchases },
			func(value *categoryMetric) int { return value.Favorites },
			func(value *categoryMetric) int { return value.Sales },
			func(value *categoryMetric) int { return value.Views },
			func(value *categoryMetric) int { return value.ListingViews },
		},
	)
}

func compareCategory(candidate, selected *categoryMetric, selectors []func(*categoryMetric) int) bool {
	for _, value := range selectors {
		candidateValue := value(candidate)
		selectedValue := value(selected)
		if candidateValue != selectedValue {
			return candidateValue > selectedValue
		}
	}
	return candidate.Name < selected.Name
}

func categoryStats(categories map[string]*categoryMetric) []domain.CategoryStat {
	ordered := make([]*categoryMetric, 0, len(categories))
	for _, category := range categories {
		if categorySignals(category) > 0 {
			ordered = append(ordered, category)
		}
	}
	sort.Slice(ordered, func(i, j int) bool {
		return combinedCategoryLess(ordered[i], ordered[j])
	})
	result := make([]domain.CategoryStat, len(ordered))
	for index, category := range ordered {
		// Count is a legacy field. It now represents explainable item-level
		// signals, not a weighted interest score.
		result[index] = domain.CategoryStat{Category: category.Name, Count: categorySignals(category)}
	}
	return result
}

func addViewedMonthlyActivity(monthly *[12]MonthlyActivity, view domain.ViewedAd, year int) {
	if len(view.ViewedAt) > 0 {
		for _, event := range view.ViewedAt {
			if !inYear(event.Time, year) {
				continue
			}
			activity := &monthly[int(event.Time.UTC().Month())-1]
			switch event.Type {
			case domain.ViewedAdEventWatch:
				activity.Views++
			case domain.ViewedAdEventLike:
				activity.Favorites++
			case domain.ViewedAdEventBuy:
				activity.Purchases++
			}
		}
		return
	}

	// Legacy profiles contain only aggregate counters. They are assigned to the
	// latest known month so the chart remains consistent with recap totals.
	if inYear(view.LastViewedAt, year) && view.ViewCount > 0 {
		monthly[int(view.LastViewedAt.UTC().Month())-1].Views += view.ViewCount
	}
	if view.IsFavorite && view.FavoritedAt != nil && inYear(*view.FavoritedAt, year) {
		monthly[int(view.FavoritedAt.UTC().Month())-1].Favorites++
	}
	if view.IsPurchased && view.PurchasedAt != nil && inYear(*view.PurchasedAt, year) {
		monthly[int(view.PurchasedAt.UTC().Month())-1].Purchases++
	}
}

func categorySignals(category *categoryMetric) int {
	return category.ViewedAds + category.Favorites + category.Purchases +
		category.Listings + category.Sales
}

func recapItem(ad domain.Ad) domain.RecapItem {
	return domain.RecapItem{
		AdID:        ad.AdID,
		Title:       ad.Title,
		Category:    ad.Category,
		Subcategory: ad.Subcategory,
		ImageURL:    ad.ImageURL,
		Price:       ad.Price,
	}
}

type yearlyViewedActivity struct {
	WatchCount        int
	Liked             bool
	Bought            bool
	UsedAvitoDelivery bool
	Times             []time.Time
}

func viewedActivityForYear(view domain.ViewedAd, year int) yearlyViewedActivity {
	activity := yearlyViewedActivity{Times: make([]time.Time, 0)}
	if len(view.ViewedAt) > 0 {
		for _, event := range view.ViewedAt {
			if !inYear(event.Time, year) {
				continue
			}
			activity.Times = append(activity.Times, event.Time)
			switch event.Type {
			case domain.ViewedAdEventWatch:
				activity.WatchCount++
			case domain.ViewedAdEventLike:
				activity.Liked = true
			case domain.ViewedAdEventBuy:
				activity.Bought = true
				activity.UsedAvitoDelivery = event.UseAvitoDelivery != nil && *event.UseAvitoDelivery
			}
		}
		return activity
	}

	// Legacy fallback for profiles saved before viewedAt[] was introduced.
	if inYear(view.LastViewedAt, year) {
		activity.WatchCount = view.ViewCount
		activity.Times = append(activity.Times, view.LastViewedAt)
	}
	if view.IsFavorite && view.FavoritedAt != nil && inYear(*view.FavoritedAt, year) {
		activity.Liked = true
		activity.Times = append(activity.Times, *view.FavoritedAt)
	}
	if view.IsPurchased && view.PurchasedAt != nil && inYear(*view.PurchasedAt, year) {
		activity.Bought = true
		activity.Times = append(activity.Times, *view.PurchasedAt)
	}
	return activity
}

func inYear(value time.Time, year int) bool {
	return !value.IsZero() && value.UTC().Year() == year
}

func addActiveDay(days map[string]struct{}, value time.Time) {
	days[value.UTC().Format("2006-01-02")] = struct{}{}
}
