package telegram

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"telegrambot/internal/e"
	"telegrambot/pkg/clients/telegram"
	"telegrambot/pkg/repository"
)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

func (p *Processor) doCmd(ctx context.Context, text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command '%s' from '%s'", text, username)

	if isAddCmd(text) {
		return p.savePage(ctx, text, chatID, username)
	}

	switch text {
	case RndCmd:
		return p.sendRandom(ctx, chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	default:
		msg := telegram.MessageConfig{
			ChatID: chatID,
			Text:   msgUnknownCommand,
		}
		return p.tg.SendMessage(msg)
	}

}

func (p *Processor) savePage(ctx context.Context, pageURL string, chatID int, username string) (err error) {
	defer func() {
		err = e.WrapIfErr("can't do cmd save page", err)
	}()

	msg := telegram.MessageConfig{
		ChatID: chatID,
	}

	page := &repository.Page{
		URL:      pageURL,
		Username: username,
	}

	isExists, err := p.repository.IsExists(ctx, page)
	if err != nil {
		return err
	}

	if isExists {
		msg.Text = msgAlreadyExists
		return p.tg.SendMessage(msg)
	}

	if err = p.repository.Save(ctx, page); err != nil {
		return err
	}

	msg.Text = msgSaved
	if err = p.tg.SendMessage(msg); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(ctx context.Context, chatID int, username string) (err error) {
	defer func() {
		err = e.WrapIfErr("can't do cmd save page", err)
	}()

	msg := telegram.MessageConfig{
		ChatID: chatID,
	}

	page, err := p.repository.PickRandom(ctx, username)
	if err != nil && !errors.Is(err, repository.ErrNoSavedPages) {
		return err
	}
	if errors.Is(err, repository.ErrNoSavedPages) {
		msg.Text = msgNoSavedPages
		return p.tg.SendMessage(msg)
	}

	msg.Text = page.URL
	if err = p.tg.SendMessage(msg); err != nil {
		return err
	}

	return p.repository.Remove(ctx, page)
}

func (p *Processor) sendHelp(chatID int) error {
	msg := telegram.MessageConfig{
		ChatID: chatID,
		Text:   msgHelp,
	}
	return p.tg.SendMessage(msg)
}

func (p *Processor) sendHello(chatID int) error {

	msg := telegram.MessageConfig{
		ChatID: chatID,
		Text:   msgHello,
	}

	return p.tg.SendMessage(msg)

}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
