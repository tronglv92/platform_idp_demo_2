package dto

import productpb "go-service-template/api/product"

type ProductFilter struct {
	Name *string `json:"name"`
	Sku  *string `json:"sku"`
}

func MapProductFilter(in *productpb.ListProductsRequest) *ProductFilter {
	return &ProductFilter{
		Name: in.Name,
		Sku:  in.Sku,
	}
}

func MapProductExportFilter(in *productpb.ExportProductsRequest) *ProductFilter {
	return &ProductFilter{}
}
