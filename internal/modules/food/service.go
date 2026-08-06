package food

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/google/uuid"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type foodService struct {
	repo FoodRepository
	log  *zap.Logger
}

func NewFoodService(repo FoodRepository, log *zap.Logger) FoodService {
	return &foodService{
		repo: repo,
		log:  log,
	}
}

func (s *foodService) GetFoods(ctx context.Context, query FoodFilterQuery) ([]FoodMasterResponse, int64, error) {
	foods, total, err := s.repo.GetFoods(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]FoodMasterResponse, len(foods))
	for i, f := range foods {
		responses[i] = ToFoodMasterResponse(&f)
	}
	return responses, total, nil
}

func (s *foodService) GetByID(ctx context.Context, id string) (*FoodMasterResponse, error) {
	food, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if food == nil {
		return nil, errors.New("food item not found")
	}
	res := ToFoodMasterResponse(food)
	return &res, nil
}

func (s *foodService) Create(ctx context.Context, req CreateFoodRequest) (*FoodMasterResponse, error) {
	if req.Name == "" {
		return nil, errors.New("food name is required")
	}
	if req.ServingSize == "" {
		return nil, errors.New("serving size is required")
	}

	status := req.Status
	if status == "" {
		status = "active"
	}
	source := req.Source
	if source == "" {
		source = "manual"
	}

	food := domain.FoodMaster{
		BaseModel: domain.BaseModel{
			ID: uuid.NewString(),
		},
		Name:           strings.TrimSpace(req.Name),
		Manufacturer:   strings.TrimSpace(req.Manufacturer),
		ServingSize:    strings.TrimSpace(req.ServingSize),
		Calories:       req.EnergyKcal,
		EnergyKcal:     req.EnergyKcal,
		ProteinG:       req.ProteinG,
		CarbohydrateG:  req.CarbohydrateG,
		FatG:           req.FatG,
		SugarG:         req.SugarG,
		SodiumMg:       req.SodiumMg,
		FiberG:         req.FiberG,
		SaturatedFatG:  req.SaturatedFatG,

		EnergyPercentageDV:       req.EnergyPercentageDV,
		ProteinPercentageDV:      req.ProteinPercentageDV,
		CarbohydratePercentageDV: req.CarbohydratePercentageDV,
		FatPercentageDV:          req.FatPercentageDV,
		SodiumPercentageDV:       req.SodiumPercentageDV,

		TotalFat:          req.TotalFat,
		SaturatedFat:      req.SaturatedFat,
		Sodium:            req.Sodium,
		Protein:           req.Protein,
		TotalCarbohydrate: req.TotalCarbohydrate,
		DietaryFiber:      req.DietaryFiber,
		Energy:            req.Energy,

		NutritionBasis: determineNutritionBasis(req.NutritionBasis, req.Manufacturer),
		Source:         source,
		Barcode:        strings.TrimSpace(req.Barcode),
		ImageURL:       strings.TrimSpace(req.ImageURL),
		Status:         status,
	}

	if err := s.repo.Create(ctx, &food); err != nil {
		return nil, err
	}

	res := ToFoodMasterResponse(&food)
	return &res, nil
}

