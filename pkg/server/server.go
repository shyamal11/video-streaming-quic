package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"drexel.edu/net-quic/pkg/pdu"
	"drexel.edu/net-quic/pkg/util"
	"github.com/quic-go/quic-go"
)

type ServerConfig struct {
	GenTLS   bool
	CertFile string
	KeyFile  string
	Address  string
	Port     int
}

type Server struct {
	cfg ServerConfig
	tls *tls.Config
	ctx context.Context
}

func NewServer(cfg ServerConfig) *Server {
	server := &Server{
		cfg: cfg,
	}
	server.tls = server.getTLS()
	server.ctx = context.TODO()
	return server
}

func (s *Server) getTLS() *tls.Config {
	if s.cfg.GenTLS {
		tlsConfig, err := util.GenerateTLSConfig()
		if err != nil {
			log.Fatal(err)
		}
		return tlsConfig
	} else {
		tlsConfig, err := util.BuildTLSConfig(s.cfg.CertFile, s.cfg.KeyFile)
		if err != nil {
			log.Fatal(err)
		}
		return tlsConfig
	}
}

func (s *Server) Run() error {
	address := fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.Port)
	listener, err := quic.ListenAddr(address, s.tls, nil)
	if err != nil {
		log.Printf("error listening: %s", err)
		return err
	}

	//SERVER LOOP
	for {
		log.Println("Accepting new session")
		sess, err := listener.Accept(s.ctx)
		if err != nil {
			log.Printf("error accepting: %s", err)
			return err
		}

		go s.streamHandler(sess)
	}
}

func (s *Server) streamHandler(sess quic.Connection) {
	for {
		log.Print("[server] waiting for client to open stream")
		stream, err := sess.AcceptStream(s.ctx)
		if err != nil {
			log.Printf("[server] stream closed: %s", err)
			break
		}

		//Handle protocol activity on stream
		s.protocolHandler(stream)
	}
}

// Update protocolHandler function
func (s *Server) protocolHandler(stream quic.Stream) error {
    buff := pdu.MakePduBuffer()

    // Open the file to save the received video
    file, err := os.Create("../received_video.mp4")
    if err != nil {
        log.Printf("[server] error creating video file: %s", err)
        return err
    }
    defer file.Close()

    for {
        n, err := stream.Read(buff)
        if err != nil {
            log.Printf("[server] Error Reading Raw Data: %s", err)
            return err
        }

        data, err := pdu.PduFromBytes(buff[:n])
        if err != nil {
            log.Printf("[server] Error decoding PDU: %s", err)
            return err
        }

        if data.Mtype == pdu.TYPE_VIDEO {
			log.Printf("[server] received video data of lengthdddd %d: %v", len(data.Data), data.Data)
            _, err = file.Write(data.Data)
            if err != nil {
                log.Printf("[server] error writing to video file: %s", err)
                return err
            }
            log.Printf("[server] received video data of length %d", len(data.Data))
        } else {
            log.Printf("[server] received non-video data")
        }
    }
}