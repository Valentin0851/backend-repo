package domain

import "time"

type ProfileSummary struct {
	ID         string
	Name       string
	JoinedAt   time.Time
	AvatarURL  string
	Stats      ProfileStats
	Highlights ProfileHighlights
	Purchases  []ProfilePurchase
	Sales      []ProfileSale
}

type ProfileStats struct {
	Likes          int
	ChatsCount     int
	PurchasesCount int
	SalesCount     int
	TotalViewCount int
	TotalSpent     int64
	TotalEarned    int64
	ReviewsCount   int
	AverageRating  *float64
}

type ProfileHighlights struct {
	FavoriteCategory       *string
	MostExpensivePurchase  *ProfilePurchase
	LeastExpensivePurchase *ProfilePurchase
	MostExpensiveSale      *ProfileSale
	LeastExpensiveSale     *ProfileSale
}

type ProfilePurchase struct {
	AdID        string
	Title       string
	Category    string
	Subcategory string
	ImageURL    string
	Price       int64
	PurchasedAt time.Time
}

type ProfileSale struct {
	AdID        string
	Title       string
	Category    string
	Subcategory string
	ImageURL    string
	Price       int64
	SoldAt      time.Time
	ViewCount   int
	Review      *Review
}

func SummarizeProfile(user *User) ProfileSummary {
	purchases := make([]ProfilePurchase, 0)
	sales := make([]ProfileSale, 0)
	totalViewCount := 0
	var totalSpent int64
	var totalEarned int64
	ratingsTotal := 0
	reviewsCount := 0

	categoryCounts := make(map[string]int)
	categoryOrder := make([]string, 0)

	for _, view := range user.Views {
		totalViewCount += view.ViewCount
		if !view.IsPurchased || view.PurchasedAt == nil {
			continue
		}
		purchase := ProfilePurchase{
			AdID:        view.AdID,
			Title:       view.Title,
			Category:    view.Category,
			Subcategory: view.Subcategory,
			ImageURL:    view.ImageURL,
			Price:       view.Price,
			PurchasedAt: *view.PurchasedAt,
		}
		purchases = append(purchases, purchase)
		totalSpent += purchase.Price
		if _, exists := categoryCounts[purchase.Category]; !exists {
			categoryOrder = append(categoryOrder, purchase.Category)
		}
		categoryCounts[purchase.Category]++
	}

	for _, ad := range user.OwnAds {
		if !ad.IsSold || ad.SoldAt == nil {
			continue
		}
		sale := ProfileSale{
			AdID:        ad.AdID,
			Title:       ad.Title,
			Category:    ad.Category,
			Subcategory: ad.Subcategory,
			ImageURL:    ad.ImageURL,
			Price:       ad.Price,
			SoldAt:      *ad.SoldAt,
			ViewCount:   ad.ViewCount,
			Review:      ad.Review,
		}
		sales = append(sales, sale)
		totalEarned += sale.Price
		if sale.Review != nil {
			reviewsCount++
			ratingsTotal += sale.Review.Rating
		}
	}

	var averageRating *float64
	if reviewsCount > 0 {
		value := float64(ratingsTotal) / float64(reviewsCount)
		averageRating = &value
	}

	return ProfileSummary{
		ID:        user.ID.String(),
		Name:      user.Name,
		JoinedAt:  user.RegisteredAt,
		AvatarURL: user.Avatar,
		Stats: ProfileStats{
			Likes:          user.Likes,
			ChatsCount:     user.ChatsCount,
			PurchasesCount: len(purchases),
			SalesCount:     len(sales),
			TotalViewCount: totalViewCount,
			TotalSpent:     totalSpent,
			TotalEarned:    totalEarned,
			ReviewsCount:   reviewsCount,
			AverageRating:  averageRating,
		},
		Highlights: ProfileHighlights{
			FavoriteCategory:       favoriteCategory(categoryOrder, categoryCounts),
			MostExpensivePurchase:  highestPurchase(purchases),
			LeastExpensivePurchase: lowestPurchase(purchases),
			MostExpensiveSale:      highestSale(sales),
			LeastExpensiveSale:     lowestSale(sales),
		},
		Purchases: purchases,
		Sales:     sales,
	}
}

func favoriteCategory(order []string, counts map[string]int) *string {
	if len(order) == 0 {
		return nil
	}
	selected := order[0]
	for _, category := range order[1:] {
		if counts[category] > counts[selected] {
			selected = category
		}
	}
	return &selected
}

func highestPurchase(items []ProfilePurchase) *ProfilePurchase {
	return selectPurchase(items, func(candidate, selected int64) bool { return candidate > selected })
}

func lowestPurchase(items []ProfilePurchase) *ProfilePurchase {
	return selectPurchase(items, func(candidate, selected int64) bool { return candidate < selected })
}

func selectPurchase(items []ProfilePurchase, replace func(int64, int64) bool) *ProfilePurchase {
	if len(items) == 0 {
		return nil
	}
	selected := items[0]
	for _, item := range items[1:] {
		if replace(item.Price, selected.Price) {
			selected = item
		}
	}
	return &selected
}

func highestSale(items []ProfileSale) *ProfileSale {
	return selectSale(items, func(candidate, selected int64) bool { return candidate > selected })
}

func lowestSale(items []ProfileSale) *ProfileSale {
	return selectSale(items, func(candidate, selected int64) bool { return candidate < selected })
}

func selectSale(items []ProfileSale, replace func(int64, int64) bool) *ProfileSale {
	if len(items) == 0 {
		return nil
	}
	selected := items[0]
	for _, item := range items[1:] {
		if replace(item.Price, selected.Price) {
			selected = item
		}
	}
	return &selected
}
