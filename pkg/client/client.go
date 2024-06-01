package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"time"

	"drexel.edu/net-quic/pkg/pdu"
	"drexel.edu/net-quic/pkg/util"
	"github.com/quic-go/quic-go"
)

type ClientConfig struct {
	ServerAddr string
	PortNumber int
	CertFile   string
}

type Client struct {
	cfg  ClientConfig
	tls  *tls.Config
	conn quic.Connection
	ctx  context.Context
}

func NewClient(cfg ClientConfig) *Client {
	cli := &Client{
		cfg: cfg,
	}

	if cfg.CertFile != "" {
		log.Printf("[cli] using cert file: %s", cfg.CertFile)
		t, err := util.BuildTLSClientConfigWithCert(cfg.CertFile)
		if err != nil {
			log.Fatal("[cli] error building TLS client config:", err)
			return nil
		}
		cli.tls = t
	} else {
		cli.tls = util.BuildTLSClientConfig()
	}

	cli.ctx = context.TODO()
	return cli
}

func (c *Client) Run() error {
	serverAddr := fmt.Sprintf("%s:%d", c.cfg.ServerAddr, c.cfg.PortNumber)
	session, err := quic.DialAddr(c.ctx, serverAddr, c.tls, nil)
	if err != nil {
		log.Printf("[cli] error dialing server %s", err)
		return err
	}
	c.conn = session
	return c.receiveVideo()
}

func (c *Client) receiveVideo() error {
	stream, err := c.conn.AcceptStream(c.ctx)
	if err != nil {
		log.Printf("[cli] error accepting stream %s", err)
		return err
	}
	defer stream.Close()



	buffer := make([]byte, pdu.MAX_PDU_SIZE)
	for {
		n, err := stream.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[cli] error reading from stream %s", err)
			return err
		}

		log.Printf("[cli] received %d bytes of video data", n)

		// Display the received data (for example, assuming stdout)
		fmt.Print(string(buffer[:n]))

		// Add some delay to simulate real-time streaming
		time.Sleep(1000 * time.Millisecond)
	}
	log.Printf("[cli] video received successfully")
	return nil
}
