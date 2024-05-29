package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"bufio"
"os"
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
	conn, err := quic.DialAddr(c.ctx, serverAddr, c.tls, nil)
	if err != nil {
		log.Printf("[cli] error dialing server %s", err)
		return err
	}
	c.conn = conn
	return c.protocolHandler()
}

func (c *Client) protocolHandler() error {
	// Open a stream
	stream, err := c.conn.OpenStreamSync(c.ctx)
	if err != nil {
		log.Printf("[cli] error opening stream %s", err)
		return err
	}
	defer stream.Close()

	// Read message from user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter message to send to server: ")
	userInput, _ := reader.ReadString('\n')

	// Create PDU with user input message
	req := pdu.NewPDU(pdu.TYPE_DATA, []byte(userInput))
	pduBytes, err := pdu.PduToBytes(req)
	if err != nil {
		log.Printf("[cli] error encoding PDU: %s", err)
		return err
	}

	// Send PDU to server
	n, err := stream.Write(pduBytes)
	if err != nil {
		log.Printf("[cli] error writing to stream %s", err)
		return err
	}
	log.Printf("[cli] wrote %d bytes to stream", n)

	// Read response from server
	buffer := pdu.MakePduBuffer()
	n, err = stream.Read(buffer)
	if err != nil {
		log.Printf("[cli] error reading from stream %s", err)
		return err
	}

	// Convert received bytes to PDU
	rsp, err := pdu.PduFromBytes(buffer[:n])
	if err != nil {
		log.Printf("[cli] error decoding PDU: %s", err)
		return err
	}

	// Log response from server
	rspDataString := string(rsp.Data)
	log.Printf("[cli] got response: %s", rspDataString)

	return nil
}
