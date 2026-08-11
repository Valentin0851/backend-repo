package generator

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
)

func buildRecapSummary(metrics ProfileMetrics) domain.RecapSummary {
	headline := "Ваш год на Авито"
	description := "Данных пока немного, но это уже первая точка вашего следующего года."

	switch {
	case metrics.Buyer.PurchasesCount > 0 && metrics.Seller.SalesCount > 0:
		headline = "Вы были по обе стороны Авито"
		description = fmt.Sprintf(
			"В %d году у вас — %s и %s. Здесь одинаково важны обе стороны.",
			metrics.Year,
			formatCount(metrics.Buyer.PurchasesCount, "покупка", "покупки", "покупок"),
			formatCount(metrics.Seller.SalesCount, "продажа", "продажи", "продаж"),
		)
	case metrics.Buyer.HasData && metrics.Seller.HasData:
		headline = "Вы искали новое и делились своим"
		description = "В профиле есть и интерес к покупкам, и собственные объявления — итоги показывают обе стороны."
	case metrics.Buyer.PurchasesCount > 0 || metrics.Buyer.HasData:
		headline = "Год новых находок"
		description = fmt.Sprintf("В %d году вы исследовали предложения и находили то, что действительно подходит.", metrics.Year)
	case metrics.Seller.SalesCount > 0 || metrics.Seller.HasData:
		headline = "Год второй жизни вещей"
		description = fmt.Sprintf("В %d году ваши объявления находили внимание и новых владельцев.", metrics.Year)
	}

	return domain.RecapSummary{
		Headline:    headline,
		Description: description,
		Buyer:       metrics.Buyer,
		Seller:      metrics.Seller,
		Combined:    metrics.Combined,
	}
}

func buildRecapCards(metrics ProfileMetrics, summary domain.RecapSummary, joinedAt time.Time) []domain.RecapCard {
	overview := domain.RecapCard{
		ID:           "year_overview",
		Kind:         "overview",
		Eyebrow:      fmt.Sprintf("Итоги %d", metrics.Year),
		Title:        summary.Headline,
		Description:  summary.Description,
		Value:        strconv.Itoa(metrics.Year),
		Shareable:    true,
		Reason:       "Обязательная первая карточка, которая объединяет активность покупателя и продавца.",
		Presentation: presentation("hero", "avito-purple", "sparkles"),
	}

	general := make([]domain.RecapCard, 0, 3)
	if metrics.Combined.MainCategory != "" {
		general = append(general, domain.RecapCard{
			ID:           "main_interest",
			Kind:         "interest",
			Eyebrow:      "Главный интерес",
			Title:        metrics.Combined.MainCategory,
			Description:  "Категория выбрана последовательно: сначала по покупкам, затем по избранному, продажам и просмотрам.",
			Value:        metrics.Combined.MainCategory,
			Shareable:    true,
			Reason:       "В категории есть подтверждённая активность пользователя.",
			Presentation: presentation("statistic", "avito-blue", "compass"),
			CTA:          categoryCTA(metrics.Combined.MainCategory),
		})
	}

	buyer := buyerCards(metrics)
	seller := sellerCards(metrics)
	combined := combinedCards(metrics, joinedAt)
	categoryChart := buildCategoryMixCard(metrics)
	activityChart := buildActivityRhythmCard(metrics)
	selected := []domain.RecapCard{overview}
	chartCandidates := make([]domain.RecapCard, 0, 2)
	if categoryChart != nil {
		chartCandidates = append(chartCandidates, *categoryChart)
		// The donut already communicates the main interest, so the equivalent
		// text card would only repeat the same insight.
		general = []domain.RecapCard{}
	}
	if activityChart != nil {
		chartCandidates = append(chartCandidates, *activityChart)
	}
	selected = appendUnique(selected, general, 1)
	selected = appendUnique(selected, chartCandidates, 2)
	selected = appendUnique(selected, buyer, 2)
	selected = appendUnique(selected, seller, 2)
	selected = appendUnique(selected, combined, 1)

	// Fill up to eight content cards from still-unused candidates. With the
	// finale this gives the frontend a compact sequence of at most nine cards.
	allCandidates := make([]domain.RecapCard, 0, len(general)+len(chartCandidates)+len(buyer)+len(seller)+len(combined))
	allCandidates = append(allCandidates, general...)
	allCandidates = append(allCandidates, chartCandidates...)
	allCandidates = append(allCandidates, buyer...)
	allCandidates = append(allCandidates, seller...)
	allCandidates = append(allCandidates, combined...)
	selected = appendUnique(selected, allCandidates, 8-len(selected))
	selected = append(selected, finalCard(metrics))
	return selected
}

