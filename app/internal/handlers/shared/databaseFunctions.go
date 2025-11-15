package shared

// TODO: Внести возможность создавать список зависимостей для shared функций

import (
	"app/internal/handlers"
	"app/pkg/database"
	"app/pkg/database/models"
	e "app/pkg/errors"
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v4"
)

func ConnectDatabase(c tele.Context, args *handlers.Arg) (*handlers.Arg, *e.ErrorInfo) {
	db := database.GetDB()

	newArgs := make(handlers.Arg)
	newArgs["db"] = db

	return &newArgs, e.Nil()
}

func GetSenderAndTargetUser(c tele.Context, args *handlers.Arg) (*handlers.Arg, *e.ErrorInfo) {
	db := (*args)["db"].(*pg.DB)
	fmt.Println(c.Sender())
	user := &models.User{
		TgID: c.Sender().ID, 
	}
	fmt.Println(user)

	target := &models.User{
		TgID: viper.GetInt64("OWNER_TG_ID"),
		IsOwner: true,
	}

	err := db.Model(target).WherePK().Select()
	if err == pg.ErrNoRows {
		_, err := db.Model(&models.User{
			TgID: viper.GetInt64("OWNER_TG_ID"),
			IsOwner: true,
		}).OnConflict("DO NOTHING").Insert()
		if err != nil {
			return args, e.FromError(err, "Failed to insert target user").WithSeverity(e.Critical).WithData(map[string]any{
				"target": target,
			})
		}
	}
	if err != nil {
		return args, e.FromError(err, "Failed to select target user").WithSeverity(e.Critical).WithData(map[string]any{
			"target": target,
		})
	}

	_, err = db.Model(user).WherePK().SelectOrInsert()
	if err != nil {
		return args, e.FromError(err, "Failed to insert user").WithSeverity(e.Critical)
	}

	(*args)["user"] = user
	(*args)["target"] = target

	return args, e.Nil()
}

func GetOrCrateThread(c tele.Context, args *handlers.Arg) (*handlers.Arg, *e.ErrorInfo) {
	db := (*args)["db"].(*pg.DB)

	fmt.Println((*args)["target"].(*models.User).TgID)
	if (*args)["user"].(*models.User).IsOwner {
		var thread models.Thread
		err := db.Model(&thread).
			Where("thread_id = ?", c.Message().ThreadID).
			Where("chat_id = ?", c.Chat().ID).
			Select()
		if err != nil {
			return args, e.FromError(err, "Failed to select thread").WithSeverity(e.Critical).WithData(map[string]any{
				"user": (*args)["user"],
			})
		}
		(*args)["thread"] = &thread
		return args, e.Nil()
	}

	chat := &models.Chat{}
	err := db.Model(chat).
		Where("chat_owner_id = ?", (*args)["target"].(*models.User).TgID).
		Select()
	if err != nil {
		return args, e.FromError(err, "Failed to select chat").WithSeverity(e.Critical).WithData(map[string]any{
			"target": (*args)["target"],
		})
	}

	(*args)["chat"] = chat

	var thread models.Thread

	err = db.Model(&thread).
		Where("chat_id = ?", chat.TgID).
		Where("associated_user_id = ?", (*args)["user"].(*models.User).TgID).
		Select()

	if err == nil {
		(*args)["thread"] = &thread
		return args, e.Nil()
	}

	if err != pg.ErrNoRows {
		return args, e.FromError(err, "Failed to select thread").WithSeverity(e.Critical).WithData(map[string]any{
			"target": (*args)["target"],
			"user": (*args)["user"],
		})
	}

	// TODO: Создать массив с IDшниками иконок и выбирать случайную
	fmt.Println(1)
	t, err := c.Bot().CreateTopic(
		&tele.Chat{
			ID: chat.TgID,
		},
		&tele.Topic{
			Name: fmt.Sprintf("@%s", c.Sender().Username),
			// IconCustomEmojiID: "5199590728270886590",
		},
	)

	if err != nil {
		return args, e.FromError(err, "Failed to create topic").WithSeverity(e.Critical).WithData(map[string]any{
			"chat": chat,
			"user": (*args)["user"],
		})
	}

	thread = models.Thread{
		ThreadID: t.ThreadID,
		ChatID: chat.TgID,
		AssociatedUserID: (*args)["user"].(*models.User).TgID,
	}

	_, err = db.Model(&thread).Insert()
	if err != nil {
		return args, e.FromError(err, "Failed to insert thread").WithSeverity(e.Critical).WithData(map[string]any{
			"thread": thread,
			"chat": chat,
			"user": (*args)["user"],
		})
	}

	(*args)["thread"] = &thread

	return args, e.Nil()
}