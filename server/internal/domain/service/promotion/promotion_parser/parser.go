package promotion_parser

import (
	"github.com/xuri/excelize/v2"
	"io"
	"server/internal/domain/entity"
	"server/pkg/failure"
	"strings"
	"unicode"
)

func ParseDoc(doc io.Reader) ([]entity.Promotion, error) {
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

		discount, isPercent := parseDiscount(strings.TrimSpace(row[2]))

		promotions = append(promotions, entity.Promotion{
			ProductCode: parseInt(row[0]),
			ProductName: strings.TrimSpace(row[1]),
			Discount:    discount,
			IsPercent:   isPercent,
		})
	}

	return promotions, nil
}

func parseDiscount(s string) (int, bool) {
	if strings.HasSuffix(s, "%") {
		return parseInt(strings.TrimSuffix(s, "%")), true
	}

	return parseInt(s), false
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
