package tg_handlers

import (
	"gopkg.in/tucnak/telebot.v2"
	"main.go/internal/domain"
	"strings"
)

const (
	AddingTaskButton             = "AddingTask"
	ChangeDataTaskButton         = "ChangeData"
	ChangeDescriptionTaskButton  = "ChangeDescription"
	ChangeStatusTaskButton       = "ChangeStatus"
	DeleteTaskButton             = "DeleteTask"
	DoneStatusTaskButton         = "DoneStatus"
	GetTaskWithStatusCompleted   = "GetTaskWithStatusCompleted"
	GetTaskWithStatusUnCompleted = "GetTaskWithStatusUnCompleted"
	GettingTaskButton            = "GetTask"
	Mistake                      = "Mistake"
	Registration                 = "Registration"
	RenamingTaskButton           = "RenamingTask"
	SetNotifications             = "SetNotification"
	SetNotificationsFrequency    = "SetNotificationsFrequency"
)

func (h Handler) CreatingKeyboardOptionsBtn(btn ...telebot.InlineButton) telebot.SendOptions {
	var TaskButtons []telebot.InlineButton
	TaskButtons = append(TaskButtons, btn...)
	keyboard := telebot.InlineKeyboardMarkup{InlineKeyboard: [][]telebot.InlineButton{TaskButtons}}
	replyMarkup := telebot.ReplyMarkup{InlineKeyboard: keyboard.InlineKeyboard}
	sendOptions := telebot.SendOptions{ReplyMarkup: &replyMarkup}
	return sendOptions
}
func ReceivingCell(Range string) (string, error) {
	parts := strings.Split(Range, ":")
	Number := strings.Split(parts[0], "A")
	return Number[1], nil
}

func (h Handler) CreateGetTaskButton(Task domain.Task) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: GettingTaskButton,
		Text:   Task.Name,
		Data:   Task.Range,
	}
}
func (h Handler) CreateAddTaskButton() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: AddingTaskButton,
		Text:   "Add Task",
	}
}
func (h Handler) CreateRenamingTaskButton(task domain.Task) telebot.InlineButton {
	ReceivedCell, err := ReceivingCell(task.Range)
	if err != nil {
		h.errChan.HandleError(err)
	}
	return telebot.InlineButton{
		Unique: RenamingTaskButton,
		Text:   "Name:" + task.Name,
		Data:   "Sheet1!A" + ReceivedCell,
	}
}
func (h Handler) CreateChangeDescriptionTaskButton(task domain.Task) telebot.InlineButton {
	ReceivedCell, err := ReceivingCell(task.Range)
	if err != nil {
		h.errChan.HandleError(err)
	}
	return telebot.InlineButton{
		Unique: ChangeDescriptionTaskButton,
		Text:   "Description:" + task.Description,
		Data:   "Sheet1!B" + ReceivedCell,
	}
}
func (h Handler) CreateChangeDataTaskButton(task domain.Task) telebot.InlineButton {
	ReceivedCell, err := ReceivingCell(task.Range)
	if err != nil {
		h.errChan.HandleError(err)
	}
	return telebot.InlineButton{
		Unique: ChangeDataTaskButton,
		Text:   "Data:" + task.Date,
		Data:   "Sheet1!C" + ReceivedCell,
	}
}
func (h Handler) CreateChangeStatusTaskButton(task domain.Task) telebot.InlineButton {
	ReceivedCell, err := ReceivingCell(task.Range)
	if err != nil {
		h.errChan.HandleError(err)
	}
	return telebot.InlineButton{
		Unique: ChangeStatusTaskButton,
		Text:   "Status:" + task.Status,
		Data:   "Sheet1!D" + ReceivedCell,
	}
}
func (h Handler) CreatingDoneStatusTaskButton(ReceivedCell string) telebot.InlineButton {

	return telebot.InlineButton{
		Unique: DoneStatusTaskButton,
		Text:   "Установить статус задачи как выполнено",
		Data:   ReceivedCell,
	}
}
func (h Handler) CreatingDeleteTaskButton(Range string) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: DeleteTaskButton,
		Text:   "Удалить данную задачу",
		Data:   Range,
	}
}
func (h Handler) CreatingGetCompletedTaskButton() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: GetTaskWithStatusCompleted,
		Text:   "Получить уже выполненые задачи",
	}

}
func (h Handler) CreatingUnCompletedTaskButton() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: GetTaskWithStatusUnCompleted,
		Text:   "Получить еще не выполненые задачи",
	}

}

func (h Handler) CreatingNotificationFrequencyTask1Minute() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotificationsFrequency,
		Text:   "Напоминания каждую минуту",
		Data:   "0.016",
	}
}
func (h Handler) CreatingNotificationFrequencyTask2Minute() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotificationsFrequency,
		Text:   "Напоминания каждые 2 минуты",
		Data:   "0.033",
	}
}
func (h Handler) CreatingNotificationFrequencyTask5Minute() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotificationsFrequency,
		Text:   "Напоминания каждые 5 минут",
		Data:   "0.083",
	}
}
func (h Handler) CreatingNotificationFrequencyTask10Minute() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotificationsFrequency,
		Text:   "Напоминания каждые 10 минут",
		Data:   "0.16",
	}
}
func (h Handler) CreatingNotificationFrequencyTask15Minute() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotificationsFrequency,
		Text:   "Напоминания каждые 15 минут",
		Data:   "0.25",
	}
}
func (h Handler) CreatingNotificationFrequencyTask30Minute() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotificationsFrequency,
		Text:   "Напоминания каждые 30 минут",
		Data:   "0.5",
	}
}
func (h Handler) CreatingNotificationFrequencyTaskHour() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotificationsFrequency,
		Text:   "Напоминания каждый час",
		Data:   "1",
	}
}
func (h Handler) CreatingNotificationFrequencyTask2Hour() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotificationsFrequency,
		Text:   "Напоминания каждые 2 часа",
		Data:   "2",
	}
}
func (h Handler) CreatingNotificationFrequencyTask3Hour() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotificationsFrequency,
		Text:   "Напоминания каждые 3 часа",
		Data:   "3",
	}
}
func (h Handler) CreatingNotificationFrequencyTaskEvery24Hours() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotificationsFrequency,
		Text:   "Напоминания каждые 24 часа",
		Data:   "24",
	}
}
func (h Handler) CreatingSetOnNotificationsTaskButton() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotifications,
		Text:   "Включить уведомления",
		Data:   "On",
	}
}
func (h Handler) CreatingSetOffNotificationsTaskButton() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: SetNotifications,
		Text:   "Выключить уведомления",
		Data:   "Off",
	}
}
func (h Handler) CreatingRegistrationButton() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: Registration,
		Text:   "Создать новый аккаунт",
	}
}
func (h Handler) CreatingMistakeRegistrationBtn() telebot.InlineButton {
	return telebot.InlineButton{
		Unique: Mistake,
		Text:   "Назад",
	}
}