func (s *foodService) Update(ctx context.Context, id string, req UpdateFoodRequest) (*FoodMasterResponse, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("food item not found")
	}

	if req.Name != "" {
		existing.Name = strings.TrimSpace(req.Name)
	}
	if req.Manufacturer != "" {
		existing.Manufacturer = strings.TrimSpace(req.Manufacturer)
	}
	if req.ServingSize != "" {
		existing.ServingSize = strings.TrimSpace(req.ServingSize)
	}

	existing.Calories = req.EnergyKcal
	existing.EnergyKcal = req.EnergyKcal
	existing.ProteinG = req.ProteinG
	existing.CarbohydrateG = req.CarbohydrateG
	existing.FatG = req.FatG
	existing.SugarG = req.SugarG
	existing.SodiumMg = req.SodiumMg
	existing.FiberG = req.FiberG
	existing.SaturatedFatG = req.SaturatedFatG

	existing.EnergyPercentageDV = req.EnergyPercentageDV
	existing.ProteinPercentageDV = req.ProteinPercentageDV
	existing.CarbohydratePercentageDV = req.CarbohydratePercentageDV
	existing.FatPercentageDV = req.FatPercentageDV
	existing.SodiumPercentageDV = req.SodiumPercentageDV

	existing.TotalFat = req.TotalFat
	existing.SaturatedFat = req.SaturatedFat
	existing.Sodium = req.Sodium
	existing.Protein = req.Protein
	existing.TotalCarbohydrate = req.TotalCarbohydrate
	existing.DietaryFiber = req.DietaryFiber
	existing.Energy = req.Energy

	if req.NutritionBasis != "" {
		existing.NutritionBasis = determineNutritionBasis(req.NutritionBasis, existing.Manufacturer)
	}
	if req.Source != "" {
		existing.Source = req.Source
	}
	if req.Barcode != "" {
		existing.Barcode = strings.TrimSpace(req.Barcode)
	}
	if req.ImageURL != "" {
		existing.ImageURL = strings.TrimSpace(req.ImageURL)
	}
	if req.Status != "" {
		existing.Status = req.Status
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	res := ToFoodMasterResponse(existing)
	return &res, nil
}

func (s *foodService) Delete(ctx context.Context, id string) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("food item not found")
	}
	return s.repo.Delete(ctx, id)
}

func (s *foodService) SearchFoods(ctx context.Context, query FoodFilterQuery) ([]FoodMasterResponse, int64, error) {
	if query.Status == "" {
		query.Status = "active"
	}
	return s.GetFoods(ctx, query)
}

func (s *foodService) PreviewExcelImport(ctx context.Context, fileBytes []byte) (*ExcelImportPreviewResponse, error) {
	s.log.Info("starting excel import preview parsing", zap.Int("bytes", len(fileBytes)))

	rowsData, err := s.readExcelRows(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(rowsData) <= 1 {
		return nil, errors.New("excel file is empty or only contains header row")
	}

	headers := rowsData[0]
	headerMap := make(map[string]int)
	for colIdx, h := range headers {
		cleanH := strings.ToLower(strings.TrimSpace(h))
		headerMap[cleanH] = colIdx
		headerMap[strings.ReplaceAll(cleanH, " ", "_")] = colIdx
		headerMap[strings.ReplaceAll(cleanH, "_", " ")] = colIdx
		headerMap[strings.ReplaceAll(strings.ReplaceAll(cleanH, " ", ""), "_", "")] = colIdx
	}

	var previewRows []ExcelImportRow
	validCount := 0
	invalidCount := 0

	for rIdx := 1; rIdx < len(rowsData); rIdx++ {
		row := rowsData[rIdx]
		if len(row) == 0 || isRowEmpty(row) {
			continue
		}

		itemReq, rowErrs := s.parseAndValidateRow(row, headerMap, rIdx+1)
		isValid := len(rowErrs) == 0

		if isValid {
			validCount++
		} else {
			invalidCount++
		}

		previewRows = append(previewRows, ExcelImportRow{
			RowIndex: rIdx + 1,
			IsValid:  isValid,
			Errors:   rowErrs,
			Data:     itemReq,
		})
	}

	return &ExcelImportPreviewResponse{
		TotalRows:   len(previewRows),
		ValidRows:   validCount,
		InvalidRows: invalidCount,
		Rows:        previewRows,
	}, nil
}

func isRowEmpty(row []string) bool {
	for _, val := range row {
		if strings.TrimSpace(val) != "" {
			return false
		}
	}
	return true
}

func (s *foodService) readExcelRows(fileBytes []byte) ([][]string, error) {
	// 1. Try parsing as XLSX (ZIP archive containing sheet XML) using standard library
	rows, err := parseXLSXBytes(fileBytes)
	if err == nil && len(rows) > 0 {
		return rows, nil
	}

	// 2. Try parsing as CSV
	r := csv.NewReader(bytes.NewReader(fileBytes))
	r.FieldsPerRecord = -1
	csvRows, csvErr := r.ReadAll()
	if csvErr == nil && len(csvRows) > 0 {
		return csvRows, nil
	}

	return nil, errors.New("invalid file format: must be valid XLSX or CSV")
}

func parseXLSXBytes(b []byte) ([][]string, error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}

	var sharedStrings []string
	for _, f := range zr.File {
		if f.Name == "xl/sharedStrings.xml" {
			rc, err := f.Open()
			if err == nil {
				sharedStrings = parseSharedStrings(rc)
				_ = rc.Close()
			}
			break
		}
	}

	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			defer rc.Close()
			return parseWorksheetXML(rc, sharedStrings)
		}
	}

	return nil, errors.New("worksheet not found in excel archive")
}

