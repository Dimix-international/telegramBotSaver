package telegram

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"telegramBotSaver/lib/e"
)

const (
	getUpdatesMethod = "getUpdates"
	sendMessageMethod = "sendMessage"
)

type Client struct {
	host     string
	basePath string //tg-bot.com/bot<token>
	client   http.Client
}

func New(host string, token string) *Client {
	return &Client{
		host: host,
		basePath: newBasePath(token),
		client: http.Client{},
	}
}

func (c *Client) Updates(offset int, limit int) (updates []Update, err error){
	defer func() {
		err = e.WrapIfErr("can't do request", err)
	} ()

	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	data, err := c.doRequest(getUpdatesMethod, q)
	if err != nil {
		return nil,  err
	}

	//распарсим ответ
	var res UpdateResponse

	if err := json.Unmarshal(data, &res); err != nil {
		return nil,  err
	}

	return res.Result, nil
}

func (c *Client) SendMessage(chatId int, text string) error {
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatId))
	q.Add("text", text)

	_, err := c.doRequest(sendMessageMethod, q)
	if err != nil {
		e.Wrap("can't send message", err)
	}

	return nil
}

func (c *Client) doRequest(method string, query url.Values) (data []byte, err error) {

	defer func() {
		err = e.WrapIfErr("can't do request", err)
	} ()

	u := url.URL{
		Scheme: "https",
		Host: c.host,
		Path: path.Join(c.basePath, method), //чтобы не было проблем со /
	}

	//формируем объект запроса
	req, err := http.NewRequest(http.MethodGet, u.String(), nil) //3-ий параметр тело запроса
	if err != nil {
		return nil, err
	}

	//передаем параметры запроса
	req.URL.RawQuery = query.Encode()

	//отправляем запрос
	resp, err := c.client.Do(req)

	if err != nil {
		return nil,  err
	}

	//закрываем тело ответа
	defer func() {_ = resp.Body.Close()} ()

	//получаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil,  err
	}

	return body, nil
}

func newBasePath(token string) string {
	return "bot" + token
}