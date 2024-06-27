package domain

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

const (
	RegistrationStepStart    = "RegistrationStepStart"
	RegistrationStepName     = "StepName"
	RegistrationStepPassword = "StepPassword"

	LoginStepStart    = "LoginStepStart"
	LoginStepPassword = "StepPassword"

	AddingTaskNameStep        = "AddingTaskNameStep"
	AddingTaskDescriptionStep = "AddingTaskDescription"
	AddingTaskDateStep        = "AddingTaskDate"
)

// ProcessAddTask---------------------------------------------------------------------------------------------------------------------------------------------------------------

type ProcessAddTask struct {
	SpreadSheetId   string
	Range           string
	TaskName        string
	TaskDescription string
	TaskDate        string
	Step            string
	ChatID          int64
}

type ProcessingAddingTasks []ProcessAddTask

// ProcessingRenamingTask---------------------------------------------------------------------------------------------------------------------------------------------------------------

type ProcessingRenamingTask struct {
	SpreadSheetId     string
	SpreadSheetCellId string
	RenameName        string
	Step              string
	ChatID            int64
}
type ProcessingRenamingTasks []ProcessingRenamingTask

// ProcessingChangingDescriptionTasks---------------------------------------------------------------------------------------------------------------------------------------------------------------!

type ProcessingChangingDescriptionTask struct {
	SpreadSheetId     string
	SpreadSheetCellId string
	ChangeDescription string
	ChatID            int64
}
type ProcessingChangingDescriptionTasks []ProcessingChangingDescriptionTask

// ProcessingChangingDataTask---------------------------------------------------------------------------------------------------------------------------------------------------------------!

type ProcessingChangingDataTask struct {
	SpreadSheetId     string
	SpreadSheetCellId string
	ChangeData        string
	ChatID            int64
}
type ProcessingChangingDataTasks []ProcessingChangingDataTask

// ProcessingChangingStatusTask!---------------------------------------------------------------------------------------------------------------------------------------------------------------!

type ProcessingChangingStatusTask struct {
	SpreadSheetId      string
	SpreadSheetCellId  string
	TaskStatusSwitcher bool
	ChatID             int64
}
type ProcessingChangingStatusTasks []ProcessingChangingStatusTask

//ToDoList !---------------------------------------------------------------------------------------------------------------------------------------------------------------!

type Task struct {
	Range       string
	Name        string
	Description string
	Status      string
	Date        string
}
type ToDoList struct {
	Tasks  Task
	ChatID int64
}
type GettingTables []ToDoList

// ProcessingRegistrationUser!---------------------------------------------------------------------------------------------------------------------------------------------------------------!

type User struct {
	Name                     string  `json:"name" bson:"name"`
	Password                 string  `json:"password" bson:"password"`
	SpreadSheetID            string  `json:"spread_sheet_id" bson:"spread_sheet_id"`
	FrequencyOfNotifications float64 `json:"frequency_of_notifications" bson:"frequency_of_notifications"`
	IsSendNotification       bool    `json:"is_send_notification" bson:"is_send_notification"`
	ChatID                   int64   `json:"chat_id" bson:"chat_id"`
	TelegramID               int64   `json:"telegram_id" bson:"_id"`
	Tasks                    []Task
}

type ProcessingRegistrationUser struct {
	Step string
	User User
}

type ProcessingRegistrationUsers []ProcessingRegistrationUser

// ProcessLoginUser!---------------------------------------------------------------------------------------------------------------------------------------------------------------!

type ProcessLoginUser struct {
	ChatID   int64
	Step     string
	Username string
}
type ProcessLoginUsers []ProcessLoginUser

// FindUserIndex -----------------------------------------------------------------------------------------------------------------------------------

func FindUserIndex(users interface{}, chatID int64) int {
	v := reflect.ValueOf(users)
	if v.Kind() != reflect.Slice {
		return -1
	}

	for i := 0; i < v.Len(); i++ {
		user := v.Index(i)
		chatIDField := user.FieldByName("ChatID")
		if !chatIDField.IsValid() || chatIDField.Kind() != reflect.Int64 {
			continue
		}
		if chatIDField.Int() == chatID {
			return i
		}
	}

	return -1
}

// ProcessingRegistrationUsers!---------------------------------------------------------------------------------------------------------------------------------------------------------------!

func (p *ProcessingRegistrationUsers) SetTelegramID(ChatID int64, value int64) {
	for i, user := range *p {
		if user.User.ChatID == ChatID {
			(*p)[i].User.TelegramID = value
		}
	}
}

