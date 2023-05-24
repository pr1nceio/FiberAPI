package db

type ShopItem struct {
	UUID        string `gorm:"primary_key;column:uuid;type:varchar;size:64;" json:"uuid"`
	Name        string `gorm:"column:name;type:varchar;size:255;" json:"name"`
	Price       int    `gorm:"column:price;type:int;" json:"price"`
	Stock       int    `gorm:"column:stock;type:int;default:0;" json:"stock"`
	InfStock    bool   `gorm:"column:infStock;type:tinyint;default:0;" json:"inf_stock"`
	Image       string `gorm:"column:image;type:varchar;size:512;" json:"image"`
	Description string `gorm:"column:description;type:varchar;size:512;" json:"description"`
	Currency    string `gorm:"column:currency;type:varchar;size:3;default:USD;" json:"currency"`
	ShopID      int    `gorm:"column:shop_id;type:int;" json:"shop_id"`
}

func (s *ShopItem) TableName() string {
	return "shopItems"
}
