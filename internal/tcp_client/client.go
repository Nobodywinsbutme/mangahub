package tcp_client

import (
	"bufio"
	"fmt"
	"net"
)

type Client struct {
	conn net.Conn
	addr string
}

func Connect(host string, port string) (*Client, error) {
	address := fmt.Sprintf("%s:%s", host, port)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		addr: address,
	}, nil
}

func (c *Client) SendProgress(userID, mangaID string, chapter int) error {
	message := fmt.Sprintf("%s|%s|%d\n", userID, mangaID, chapter)
	_, err := c.conn.Write([]byte(message))
	return err
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func ListenForUpdates(client *Client, handler func(string) error) error {
	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		msg := scanner.Text()
		if err := handler(msg); err != nil {
			return err
		}
	}
	return scanner.Err()
}