func (p *ProcessingRegistrationUsers) SetSendNotifications(ChatID int64, value bool) {
	for i, user := range *p {
		if user.User.ChatID == ChatID {
			(*p)[i].User.IsSendNotification = value
		}
	}
}

func (p *ProcessingRegistrationUsers) SetNotificationFrequency(ChatID int64, Frequency float64) {
	for i, user := range *p {
		if user.User.ChatID == ChatID {
			(*p)[i].User.FrequencyOfNotifications = Frequency
		}
	}
}

func (p *ProcessingRegistrationUsers) SetSpreadSheetID(ChatID int64, spreadSheetID string) {

	for i, user := range *p {
		if user.User.ChatID == ChatID {
			(*p)[i].User.SpreadSheetID = spreadSheetID
		}
	}
}

func (p *ProcessingRegistrationUsers) UpdateRegistrationStep(ChatID int64, step string) {
	for i, user := range *p {
		if user.User.ChatID == ChatID {
			(*p)[i].Step = step
		}
	}
}

func (p *ProcessingRegistrationUsers) SetName(ChatID int64, name string) {
	for i, user := range *p {
		if user.User.ChatID == ChatID {
			(*p)[i].User.Name = name
		}
	}
}

func (p *ProcessingRegistrationUsers) SetPassword(ChatID int64, password string) {
	for i, user := range *p {
		if user.User.ChatID == ChatID {
			(*p)[i].User.Password = password
		}
	}
}

func (p *ProcessingRegistrationUsers) IfExist(ChatID int64) bool {
	for _, user := range *p {
		if user.User.ChatID == ChatID {
			return true
		}
	}
	return false
}

func (p *ProcessingRegistrationUsers) GetOrCreate(ChatID int64) ProcessingRegistrationUser {
	for _, usr := range *p {
		if usr.User.ChatID == ChatID {
			return usr
		}
	}
	NewUser := ProcessingRegistrationUser{
		Step: RegistrationStepStart,
	}
	NewUser.User.ChatID = ChatID
	*p = append(*p, NewUser)
	return NewUser
}

func (p *ProcessingRegistrationUsers) Delete(ChatID int64) {
	for i, user := range *p {
		if user.User.ChatID == ChatID {
			*p = append((*p)[:i], (*p)[i+1:]...)
		}
	}
}

// ProcessLoginUsers !------------------------------------------------------------------------------------------------------------!//

func (p *ProcessLoginUsers) UpdateStep(Step string, chatID int64) {
	if idx := FindUserIndex(*p, chatID); idx != -1 {
		(*p)[idx].Step = Step
	}
}
func (p *ProcessLoginUsers) IfExist(ChatID int64) bool {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		return true
	}
	return false
}
func (p *ProcessLoginUsers) GetOrCreate(ChatID int64) ProcessLoginUser {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		return (*p)[idx]
	}
	NewUser := ProcessLoginUser{
		ChatID: ChatID,
		Step:   LoginStepStart,
	}
	*p = append(*p, NewUser)
	return NewUser
}

func (p *ProcessLoginUsers) Unmarshal(data []byte) (User, error) {
	var UserStruct User
	if err := json.Unmarshal(data, &UserStruct); err != nil {
		return User{}, err
	}
	return UserStruct, nil
}
func (p *ProcessLoginUsers) Delete(ChatID int64) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		*p = append((*p)[:idx], (*p)[idx+1:]...)
	}
}

// ProcessingRenamingTasks!---------------------------------------------------------------------------------------------------------------------------------------------------------------!

func (p *ProcessingRenamingTasks) AddNewName(ChatID int64, NewName string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].RenameName = NewName
	}
}

func (p *ProcessingRenamingTasks) IfExist(ChatID int64) bool {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		return true
	}
	return false
}
func (p *ProcessingRenamingTasks) GetOrCreate(ChatId int64) ProcessingRenamingTask {
	if idx := FindUserIndex(*p, ChatId); idx != -1 {
		return (*p)[idx]
	}
	NewProcess := ProcessingRenamingTask{
		ChatID: ChatId,
	}
	*p = append(*p, NewProcess)
	return NewProcess
}
func (p *ProcessingRenamingTasks) SetSpreadSheetId(ChatId int64, SpreadSheetId string) {
	if idx := FindUserIndex(*p, ChatId); idx != -1 {
		(*p)[idx].SpreadSheetId = SpreadSheetId
	}
}