type sstXML struct {
	XMLName xml.Name `xml:"sst"`
	SI      []struct {
		T string `xml:"t"`
		R []struct {
			T string `xml:"t"`
		} `xml:"r"`
	} `xml:"si"`
}

func parseSharedStrings(r io.Reader) []string {
	var sst sstXML
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&sst); err != nil {
		return nil
	}
	var res []string
	for _, item := range sst.SI {
		if item.T != "" {
			res = append(res, item.T)
		} else {
			var sb strings.Builder
			for _, run := range item.R {
				sb.WriteString(run.T)
			}
			res = append(res, sb.String())
		}
	}
	return res
}

type worksheetXML struct {
	XMLName   xml.Name `xml:"worksheet"`
	SheetData struct {
		Rows []struct {
			C []struct {
				T  string `xml:"t,attr"`
				V  string `xml:"v"`
				Is struct {
					T string `xml:"t"`
				} `xml:"is"`
			} `xml:"c"`
		} `xml:"row"`
	} `xml:"sheetData"`
}

func parseWorksheetXML(r io.Reader, sharedStrings []string) ([][]string, error) {
	var ws worksheetXML
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&ws); err != nil {
		return nil, err
	}

	var result [][]string
	for _, row := range ws.SheetData.Rows {
		var rowVals []string
		for _, cell := range row.C {
			val := cell.V
			if cell.T == "s" {
				if idx, err := strconv.Atoi(val); err == nil && idx >= 0 && idx < len(sharedStrings) {
					val = sharedStrings[idx]
				}
			} else if cell.T == "inlineStr" {
				val = cell.Is.T
			}
			rowVals = append(rowVals, val)
		}
		if len(rowVals) > 0 {
			result = append(result, rowVals)
		}
	}
	return result, nil
}

