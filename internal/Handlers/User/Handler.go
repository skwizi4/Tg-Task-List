package User

import (
	"fmt"
	"github.com/skwizi4/lib/ErrChan"
	gpt3 "github.com/skwizi4/lib/Gpt-3"
	"github.com/skwizi4/lib/Redis"
	"github.com/skwizi4/lib/logs"
	"gopkg.in/tucnak/telebot.v2"
	"main.go/internal/Repo/MongoDB"
	"main.go/internal/Services/ActiveUsers"
	googleSheets "main.go/internal/Services/google-sheets"
	"main.go/internal/domain"
	"strconv"
	"time"
)

type Handler struct {
	OnlineUsers                 ActiveUsers.Cache
	ErrorChannel                *ErrChan.ErrorChannel
	GoogleSheetsAPI             googleSheets.SheetsInterface
	Gpt3                        gpt3.GPT3
	mongo                       MongoDB.Mongo
	logs                        logs.GoLogger
	processingLoginUsers        *domain.ProcessLoginUsers
	processingRegistrationUsers *domain.ProcessingRegistrationUsers
	Redis                       Redis.Redis
	tgbot                       *telebot.Bot
}

func New(
	OnlineUsers ActiveUsers.Cache,
	ErrorChannel *ErrChan.ErrorChannel,
	GoogleSheetsAPI googleSheets.SheetsInterface,
	Gpt3 gpt3.GPT3,
	Mongo MongoDB.Mongo,
	logs logs.GoLogger,
	processingLoginUser *domain.ProcessLoginUsers,
	processingRegistrationUser *domain.ProcessingRegistrationUsers,
	Redis Redis.Redis,
	tgbot *telebot.Bot,
) Handler {
	return Handler{
		tgbot:                       tgbot,
		processingRegistrationUsers: processingRegistrationUser,
		processingLoginUsers:        processingLoginUser,
		Redis:                       Redis,
		logs:                        logs,
		GoogleSheetsAPI:             GoogleSheetsAPI,
		ErrorChannel:                ErrorChannel,
		Gpt3:                        Gpt3,
		mongo:                       Mongo,
		OnlineUsers:                 OnlineUsers,
	}
}

// ------------------------------------------------Registration-------------------------------------------------------------------------------------

func (h Handler) CreateUser(msg *telebot.Message) error {
	processUser := h.processingRegistrationUsers.GetOrCreate(msg.Chat.ID)
	switch processUser.Step {
	case domain.RegistrationStepStart:
		h.processingRegistrationUsers.GetOrCreate(msg.Chat.ID)
		if _, err := h.tgbot.Send(msg.Sender, "Введите свое имя"); err != nil {
			h.processingRegistrationUsers.Delete(msg.Chat.ID)
			return err
		}
		h.processingRegistrationUsers.UpdateRegistrationStep(msg.Chat.ID, domain.RegistrationStepName)
	case domain.RegistrationStepName:
		if msg.Text == "/exit" {
			if _, err := h.tgbot.Send(msg.Sender, "Регистрация прервана"); err != nil {
				h.processingRegistrationUsers.Delete(msg.Chat.ID)
				return err
			}
			h.processingRegistrationUsers.Delete(msg.Chat.ID)
			return nil
		}
		h.processingRegistrationUsers.SetName(msg.Chat.ID, msg.Text)
		if _, err := h.tgbot.Send(msg.Sender, "Введите пароль"); err != nil {
			h.processingRegistrationUsers.Delete(msg.Chat.ID)
			return err
		}
		h.processingRegistrationUsers.UpdateRegistrationStep(msg.Chat.ID, domain.RegistrationStepPassword)
	case domain.RegistrationStepPassword:
		if msg.Text == "/exit" {
			if _, err := h.tgbot.Send(msg.Sender, "Регистрация прервана"); err != nil {
				h.processingRegistrationUsers.Delete(msg.Chat.ID)
				return err
			}
			h.processingRegistrationUsers.Delete(msg.Chat.ID)
			return nil
		}
		h.processingRegistrationUsers.SetPassword(msg.Chat.ID, msg.Text)
		Users := h.processingRegistrationUsers.GetOrCreate(msg.Chat.ID)
		id, err := h.GoogleSheetsAPI.CreateSpreadsheet(Users.User.Name + strconv.Itoa(int(msg.Sender.ID)))
		if err != nil {
			h.processingRegistrationUsers.Delete(msg.Chat.ID)
			h.logs.ErrorFrmt("ERROR IS OCCURRED IN CREATING SPREADSHEET")
		}
		DataNow := time.Now().Format("2006-01-02 15:04:05")
		task := []interface{}{"Название Задачи", "Описание задачи", "Дата выполненеия задачи" + "*" + DataNow + "*", "Не выполнено"}
		values := [][]interface{}{task}
		if err = h.GoogleSheetsAPI.AddTask(id, "A1:D49", values); err != nil {
			_, _ = h.tgbot.Send(msg.Sender, "Sorry, something wrong with connecting to our DataBase(")
			h.processingRegistrationUsers.Delete(msg.Chat.ID)
			return err
		}
		h.processingRegistrationUsers.SetSpreadSheetID(msg.Chat.ID, id)
		h.processingRegistrationUsers.SetNotificationFrequency(msg.Chat.ID, 0)
		h.processingRegistrationUsers.SetSendNotifications(msg.Chat.ID, false)
		h.processingRegistrationUsers.SetTelegramID(msg.Chat.ID, msg.Sender.ID)
		if res, err := h.mongo.Get(msg.Sender.ID); err == nil && res != nil {
			if err := h.mongo.Update(msg.Sender.ID, h.processingRegistrationUsers.GetOrCreate(msg.Chat.ID).User); err != nil {
				h.processingRegistrationUsers.Delete(msg.Chat.ID)
				return err
			}
		} else {
			if err = h.mongo.Create(h.processingRegistrationUsers.GetOrCreate(msg.Chat.ID).User); err != nil {
				h.processingRegistrationUsers.Delete(msg.Chat.ID)
				return err
			}
		}

		if _, err = h.tgbot.Send(msg.Sender, "вы успешно зарегестрировались!"); err != nil {
			h.processingRegistrationUsers.Delete(msg.Chat.ID)
			return err
		}
		h.logs.Info("Registration successfully")
		h.processingRegistrationUsers.Delete(msg.Chat.ID)

	}
	return nil
}

