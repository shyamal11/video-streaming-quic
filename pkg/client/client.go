package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"drexel.edu/net-quic/pkg/pdu"
	"drexel.edu/net-quic/pkg/util"
	"github.com/quic-go/quic-go"
	"golang.org/x/term"
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
	return c.receiveVideo()
}

func (c *Client) receiveVideo() error {
	// Accept the QUIC stream
	stream, err := c.conn.OpenStreamSync(c.ctx)
	if err != nil {
		log.Printf("[cli] error opening stream %s", err)
		return err
	}

	n, err := stream.Write([]byte("hello from client"))
	if err != nil {
		log.Printf("[cli] error writing to stream %s", err)
		return err
	}
	log.Printf("[cli] wrote %d bytes to stream", n)

	initData := make([]byte, 1024)
	n, err = stream.Read(initData)
	if err != nil {
		log.Printf("[cli] error reading from stream %s", err)
		return err
	}

	log.Printf("decoded message from server: %s", initData[:n])

	// // Set terminal to raw mode to capture single key presses
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Printf("Error setting terminal to raw mode: %v", err)
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// for {
	//     // Capture a single key press
	buf := make([]byte, 1)
	_, err = os.Stdin.Read(buf)
	if err != nil {
		log.Printf("Error reading key press: %v", err)
		return err
	}

	// Send key press to server
	_, err = stream.Write(buf)
	if err != nil {
		log.Printf("Error sending choice: %v", err)
		return err
	}

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
		for {
			pdu, err := pdu.PduFromFramedBytes(stream)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Printf("Error decoding PDU: %v", err)
				return
			}

			// Print received packet number
			fmt.Printf("Received packet number: %d\n", pdu.PacketNo)

			// Write video data to FFplay stdin
			_, err = inpipe.Write(pdu.Data)
			if err != nil {
				log.Printf("Error writing to FFplay: %v", err)
				return
			}

		}
		log.Println("Received Packet successfully")
	}()

	// Wait for FFplay process to complete
	err = ffmpeg.Wait()
	if err != nil {
		log.Printf("Error waiting for FFplay: %v", err)
		return err
	}

	return nil
}