func (s *foodService) parseAndValidateRow(row []string, headerMap map[string]int, rowNum int) (CreateFoodRequest, []string) {
	var item CreateFoodRequest
	var errs []string

	getVal := func(keys ...string) string {
		for _, k := range keys {
			cleanKey := strings.ToLower(strings.TrimSpace(k))
			if idx, ok := headerMap[cleanKey]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			compact := strings.ReplaceAll(strings.ReplaceAll(cleanKey, " ", ""), "_", "")
			if idx, ok := headerMap[compact]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
		}
		return ""
	}

	parseFloat := func(field string, keys ...string) float64 {
		valStr := getVal(keys...)
		if valStr == "" {
			return 0
		}
		valStr = strings.ReplaceAll(valStr, ",", ".")
		v, err := strconv.ParseFloat(valStr, 64)
		if err != nil || v < 0 {
			errs = append(errs, fmt.Sprintf("%s must be a valid non-negative number (got '%s')", field, valStr))
			return 0
		}
		return math.Round(v*100) / 100
	}

	// Required fields: name, manufacturer, serving_size
	item.Name = getVal("name", "nama", "nama_makanan", "food_name", "produk")
	if item.Name == "" {
		errs = append(errs, "name is required")
	}

	item.Manufacturer = getVal("manufacturer", "produsen", "merk", "brand", "perusahaan")
	// manufacturer boleh kosong — akan dianggap sebagai "Tidak Diketahui"
	if item.Manufacturer == "" {
		item.Manufacturer = "Tidak Diketahui"
	}

	// Tentukan NutritionBasis dan ServingSize berdasarkan manufacturer
	// Hanya dua pilihan:
	//   - Produsen diketahui → PER_PACKAGE
	//   - Produsen tidak diketahui → PER_100G (Per 100 g BDD)
	item.NutritionBasis = determineNutritionBasis("", item.Manufacturer)
	if item.NutritionBasis == "PER_100G" {
		item.ServingSize = "Per 100 g BDD (Berat Dapat Dimakan)"
	} else {
		// Per kemasan: ambil dari kolom serving_size di Excel jika ada, sebagai informasi tambahan
		rawServing := getVal(
			"serving_size", "serving size", "servingsize",
			"ukuran_saji", "ukuran saji", "ukuransaji",
			"ukuran_porsi", "ukuran porsi", "ukuranporsi",
			"takaran_saji", "takaran saji", "takaransaji",
			"porsi", "serving",
		)
		if rawServing != "" {
			item.ServingSize = rawServing
		} else {
			item.ServingSize = "Per Kemasan"
		}
	}

	item.Barcode = getVal("barcode")
	item.Source = "excel_import"
	item.Status = "active"

	// Nutrition values
	item.EnergyKcal = parseFloat("energy_kcal", "energy_kcal", "energi", "kalori", "calories")
	item.ProteinG = parseFloat("protein_g", "protein_g", "protein")
	item.CarbohydrateG = parseFloat("carbohydrate_g", "carbohydrate_g", "karbohidrat_total", "carbs", "carbohydrate")
	item.FatG = parseFloat("fat_g", "fat_g", "lemak_total", "fat")
	item.SugarG = parseFloat("sugar_g", "sugar_g", "gula", "sugar")
	item.SodiumMg = parseFloat("sodium_mg", "sodium_mg", "natrium", "sodium")
	item.FiberG = parseFloat("fiber_g", "fiber_g", "serat_pangan", "fiber")
	item.SaturatedFatG = parseFloat("saturated_fat_g", "saturated_fat_g", "lemak_jenuh")

	// Daily values %DV
	item.EnergyPercentageDV = parseFloat("energy_percentage_dv", "energy_percentage_dv")
	item.ProteinPercentageDV = parseFloat("protein_percentage_dv", "protein_percentage_dv")
	item.CarbohydratePercentageDV = parseFloat("carbohydrate_percentage_dv", "carbohydrate_percentage_dv")
	item.FatPercentageDV = parseFloat("fat_percentage_dv", "fat_percentage_dv")
	item.SodiumPercentageDV = parseFloat("sodium_percentage_dv", "sodium_percentage_dv")

	// Label fields
	item.TotalFat = parseFloat("total_fat", "lemak_total", "total_fat")
	item.SaturatedFat = parseFloat("saturated_fat", "lemak_jenuh", "saturated_fat")
	item.Sodium = parseFloat("sodium", "natrium", "sodium")
	item.Protein = parseFloat("protein", "protein")
	item.TotalCarbohydrate = parseFloat("total_carbohydrate", "karbohidrat_total", "total_carbohydrate")
	item.DietaryFiber = parseFloat("dietary_fiber", "serat_pangan", "dietary_fiber")
	item.Energy = parseFloat("energy", "energi", "energy")

	return item, errs
}

