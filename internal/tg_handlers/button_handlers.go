package tg_handlers

import (
	"errors"
	"fmt"
	"gopkg.in/tucnak/telebot.v2"
	"main.go/internal/domain"
	"strconv"
)

func (h Handler) HandleAddTaskButton(c *telebot.Callback) error {
	if _, err := h.Users.Get(c.Sender.ID); err != nil {
		if _, err := h.tgbot.Send(c.Sender, "Вы не авторизованы, введите /login и пройдите авторизацию пожалуйста"); err != nil {
			return err
		}
		return errors.New("user isn't login")
	}
	User, err := h.Users.Get(c.Sender.ID)
	if err != nil {
		return err
	}
	if _, err = h.tgbot.Send(c.Sender, "Введите название вашей задачи"); err != nil {
		h.errChan.HandleError(err)
		return err
	}
	h.ProcessingAddingTask.GetOrCreate(c.Message.Chat.ID)
	h.ProcessingAddingTask.UpdateStep(c.Message.Chat.ID, domain.AddingTaskNameStep)
	h.ProcessingAddingTask.SetSpreadSheetId(c.Message.Chat.ID, User.SpreadSheetID)
	return nil
}

func (h Handler) HandleChangeNotificationsFrequency(c *telebot.Callback) error {
	user, err := h.Mongo.Get(c.Sender.ID)
	if err != nil {
		return err
	}
	Frequency, err := strconv.ParseFloat(c.Data, 64)
	if err != nil {
		return err
	}
	user.FrequencyOfNotifications = Frequency
	if err = h.Mongo.Update(c.Sender.ID, *user); err != nil {
		return err
	}
	if _, err = h.tgbot.Send(c.Sender, "Операция выполнена успешно"); err != nil {
		return err
	}
	err = h.Users.Update(c.Sender.ID, *user)
	return err
}

func (h Handler) HandleChangingTaskDataButton(c *telebot.Callback) error {
	User, err := h.Users.Get(c.Sender.ID)
	if err != nil {
		return err
	}
	if _, err = h.tgbot.Send(c.Sender, "Введите то на что вы хотите заменить дату выполнения вашей задачи"); err != nil {
		h.errChan.HandleError(err)
		return err
	}
	h.ProcessingChangingDataTasks.GetOrCreate(c.Message.Chat.ID)
	h.ProcessingChangingDataTasks.SetSpreadSheetId(c.Message.Chat.ID, User.SpreadSheetID)
	h.ProcessingChangingDataTasks.AddSpreadSheetCellId(c.Message.Chat.ID, c.Data)
	return nil
}

func (h Handler) HandleChangingTaskDescriptionButton(c *telebot.Callback) error {
	User, err := h.Users.Get(c.Sender.ID)
	if err != nil {
		return err
	}
	if _, err = h.tgbot.Send(c.Sender, "Введите то на что вы хотите заменить описание вашей задачи"); err != nil {
		h.errChan.HandleError(err)
		return err
	}
	h.ProcessingChangingDescriptionTasks.GetOrCreate(c.Message.Chat.ID)
	h.ProcessingChangingDescriptionTasks.SetSpreadSheetId(c.Message.Chat.ID, User.SpreadSheetID)
	h.ProcessingChangingDescriptionTasks.AddSpreadSheetCellId(c.Message.Chat.ID, c.Data)
	return nil
}

func (h Handler) HandleDeleteTaskButton(c *telebot.Callback) error {
	User, err := h.Users.Get(c.Sender.ID)
	if err != nil {
		return err
	}
	Index, err := ReceivingCell(c.Data)
	if err != nil {
		return err
	}
	I, err := strconv.Atoi(Index)
	if err = h.GoogleSheetsApi.ClearTask(User.SpreadSheetID, int64(I)); err != nil {
		return err
	}
	if _, err = h.tgbot.Send(c.Sender, "Задача успешно удалена"); err != nil {
		return err
	}
	return nil
}