func buyerCards(metrics ProfileMetrics) []domain.RecapCard {
	if !metrics.Buyer.HasData {
		return []domain.RecapCard{}
	}
	cards := make([]domain.RecapCard, 0, 5)
	if metrics.Buyer.ViewedAdsCount > 0 {
		cards = append(cards, domain.RecapCard{
			ID:           "viewed_findings",
			Kind:         "buyer",
			Eyebrow:      "Ваш поиск",
			Title:        "Столько находок вы изучили",
			Description:  fmt.Sprintf("К этим объявлениям вы обратились %s.", formatCount(metrics.Buyer.TotalViews, "раз", "раза", "раз")),
			Value:        formatCount(metrics.Buyer.ViewedAdsCount, "объявление", "объявления", "объявлений"),
			Shareable:    true,
			Reason:       "Есть watch-события в выбранном году; для старых профилей используется lastViewedAt.",
			Presentation: presentation("statistic", "avito-blue", "eye"),
			CTA:          &domain.RecapCardCTA{Label: "Посмотреть свежие варианты", Action: "open_recommendations"},
		})
	}
	if metrics.Buyer.AvitoDeliveryPurchases > 0 {
		cards = append(cards, domain.RecapCard{
			ID:           "avito_delivery",
			Kind:         "buyer",
			Eyebrow:      "Безопасные сделки",
			Title:        "Авито Доставка была рядом",
			Description:  "Считаем покупки, у которых buy-событие отмечено useAvitoDelivery=true.",
			Value:        formatCount(metrics.Buyer.AvitoDeliveryPurchases, "покупка", "покупки", "покупок"),
			Shareable:    true,
			Reason:       "Есть подтверждённые покупки с использованием Авито Доставки.",
			Presentation: presentation("statistic", "avito-green", "delivery"),
		})
	}
	if metrics.Buyer.FavoritesCount > 0 {
		cards = append(cards, domain.RecapCard{
			ID:           "favorites",
			Kind:         "buyer",
			Eyebrow:      "Избранное",
			Title:        "Вы сохранили на потом",
			Description:  "Иногда лучший выбор начинается с паузы и возможности спокойно сравнить варианты.",
			Value:        formatCount(metrics.Buyer.FavoritesCount, "находка", "находки", "находок"),
			Shareable:    true,
			Reason:       "Есть избранные объявления с favoritedAt в выбранном году.",
			Presentation: presentation("statistic", "avito-red", "heart"),
			CTA:          &domain.RecapCardCTA{Label: "Вернуться в избранное", Action: "open_favorites"},
		})
	}
	if metrics.Buyer.LargestPurchase != nil {
		purchase := metrics.Buyer.LargestPurchase
		cards = append(cards, domain.RecapCard{
			ID:           "largest_purchase",
			Kind:         "buyer",
			Eyebrow:      "Крупная находка года",
			Title:        purchase.Title,
			Description:  "Самая крупная подтверждённая покупка года. Эта карточка не попадает в публичный share-ответ.",
			Value:        formatMoney(purchase.Price),
			ImageURL:     purchase.ImageURL,
			Shareable:    false,
			Reason:       "Выбрана покупка с максимальной ценой среди purchasedAt выбранного года.",
			Presentation: presentation("product", "avito-purple", "trophy"),
			CTA:          listingCTA(purchase.AdID, purchase.Category, "open_listing", "Открыть объявление"),
		})
	}
	if metrics.Buyer.ChatsCount > 0 {
		cards = append(cards, domain.RecapCard{
			ID:           "chats",
			Kind:         "buyer",
			Eyebrow:      "Диалоги",
			Title:        "Вы не боялись уточнять",
			Description:  "Это текущее агрегированное число диалогов профиля — без содержания и имён собеседников.",
			Value:        formatCount(metrics.Buyer.ChatsCount, "диалог", "диалога", "диалогов"),
			Shareable:    false,
			Reason:       "В профиле заполнено агрегированное поле chatsCount.",
			Presentation: presentation("statistic", "avito-blue", "message"),
			CTA:          &domain.RecapCardCTA{Label: "Открыть сообщения", Action: "open_chats"},
		})
	}
	if metrics.Buyer.PurchasesCount > 1 {
		cards = append(cards, domain.RecapCard{
			ID:           "purchases",
			Kind:         "buyer",
			Eyebrow:      "Ваши находки",
			Title:        "Выборы стали покупками",
			Description:  "Считаем только объявления с подтверждённой датой покупки в выбранном году.",
			Value:        formatCount(metrics.Buyer.PurchasesCount, "покупка", "покупки", "покупок"),
			Shareable:    true,
			Reason:       "За год есть больше одной подтверждённой покупки.",
			Presentation: presentation("statistic", "avito-green", "bag"),
		})
	}
	return cards
}

