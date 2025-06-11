package telegram

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"telegrambot/internal/e"
)

const (
	getUpdatesMethod    = "getUpdates"
	sendMessageMethod   = "sendMessage"
	answerCallbackQuery = "answerCallbackQuery"
)

type Client struct {
	host     string
	basePath string
	client   http.Client
}

func New(host string, token string) *Client {
	return &Client{
		host:     host,
		basePath: newBasePath(token),
		client:   http.Client{},
	}
}

func newBasePath(token string) string {
	return "bot" + token
}

func (c *Client) Updates(offset int, limit int) ([]Update, error) {
	query := url.Values{}
	query.Add("offset", strconv.Itoa(offset))
	query.Add("limit", strconv.Itoa(limit))

	data, err := c.doRequest(getUpdatesMethod, query)
	if err != nil {
		return nil, err
	}

	var res UpdateResponse

	if err = json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (c *Client) SendMessage(msg MessageConfig) (err error) {
	query := url.Values{}
	query.Add("chat_id", strconv.Itoa(msg.ChatID))
	query.Add("text", msg.Text)

	if msg.ReplyMarkup != nil {
		replyMarkup, err := json.Marshal(msg.ReplyMarkup)
		if err != nil {
			return e.Wrap("can't marshal msg.ReplyMarkup", err)
		}
		query.Add("reply_markup", string(replyMarkup))
	}

	_, err = c.doRequest(sendMessageMethod, query)
	if err != nil {
		return e.Wrap("cannot send message", err)
	}

	return nil
}

func (c *Client) AnswerCallbackQuery(ans CallbackQueryConfig) (err error) {
	query := url.Values{}
	query.Add("callback_query_id", ans.CallbackQueryId)
	query.Add("text", *ans.Text)
	if ans.ShowAlert != nil && *ans.ShowAlert {
		query.Add("show_alert", "true")
	}

	_, err = c.doRequest(answerCallbackQuery, query)
	if err != nil {
		return e.Wrap("cannot answer to callback_query", err)
	}

	return nil
}

func (c *Client) doRequest(method string, query url.Values) (data []byte, err error) {
	defer func() {
		err = e.WrapIfErr("cannot send http request", err)
	}()

	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basePath, method),
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = e.WrapIfErr("can't close response body", resp.Body.Close())
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
