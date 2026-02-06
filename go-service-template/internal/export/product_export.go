package export

import (
	"context"
	"go-service-template/helper/excel"
	"go-service-template/helper/exporter"
	"go-service-template/helper/model"
	rp "go-service-template/internal/repository"
	"go-service-template/internal/types/dto"

	"github.com/xuri/excelize/v2"
)

type productHandler struct {
	productRepo rp.ProductRepository
	filters     *dto.ProductFilter
}

func NewOutboundTicketHandler(productRepo rp.ProductRepository, filters *dto.ProductFilter) exporter.ExcelHandler {
	return &productHandler{
		productRepo: productRepo,
		filters:     filters,
	}
}

func (s *productHandler) headers() []excel.Header {
	return []excel.Header{
		{Name: "ID", Width: 20},
		{Name: "SKU", Width: 25},
		{Name: "Name", Width: 20},
	}
}

func (s *productHandler) Build(ctx context.Context, writer *excel.ExcelWriter) (*excelize.File, error) {
	sw, err := writer.File().NewStreamWriter(writer.SheetName())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = sw.Flush()
	}()

	if err = writer.SetHeader(sw, s.headers(), writer.HeaderStyle()); err != nil {
		return nil, err
	}

	var (
		prevLink, nextLink string
		currentRow         = 2
	)
	for {
		results, cursor, err := s.productRepo.FindWithCursor(
			ctx,
			model.DefaultLimit,
			prevLink,
			nextLink,
			"Id",
			model.SortOrderDefault,
			s.filters,
		)
		if err != nil {
			return nil, err
		}
		if len(results) == 0 {
			break
		}

		for _, val := range results {

			values := []any{
				val.Id,   // ID
				val.Sku,  // SKU
				val.Name, // Name
			}
			if err = writer.SetRow(sw, currentRow, values); err != nil {
				return nil, err
			}
			currentRow++
		}

		if len(cursor.Next) == 0 {
			break
		}
		nextLink = cursor.Next
		prevLink = cursor.Prev
	}

	return writer.File(), nil
}
