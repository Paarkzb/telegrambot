package telegram

import (
	"context"
	"errors"
	"fmt"
	"telegrambot/internal/e"
	"telegrambot/pkg/clients/telegram"
	"telegrambot/pkg/events"
	"telegrambot/pkg/repository"
	"telegrambot/pkg/state"
)

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

type Processor struct {
	tg         *telegram.Client
	offset     int
	repository repository.Repository
	cache      state.Cache
}

type Meta struct {
	ChatID          int    `json:"chat_id"`
	Username        string `json:"username"`
	CallbackQueryId string `json:"callback_query_id"`
}

func New(client *telegram.Client, repository repository.Repository, cache state.Cache) *Processor {
	return &Processor{tg: client, repository: repository, cache: cache}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	for _, u := range updates {
		res = append(res, event(u))
	}

	p.offset = updates[len(updates)-1].UpdateId + 1

	return res, nil
}

func (p *Processor) Process(ctx context.Context, event events.Event) error {
	metaInfo, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}
	switch event.Type {
	case events.Message:
		//chatState, err := p.cache.GetState(ctx, fmt.Sprintf("%d", metaInfo.ChatID))
		//if err != nil {
		//	return e.Wrap("can't get state", err)
		//}
		//log.Println("chatState:", chatState)

		return p.processMessage(ctx, event)
	case events.CallbackQuery:
		err = p.cache.SetState(ctx, fmt.Sprintf("%d", metaInfo.ChatID), "waiting_for_input")
		if err != nil {
			return e.Wrap("can't set state state", err)
		}
		return p.processCallbackQuery(ctx, event)
	default:
		return e.Wrap("can't process message", ErrUnknownEventType)
	}

}

func (p *Processor) processMessage(ctx context.Context, event events.Event) error {
	metaInfo, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}

	if err := p.doCmd(ctx, event.Text, metaInfo.ChatID, metaInfo.Username); err != nil {
		return e.Wrap("can't process message", err)
	}

	return nil
}

func (p *Processor) processCallbackQuery(ctx context.Context, event events.Event) error {
	metaInfo, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}

	ans := telegram.CallbackQueryConfig{
		CallbackQueryId: metaInfo.CallbackQueryId,
		Text:            &event.Text,
	}

	return p.tg.AnswerCallbackQuery(ans)
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}

	return res, nil
}

func event(update telegram.Update) events.Event {
	updateType := fetchType(update)

	res := events.Event{
		Type: updateType,
	}

	if updateType == events.Message {
		res.Meta = Meta{
			ChatID:   update.Message.Chat.ID,
			Username: update.Message.From.Username,
		}

		res.Text = fetchText(update)
	}

	if updateType == events.CallbackQuery {
		res.Meta = Meta{
			ChatID:          update.CallbackQuery.Message.Chat.ID,
			Username:        update.CallbackQuery.From.Username,
			CallbackQueryId: update.CallbackQuery.ID,
		}

		res.Text = fetchCallbackQueryData(update)
	}

	return res
}

func fetchText(update telegram.Update) string {
	if update.Message == nil {
		return ""
	}

	return update.Message.Text
}

func fetchCallbackQueryData(update telegram.Update) string {
	if update.CallbackQuery == nil {
		return ""
	}

	return *update.CallbackQuery.Data
}

func fetchType(update telegram.Update) events.Type {
	if update.Message != nil {
		return events.Message
	}
	if update.CallbackQuery != nil {
		return events.CallbackQuery
	}
	return events.Unknown
}
