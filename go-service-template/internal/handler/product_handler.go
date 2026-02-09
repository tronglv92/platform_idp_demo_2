package handler

import (
	"context"
	productpb "go-service-template/api/product"
	"go-service-template/helper/errors"
	pmcpb "go-service-template/helper/pmcpb/protobuf"
	"go-service-template/internal/registry"
	"go-service-template/internal/service"
	"os"
	"time"

	"go-service-template/helper/queue"
	consumerq "go-service-template/helper/queue/consumer"
	kconf "go-service-template/helper/queue/kafka"

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

func (h *ProductHandler) CheckKafka(ctx context.Context, req *productpb.CheckKafkaRequest) (*productpb.CheckKafkaResponse, error) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	// Build queue producer using internal helper
	cfg := queue.Config{
		Stack: queue.KafkaDriver,
		Kafka: kconf.Config{
			Brokers: []string{broker},
			Topic:   "health-check",
		},
	}

	prod, err := queue.New(cfg)
	if err != nil {
		return &productpb.CheckKafkaResponse{Status: false, Message: err.Error()}, nil
	}
	defer prod.Close()

	// send a small health payload with short timeout
	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	payload := &consumerq.Payload{
		QueueName: "health-check",
		Message:   "ping",
	}

	if err := prod.SendCtx(dialCtx, payload); err != nil {
		return &productpb.CheckKafkaResponse{Status: false, Message: err.Error()}, nil
	}

	return &productpb.CheckKafkaResponse{Status: true, Message: "connected"}, nil
}
