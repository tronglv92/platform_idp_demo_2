package repository

import (
	"context"
	"go-service-template/helper/db"
	"go-service-template/helper/model"
	baseRepo "go-service-template/helper/sql/repository"
	"go-service-template/internal/types/dto"
	"go-service-template/internal/types/entity"

	"gorm.io/gorm"
)

type productRepo struct {
	baseRepo.Repository[entity.Product]
}

type ProductRepository interface {
	baseRepo.Repository[entity.Product]
	CountBySku(ctx context.Context, code string) (int64, error)
	FindWithPagination(ctx context.Context, limit, page int, sortOrder, sortBy string, filters *dto.ProductFilter) ([]*entity.Product, model.Pagination, error)
	FindWithCursor(ctx context.Context, limit int, prev, next, sortBy, sortOrder string, filters *dto.ProductFilter) ([]*entity.Product, *model.Cursor, error)
}

func NewProductRepository(db db.Database) ProductRepository {
	return &productRepo{
		baseRepo.NewRepository[entity.Product](db),
	}
}

func (r *productRepo) FindWithPagination(ctx context.Context, limit, page int, sortOrder, sortBy string, filters *dto.ProductFilter) ([]*entity.Product, model.Pagination, error) {
	return r.QueryWithPagination(ctx, limit, page,
		r.WithOrder(sortOrder, sortBy),
		r.WithFilter(filters),
	)
}

func (r *productRepo) FindWithCursor(ctx context.Context, limit int, prev, next, sortBy, sortOrder string, filters *dto.ProductFilter) ([]*entity.Product, *model.Cursor, error) {
	return r.QueryWithCursor(ctx, limit, prev, next, sortBy, sortOrder, r.WithFilter(filters))
}

func (r *productRepo) CountBySku(ctx context.Context, code string) (int64, error) {
	return r.Count(ctx, func(g *gorm.DB) *gorm.DB {
		return g.Where("sku=?", code)
	})
}

func (r *productRepo) WithFilter(filters *dto.ProductFilter) baseRepo.QueryOpt {
	return func(g *gorm.DB) *gorm.DB {
		if filters.Name != nil {
			g = g.Where("names LIKE ?", "%"+*filters.Name+"%")
		}
		if filters.Sku != nil {
			g = g.Where("sku=?", *filters.Sku)
		}
		return g
	}
}
