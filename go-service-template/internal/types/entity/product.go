package entity

import (
	"go-service-template/helper/model"
	baseModel "go-service-template/helper/sql/entity"
)

type Product struct {
	baseModel.IdModel
	Name   string       `gorm:"type:varchar(255)"`
	Sku    string       `gorm:"type:varchar(10);uniqueIndex"`
	Status model.Status `gorm:"column:status;default:0"`
	Price  float64      `gorm:"column:price;default:0"`
}

func (Product) TableName() string {
	return "products"
}
