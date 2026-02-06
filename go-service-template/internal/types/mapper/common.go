package mapper

import (
	"go-service-template/helper/model"
	pmcpb "go-service-template/helper/pmcpb/protobuf"
	baseEntity "go-service-template/helper/sql/entity"
	"go-service-template/helper/toolkit/timex"
)

func MapAuditMetaData(item *baseEntity.IdModel) *pmcpb.AuditMetadata {
	if item == nil {
		return nil
	}
	return &pmcpb.AuditMetadata{
		CreatedBy:    item.CreatedBy,
		CreatedByUid: item.CreatedByUId,
		UpdatedBy:    item.UpdatedBy,
		UpdatedByUid: item.UpdatedByUId,
		CreatedAt:    timex.TimeToProto(item.CreatedAt),
		UpdatedAt:    timex.TimeToProto(item.UpdatedAt),
	}
}

func MapPaginationData(pagination model.Pagination) *pmcpb.Pagination {
	return &pmcpb.Pagination{
		Limit:        pagination.GetLimit(),
		Page:         pagination.GetPage(),
		TotalPage:    pagination.GetTotalPage(),
		TotalRecords: pagination.GetTotalRecords(),
	}
}
