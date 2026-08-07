package domain

type Achievement struct {
	Slug        string
	Title       string
	Description string
	Icon        string
	Category    string
}

var DefaultAchievements = []Achievement{
	{
		Slug:        "curious",
		Title:       "Любопытный",
		Description: "Просмотрел более 500 объявлений за год",
		Icon:        "👀",
		Category:    "views",
	},
	{
		Slug:        "explorer",
		Title:       "Исследователь",
		Description: "Просмотрел более 1000 объявлений за год",
		Icon:        "🔍",
		Category:    "views",
	},
	{
		Slug:        "social_butterfly",
		Title:       "Душа компании",
		Description: "Отправил более 50 сообщений за год",
		Icon:        "💬",
		Category:    "social",
	},
	{
		Slug:        "seller_master",
		Title:       "Мастер продаж",
		Description: "Продал более 5 товаров за год",
		Icon:        "🏆",
		Category:    "sales",
	},
	{
		Slug:        "shopaholic",
		Title:       "Шопоголик",
		Description: "Купил более 10 товаров за год",
		Icon:        "🛍️",
		Category:    "sales",
	},
	{
		Slug:        "veteran",
		Title:       "Ветеран",
		Description: "Был активен более 300 дней в году",
		Icon:        "⭐",
		Category:    "activity",
	},
	{
		Slug:        "enthusiast",
		Title:       "Энтузиаст",
		Description: "Был активен более 100 дней в году",
		Icon:        "🔥",
		Category:    "activity",
	},
}
