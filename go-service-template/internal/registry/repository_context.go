package registry

import (
	"go-service-template/helper/db"
	rp "go-service-template/internal/repository"
)

// RepositoryContext - Database repository access
type RepositoryContext interface {
	GetProductRepo() rp.ProductRepository
}

type repositoryContext struct {
	sqlConn     db.Database
	productRepo rp.ProductRepository
}

func NewRepositoryContext(sqlConn db.Database) RepositoryContext {
	return &repositoryContext{
		sqlConn:     sqlConn,
		productRepo: rp.NewProductRepository(sqlConn),
	}
}

func (r *repositoryContext) GetProductRepo() rp.ProductRepository {
	return r.productRepo
}
