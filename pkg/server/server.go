package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"drexel.edu/net-quic/pkg/pdu"
	"drexel.edu/net-quic/pkg/util"
	"github.com/quic-go/quic-go"
)

// ServerConfig holds configuration settings for the server
type ServerConfig struct {
	GenTLS   bool   // Flag to generate TLS config
	CertFile string // Certificate file path
	KeyFile  string // Key file path
	Address  string // Server address
	Port     int    // Server port
}

// Server represents the QUIC server
type Server struct {
	cfg       ServerConfig
	tls       *tls.Config
	conn      quic.Connection
	ctx       context.Context
	clientID  int
	clientMux sync.Mutex
}

// NewServer initializes a new Server instance
func NewServer(cfg ServerConfig) *Server {
	server := &Server{
		cfg:      cfg,
		clientID: 0,
	}
	server.tls = server.getTLS()
	server.ctx = context.TODO()
	return server
}

// getTLS returns the TLS configuration based on the server settings
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

// Run starts the server and listens for incoming connections
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

		s.clientMux.Lock()
		s.clientID++
		clientID := s.clientID
		s.clientMux.Unlock()

		go s.streamHandler(sess, clientID)
	}
}

// streamHandler handles incoming streams for a given session
func (s *Server) streamHandler(sess quic.Connection, clientID int) {
	for {
		log.Print("[server] waiting for client to open stream")
		stream, err := sess.AcceptStream(s.ctx)
		if err != nil {
			log.Printf("[server] stream closed: %s", err)
			break
		}

		// Handle initialization handshake
		err = s.receiveInitializationData(stream, clientID)
		if err != nil {
			log.Printf("[server] error receiving initialization data: %s", err)
			break
		}
	}
}

// receiveInitializationData handles the initial data exchange with the client
func (s *Server) receiveInitializationData(stream quic.Stream, clientID int) error {

	// Read initialization message from the client
	initData := make([]byte, 1024)
	n, err := stream.Read(initData)
	if err != nil {
		return err
	}

	log.Printf("[server] received initialization message from client %d: %s", clientID, initData[:n])

	// User input choice received from client
	welcomeMsg := fmt.Sprintf("Client %d: What would you like to watch? \n 1: big buck bunny \n 2: Sailing Boat \n 3: Toy Train \n Enter input as number choice!", clientID)
	stream.Write([]byte(welcomeMsg))
	log.Printf("[server] sent to client %d: %s", clientID, welcomeMsg)

	// Check if the initialization message indicates the start of video transmission
	initData = make([]byte, 1)
	n, err = stream.Read(initData)
	if err != nil {
		return err
	}
	log.Printf("[server] received choice from client %d: %s", clientID, initData[:n])
	err = s.readVideo(stream, string(initData[:n]), clientID)
	if err != nil {
		log.Printf("[server] error reading video file for client %d: %s", clientID, err)
		return err
	}

	return nil
}

// readVideo reads and streams the video file based on client's choice
func (s *Server) readVideo(stream quic.Stream, inputChoice string, clientID int) error {

	defer stream.Close()
	filePath := "test.mp4"
	if inputChoice == "1" {
		filePath = "test.mp4"
	}
	if inputChoice == "2" {
		filePath = "shipvideo.mp4"
	}
	if inputChoice == "3" {
		filePath = "trainvideo.mp4"
	}

	// filePath := "../test.mp4"

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[server] error opening video file for client %d: %s", clientID, err)
		return err
	}
	defer file.Close()

	// Initializing a buffer to read chunks of the video file.
	buffer := make([]byte, pdu.MAX_PDU_SIZE)
	currentPacketNo := uint32(1)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[server] error reading from video file for client %d: %s", clientID, err)
			return err
		}

		// Create a PDU with the type DATA, the current packet number, and the data read.
		pdu := pdu.NewPDU(pdu.TYPE_DATA, currentPacketNo, buffer[:n])

		log.Printf("[server] Sending %d bytes of video data to client %d (Packet Number: %d)", n, clientID, currentPacketNo)

		// Convert the PDU to a byte slice for transmission.

		pduBytes, err := pdu.ToFramedBytes()

		if err != nil {
			log.Printf("[server] error encoding PDU for client %d: %s", clientID, err)
			return err
		}

		// Write the PDU byte slice to the stream.
		_, err = stream.Write(pduBytes)
		if err != nil {
			log.Printf("[server] error writing to stream for client %d: %s", clientID, err)
			return err
		}

		currentPacketNo++
	}
	log.Printf("[server] video sent successfully to client %d", clientID)
	return nil
}
