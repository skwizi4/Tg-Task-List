package App

import (
	"context"
	"github.com/go-playground/validator/v10"
	_ "github.com/redis/go-redis/v9"
	"github.com/skwizi4/lib/ErrChan"
	gpt3 "github.com/skwizi4/lib/Gpt-3"
	"github.com/skwizi4/lib/Redis"
	logger "github.com/skwizi4/lib/logs"
	tele "gopkg.in/tucnak/telebot.v2"
	"main.go/internal/Config"
	"main.go/internal/Handlers"
	taskHandler "main.go/internal/Handlers/Tasks"
	userHandler "main.go/internal/Handlers/User"
	googlesheets2 "main.go/internal/Services/google-sheets"
	"main.go/internal/domain"
	"main.go/internal/tg_handlers"
	"sync"
	"time"
)

type App struct {
	ActiveUsers                        domain.ActiveUsers
	bot                                *tele.Bot
	ChatGpt                            gpt3.GPT3
	config                             Config.Config
	ErrChan                            *ErrChan.ErrorChannel
	GoogleSheetsApi                    googlesheets2.SheetsInterface
	logger                             logger.GoLogger
	mutex                              sync.Mutex
	ProcessLoginUsers                  domain.ProcessLoginUsers
	ProcessingAddingTasks              domain.ProcessingAddingTasks
	ProcessingChangingDataTasks        domain.ProcessingChangingDataTasks
	ProcessingChangingDescriptionTasks domain.ProcessingChangingDescriptionTasks
	ProcessingChangingStatusTasks      domain.ProcessingChangingStatusTasks
	ProcessingRegistrationUsers        domain.ProcessingRegistrationUsers
	ProcessingRenamingTasks            domain.ProcessingRenamingTasks
	redis                              Redis.Redis
	ServiceName                        string
	taskHandler                        Handlers.TaskHandler
	tgHandler                          tg_handlers.Handler
	UserHandler                        Handlers.UserHandler
	validator                          *validator.Validate
}

func New(name string) App {
	return App{
		ServiceName: name,
	}
}

func (a *App) Run(ctx context.Context) {
	a.InitLogger()
	a.InitErrorHandler(ctx)
	a.initValidator()
	a.populateConfig()
	a.initRedis()
	a.InitGoogleSheetsClient()
	a.InitChatGptClient()
	a.InitTgBot()
	a.InitHandlers()
	a.ListenTgBot()
}

func (a *App) InitErrorHandler(ctx context.Context) {
	a.ErrChan = ErrChan.InitErrChan(10, a.logger)
	go func() {
		for {
			select {
			case <-ctx.Done():
				a.ErrChan.Close()
				return
			}
		}
	}()
	a.ErrChan.Start()
	a.logger.InfoFrmt("InitErrorHandler-Successfully")
}

func (a *App) InitHandlers() {
	a.UserHandler = userHandler.New(&a.ActiveUsers, a.ErrChan, a.GoogleSheetsApi, a.ChatGpt,
		a.logger, &a.ProcessLoginUsers, &a.ProcessingRegistrationUsers, a.redis, a.bot)
	a.taskHandler = taskHandler.New(a.bot, a.ErrChan, a.GoogleSheetsApi, a.logger,
		&a.ProcessingAddingTasks, &a.ProcessingChangingDataTasks, &a.ProcessingChangingDescriptionTasks,
		&a.ProcessingChangingStatusTasks, &a.ProcessingRenamingTasks, &a.ActiveUsers)
	a.tgHandler = tg_handlers.New(a.bot, a.UserHandler, a.taskHandler, a.logger, a.ErrChan, &a.ActiveUsers, &a.ProcessingRenamingTasks, &a.ProcessingChangingDescriptionTasks,
		&a.ProcessingChangingDataTasks, &a.ProcessingChangingStatusTasks, a.GoogleSheetsApi, &a.ProcessingAddingTasks, a.redis, &a.ProcessingRegistrationUsers)
	a.logger.InfoFrmt("InitHandlers-Successfully")
}

func (a *App) InitLogger() {
	a.logger = logger.InitLogger()
	a.logger.InfoFrmt("InitLogger-Successfully")
}

func (a *App) initValidator() {
	a.validator = validator.New()
	a.logger.InfoFrmt("initValidator-Successfully")
}

func (a *App) populateConfig() {
	cfg, err := Config.ParseConfig("C:\\golang\\src\\TG-ToDoList\\config.json")
	if err != nil {
		a.logger.ErrorFrmt("error in parsing config: %s", err)
	}
	err = cfg.ValidateConfig(a.validator)
	if err != nil {
		a.logger.ErrorFrmt("error in config validation: %s", err)
	}
	a.config = *cfg
	a.logger.InfoFrmt("InitConfig-Successfully")
}

func (a *App) initRedis() {
	a.redis = Redis.New(a.config.Redis.Password, a.config.Redis.DB, a.config.Redis.Addr)
	a.logger.InfoFrmt("InitRedis-Successfully")
}

