package mapper

import (
	productpb "go-service-template/api/product"
	"go-service-template/internal/types/entity"

	"go-service-template/helper/model"
)

func ProductMapToResponse(item *entity.Product) *productpb.GetProductResponse {
	return &productpb.GetProductResponse{
		Id:     item.Id,
		Uid:    item.UId,
		Sku:    item.Sku,
		Name:   item.Name,
		Status: productpb.Status(item.Status),
		Price:  item.Price,
	}
}

func ProductMapToResponses(items []*entity.Product, pagination model.Pagination) *productpb.ListProductsResponse {
	var results []*productpb.GetProductResponse
	for _, val := range items {
		results = append(results, ProductMapToResponse(val))
	}
	return &productpb.ListProductsResponse{
		Products:   results,
		Pagination: MapPaginationData(pagination),
	}
}
