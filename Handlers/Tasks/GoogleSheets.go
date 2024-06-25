package Tasks

import (
	"fmt"
	"github.com/skwizi4/lib/ErrChan"
	"github.com/skwizi4/lib/logs"
	"gopkg.in/tucnak/telebot.v2"
	googleSheets "main.go/internal/Services/google-sheets"
	"main.go/internal/domain"
	"time"
)

const (
	ErrorConnectingToGoogleSheetsApi = "Ошибка в подключении к GoogleSheetsApi"
)

type Handler struct {
	bot                                *telebot.Bot
	ErrChan                            *ErrChan.ErrorChannel
	GoogleSheetsApi                    googleSheets.SheetsInterface
	logger                             logs.GoLogger
	ProcessingAddingTasks              *domain.ProcessingAddingTasks
	ProcessingChangingDataTasks        *domain.ProcessingChangingDataTasks
	ProcessingChangingDescriptionTasks *domain.ProcessingChangingDescriptionTasks
	ProcessingChangingStatusTasks      *domain.ProcessingChangingStatusTasks
	ProcessingRenamingTasks            *domain.ProcessingRenamingTasks
	Users                              *domain.ActiveUsers
}

func New(
	bot *telebot.Bot,
	ErrChan *ErrChan.ErrorChannel,
	GoogleSheetsApi googleSheets.SheetsInterface,
	logger logs.GoLogger,
	ProcessingAddingTasks *domain.ProcessingAddingTasks,
	ProcessingChangingDataTasks *domain.ProcessingChangingDataTasks,
	ProcessingChangingDescriptionTasks *domain.ProcessingChangingDescriptionTasks,
	ProcessingChangingStatusTasks *domain.ProcessingChangingStatusTasks,
	ProcessingRenamingTasks *domain.ProcessingRenamingTasks,
	Users *domain.ActiveUsers,
) Handler {
	return Handler{
		bot:                                bot,
		ErrChan:                            ErrChan,
		GoogleSheetsApi:                    GoogleSheetsApi,
		logger:                             logger,
		ProcessingAddingTasks:              ProcessingAddingTasks,
		ProcessingChangingDataTasks:        ProcessingChangingDataTasks,
		ProcessingChangingDescriptionTasks: ProcessingChangingDescriptionTasks,
		ProcessingChangingStatusTasks:      ProcessingChangingStatusTasks,
		ProcessingRenamingTasks:            ProcessingRenamingTasks,
		Users:                              Users,
	}
}

func (h Handler) sendLoginPrompt(msg *telebot.Message) error {
	if _, err := h.bot.Send(msg.Sender, "Вы не авторизированы, введите /login и пройдите авторизацию пожалуйста"); err != nil {
		return err
	}
	return nil
}

//---------------------------------------  GetTask Handler--------------------------------------------------------------------------------