func sellerCards(metrics ProfileMetrics) []domain.RecapCard {
	if !metrics.Seller.HasData {
		return []domain.RecapCard{}
	}
	cards := make([]domain.RecapCard, 0, 8)
	if metrics.Seller.ListingsCount > 0 {
		description := fmt.Sprintf(
			"В %d году опубликовано %d: сейчас активно %d, в архиве — %d.",
			metrics.Year,
			metrics.Seller.ListingsCount,
			metrics.Seller.ActiveListings,
			metrics.Seller.ArchivedListings,
		)
		if !metrics.SellerListingsAreAnnual {
			description = fmt.Sprintf("Сейчас активно %d, в архиве — %d. Для старых записей без publishedAt это снимок профиля.", metrics.Seller.ActiveListings, metrics.Seller.ArchivedListings)
		}
		cards = append(cards, domain.RecapCard{
			ID:           "seller_portfolio",
			Kind:         "seller",
			Eyebrow:      "Вы — продавец",
			Title:        "Ваши объявления",
			Description:  description,
			Value:        formatCount(metrics.Seller.ListingsCount, "объявление", "объявления", "объявлений"),
			Shareable:    true,
			Reason:       "В профиле есть хотя бы одно ownAds.",
			Presentation: presentation("statistic", "avito-green", "megaphone"),
			CTA:          &domain.RecapCardCTA{Label: "Разместить новое", Action: "create_listing"},
		})
	}
	if metrics.Seller.StarListing != nil {
		listing := metrics.Seller.StarListing
		description := fmt.Sprintf("Самое просматриваемое среди объявлений, опубликованных в %d году.", metrics.Year)
		if !metrics.SellerListingsAreAnnual {
			description = "Самое просматриваемое объявление среди тех, которые сейчас есть в профиле."
		}
		cards = append(cards, domain.RecapCard{
			ID:           "star_listing",
			Kind:         "seller",
			Eyebrow:      "Объявление-звезда",
			Title:        listing.Title,
			Description:  description,
			Value:        formatCount(max(metrics.StarListingViews, 0), "просмотр", "просмотра", "просмотров"),
			ImageURL:     listing.ImageURL,
			Shareable:    true,
			Reason:       "Выбрано ownAds с максимальным viewCount.",
			Presentation: presentation("product", "avito-purple", "star"),
			CTA:          listingCTA(listing.AdID, listing.Category, "open_own_listing", "Открыть объявление"),
		})
	}
	if metrics.Seller.SalesCount > 0 {
		cards = append(cards, domain.RecapCard{
			ID:           "second_life",
			Kind:         "seller",
			Eyebrow:      "Вторая жизнь вещей",
			Title:        "Ваши вещи нашли новых владельцев",
			Description:  "Считаем объявления с подтверждённой датой продажи в выбранном году.",
			Value:        formatCount(metrics.Seller.SalesCount, "продажа", "продажи", "продаж"),
			Shareable:    true,
			Reason:       "Есть ownAds с isSold и soldAt в выбранном году.",
			Presentation: presentation("statistic", "avito-green", "recycle"),
			CTA:          &domain.RecapCardCTA{Label: "Дать вещи вторую жизнь", Action: "create_listing"},
		})
	}
	if metrics.Seller.ListingViews > 0 {
		description := fmt.Sprintf("Сумма viewCount объявлений, опубликованных в %d году.", metrics.Year)
		if !metrics.SellerListingsAreAnnual {
			description = "Сумма viewCount всех объявлений в текущем профиле."
		}
		cards = append(cards, domain.RecapCard{
			ID:           "listing_views",
			Kind:         "seller",
			Eyebrow:      "Внимание покупателей",
			Title:        "Ваши объявления заметили",
			Description:  description,
			Value:        formatCount(metrics.Seller.ListingViews, "просмотр", "просмотра", "просмотров"),
			Shareable:    true,
			Reason:       "У собственных объявлений есть просмотры.",
			Presentation: presentation("statistic", "avito-blue", "eye"),
		})
	}
	if metrics.Seller.FavoritesReceived > 0 {
		cards = append(cards, domain.RecapCard{
			ID:           "seller_favorites",
			Kind:         "seller",
			Eyebrow:      "Магнит для избранного",
			Title:        "Ваши предложения нравятся",
			Description:  fmt.Sprintf("Сумма favoritesCount объявлений, опубликованных в %d году.", metrics.Year),
			Value:        formatCount(metrics.Seller.FavoritesReceived, "добавление", "добавления", "добавлений"),
			Shareable:    true,
			Reason:       "У собственных объявлений favoritesCount больше нуля.",
			Presentation: presentation("statistic", "avito-red", "heart"),
		})
	} else if metrics.Seller.LikesReceived > 0 {
		cards = append(cards, domain.RecapCard{
			ID:           "seller_likes_legacy",
			Kind:         "seller",
			Eyebrow:      "Магнит для избранного",
			Title:        "Ваши предложения нравятся",
			Description:  "Для старого профиля используем агрегированное поле likes.",
			Value:        formatCount(metrics.Seller.LikesReceived, "отметка", "отметки", "отметок"),
			Shareable:    true,
			Reason:       "В старом профиле likes больше нуля, а favoritesCount ещё не заполнен.",
			Presentation: presentation("statistic", "avito-red", "heart"),
		})
	}
	if metrics.Seller.ContactsReceived > 0 {
		cards = append(cards, domain.RecapCard{
			ID:           "seller_contacts",
			Kind:         "seller",
			Eyebrow:      "Начало разговора",
			Title:        "Ваши объявления вызывали вопросы",
			Description:  fmt.Sprintf("Сумма contactsCount объявлений, опубликованных в %d году.", metrics.Year),
			Value:        formatCount(metrics.Seller.ContactsReceived, "контакт", "контакта", "контактов"),
			Shareable:    true,
			Reason:       "У собственных объявлений contactsCount больше нуля.",
			Presentation: presentation("statistic", "avito-blue", "message"),
			CTA:          &domain.RecapCardCTA{Label: "Открыть свои объявления", Action: "open_own_listings"},
		})
	}
	if metrics.Seller.ReviewsCount > 0 && metrics.Seller.AverageRating != nil {
		cards = append(cards, domain.RecapCard{
			ID:           "reviews",
			Kind:         "seller",
			Eyebrow:      "Что пишут люди",
			Title:        "Сделки оставили хорошее впечатление",
			Description:  fmt.Sprintf("Средняя оценка по %s выбранного года.", formatCount(metrics.Seller.ReviewsCount, "отзыву", "отзывам", "отзывам")),
			Value:        fmt.Sprintf("%.1f из 5", *metrics.Seller.AverageRating),
			Shareable:    true,
			Reason:       "Есть отзывы с createdAt в выбранном году.",
			Presentation: presentation("statistic", "avito-orange", "review"),
		})
	}
	return cards
}

