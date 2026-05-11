package udp_client

import (
	"fmt"
	"net"
)

type Client struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

func Connect(host string, port string) (*Client, error) {
	address := fmt.Sprintf("%s:%s", host, port)

	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		addr: udpAddr,
	}, nil
}

func (c *Client) Register(username string) error {
	message := fmt.Sprintf("REGISTER|%s", username)
	_, err := c.conn.Write([]byte(message))
	return err
}

func (c *Client) ListenForNotifications(handler func(string) error) error {
	buffer := make([]byte, 1024)

	for {
		n, _, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			return err
		}

		msg := string(buffer[:n])
		if err := handler(msg); err != nil {
			return err
		}
	}
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) SendNotification(title string, chapter int) error {
	message := fmt.Sprintf("NOTIFY|%s|%d", title, chapter)
	_, err := c.conn.Write([]byte(message))
	return err
}
