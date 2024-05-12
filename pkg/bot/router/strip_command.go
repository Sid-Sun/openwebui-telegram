package router

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

func StripCommand(command string) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			_, after, found := strings.Cut(c.Message().Text, command+" ")
			if found {
				c.Message().Text = after
			}
			return next(c)
		}
	}
}