func (s *foodService) ConfirmExcelImport(ctx context.Context, req ExcelImportConfirmRequest) (*ExcelImportConfirmResponse, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("no items provided to import")
	}

	s.log.Info("confirming food excel bulk import", zap.Int("items_count", len(req.Items)))

	var domainFoods []domain.FoodMaster
	for _, item := range req.Items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		status := item.Status
		if status == "" {
			status = "active"
		}
		manufacturer := strings.TrimSpace(item.Manufacturer)
		if manufacturer == "" {
			manufacturer = "Tidak Diketahui"
		}

		nutritionBasis := determineNutritionBasis("", manufacturer)
		var servingSize string
		if nutritionBasis == "PER_100G" {
			servingSize = "Per 100 g BDD (Berat Dapat Dimakan)"
		} else {
			servingSize = strings.TrimSpace(item.ServingSize)
			if servingSize == "" {
				servingSize = "Per Kemasan"
			}
		}

		domainFoods = append(domainFoods, domain.FoodMaster{
			BaseModel: domain.BaseModel{
				ID: uuid.NewString(),
			},
			Name:          strings.TrimSpace(item.Name),
			Manufacturer:  manufacturer,
			ServingSize:   servingSize,
			Calories:      item.EnergyKcal,
			EnergyKcal:    item.EnergyKcal,
			ProteinG:      item.ProteinG,
			CarbohydrateG: item.CarbohydrateG,
			FatG:          item.FatG,
			SugarG:        item.SugarG,
			SodiumMg:      item.SodiumMg,
			FiberG:        item.FiberG,
			SaturatedFatG: item.SaturatedFatG,

			EnergyPercentageDV:       item.EnergyPercentageDV,
			ProteinPercentageDV:      item.ProteinPercentageDV,
			CarbohydratePercentageDV: item.CarbohydratePercentageDV,
			FatPercentageDV:          item.FatPercentageDV,
			SodiumPercentageDV:       item.SodiumPercentageDV,

			TotalFat:          item.TotalFat,
			SaturatedFat:      item.SaturatedFat,
			Sodium:            item.Sodium,
			Protein:           item.Protein,
			TotalCarbohydrate: item.TotalCarbohydrate,
			DietaryFiber:      item.DietaryFiber,
			Energy:            item.Energy,

			NutritionBasis: nutritionBasis,
			Source:         "excel_import",
			Barcode:        strings.TrimSpace(item.Barcode),
			ImageURL:       strings.TrimSpace(item.ImageURL),
			Status:         status,
		})
	}

	inserted, err := s.repo.BulkInsertInBatches(ctx, domainFoods, 500)
	if err != nil {
		return nil, fmt.Errorf("bulk insert failed: %w", err)
	}

	return &ExcelImportConfirmResponse{
		SuccessCount: inserted,
		FailedCount:  len(req.Items) - inserted,
	}, nil
}

func (s *foodService) ExportFoods(ctx context.Context, query FoodFilterQuery, format string) ([]byte, string, string, error) {
	foods, err := s.repo.GetAllForExport(ctx, query)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to fetch foods for export: %w", err)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	_ = w.Write([]string{
		"name", "manufacturer", "serving_size", "energy_kcal", "protein_g",
		"carbohydrate_g", "fat_g", "sugar_g", "sodium_mg", "fiber_g",
		"saturated_fat_g", "barcode", "status",
	})

	for _, f := range foods {
		_ = w.Write([]string{
			f.Name, f.Manufacturer, f.ServingSize,
			fmt.Sprintf("%.2f", f.EnergyKcal), fmt.Sprintf("%.2f", f.ProteinG),
			fmt.Sprintf("%.2f", f.CarbohydrateG), fmt.Sprintf("%.2f", f.FatG),
			fmt.Sprintf("%.2f", f.SugarG), fmt.Sprintf("%.2f", f.SodiumMg),
			fmt.Sprintf("%.2f", f.FiberG), fmt.Sprintf("%.2f", f.SaturatedFatG),
			f.Barcode, f.Status,
		})
	}
	w.Flush()

	if strings.ToLower(format) == "xlsx" {
		return buf.Bytes(), "application/vnd.ms-excel", "master_data_makanan.csv", nil
	}

	return buf.Bytes(), "text/csv", "master_data_makanan.csv", nil
}

func (s *foodService) GetStats(ctx context.Context) (*FoodStatsResponse, error) {
	return s.repo.GetStats(ctx)
}

func determineNutritionBasis(explicitBasis, manufacturer string) string {
	b := strings.TrimSpace(strings.ToUpper(explicitBasis))
	if b == "PER_100G" || b == "PER_SERVING" || b == "PER_PACKAGE" {
		return b
	}
	m := strings.TrimSpace(manufacturer)
	if m == "" || strings.EqualFold(m, "tidak diketahui") || strings.EqualFold(m, "tidak ada") || strings.EqualFold(m, "-") {
		return "PER_100G"
	}
	return "PER_PACKAGE"
}
