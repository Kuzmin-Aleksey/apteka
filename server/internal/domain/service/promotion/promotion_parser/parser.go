package promotion_parser

import (
	"github.com/xuri/excelize/v2"
	"golang.org/x/net/context"
	"io"
	"server/internal/domain/entity"
	"server/pkg/contextx"
	"server/pkg/failure"
	"strings"
	"unicode"
)

func ParseDoc(ctx context.Context, doc io.Reader) ([]entity.Promotion, error) {
	f, err := excelize.OpenReader(doc)
	if err != nil {
		return nil, failure.NewInvalidFileError(err.Error())
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, failure.NewInvalidFileError("no sheets in file")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, failure.NewInvalidFileError(err.Error())
	}

	if len(rows) == 0 {
		return nil, failure.NewInvalidFileError("no rows in sheet")
	}

	var promotions []entity.Promotion

	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		if len(row[0]) == 0 || len(row[1]) == 0 || len(row[2]) == 0 {
			continue
		}

		productCodes := parseProductCodes(row[0])
		prodName := strings.TrimSpace(row[1])
		discount, isPercent := parseDiscount(strings.TrimSpace(row[2]))

		for _, productCode := range productCodes {
			promotions = append(promotions, entity.Promotion{
				ProductCode: productCode,
				ProductName: prodName,
				Discount:    discount,
				IsPercent:   isPercent,
			})
		}

		if len(productCodes) == 0 {
			contextx.GetLoggerOrDefault(ctx).Warn("ParseDoc - product codes not found", "row", row)
		}
	}

	return promotions, nil
}

func parseDiscount(s string) (int, bool) {
	if strings.HasSuffix(s, "%") {
		return parseInt(strings.TrimSuffix(s, "%")), true
	}

	return parseInt(s), false
}

func parseProductCodes(s string) []int {
	var codes []int

	currentCode := 0

	for _, c := range s {
		if unicode.IsDigit(c) {
			currentCode = currentCode*10 + int(c-'0')
		} else if currentCode != 0 {
			codes = append(codes, currentCode)
			currentCode = 0
		}
	}
	if currentCode != 0 {
		codes = append(codes, currentCode)
	}

	return codes
}

func parseInt(s string) int {
	var n int
	for _, c := range s {
		if unicode.IsDigit(c) {
			n = n*10 + int(c-'0')
		}
	}

	return n
}
