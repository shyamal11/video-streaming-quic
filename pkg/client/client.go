package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	

	"os/exec"
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
	// Accept the QUIC stream
	stream, err := c.conn.AcceptStream(c.ctx)
	if err != nil {
		log.Printf("[cli] error accepting stream: %s", err)
		return err
	}
	defer stream.Close()

	// Command to run FFplay
	ffmpeg := exec.Command("ffplay", "-f", "mp4", "-i", "pipe:")
	inpipe, err := ffmpeg.StdinPipe()
	if err != nil {
		log.Printf("Error creating pipe: %v", err)
		return err
	}
	defer inpipe.Close()

	// Start FFplay process
	err = ffmpeg.Start()
	if err != nil {
		log.Printf("Error starting FFplay: %v", err)
		return err
	}

	// Copy data from stream to FFmpeg process concurrently
	go func() {
		defer stream.Close()
		_, err := io.Copy(inpipe, stream)
		if err != nil {
			log.Printf("Error copying data to FFmpeg: %v", err)
			return
		}
	}()

	// Wait for FFplay process to finish
	err = ffmpeg.Wait()
	if err != nil {
		log.Printf("FFplay exited with error: %v", err)
		return err
	}

	return nil
}


