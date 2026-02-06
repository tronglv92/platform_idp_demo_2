package mapper

import (
	"strconv"
	"time"
)

type ProductInformation struct {
	BigCategory      string
	CategoryOne      string
	CategoryTwo      string
	CategoryThree    string
	ItemId           string
	ItemName         string
	CoreRange        string
	StarGroup        string
	NumberOfStars    int32
	PurchMultiple    int32
	OwnBrand         bool
	BrandId          string
	Brand            string
	VendorId         string
	VendorName       string
	ManufacturerId   string
	ManufacturerName string
	ProcurementType  string
	DistributionType string
	ABCTransaction   string
	MedicineType     string
	InventUnit       string
	PurchUnit        string
	PurchConvert     float64
	IsStopped        bool
	StopDate         *time.Time
	BrandItem        string
	Description      string
	LaunchingNPD     string
	Origin           string
	POGUnit          string
	POGWidth         float64
	POGDepth         float64
	POGHeight        float64
	StorageCondition string
	SpecialGroup     string
	Dosage           string
	HotlineSKU       bool
	ATC1             string
	ATCVN1           string
	ATC2             string
	ATCVN2           string
	ATC3             string
	ATCVN3           string
	ATC4             string
	ATCVN4           string
	ATC5             string
	ATCVN5           string
}

func MapRowToProductInformation(row []string) *ProductInformation {
	p := &ProductInformation{}

	p.BigCategory = row[0]
	p.CategoryOne = row[1]
	p.CategoryTwo = row[2]
	p.CategoryThree = row[3]

	p.ItemId = row[4]
	p.ItemName = row[5]
	p.CoreRange = row[6]
	p.StarGroup = row[7]
	//p.NumberOfStars, _ = strconv.Atoi(row[8])
	//p.PurchMultiple, _ = strconv.Atoi(row[9])
	//p.OwnBrand = strings.EqualFold(row[10], "true") || row[10] == "1"

	// Brand & Vendor
	p.BrandId = row[11]
	p.Brand = row[12]
	p.VendorId = row[13]
	p.VendorName = row[14]
	p.ManufacturerId = row[15]
	p.ManufacturerName = row[16]

	// Procurement & Distribution
	p.ProcurementType = row[17]
	p.DistributionType = row[18]
	p.ABCTransaction = row[19]
	p.MedicineType = row[20]

	// Units
	p.InventUnit = row[21]
	p.PurchUnit = row[22]
	p.PurchConvert, _ = strconv.ParseFloat(row[23], 64)

	// Status
	//p.IsStopped = strings.EqualFold(row[24], "true") || row[24] == "1"
	//if row[28] != "" {
	//	if t, err := time.Parse("2006-01-02", row[28]); err == nil {
	//		p.StopDate = &t
	//	}
	//}

	// Misc
	p.BrandItem = row[25]
	p.Description = row[26]
	p.LaunchingNPD = row[27]
	p.Origin = row[28]
	//p.StopDate = row[29]
	p.POGUnit = row[30]
	//p.POGWidth, _ = strconv.ParseFloat(row[31], 64)
	//p.POGDepth, _ = strconv.ParseFloat(row[32], 64)
	//p.POGHeight, _ = strconv.ParseFloat(row[33], 64)
	p.StorageCondition = row[34]
	p.SpecialGroup = row[35]
	p.Dosage = row[36]

	return p
}