func combinedCards(metrics ProfileMetrics, joinedAt time.Time) []domain.RecapCard {
	cards := make([]domain.RecapCard, 0, 3)
	if metrics.Buyer.HasData && metrics.Seller.HasData {
		cards = append(cards, domain.RecapCard{
			ID:           "both_sides",
			Kind:         "combined",
			Eyebrow:      "Две стороны одного года",
			Title:        "Вы и находили, и помогали находить",
			Description:  "Покупатель и продавец не роли навсегда: в этих итогах они существуют одновременно.",
			Value:        formatCount(metrics.Combined.Deals, "сделка", "сделки", "сделок"),
			Shareable:    true,
			Reason:       "Есть данные и со стороны покупателя, и со стороны продавца.",
			Presentation: presentation("split", "avito-purple", "switch"),
		})
	}
	if metrics.Combined.Categories > 1 {
		cards = append(cards, domain.RecapCard{
			ID:           "interest_circle",
			Kind:         "combined",
			Eyebrow:      "Круг интересов",
			Title:        "Ваш Авито шире одной категории",
			Description:  "Считаем категории с реальными просмотрами, избранным, покупками, продажами или объявлениями.",
			Value:        formatCount(metrics.Combined.Categories, "категория", "категории", "категорий"),
			Shareable:    true,
			Reason:       "В профиле есть активность более чем в одной категории.",
			Presentation: presentation("orbit", "avito-blue", "categories"),
			CTA:          &domain.RecapCardCTA{Label: "Удивить меня", Action: "open_recommendations"},
		})
	}
	if !joinedAt.IsZero() && joinedAt.Year() <= metrics.Year {
		years := metrics.Year - joinedAt.Year() + 1
		cards = append(cards, domain.RecapCard{
			ID:           "avito_history",
			Kind:         "combined",
			Eyebrow:      "Вместе с Авито",
			Title:        "Это уже часть вашей истории",
			Description:  fmt.Sprintf("Профиль появился в %d году.", joinedAt.Year()),
			Value:        formatCount(years, "год", "года", "лет"),
			Shareable:    true,
			Reason:       "В профиле есть joinedAt не позже года итогов.",
			Presentation: presentation("statistic", "avito-orange", "calendar"),
		})
	}
	return cards
}

