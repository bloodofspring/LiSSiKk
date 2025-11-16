package client

import (
	"app/internal/handlers/commandAddChat"
	"app/internal/handlers/commandBlock"
	"app/internal/handlers/commandStart"
	"app/internal/handlers/newMessage"
	e "app/pkg/errors"

	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v4"
)

func OnlyOwnerMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		ownerID := viper.GetInt64("OWNER_TG_ID")
		sender := c.Sender()
		if sender != nil && sender.ID != ownerID {
			return c.Reply("You are not the owner of the bot")
		}
		
		return next(c)
	}
}

func LoadHandlers(bot *tele.Bot) *e.ErrorInfo {
	startChain := commandstart.CommandStartChain()
	addChatChain := commandaddchat.CommandAddChatChain()
	blockChain := commandblock.CommandBlockUserChain()
	newMessageChain := newmessage.NewMessageChain()

	bot.Handle("/start", startChain.Run)
	bot.Handle("/initchat", addChatChain.Run, OnlyOwnerMiddleware)
	bot.Handle("/block", blockChain.Run, OnlyOwnerMiddleware)
	for _, event := range []string{
		tele.OnText,
		tele.OnPhoto,
		tele.OnVideo,
		tele.OnDocument,
		tele.OnAudio,
		tele.OnVoice,
		tele.OnVideoNote,
		tele.OnSticker,
		tele.OnAnimation,
		tele.OnLocation,
		tele.OnContact,
		tele.OnVenue,
		tele.OnPoll,
		tele.OnDice,
		tele.OnGame,
		tele.OnInvoice,
	} {
		bot.Handle(
			event,
			newMessageChain.Run,
			func(next tele.HandlerFunc) tele.HandlerFunc {
				return func(c tele.Context) error {
					ownerID := viper.GetInt64("OWNER_TG_ID")
					sender := c.Sender()
					if sender != nil && sender.ID == ownerID && c.Chat() != nil && c.Chat().Type == tele.ChatPrivate {
						return nil
					}
					
					return next(c)
				}
			},
		)
	}

	return e.Nil()
}
