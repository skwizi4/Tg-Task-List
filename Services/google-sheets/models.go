package google_sheets

import (
	"google.golang.org/api/sheets/v4"
)

const (
	sheetRange      = "A1:D49"
	TaskName        = 0
	TaskDescription = 1
	TaskDate        = 2
	TaskStatus      = 3
)

type GoogleSheetsClient struct {
	srv *sheets.Service
}
