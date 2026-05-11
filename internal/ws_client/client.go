package ws_client

import (
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
}

func Connect(host string, port string, username string) (*Client, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   "/ws",
	}
	q := u.Query()
	q.Set("username", username)
	u.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

type Message struct {
	Username  string `json:"username"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

func (c *Client) Send(text string) error {
	type outbound struct {
		Text string `json:"text"`
	}
	msg := outbound{Text: text}
	return c.conn.WriteJSON(msg)
}

func (c *Client) Receive(handler func(Message) error) error {
	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			return err
		}
		if err := handler(msg); err != nil {
			return err
		}
	}
}