func finalCard(metrics ProfileMetrics) domain.RecapCard {
	title := "Впереди новые находки"
	description := "Продолжите с того, что оказалось важным именно для вас."
	cta := &domain.RecapCardCTA{Label: "Посмотреть рекомендации", Action: "open_recommendations"}
	if metrics.Seller.HasData && !metrics.Buyer.HasData {
		title = "Следующая вещь уже ждёт нового владельца"
		description = "Новое объявление может стать звездой следующих итогов."
		cta = &domain.RecapCardCTA{Label: "Разместить объявление", Action: "create_listing"}
	}
	return domain.RecapCard{
		ID:           "next_step",
		Kind:         "final",
		Eyebrow:      "Продолжение следует",
		Title:        title,
		Description:  description,
		Shareable:    true,
		Reason:       "Обязательная финальная карточка с одним понятным следующим действием.",
		Presentation: presentation("finale", "avito-purple", "arrow"),
		CTA:          cta,
	}
}

func appendUnique(selected, candidates []domain.RecapCard, limit int) []domain.RecapCard {
	if limit <= 0 {
		return selected
	}
	existing := make(map[string]struct{}, len(selected))
	for _, card := range selected {
		existing[card.ID] = struct{}{}
	}
	added := 0
	for _, candidate := range candidates {
		if added >= limit {
			break
		}
		if _, exists := existing[candidate.ID]; exists {
			continue
		}
		selected = append(selected, candidate)
		existing[candidate.ID] = struct{}{}
		added++
	}
	return selected
}

func presentation(layout, theme, icon string) domain.RecapCardPresentation {
	return domain.RecapCardPresentation{Layout: layout, Theme: theme, Icon: icon}
}

func categoryCTA(category string) *domain.RecapCardCTA {
	return &domain.RecapCardCTA{
		Label:  "Открыть свежие объявления",
		Action: "open_category",
		Params: map[string]string{"category": category},
	}
}

func listingCTA(adID, category, action, label string) *domain.RecapCardCTA {
	if adID == "" {
		return categoryCTA(category)
	}
	return &domain.RecapCardCTA{
		Label:  label,
		Action: action,
		Params: map[string]string{"ad_id": adID},
	}
}

func formatCount(value int, one, few, many string) string {
	form := many
	lastTwo := value % 100
	last := value % 10
	if lastTwo < 11 || lastTwo > 14 {
		switch last {
		case 1:
			form = one
		case 2, 3, 4:
			form = few
		}
	}
	return fmt.Sprintf("%d %s", value, form)
}

func formatMoney(value int64) string {
	digits := strconv.FormatInt(value, 10)
	parts := make([]string, 0, (len(digits)+2)/3)
	for len(digits) > 3 {
		parts = append([]string{digits[len(digits)-3:]}, parts...)
		digits = digits[:len(digits)-3]
	}
	parts = append([]string{digits}, parts...)
	return strings.Join(parts, " ") + " ₽"
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}
