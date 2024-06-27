package tg_handlers

import (
	"github.com/skwizi4/lib/ErrChan"
	"github.com/skwizi4/lib/Redis"
	"github.com/skwizi4/lib/logs"
	"gopkg.in/tucnak/telebot.v2"
	"main.go/internal/Handlers"
	"main.go/internal/Repo/MongoDB"
	"main.go/internal/Services/ActiveUsers"
	GoogleSheets "main.go/internal/Services/google-sheets"
	"main.go/internal/domain"
)

type Handler struct {
	GoogleSheetsApi                    GoogleSheets.SheetsInterface
	ProcessingAddingTask               *domain.ProcessingAddingTasks
	ProcessingChangingDataTasks        *domain.ProcessingChangingDataTasks
	ProcessingChangingDescriptionTasks *domain.ProcessingChangingDescriptionTasks
	ProcessingChangingStatusTasks      *domain.ProcessingChangingStatusTasks
	ProcessingRenamingTasks            *domain.ProcessingRenamingTasks
	Redis                              Redis.Redis
	Mongo                              MongoDB.Mongo
	UserHandler                        Handlers.UserHandler
	Users                              ActiveUsers.Cache
	errChan                            *ErrChan.ErrorChannel
	logger                             logs.GoLogger
	taskHandler                        Handlers.TaskHandler
	tgbot                              *telebot.Bot
	processingRegistrationUsers        *domain.ProcessingRegistrationUsers
}

func New(tgbot *telebot.Bot, userHandler Handlers.UserHandler, taskHandler Handlers.TaskHandler, logger logs.GoLogger, errChan *ErrChan.ErrorChannel, Users ActiveUsers.Cache,
	tasks *domain.ProcessingRenamingTasks, descriptionTasks *domain.ProcessingChangingDescriptionTasks, dataTasks *domain.ProcessingChangingDataTasks,
	statusTasks *domain.ProcessingChangingStatusTasks, GoogleSheetsApi GoogleSheets.SheetsInterface, ProcessingAddingTask *domain.ProcessingAddingTasks, redis Redis.Redis, Mongo MongoDB.Mongo,
	ProcessingRegistrationUsers *domain.ProcessingRegistrationUsers) Handler {
	return Handler{
		tgbot:                              tgbot,
		UserHandler:                        userHandler,
		taskHandler:                        taskHandler,
		logger:                             logger,
		errChan:                            errChan,
		Users:                              Users,
		ProcessingRenamingTasks:            tasks,
		ProcessingChangingDescriptionTasks: descriptionTasks,
		ProcessingChangingDataTasks:        dataTasks,
		ProcessingChangingStatusTasks:      statusTasks,
		GoogleSheetsApi:                    GoogleSheetsApi,
		ProcessingAddingTask:               ProcessingAddingTask,
		Redis:                              redis,
		processingRegistrationUsers:        ProcessingRegistrationUsers,
		Mongo:                              Mongo,
	}
}

//UserHandlers
// -----------------------------------------------------------------------------------Registration-----------------------------------------------------------------------------------

func (h Handler) RegistrationHandler(msg *telebot.Message) {
	if h.processingRegistrationUsers.IfExist(msg.Chat.ID) {
		if err := h.UserHandler.CreateUser(msg); err != nil {
			h.errChan.HandleError(err)
		}
		return
	}

	if _, err := h.Users.Get(msg.Sender.ID); err == nil { //Обработчик случая, когда пользвотель уже имеет аккаунт и авторизован
		Options := h.CreatingKeyboardOptionsBtn(h.CreatingRegistrationButton(), h.CreatingMistakeRegistrationBtn())
		if _, err = h.tgbot.Send(msg.Sender, "Вы уже авторизованы, вы дейстивтельно хотите создать новый акканут? Задачи, добавлене вам на этом аккаунте восстановить будет невозможно", &Options); err != nil {
			h.errChan.HandleError(err)
		}

	} else if _, err = h.Users.Get(msg.Sender.ID); err != nil { // Обработчик случае когда пользовтель не авторизирован/ не создан новый аккаунт
		if _, err = h.Mongo.Get(msg.Sender.ID); err != nil {
			if err = h.UserHandler.CreateUser(msg); err != nil {
				h.errChan.HandleError(err)
			}
		} else {
			Options := h.CreatingKeyboardOptionsBtn(h.CreatingRegistrationButton(), h.CreatingMistakeRegistrationBtn())
			if _, err := h.tgbot.Send(msg.Sender, "У вас уже есть аккаунт (Можете авторизироваться используя команду /login) , вы дейстивтельно хотите создать новый акканут? Задачи, добавлене вам на этом аккаунте восстановить будет невозможно", &Options); err != nil {
				h.errChan.HandleError(err)
			}
		}

	}
}