// ------------------------------------------------Login-------------------------------------------------------------------------------------

func (h Handler) LoginUser(chatId int64, msg *telebot.Message) error {
	processUser := h.processingLoginUsers.GetOrCreate(chatId)
	switch processUser.Step {
	case domain.LoginStepStart:
		if _, err := h.tgbot.Send(msg.Sender, "введите пароль"); err != nil {
			h.processingLoginUsers.Delete(chatId)
			return err
		}
		h.processingLoginUsers.UpdateStep(domain.LoginStepPassword, msg.Chat.ID)
	case domain.LoginStepPassword:
		if msg.Text == "/exit" {
			if _, err := h.tgbot.Send(msg.Sender, "Процесс авторизации прерван"); err != nil {
				h.processingLoginUsers.Delete(chatId)
				return err
			}
			h.processingLoginUsers.Delete(chatId)
			return nil
		}
		user, err := h.mongo.Get(msg.Sender.ID)
		if err != nil {
			h.processingLoginUsers.Delete(chatId)
			return err
		}
		if msg.Text != user.Password {
			if _, err = h.tgbot.Send(msg.Sender, "Вы ввели непарвильный пароль"); err != nil {
				h.processingLoginUsers.Delete(chatId)
				return err
			}
			h.processingLoginUsers.Delete(chatId)
			return nil
		}
		if _, err = h.tgbot.Send(msg.Sender, fmt.Sprintf("Привет %s", user.Name)); err != nil {
			h.processingLoginUsers.Delete(chatId)
			return err
		}
		//ToDo Create users cache to Redis 1. CRUD Methods
		if err = h.OnlineUsers.Create(msg.Sender.ID, *user); err != nil {
			h.processingLoginUsers.Delete(chatId)
			return err
		}
		h.processingLoginUsers.Delete(msg.Chat.ID)

	}
	return nil
}

// !!!!!!------------------------------------------------GetProfile--------------------------------------------------------------------------!!!!

func (h Handler) GetProfile(UncompletedTaskQuantity, CompletedTaskQuantity int, UserName string, IsSendingNotification bool, NotificationsFrequency float64, msg *telebot.Message) error {
	var Notifications string
	var Frequency float64
	if IsSendingNotification == true {
		Notifications = "Включено"
		Frequency = NotificationsFrequency
		formatStr := fmt.Sprintf("Ваш профиль: \n Имя: %s \n Напоминания: %s \n Частота напоминаний: %.2f \n Кол-во сделаных задач: %d \n Кол-во не сделаных задач: %d \n", UserName, Notifications, Frequency, CompletedTaskQuantity, UncompletedTaskQuantity)
		if _, err := h.tgbot.Send(msg.Sender, formatStr); err != nil {
			return err
		}
		return nil
	} else {
		Notifications = "Выключено"
	}
	formatStr := fmt.Sprintf("Ваш профиль: \n Имя: %s \n Напоминания: %s \n Кол-во сделаных задач: %d \n Кол-во не сделаных задач: %d \n", UserName, Notifications, CompletedTaskQuantity, UncompletedTaskQuantity)
	if _, err := h.tgbot.Send(msg.Sender, formatStr); err != nil {
		return err
	}
	return nil
}

// !!!!!!------------------------------------------------SendNotifications--------------------------------------------------------------------------!!!!

func (h Handler) SendNotifications(chatId int64) error {
	//User := h.ActiveUsers.GetOrCreate(chatId, "")
	//tasks := User.Tasks
	//if tasks == nil {
	//	if _, err := h.tgbot.Send(telebot.ChatID(chatId), "Вы еще не получили задачи, для того чтобы получить напоминание о задачах, нажмите на команду /GetTasks.Если у вас нет задач, то добавьте их."); err != nil {
	//		return err
	//	}
	//	h.ActiveUsers.IsSendingNotifications(chatId, true)
	//	return errors.New("user has no tasks yet")
	//}
	//prompt := "У меня есть следующие задачи:\n"
	//for _, task := range tasks {
	//	prompt += fmt.Sprintf("Задача: %s, Важность: %s, Дата: %s\n", task.Name, task.Description, task.Date)
	//}
	//prompt += "Какую задачу мне следует выбрать для напоминания, учитывая их важность и дату?"
	//
	//Resp, err := h.Gpt3.Request(prompt)
	//if err != nil {
	//	return err
	//}
	//if _, err = h.tgbot.Send(telebot.ChatID(chatId), Resp); err != nil {
	//	return err
	//}
	//h.ActiveUsers.IsSendingNotifications(chatId, true)
	return nil
}