func (h Handler) GetCompletedTask(ChatId int64, Sender *telebot.User) ([]domain.Task, error) {
	user := h.Users.GetOrCreate(ChatId, Sender.Username)
	tasks, err := h.GoogleSheetsApi.GetCompletedTask(user.SpreadSheetID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
func (h Handler) GetUnCompletedTask(ChatId int64, SenderUsername *telebot.User) ([]domain.Task, error) {
	user := h.Users.GetOrCreate(ChatId, SenderUsername.Username)
	allTasks, tasks, err := h.GoogleSheetsApi.GetUnCompletedTasks(user.SpreadSheetID)
	if err != nil {
		return nil, err
	}
	h.Users.AddTasks(ChatId, allTasks)
	return tasks, nil
}

//---------------------------------------  ChangeTask handlers--------------------------------------------------------------------------------

func (h Handler) RenameTask(msg *telebot.Message) error {
	h.ProcessingRenamingTasks.AddNewName(msg.Chat.ID, msg.Text)
	Process := h.ProcessingRenamingTasks.GetOrCreate(msg.Chat.ID)
	if errSheets := h.GoogleSheetsApi.RenamingCell(Process.SpreadSheetId, Process.SpreadSheetCellId, Process.RenameName); errSheets != nil {
		if _, err := h.bot.Send(msg.Sender, ErrorConnectingToGoogleSheetsApi); err != nil {
			return err
		}
		return errSheets
	}
	h.ProcessingRenamingTasks.Delete(msg.Chat.ID)
	return nil
}

func (h Handler) ChangeDescription(msg *telebot.Message) error {
	h.ProcessingChangingDescriptionTasks.AddNewDescription(msg.Chat.ID, msg.Text)
	Process := h.ProcessingChangingDescriptionTasks.GetOrCreate(msg.Chat.ID)
	if errSheets := h.GoogleSheetsApi.RenamingCell(Process.SpreadSheetId, Process.SpreadSheetCellId, Process.ChangeDescription); errSheets != nil {
		if _, err := h.bot.Send(msg.Sender, ErrorConnectingToGoogleSheetsApi); err != nil {
			return err
		}
		return errSheets
	}
	h.ProcessingChangingDescriptionTasks.Delete(msg.Chat.ID)
	return nil
}
func (h Handler) ChangeData(msg *telebot.Message) error {
	h.ProcessingChangingDataTasks.AddNewData(msg.Chat.ID, msg.Text)
	Process := h.ProcessingChangingDataTasks.GetOrCreate(msg.Chat.ID)
	if errSheets := h.GoogleSheetsApi.RenamingCell(Process.SpreadSheetId, Process.SpreadSheetCellId, Process.ChangeData); errSheets != nil {
		if _, err := h.bot.Send(msg.Sender, ErrorConnectingToGoogleSheetsApi); err != nil {
			return err
		}
		return errSheets
	}

	h.ProcessingChangingDataTasks.Delete(msg.Chat.ID)
	return nil
}

//---------------------------------------  AddTask Handler--------------------------------------------------------------------------------

func (h Handler) AddTask(msg *telebot.Message) error {
	if !h.Users.IfExist(msg.Chat.ID) {
		if err := h.sendLoginPrompt(msg); err != nil {
			return err
		}
	}
	Process := h.ProcessingAddingTasks.GetOrCreate(msg.Chat.ID)
	User := h.Users.GetOrCreate(msg.Chat.ID, msg.Sender.Username)
	switch Process.Step {
	case domain.AddingTaskNameStep:
		if msg.Text != "/exit" {
			h.ProcessingAddingTasks.SetTaskName(msg.Chat.ID, msg.Text)
			if _, err := h.bot.Send(msg.Sender, "Введите описание задачи:"); err != nil {
				h.ProcessingAddingTasks.Delete(msg.Chat.ID)
				return err
			}
			h.ProcessingAddingTasks.UpdateStep(msg.Chat.ID, domain.AddingTaskDescriptionStep)
			return nil
		}
		h.ProcessingAddingTasks.Delete(msg.Chat.ID)
		return nil
	case domain.AddingTaskDescriptionStep:
		if msg.Text != "/exit" {
			h.ProcessingAddingTasks.SetTaskDescription(msg.Chat.ID, msg.Text)
			if _, err := h.bot.Send(msg.Sender, "Введите время до которого вам нужно сделать задачу"); err != nil {
				h.ProcessingAddingTasks.Delete(msg.Chat.ID)
				return err
			}
			h.ProcessingAddingTasks.UpdateStep(msg.Chat.ID, domain.AddingTaskDateStep)
			return nil
		}
		h.ProcessingAddingTasks.Delete(msg.Chat.ID)
		return nil
	case domain.AddingTaskDateStep:
		if msg.Text != "/exit" {
			DataNow := time.Now().Format("2006-01-02 15:04:05")
			h.ProcessingAddingTasks.SetTaskDate(msg.Chat.ID, msg.Text+"*"+DataNow+"*")
			if len(User.Tasks) != 0 {
				if err := h.ProcessingAddingTasks.SetRangeSheet(msg.Chat.ID, User.Tasks[len(User.Tasks)-1].Range); err != nil {
					return err
				}
			} else if len(User.Tasks) == 0 {
				if err := h.ProcessingAddingTasks.SetRangeSheet(msg.Chat.ID, "Sheet1!A1:D1"); err != nil {
					return err
				}
			}
			Process = h.ProcessingAddingTasks.GetOrCreate(msg.Chat.ID)
			str := fmt.Sprintf("Sheet1!A%s:D%s", Process.Range, Process.Range)
			values := [][]interface{}{
				{Process.TaskName, Process.TaskDescription, Process.TaskDate, "Не выполнено"},
			}
			if err := h.GoogleSheetsApi.AddTask(Process.SpreadSheetId, str, values); err != nil {
				return err
			}
			if _, err := h.bot.Send(msg.Sender, "Операция выполнена успешно, задача добавлена. Введите /GetTasks для просмотра новой задачи"); err != nil {
				return err
			}
			h.ProcessingAddingTasks.Delete(msg.Chat.ID)

		}

	}
	return nil
}
