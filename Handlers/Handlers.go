package Handlers

import (
	"gopkg.in/tucnak/telebot.v2"
	"main.go/internal/domain"
)

type (
	UserHandler interface {
		CreateUser(msg *telebot.Message) error
		LoginUser(ChatId int64, msg *telebot.Message) error
		SendNotifications(chatId int64) error
		GetProfile(UncompletedTaskQuantity, CompletedTaskQuantity int, UserName string, IsSendingNotification bool, NotificationsFrequency float64, msg *telebot.Message) error
	}
	TaskHandler interface {
		GetCompletedTask(ChatId int64, SenderUsername *telebot.User) ([]domain.Task, error)
		GetUnCompletedTask(ChatId int64, SenderUsername *telebot.User) ([]domain.Task, error)
		RenameTask(msg *telebot.Message) error
		ChangeDescription(msg *telebot.Message) error
		ChangeData(msg *telebot.Message) error
		AddTask(msg *telebot.Message) error
	}
)
