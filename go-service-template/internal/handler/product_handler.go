package handler

import (
	"context"
	productpb "go-service-template/api/product"
	"go-service-template/helper/errors"
	pmcpb "go-service-template/helper/pmcpb/protobuf"
	"go-service-template/internal/registry"
	"go-service-template/internal/service"

	"google.golang.org/protobuf/types/known/emptypb"
)

type ProductHandler struct {
	productSvc service.ProductService

	productpb.UnimplementedProductServer
}

func NewProductHandler(reg registry.ServiceContext) *ProductHandler {
	return &ProductHandler{
		productSvc: reg.GetServiceFactory().GetProductService(),
	}
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.BadRequest(err)
	}
	return h.productSvc.GetProduct(ctx, req)
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.BadRequest(err)
	}
	return h.productSvc.ListProducts(ctx, req)
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.BadRequest(err)
	}
	return h.productSvc.CreateProduct(ctx, req)
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.BadRequest(err)
	}
	return h.productSvc.UpdateProduct(ctx, req)
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.BadRequest(err)
	}
	return h.productSvc.DeleteProduct(ctx, req)
}

func (h *ProductHandler) ExportProducts(ctx context.Context, req *productpb.ExportProductsRequest) (*pmcpb.FileInfo, error) {
	return h.productSvc.ExportProducts(ctx, req)
}

func (h *ProductHandler) GetProductFromCache(ctx context.Context, req *productpb.GetProductFromCacheRequest) (*productpb.GetProductFromCacheResponse, error) {

	if err := req.Validate(); err != nil {
		return nil, errors.BadRequest(err)
	}
	return h.productSvc.GetProductFromCache(ctx, req)
}