func (a *App) InitGoogleSheetsClient() {
	sheetsClient := googlesheets2.GoogleSheetsClient{}
	if err := sheetsClient.InitGoogleSheetsClient(); err != nil {
		a.logger.ErrorFrmt("error in initializing Google sheets client: %s", err)
	}
	a.GoogleSheetsApi = &sheetsClient
	a.logger.InfoFrmt("InitGoogleSpreadSheet-Successfully")
}

func (a *App) InitChatGptClient() {
	a.ChatGpt = gpt3.InitGP3(a.config.Gpt.Model, a.config.Gpt.ApiKey)
	a.logger.InfoFrmt("InitChatGpt-Successfully")
}

func (a *App) InitTgBot() {
	botSettings := tele.Settings{
		Token:  a.config.Token.Token,
		Poller: &tele.LongPoller{Timeout: 1 * time.Second},
	}
	var err error
	if a.bot, err = tele.NewBot(botSettings); err != nil {
		a.logger.ErrorFrmt("Error Is occurred in InitTgBot, error: %s", err)
	}
	a.logger.InfoFrmt("InitTgBot-Successfully")
}

func (a *App) ListenTgBot() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	go a.bot.Handle("/registration", func(msg *tele.Message) {
		go a.tgHandler.RegistrationHandler(msg)
	})
	a.bot.Handle("/login", func(msg *tele.Message) {
		go a.tgHandler.LoginHandler(msg)
	})
	a.bot.Handle("/GetTasks", func(msg *tele.Message) {
		go a.tgHandler.GetTasks(msg)
	})
	a.bot.Handle("/ChangeNotifications", func(msg *tele.Message) {
		go a.tgHandler.ChangeFrequencyOfNotifications(msg)
	})
	a.bot.Handle("/SetNotifications", func(msg *tele.Message) {
		go a.tgHandler.SetNotifications(msg)
	})
	a.bot.Handle("/GetProfile", func(msg *tele.Message) {
		go a.tgHandler.ProfileHandler(msg)
	})

	a.bot.Handle(tele.OnText, func(msg *tele.Message) {
		if a.ProcessingRegistrationUsers.IfExist(msg.Chat.ID) {
			go a.tgHandler.RegistrationHandler(msg)
		} else if a.ProcessLoginUsers.IfExist(msg.Chat.ID) {
			go a.tgHandler.LoginHandler(msg)
		} else if a.ProcessingRenamingTasks.IfExist(msg.Chat.ID) {
			go a.tgHandler.RenamingTaskHandler(msg)
		} else if a.ProcessingChangingDescriptionTasks.IfExist(msg.Chat.ID) {
			go a.tgHandler.ChangingDescriptionHandler(msg)
		} else if a.ProcessingChangingDataTasks.IfExist(msg.Chat.ID) {
			go a.tgHandler.ChangingDataHandler(msg)
		} else if a.ProcessingAddingTasks.IfExist(msg.Chat.ID) {
			go a.tgHandler.AddingTaskHandler(msg)
		}
	})

	a.bot.Handle(&tele.Btn{Unique: tg_handlers.GettingTaskButton}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleTaskButton(c); err != nil {
			a.ErrChan.HandleError(err)
		}

	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.RenamingTaskButton}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleRenamingTaskButton(c); err != nil {
			a.ErrChan.HandleError(err)
		}

	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.ChangeDescriptionTaskButton}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleChangingTaskDescriptionButton(c); err != nil {
			a.ErrChan.HandleError(err)
		}

	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.ChangeDataTaskButton}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleChangingTaskDataButton(c); err != nil {
			a.ErrChan.HandleError(err)
		}
	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.ChangeStatusTaskButton}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleSettingTaskStatusButton(c); err != nil {
			a.ErrChan.HandleError(err)
		}

	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.DoneStatusTaskButton}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleSettingStatusCompletedTaskButton(c); err != nil {
			a.ErrChan.HandleError(err)
		}

	})

	a.bot.Handle(&tele.Btn{Unique: tg_handlers.AddingTaskButton}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleAddTaskButton(c); err != nil {
			a.ErrChan.HandleError(err)
		}
	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.DeleteTaskButton}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleDeleteTaskButton(c); err != nil {
			a.ErrChan.HandleError(err)
		}

	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.GetTaskWithStatusCompleted}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleGetTaskCompletedButton(c); err != nil {
			a.ErrChan.HandleError(err)
		}

	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.GetTaskWithStatusUnCompleted}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleGetTaskUnCompletedButton(c); err != nil {
			a.ErrChan.HandleError(err)
		}
	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.SetNotificationsFrequency}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleChangeNotificationsFrequency(c); err != nil {
			a.ErrChan.HandleError(err)
		}
	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.SetNotifications}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleSetNotificationsOffOrOn(c); err != nil {
			a.ErrChan.HandleError(err)
		}
	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.Registration}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleRegistrationStart(c); err != nil {
			a.ErrChan.HandleError(err)
		}
	})
	a.bot.Handle(&tele.Btn{Unique: tg_handlers.Mistake}, func(c *tele.Callback) {
		if err := a.tgHandler.HandleMistake(c); err != nil {
			a.ErrChan.HandleError(err)
		}
	})
	/// go a.tgHandler.NotifyUsers()

	a.bot.Start()
}
