package service

import (
	"context"
	"fmt"
	productpb "go-service-template/api/product"
	"go-service-template/helper/cache"
	"go-service-template/helper/errors"
	"go-service-template/helper/excel"
	"go-service-template/helper/identity"
	"go-service-template/helper/model"
	pmcpb "go-service-template/helper/pmcpb/protobuf"
	"go-service-template/internal/export"
	rp "go-service-template/internal/repository"
	"go-service-template/internal/types/dto"
	"go-service-template/internal/types/entity"
	"go-service-template/internal/types/mapper"
	"time"

	"github.com/zeromicro/go-zero/core/jsonx"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProductService interface {
	GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error)
	ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error)
	CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*emptypb.Empty, error)
	UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*emptypb.Empty, error)
	DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*emptypb.Empty, error)
	ExportProducts(ctx context.Context, req *productpb.ExportProductsRequest) (*pmcpb.FileInfo, error)
	GetProductFromCache(ctx context.Context, req *productpb.GetProductFromCacheRequest) (*productpb.GetProductFromCacheResponse, error)
}

type productSvcImpl struct {
	productRepo rp.ProductRepository
	cacheClient cache.Cache
}

func NewProductService(productRepo rp.ProductRepository, cacheClient cache.Cache) ProductService {
	return &productSvcImpl{
		productRepo: productRepo,
		cacheClient: cacheClient,
	}
}

func (s *productSvcImpl) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	result, err := s.productRepo.FindByUid(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	resp := mapper.ProductMapToResponse(result)

	// Cache the product by SKU for "GetProductFromCache"
	if result.Sku != "" {
		data, _ := jsonx.Marshal(resp)
		_ = s.cacheClient.SetWithExpireCtx(ctx, result.Sku, string(data), 10*time.Minute)
	}

	return resp, nil
}

func (s *productSvcImpl) ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
	results, pagination, err := s.productRepo.FindWithPagination(ctx,
		int(req.GetLimit()),
		int(req.GetPage()),
		req.GetSortOrder(),
		req.GetSortBy(),
		dto.MapProductFilter(req),
	)
	if err != nil {
		return nil, err
	}

	return mapper.ProductMapToResponses(results, pagination), nil
}

func (s *productSvcImpl) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*emptypb.Empty, error) {
	total, err := s.productRepo.CountBySku(ctx, req.GetSku())
	if err == nil && total > 0 {
		return nil, errors.DuplicateData
	}

	err = s.productRepo.Create(ctx, &entity.Product{
		Name:   req.Name,
		Sku:    req.Sku,
		Status: model.Status(req.Status),
		Price:  req.Price,
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *productSvcImpl) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*emptypb.Empty, error) {
	item, err := s.productRepo.FindByUid(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{
		"name":   req.Name,
		"status": req.Status,
		"price":  req.Price,
	}
	if e := s.productRepo.UpdateById(ctx, m, item.Id); e != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *productSvcImpl) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*emptypb.Empty, error) {
	item, err := s.productRepo.FindByUid(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, s.productRepo.DeleteById(ctx, item.Id)
}

func (s *productSvcImpl) ExportProducts(ctx context.Context, req *productpb.ExportProductsRequest) (*pmcpb.FileInfo, error) {
	if req.GetExportType() == int32(productpb.EXPORT_TYPE_URL) {
		_, err := identity.UserClaimsFromContext(ctx)
		if err != nil {
			return nil, err
		}

		//err = h.appContext.GetProducer().Send(&baseConsumer.Payload{
		//	TraceId:   trace.TraceIDFromContext(ctx),
		//	QueueName: constant.ExportSendQueue,
		//	Message: event.ExportEvent{
		//		Model:         export.OutboundTicketModel,
		//		EmployeeId:    u.GetId(),
		//		EmployeeEmail: u.GetEmail(),
		//		Filters:       dto.MapExportToOutboundTicketFilter(req),
		//	},
		//})
		//if err != nil {
		//	return nil, err
		//}
		return &pmcpb.FileInfo{}, nil
	}

	writer, err := export.Factory(s.productRepo, export.ProductModel, map[string]any{
		"filters": dto.MapProductExportFilter(req),
	})
	if err != nil {
		return nil, err
	}

	f, err := writer.Export(ctx, excel.NewExcelWriter())
	if err != nil {
		return nil, err
	}
	return &pmcpb.FileInfo{
		FileData: f.Bytes(),
	}, nil
}

func (s *productSvcImpl) GetProductFromCache(ctx context.Context, req *productpb.GetProductFromCacheRequest) (*productpb.GetProductFromCacheResponse, error) {
	sku := req.GetSku()
	if sku == "" {
		return nil, errors.BadRequest(fmt.Errorf("sku is required"))
	}

	// GET from cache
	var val string
	err := s.cacheClient.GetCtx(ctx, sku, &val)
	if err != nil {
		if s.cacheClient.IsNotFound(err) {
			// Cache miss: In a real scenario, we might want to query DB here.
			// But for "GetProductFromCache", returning miss is also fine.
			return &productpb.GetProductFromCacheResponse{
				Sku:       sku,
				Data:      "",
				FromCache: false,
			}, nil
		}
		return nil, err
	}

	return &productpb.GetProductFromCacheResponse{
		Sku:       sku,
		Data:      val,
		FromCache: true,
	}, nil
}
