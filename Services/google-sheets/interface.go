package google_sheets

import "main.go/internal/domain"

type SheetsInterface interface {
	AddTask(spreadsheetId, sheetRange string, values [][]interface{}) error
	InitGoogleSheetsClient() error
	RenamingCell(spreadSheetId, sheetId, value string) error
	CreateSpreadsheet(title string) (string, error)
	GetCompletedTask(spreadsheetId string) ([]domain.Task, error)
	GetUnCompletedTasks(spreadsheetId string) ([]domain.Task, []domain.Task, error)
	ClearTask(spreadsheetId string, sheetRange int64) error
}
