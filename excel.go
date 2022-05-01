package csfutil

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"strconv"
)

const (
	sheet            = "Sheet1"
	labelColumn      = "A"
	valueColumn      = "B"
	extraValueColumn = "C"
)

func ExportExcel(csf *CSFUtil, output string) error {
	f := excelize.NewFile()
	defer f.Close()

	for i := 0; i < len(csf.Order); i++ {
		row := strconv.Itoa(i + 1)

		if err := f.SetCellValue(sheet, labelColumn+row, csf.Values[csf.Order[i]].Label.ValueString()); err != nil {
			return err
		}

		if err := f.SetCellValue(sheet, valueColumn+row, csf.Values[csf.Order[i]].Value.ValueString()); err != nil {
			return err
		}

		if csf.Values[csf.Order[i]].Value.HaveExtra {
			if err := f.SetCellValue(sheet, extraValueColumn+row, csf.Values[csf.Order[i]].Value.ExtraValueString()); err != nil {
				return err
			}
		}
	}

	return f.SaveAs(output)
}

func ImportExcel(input, output string) error {
	f, err := excelize.OpenFile(input)
	if err != nil {
		return err
	}
	defer f.Close()

	csf, err := Open(output)
	if err != nil {
		return err
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return err
	}

	if len(rows) < 1 {
		return fmt.Errorf("no data to read")
	}

	for i, row := range rows {
		switch len(row) {
		case 1:
			csf.WriteLabelValue(NewLabelValue(row[0], ""), false)
		case 2:
			csf.WriteLabelValue(NewLabelValue(row[0], row[1]), false)
		case 3:
			csf.WriteLabelValue(NewLabelValue(row[0], row[1], row[2]), false)
		default:
			return fmt.Errorf("wrong column count, want 2~3, got %d, at row: %d", len(row), i)
		}
	}

	return csf.Save()
}