func (h Handler) HandleGetTaskCompletedButton(c *telebot.Callback) error {
	Tasks, err := h.taskHandler.GetCompletedTask(c.Sender)
	if err != nil {
		return err
	}
	if Tasks != nil {
		h.SendTasksWithoutButtons(Tasks, c.Message.Chat.ID)

	} else {
		if _, err = h.tgbot.Send(c.Sender, "Вы еще не выполнили не одну задачу("); err != nil {
			return err
		}
	}
	return nil
}

func (h Handler) HandleGetTaskUnCompletedButton(c *telebot.Callback) error {
	tasks, err := h.taskHandler.GetUnCompletedTask(c.Sender)
	if err != nil {
		return err
	}
	if tasks != nil {
		h.SendTasksWithButtons(tasks, c.Message.Chat.ID)

	} else {
		var TaskButtons []telebot.InlineButton
		TaskButtons = append(TaskButtons, h.CreateAddTaskButton())
		keyboard := telebot.InlineKeyboardMarkup{InlineKeyboard: [][]telebot.InlineButton{TaskButtons}}
		replyMarkup := telebot.ReplyMarkup{InlineKeyboard: keyboard.InlineKeyboard}
		sendOptions := telebot.SendOptions{ReplyMarkup: &replyMarkup}
		if _, err = h.tgbot.Send(c.Sender, "Вы еще не добавили новую задачу", &sendOptions); err != nil {
			return err
		}
	}
	return nil
}

func (h Handler) HandleMistake(c *telebot.Callback) error {
	if _, err := h.tgbot.Send(c.Sender, "Процесс  отменен"); err != nil {
		return err
	}
	return nil
}

func (h Handler) HandleRegistrationStart(c *telebot.Callback) error {
	h.processingRegistrationUsers.GetOrCreate(c.Message.Chat.ID)
	if _, err := h.tgbot.Send(c.Sender, "Введите свое имя"); err != nil {
		h.processingRegistrationUsers.Delete(c.Message.Chat.ID)
		return err
	}
	h.processingRegistrationUsers.UpdateRegistrationStep(c.Message.Chat.ID, domain.RegistrationStepName)
	return nil
}

func (h Handler) HandleRenamingTaskButton(c *telebot.Callback) error {
	User, err := h.Users.Get(c.Sender.ID)
	if err != nil {
		return err
	}

	if _, err = h.tgbot.Send(c.Sender, "Введите то на что вы хотите заменить название вашей задачи"); err != nil {
		h.errChan.HandleError(err)
		return err
	}

	h.ProcessingRenamingTasks.GetOrCreate(c.Message.Chat.ID)
	h.ProcessingRenamingTasks.SetSpreadSheetId(c.Message.Chat.ID, User.SpreadSheetID)
	h.ProcessingRenamingTasks.AddSpreadSheetCellId(c.Message.Chat.ID, c.Data)

	return nil
}

func (h Handler) HandleSetNotificationsOffOrOn(c *telebot.Callback) error {
	user, err := h.Mongo.Get(c.Sender.ID)
	if err != nil {
		return err
	}
	var IsSend bool
	if c.Data == "On" {
		IsSend = true
	} else if c.Data == "Off" {
		IsSend = false
	}
	user.IsSendNotification = IsSend

	if err = h.Mongo.Update(c.Sender.ID, *user); err != nil {
		return err
	}
	if _, err = h.tgbot.Send(c.Sender, "Операция выполнена успешно"); err != nil {
		return err
	}
	err = h.Users.Update(c.Sender.ID, *user)

	return err
}

func (h Handler) HandleSettingStatusCompletedTaskButton(c *telebot.Callback) error {
	User, err := h.Users.Get(c.Sender.ID)
	if err != nil {
		return err
	}
	if err = h.GoogleSheetsApi.RenamingCell(User.SpreadSheetID, c.Data, "Выполнено"); err != nil {
		return err
	}

	if _, err = h.tgbot.Send(c.Sender, "Операция выполнена успешно!"); err != nil {
		h.errChan.HandleError(err)
	}

	return nil
}