// -----------------------------------------------------------------------------------Authorize-----------------------------------------------------------------------------------

func (h Handler) LoginHandler(msg *telebot.Message) {
	if _, err := h.Users.Get(msg.Sender.ID); err == nil {
		if _, err = h.tgbot.Send(msg.Sender, "Вы уже прошли авторизацию, если вы хотите удалить аккаунт, просто зарегестрируйте новый, все ваши старые данные будут удалены"); err != nil {
			h.errChan.HandleError(err)
		}
		return
	}
	if err := h.UserHandler.LoginUser(msg.Chat.ID, msg); err != nil {
		h.errChan.HandleError(err)
	}
}

// -----------------------------------------------------------------------------------GetProfile-----------------------------------------------------------------------------------

func (h Handler) ProfileHandler(msg *telebot.Message) {
	if _, err := h.Users.Get(msg.Sender.ID); err != nil {
		if _, err := h.tgbot.Send(msg.Sender, "Вы не авторизованы, введите /login и пройдите авторизацию пожалуйста"); err != nil {
			h.errChan.HandleError(err)
		}
		return
	}
	CompletedTask, err := h.taskHandler.GetCompletedTask(msg.Sender)
	if err != nil {
		h.errChan.HandleError(err)
	}
	UncompletedTask, err := h.taskHandler.GetUnCompletedTask(msg.Sender)
	if err != nil {
		h.errChan.HandleError(err)
	}
	UncompletedTaskQuantity := len(UncompletedTask)
	CompletedTaskQuantity := len(CompletedTask)
	user, err := h.Mongo.Get(msg.Sender.ID)
	if err != nil {
		h.errChan.HandleError(err)
	}

	if err = h.UserHandler.GetProfile(UncompletedTaskQuantity, CompletedTaskQuantity, user.Name, user.IsSendNotification, user.FrequencyOfNotifications, msg); err != nil {
		h.errChan.HandleError(err)
	}

}

// TasksHandlers

// -----------------------------------------------------------------------------------GetTasks-----------------------------------------------------------------------------------

func (h Handler) GetTasks(msg *telebot.Message) {
	if _, err := h.Users.Get(msg.Sender.ID); err != nil {
		if _, err := h.tgbot.Send(msg.Sender, "Вы не авторизованы, введите /login и пройдите авторизацию пожалуйста"); err != nil {
			h.errChan.HandleError(err)
		}
		return
	}
	var TaskButtons []telebot.InlineButton
	TaskButtons = append(TaskButtons, h.CreatingGetCompletedTaskButton(), h.CreatingUnCompletedTaskButton())
	keyboard := telebot.InlineKeyboardMarkup{InlineKeyboard: [][]telebot.InlineButton{TaskButtons}}
	replyMarkup := telebot.ReplyMarkup{InlineKeyboard: keyboard.InlineKeyboard}
	sendOptions := telebot.SendOptions{ReplyMarkup: &replyMarkup}
	if _, err := h.tgbot.Send(msg.Sender, "Выберите одну из опций", &sendOptions); err != nil {
		h.errChan.HandleError(err)
	}
}

// -----------------------------------------------------------------------------------RenameTask-----------------------------------------------------------------------------------

func (h Handler) RenamingTaskHandler(msg *telebot.Message) {
	if err := h.taskHandler.RenameTask(msg); err != nil {
		h.errChan.HandleError(err)
	}
	tasks, err := h.taskHandler.GetUnCompletedTask(msg.Sender)
	if err != nil {
		h.errChan.HandleError(err)
	}
	h.SendTasksWithButtons(tasks, msg.Chat.ID)

}

// -----------------------------------------------------------------------------------ChangeDescriptionOfTask-----------------------------------------------------------------------------------

func (h Handler) ChangingDescriptionHandler(msg *telebot.Message) {
	if err := h.taskHandler.ChangeDescription(msg); err != nil {
		h.errChan.HandleError(err)
	}
	tasks, err := h.taskHandler.GetUnCompletedTask(msg.Sender)
	if err != nil {
		h.errChan.HandleError(err)
	}
	h.SendTasksWithButtons(tasks, msg.Chat.ID)
}

// -----------------------------------------------------------------------------------ChangeDataOfTask-----------------------------------------------------------------------------------

func (h Handler) ChangingDataHandler(msg *telebot.Message) {
	if err := h.taskHandler.ChangeData(msg); err != nil {
		h.errChan.HandleError(err)
	}
	tasks, err := h.taskHandler.GetUnCompletedTask(msg.Sender)
	if err != nil {
		h.errChan.HandleError(err)
	}
	h.SendTasksWithButtons(tasks, msg.Chat.ID)
}