func (p *ProcessingRenamingTasks) AddSpreadSheetCellId(ChatID int64, cellId string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].SpreadSheetCellId = cellId
	}
}
func (p *ProcessingRenamingTasks) Delete(ChatID int64) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		*p = append((*p)[:idx], (*p)[idx+1:]...)
	}
}

// ProcessingChangingDescriptionTasks-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

func (p *ProcessingChangingDescriptionTasks) GetOrCreate(ChatID int64) ProcessingChangingDescriptionTask {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		return (*p)[idx]
	}
	NewProcess := ProcessingChangingDescriptionTask{
		ChatID: ChatID,
	}
	*p = append(*p, NewProcess)
	return NewProcess
}
func (p *ProcessingChangingDescriptionTasks) SetSpreadSheetId(ChatID int64, SpreadSheetId string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].SpreadSheetId = SpreadSheetId
	}
}
func (p *ProcessingChangingDescriptionTasks) IfExist(ChatID int64) bool {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		return true
	}
	return false
}
func (p *ProcessingChangingDescriptionTasks) AddNewDescription(ChatID int64, NewDescription string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].ChangeDescription = NewDescription
	}
}

func (p *ProcessingChangingDescriptionTasks) AddSpreadSheetCellId(ChatID int64, cellId string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].SpreadSheetCellId = cellId
	}
}
func (p *ProcessingChangingDescriptionTasks) Delete(ChatID int64) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		*p = append((*p)[:idx], (*p)[idx+1:]...)
	}
}

// ProcessingChangingDataTasks-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

func (p *ProcessingChangingDataTasks) GetOrCreate(ChatID int64) ProcessingChangingDataTask {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		return (*p)[idx]
	}
	NewProcess := ProcessingChangingDataTask{
		ChatID: ChatID,
	}
	*p = append(*p, NewProcess)
	return NewProcess
}
func (p *ProcessingChangingDataTasks) SetSpreadSheetId(ChatID int64, SpreadSheetId string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].SpreadSheetId = SpreadSheetId
	}
}
func (p *ProcessingChangingDataTasks) IfExist(ChatID int64) bool {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		return true
	}
	return false
}
func (p *ProcessingChangingDataTasks) AddNewData(ChatID int64, NewData string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].ChangeData = NewData
	}
}

func (p *ProcessingChangingDataTasks) AddSpreadSheetCellId(ChatID int64, cellId string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].SpreadSheetCellId = cellId
	}
}
func (p *ProcessingChangingDataTasks) Delete(ChatID int64) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		*p = append((*p)[:idx], (*p)[idx+1:]...)
	}
}

//Processing Adding Task-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

func (p *ProcessingAddingTasks) GetOrCreate(ChatID int64) ProcessAddTask {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		return (*p)[idx]
	}
	NewProcess := ProcessAddTask{
		ChatID: ChatID,
	}
	*p = append(*p, NewProcess)
	return NewProcess
}
func (p *ProcessingAddingTasks) IfExist(ChatID int64) bool {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		return true
	}
	return false
}
func (p *ProcessingAddingTasks) SetTaskName(ChatID int64, taskName string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].TaskName = taskName
	}

}

func (p *ProcessingAddingTasks) SetTaskDescription(ChatID int64, taskDescription string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].TaskDescription = taskDescription
	}
}
func (p *ProcessingAddingTasks) SetTaskDate(ChatID int64, taskDate string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].TaskDate = taskDate
	}
}
func (p *ProcessingAddingTasks) SetSpreadSheetId(ChatID int64, SpreadSheetId string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].SpreadSheetId = SpreadSheetId
	}
}
func (p *ProcessingAddingTasks) Delete(ChatID int64) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		*p = append((*p)[:idx], (*p)[idx+1:]...)
	}
}
func (p *ProcessingAddingTasks) SetRangeSheet(ChatID int64, taskRange string) error {
	parts := strings.Split(taskRange, ":")
	Number := strings.Split(parts[0], "A")
	num, err := strconv.Atoi(Number[1])
	if err != nil {
		return err
	}
	num++
	RangeSheetId := strconv.Itoa(num)
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].Range = RangeSheetId
	}
	return nil
}
func (p *ProcessingAddingTasks) UpdateStep(ChatID int64, step string) {
	if idx := FindUserIndex(*p, ChatID); idx != -1 {
		(*p)[idx].Step = step
	}

}
