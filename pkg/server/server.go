package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
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
	cfg  ServerConfig
	tls  *tls.Config
	conn quic.Connection
	ctx  context.Context
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

	// SERVER LOOP
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

		        // Handle initialization handshake
        err = s.receiveInitializationData(stream)
        if err != nil {
            log.Printf("[server] error receiving initialization data: %s", err)
            break
        }

		//Handle protocol activity on stream
		s.readVideo(stream)
	}
}



func (s *Server) receiveInitializationData(stream quic.Stream) error {
    // Read initialization message from the client
    initData := make([]byte, 1024)
    n, err := stream.Read(initData)
    if err != nil {
        return err
    }

    log.Printf("[server] received initialization message: %s", initData[:n])

    // Check if the initialization message indicates the start of video transmission
    if string(initData[:n]) == "hi" {
        // Start reading the video file and sending its contents over the stream
        err = s.readVideo(stream)
        if err != nil {
            log.Printf("[server] error reading video file: %s", err)
            return err
        }
    }

    return nil
}

func (s *Server) readVideo(stream  quic.Stream) error {
	

	defer stream.Close()

	filePath := "../test.mp4" // Path to your video file

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[server] error opening video file: %s", err)
		return err
	}
	defer file.Close()

	buffer := make([]byte, pdu.MAX_PDU_SIZE)
	currentPacketNo := uint32(1)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[server] error reading from video file: %s", err)
			return err
		}

		

		

		// Create PDU
		pdu := pdu.NewPDU(pdu.TYPE_DATA, currentPacketNo, buffer[:n])

		log.Printf("[server] Sending %d bytes of video data (Packet Number: %d)", n, currentPacketNo)
		
		
		
		pduBytes, err := pdu.ToFramedBytes()
		
		if err != nil {
			log.Printf("[server] error encoding PDU: %s", err)
			return err
		}

		_, err = stream.Write(pduBytes)
		if err != nil {
			log.Printf("[server] error writing to stream: %s", err)
			return err
		}

		currentPacketNo++ 
	}
	log.Printf("[server] video sent successfully")
	return nil
}

// package server

// import (
// 	"context"
// 	"crypto/tls"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"

// 	"drexel.edu/net-quic/pkg/pdu"
// 	"drexel.edu/net-quic/pkg/util"
// 	"github.com/quic-go/quic-go"
// )

// type ServerConfig struct {
// 	GenTLS   bool
// 	CertFile string
// 	KeyFile  string
// 	Address  string
// 	Port     int
// }

// type Server struct {
// 	cfg  ServerConfig
// 	tls  *tls.Config
// 	conn quic.Connection
// 	ctx  context.Context
// }

// func NewServer(cfg ServerConfig) *Server {
// 	server := &Server{
// 		cfg: cfg,
// 	}
// 	server.tls = server.getTLS()
// 	server.ctx = context.TODO()
// 	return server
// }

// func (s *Server) getTLS() *tls.Config {
// 	if s.cfg.GenTLS {
// 		tlsConfig, err := util.GenerateTLSConfig()
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		return tlsConfig
// 	} else {
// 		tlsConfig, err := util.BuildTLSConfig(s.cfg.CertFile, s.cfg.KeyFile)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		return tlsConfig
// 	}
// }

// func (s *Server) Run() error {
// 	address := fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.Port)
// 	listener, err := quic.ListenAddr(address, s.tls, nil)
// 	if err != nil {
// 		log.Printf("error listening: %s", err)
// 		return err
// 	}

// 	//SERVER LOOP
// 	for {
// 		log.Println("Accepting new session")
// 		sess, err := listener.Accept(s.ctx)
// 		if err != nil {
// 			log.Printf("error accepting: %s", err)
// 			return err
// 		}

// 		go s.streamHandler(sess)
// 	}
// }

// func (s *Server) streamHandler(sess quic.Connection) {
// 	stream, err := sess.OpenStream()
// 	if err != nil {
// 		log.Printf("[server] error opening stream: %s", err)
// 		return
// 	}
// 	defer stream.Close()

// 	filePath := "test.mp4" // Path to your video file

// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		log.Printf("[server] error opening video file: %s", err)
// 		return
// 	}
// 	defer file.Close()

// 	buffer := make([]byte, pdu.MAX_PDU_SIZE)
// 	for {
// 		n, err := file.Read(buffer)
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			log.Printf("[server] error reading from video file: %s", err)
// 			return
// 		}

// 		log.Printf("[server] sending %d bytes of video data", n)

// 		_, err = stream.Write(buffer[:n])
// 		if err != nil {
// 			log.Printf("[server]error writing to stream: %s", err)
// 			return
// 		}
// 	}
// 	log.Printf("[server] video sent successfully")
// }