func (h Handler) HandleSettingTaskStatusButton(c *telebot.Callback) error {
	var TaskButtons []telebot.InlineButton
	TaskButtons = append(TaskButtons, h.CreatingDoneStatusTaskButton(c.Data), h.CreatingMistakeRegistrationBtn())
	keyboard := telebot.InlineKeyboardMarkup{InlineKeyboard: [][]telebot.InlineButton{TaskButtons}}
	replyMarkup := telebot.ReplyMarkup{InlineKeyboard: keyboard.InlineKeyboard}
	sendOptions := telebot.SendOptions{ReplyMarkup: &replyMarkup}
	if _, err := h.tgbot.Send(c.Sender, "Вы уже выполнили задачу?", &sendOptions); err != nil {
		return err
	}
	return nil
}

func (h Handler) HandleTaskButton(c *telebot.Callback) error {
	User, err := h.Users.Get(c.Sender.ID)
	if err != nil {
		return err
	}
	Tasks := User.Tasks
	for _, Task := range Tasks {
		if Task.Range == c.Data {
			h.SendTaskWithButtons(Task, c.Message.Chat.ID)
		}

	}
	return nil
}

func (h Handler) SendTaskWithButtons(task domain.Task, chatID int64) {
	var TaskButtons []telebot.InlineButton
	TaskButtons = append(TaskButtons, h.CreateRenamingTaskButton(task), h.CreateChangeDescriptionTaskButton(task), h.CreateChangeDataTaskButton(task), h.CreateChangeStatusTaskButton(task),
		h.CreatingDeleteTaskButton(task.Range))
	keyboard := telebot.InlineKeyboardMarkup{InlineKeyboard: [][]telebot.InlineButton{TaskButtons}}
	replyMarkup := telebot.ReplyMarkup{InlineKeyboard: keyboard.InlineKeyboard}
	sendOptions := telebot.SendOptions{ReplyMarkup: &replyMarkup}
	TaskStr := fmt.Sprintf("Название задачи: %s \n Описание задачи: %s \n Срок задачи: %s \n Статус задачи: %s", task.Name, task.Description, task.Date, task.Status)
	if _, err := h.tgbot.Send(telebot.ChatID(chatID), TaskStr, &sendOptions); err != nil {
		h.errChan.HandleError(err)
	}
}

func (h Handler) SendTasksWithButtons(tasks []domain.Task, chatID int64) {
	var taskButtons []telebot.InlineButton
	for _, task := range tasks {
		taskButtons = append(taskButtons, h.CreateGetTaskButton(task))
	}
	taskButtons = append(taskButtons, h.CreateAddTaskButton())
	keyboard := telebot.InlineKeyboardMarkup{InlineKeyboard: [][]telebot.InlineButton{taskButtons}}
	replyMarkup := telebot.ReplyMarkup{InlineKeyboard: keyboard.InlineKeyboard}
	sendOptions := telebot.SendOptions{ReplyMarkup: &replyMarkup}
	if _, err := h.tgbot.Send(telebot.ChatID(chatID), "Список задач:", &sendOptions); err != nil {
		h.errChan.HandleError(err)
	}
}

func (h Handler) SendTasksWithoutButtons(tasks []domain.Task, chatID int64) {
	for _, task := range tasks {
		formatStr := fmt.Sprintf("Название задачи: %s \n Описание задачи: %s \n Дата задачи: %s \n Статус задачи: %s \n\n\n", task.Name, task.Description, task.Date, task.Status)
		if _, err := h.tgbot.Send(telebot.ChatID(chatID), formatStr); err != nil {
			h.errChan.HandleError(err)
		}
	}
}
