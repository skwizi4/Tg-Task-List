package google_sheets

import (
	"context"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"log"
	"main.go/internal/domain"
	"os"
)

func (c *GoogleSheetsClient) InitGoogleSheetsClient() error {
	ctx := context.Background()
	b, err := os.ReadFile("C:\\golang\\src\\TG-ToDoList\\to-do-list-424109-87214648eef9.json")
	if err != nil {
		return err
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return err
	}

	client := config.Client(ctx)
	c.srv, err = sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}
	return nil
}

func (c GoogleSheetsClient) AddTask(spreadsheetId, sheetRange string, values [][]interface{}) error {
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, values...)
	if _, err := c.srv.Spreadsheets.Values.Update(spreadsheetId, sheetRange, &vr).ValueInputOption("RAW").Do(); err != nil {
		return err
	}
	return nil
}

func (c *GoogleSheetsClient) CreateSpreadsheet(title string) (string, error) {
	ctx := context.Background()
	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
	}
	if c.srv == nil {
		return "", nil
	}
	createSpreadsheetResp, err := c.srv.Spreadsheets.Create(spreadsheet).Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return createSpreadsheetResp.SpreadsheetId, nil
}
func (c GoogleSheetsClient) GetCompletedTask(spreadsheetId string) ([]domain.Task, error) {
	resp, err := c.srv.Spreadsheets.Values.Get(spreadsheetId, sheetRange).Do()
	if err != nil {
		log.Fatalf("Не удалось получить данные из таблицы: %v", err)
		return nil, err
	}
	var tasks []domain.Task
	if len(resp.Values) > 0 {
		for i, row := range resp.Values {
			if len(row) == 0 || (len(row) == 1 && row[0] == "") {
				continue
			}
			task := domain.Task{
				Name:        getStringFromCell(row, TaskName),
				Description: getStringFromCell(row, TaskDescription),
				Date:        getStringFromCell(row, TaskDate),
				Status:      getStringFromCell(row, TaskStatus),
				Range:       fmt.Sprintf("Sheet1!A%d:D%d", i+1, i+1),
			}
			if task.Status == "Выполнено" {
				tasks = append(tasks, task)
			}
		}
	}
	return tasks, nil
}
func (c GoogleSheetsClient) GetUnCompletedTasks(spreadsheetId string) ([]domain.Task, []domain.Task, error) {
	resp, err := c.srv.Spreadsheets.Values.Get(spreadsheetId, sheetRange).Do()
	if err != nil {
		log.Fatalf("Не удалось получить данные из таблицы: %v", err)
		return nil, nil, err
	}
	var tasks []domain.Task
	var AllTasks []domain.Task
	if len(resp.Values) > 0 {
		for i, row := range resp.Values {
			if len(row) == 0 || (len(row) == 1 && row[0] == "") {
				continue
			}
			task := domain.Task{
				Name:        getStringFromCell(row, TaskName),
				Description: getStringFromCell(row, TaskDescription),
				Date:        getStringFromCell(row, TaskDate),
				Status:      getStringFromCell(row, TaskStatus),
				Range:       fmt.Sprintf("Sheet1!A%d:D%d", i+1, i+1),
			}
			AllTasks = append(AllTasks, task)
			if task.Status == "Не выполнено" {
				tasks = append(tasks, task)
			}
		}
	}
	return AllTasks, tasks, nil
}

func (c GoogleSheetsClient) RenamingCell(spreadSheetId, sheetId, value string) error {
	var vr sheets.ValueRange
	myval := []interface{}{value}
	vr.Values = append(vr.Values, myval)
	_, err := c.srv.Spreadsheets.Values.Update(spreadSheetId, sheetId, &vr).ValueInputOption("RAW").Do()
	if err != nil {
		return err
	}
	return nil
}
func (c GoogleSheetsClient) ClearTask(spreadsheetId string, sheetRange int64) error {

	gridRange := &sheets.GridRange{
		StartRowIndex:    sheetRange - 1,
		EndRowIndex:      sheetRange,
		StartColumnIndex: 0,
		EndColumnIndex:   4,
	}

	deleteRangeRequest := &sheets.DeleteRangeRequest{
		Range:          gridRange,
		ShiftDimension: "ROWS",
	}

	request := &sheets.Request{
		DeleteRange: deleteRangeRequest,
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{request},
	}

	_, err := c.srv.Spreadsheets.BatchUpdate(spreadsheetId, batchUpdateRequest).Do()
	if err != nil {
		return err
	}
	return nil
}

//	func parseRange(sheetRange string) (startCell, endCell string) {
//		parts := strings.Split(sheetRange, ":")
//		return parts[0], parts[1]
//	}
//
//	func cellToGrid(cell string) (row, col int) {
//		col = 0
//		row = 0
//		for i := 0; i < len(cell); i++ {
//			if cell[i] >= 'A' && cell[i] <= 'Z' {
//				col = col*26 + int(cell[i]-'A'+1)
//			} else if cell[i] >= '0' && cell[i] <= '9' {
//				row = row*10 + int(cell[i]-'0')
//			}
//		}
//		return row - 1, col - 1
//	}
func getStringFromCell(row []interface{}, index int) string {
	if len(row) > index {
		if val, ok := row[index].(string); ok {
			return val
		}
	}
	return ""
}
