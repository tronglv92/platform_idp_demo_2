package export

import (
	"fmt"
	"go-service-template/helper/exporter"
	"go-service-template/helper/toolkit/jsonx"
	"go-service-template/helper/toolkit/timex"
	"go-service-template/internal/repository"
	"go-service-template/internal/types/dto"
	"time"
)

const (
	ProductModel = "export-product"
)

func Factory(ticketRepo repository.ProductRepository, model string, filters any) (exporter.Exporter, error) {
	var writer exporter.Exporter
	switch model {
	case ProductModel:
		params, err := jsonx.MapToStruct[dto.ProductFilter](filters)
		if err != nil {
			return nil, err
		}
		writer = exporter.NewExcelExporter(
			fmt.Sprintf("product-%s.xlsx", timex.Now().Format(time.DateOnly)),
			NewOutboundTicketHandler(ticketRepo, params),
		)
	default:
		return nil, fmt.Errorf("unsupported export model: %s", model)
	}
	return writer, nil
}
