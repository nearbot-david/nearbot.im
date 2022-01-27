package handlers

import tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type HandlerFunc func(bot *tg.BotAPI, update *tg.Update)
