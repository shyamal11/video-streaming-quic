# video-streaming-quic
A video streaming protocol implementation  using QUIC.

# GROUP 

- **Evan Velagaleti:** ev379@drexel.edu
- **Geeta Kukreja:** gk457@drexel.edu
- **Shyamal Mishra:** sm4825@drexel.edu

## Overview

Our protocol is designed to stream videos from a server to multiple clients using the QUIC protocol. QUIC provides several advantages over traditional protocols, including reduced latency, improved connection resilience, and multiplexing capabilities, which makes it ideal for real-time video streaming applications.

## Features of QUIC in Our Protocol

### Authentication

We've incorporated TLS to secure our connections. The client uses a certificate file for authentication, ensuring that only authorized clients can connect to the server. If a certificate file is not provided, the client defaults to using a basic TLS configuration.

### Multi-client Support

Our server can handle multiple clients simultaneously. Each client initiates a connection and opens a stream to communicate with the server. The server can manage multiple streams concurrently, allowing multiple clients to stream videos at the same time.

### Error Handling

Error handling is critical in our protocol. Both the client and server log any errors encountered during communication. If an error occurs, such as an issue with reading from a stream or decoding a PDU, it is logged.
### Chat Messages from Client to Server

Clients can send chat messages to the server. This is part of the initialization handshake, where the client sends a "hello from client" message. The server responds with a menu of video options for the client to choose from.

### Video Controls

Clients can control video playback using keyboard inputs. They can pause and play video stream. This control is implemented by sending specific command from the client to the server, which processes them and adjusts the video stream by stopping the video reading and decoding stream accordingly. WHen Played again, the stream bytes continues. 

## Stateful Protocol with DFA

Both the client and server implement a stateful protocol, ensuring that the protocol adheres to our deterministic finite automaton (DFA). This ensures that both ends of the communication can validate the state of the connection at any given time.

### States of the DFA

Our DFA includes the following states:

1. **Initial State**: The client sends a "hello" message to the server.
2. **Waiting for Server Response**: The server responds with a video menu.
3. **Waiting for Client Choice**: The client sends their video choice.
4. **Streaming Video**: The server streams the video to the client.
5. **Handling Controls**: The client sends control command ie PAUSE and PLAY.
6. **Error State**: If an error occurs, the connection logs the error and attempts to recover.
7. **Waiting for Acknowledgments**: The server waits for acknowledgments from the client to ensure data integrity.

## Configuration and Execution

### Server Configuration

The server binds to a hardcoded port number, which you can specify in the configuration file. It uses the following settings:

- **GenTLS**: Flag to generate TLS config.
- **CertFile**: Path to the certificate file.
- **KeyFile**: Path to the key file.
- **Address**: Server address.
- **Port**: Server port.

### Client Configuration

The client can specify the server's hostname or IP address via a configuration file or command line arguments. The configuration includes:

- **ServerAddr**: Server address.
- **PortNumber**: Server port.
- **CertFile**: Path to the certificate file.

### Initial configuration in Echo:
PORT_NUMBER  = 4242
SERVER PARAMETERS:
SERVER_IP = "0.0.0.0"
CLIENT PARAMETERS 
SERVER_ADDR = "localhost"

### Command Line Output:

With the command line, the user enters the video choice, and the FFMPEG player is used to decode the bytes and stream the video.

---

### There is a single binary that is used to run both the client and the server

- **server:** go run cmd/echo/echo.go -server
- **client:** go run cmd/echo/echo.go -client

This should be done on separate terminals, and can handle multiple clients. 
When the client starts, **enter a number on Client CLI** which requestes the required video.  
Every client will open the video player, and the **SpaceBar key** to **play / pause ** each stream. 

