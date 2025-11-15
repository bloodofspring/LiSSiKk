package commandstart

import (
	"app/internal/handlers"
	"app/internal/handlers/shared"
	e "app/pkg/errors"
	"fmt"
	"time"

	tele "gopkg.in/telebot.v4"
)

func CommandStartChain() *handlers.HandlerChain {
	return handlers.HandlerChain{}.Init(
		10*time.Second,
		shared.GetOrCrateThread,
		handlers.InitChainHandler(SendGreetingMessage),
	)
}


func SendGreetingMessage(c tele.Context, args *handlers.Arg) (*handlers.Arg, *e.ErrorInfo) {
	err := c.Reply(fmt.Sprintf("Привет, %s!\n\n", c.Sender().FirstName))
	if err != nil {
		return args, e.FromError(err, "Failed to send greeting message").WithSeverity(e.Critical).WithData(map[string]any{
			"user": (*args)["user"],
		})
	}
	c.Send("-*- Для того чтобы начать просто отправь любое сообщение -*-")

	return args, e.Nil()
}
