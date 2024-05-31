package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"encoding/hex"
	"io"
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
    stream, err := c.conn.OpenStreamSync(c.ctx)
    if err != nil {
        log.Printf("[cli] error opening stream %s", err)
        return err
    }
    defer stream.Close()

    // Open the video file
    file, err := os.Open("../test.mp4") // Replace with the path to your video file
    if err != nil {
        log.Printf("[cli] error opening video file %s", err)
        return err
    }
    defer file.Close()

    buffer := make([]byte, pdu.MAX_PDU_SIZE)
    for {
        n, err := file.Read(buffer)
        if err != nil {
            if err == io.EOF {
                break
            }
            log.Printf("[cli] error reading from file %s", err)
            return err
        }

		log.Printf("[cli] read %d bytes from file: %s", n, hex.EncodeToString(buffer[:n]))

        videoPDU := pdu.NewPDU(pdu.TYPE_VIDEO, buffer[:n])
        pduBytes, err := pdu.PduToBytes(videoPDU)
        if err != nil {
            log.Printf("[cli] error encoding PDU: %s", err)
            return err
        }

        _, err = stream.Write(pduBytes)
        if err != nil {
            log.Printf("[cli] error writing to stream %s", err)
            return err
        }
    }
    log.Printf("[cli] video streaming completed")
    return nil
}
