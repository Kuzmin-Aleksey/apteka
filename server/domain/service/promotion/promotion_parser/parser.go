package promotion_parser

import (
	"github.com/xuri/excelize/v2"
	"io"
	"server/domain/models"
	"unicode"
)

func ParseDoc(doc io.Reader) ([]models.Promotion, error) {
	f, err := excelize.OpenReader(doc)
	if err != nil {
		return nil, models.NewError(models.ErrInvalidFile, err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, models.NewError(models.ErrInvalidFile, "no sheets in file")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, models.NewError(models.ErrInvalidFile, err)
	}

	if len(rows) == 0 {
		return nil, models.NewError(models.ErrInvalidFile, "no rows in sheet")
	}

	var promotions []models.Promotion

	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		if len(row[0]) == 0 || len(row[1]) == 0 || len(row[2]) == 0 {
			continue
		}
		promotions = append(promotions, models.Promotion{
			ProductCode: parseInt(row[0]),
			ProductName: row[1],
			Discount:    parseInt(row[2]),
		})
	}

	return promotions, nil
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