// -----------------------------------------------------------------------------------AddTask-----------------------------------------------------------------------------------

func (h Handler) AddingTaskHandler(msg *telebot.Message) {
	if err := h.taskHandler.AddTask(msg); err != nil {
		h.errChan.HandleError(err)
	}
}

//-----------------------------------------------------------------------------------ChangeFrequencyOfNotifications-----------------------------------------------------------------------------------

func (h Handler) ChangeFrequencyOfNotifications(msg *telebot.Message) {
	if _, err := h.Users.Get(msg.Sender.ID); err != nil {
		if _, err := h.tgbot.Send(msg.Sender, "Вы не авторизованы, введите /login и пройдите авторизацию пожалуйста"); err != nil {
			h.errChan.HandleError(err)
		}
		return
	}
	var TaskButtons []telebot.InlineButton
	TaskButtons = append(TaskButtons, h.CreatingNotificationFrequencyTask1Minute(), h.CreatingNotificationFrequencyTask2Minute(), h.CreatingNotificationFrequencyTask5Minute(), h.CreatingNotificationFrequencyTask15Minute(),
		h.CreatingNotificationFrequencyTask30Minute(), h.CreatingNotificationFrequencyTaskHour(), h.CreatingNotificationFrequencyTask2Hour(), h.CreatingNotificationFrequencyTask3Hour(), h.CreatingNotificationFrequencyTaskEvery24Hours())
	keyboard := telebot.InlineKeyboardMarkup{InlineKeyboard: [][]telebot.InlineButton{TaskButtons}}
	replyMarkup := telebot.ReplyMarkup{InlineKeyboard: keyboard.InlineKeyboard}
	sendOptions := telebot.SendOptions{ReplyMarkup: &replyMarkup}
	if _, err := h.tgbot.Send(msg.Sender, "Выберите как часто бот будет напоминать вам о ваших задачах", &sendOptions); err != nil {
		h.errChan.HandleError(err)
	}
}

//-----------------------------------------------------------------------------------SetNotifications-----------------------------------------------------------------------------------

func (h Handler) SetNotifications(msg *telebot.Message) {
	if _, err := h.Users.Get(msg.Sender.ID); err != nil {
		if _, err := h.tgbot.Send(msg.Sender, "Вы не авторизованы, введите /login и пройдите авторизацию пожалуйста"); err != nil {
			h.errChan.HandleError(err)
		}
		return
	}
	var TaskButtons []telebot.InlineButton
	TaskButtons = append(TaskButtons, h.CreatingSetOnNotificationsTaskButton(), h.CreatingSetOffNotificationsTaskButton())
	keyboard := telebot.InlineKeyboardMarkup{InlineKeyboard: [][]telebot.InlineButton{TaskButtons}}
	replyMarkup := telebot.ReplyMarkup{InlineKeyboard: keyboard.InlineKeyboard}
	sendOptions := telebot.SendOptions{ReplyMarkup: &replyMarkup}
	if _, err := h.tgbot.Send(msg.Sender, "Выберите одну из опций", &sendOptions); err != nil {
		h.errChan.HandleError(err)
	}
	return
}

//-----------------------------------------------------------------------------------NotifyUsers-----------------------------------------------------------------------------------

//func (h Handler) NotifyUsers() {
//	for {
//		for _, user := range *h.Users {
//			if user.StopNotificationChan == nil {
//				user.StopNotificationChan = make(chan struct{})
//			}
//			if user.NotificationFrequency != 0 && user.IsSendingNotifications == true {
//
//				h.Users.IsSendingNotifications(user.ChatID, false)
//
//				go func(user domain.ActiveUser) {
//					var duration time.Duration
//					switch user.NotificationFrequency {
//					case 24:
//						duration = 24 * time.Hour
//					case 3:
//						duration = 3 * time.Hour
//					case 2:
//						duration = 2 * time.Hour
//					case 1:
//						duration = time.Hour
//					case 0.5:
//						duration = 30 * time.Minute
//					case 0.25:
//						duration = 15 * time.Minute
//					case 0.16:
//						duration = 10 * time.Minute
//					case 0.083:
//						duration = 5 * time.Minute
//					case 0.033:
//						duration = 2 * time.Minute
//					case 0.016:
//						duration = time.Minute
//					}
//
//					ticker := time.NewTicker(duration)
//					defer ticker.Stop()
//
//					for {
//						select {
//						case <-ticker.C:
//							err := h.UserHandler.SendNotifications(user.ChatID)
//							if err != nil {
//								h.errChan.HandleError(err)
//							}
//						case <-user.StopNotificationChan:
//							return
//						}
//					}
//				}(user)
//			}
//		}
//
//		time.Sleep(10 * time.Second)
//	}
//}
