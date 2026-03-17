// Package mcp implements the Model Context Protocol
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"

	"github.com/gorilla/websocket"
)

// Transport handles communication for MCP
type Transport interface {
	Send(data []byte) error
	Receive() ([]byte, error)
	Close() error
}

// StdioTransport implements transport over stdin/stdout
type StdioTransport struct {
	encoder *json.Encoder
	decoder *json.Decoder
	closer  io.Closer
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(in io.Reader, out io.Writer, closer io.Closer) *StdioTransport {
	return &StdioTransport{
		encoder: json.NewEncoder(out),
		decoder: json.NewDecoder(in),
		closer:  closer,
	}
}

// Send sends data over stdout
func (t *StdioTransport) Send(data []byte) error {
	return t.encoder.Encode(json.RawMessage(data))
}

// Receive receives data from stdin
func (t *StdioTransport) Receive() ([]byte, error) {
	var raw json.RawMessage
	if err := t.decoder.Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	if t.closer != nil {
		return t.closer.Close()
	}
	return nil
}

// WebSocketTransport implements transport over WebSocket
type WebSocketTransport struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// NewWebSocketTransport creates a new WebSocket transport
func NewWebSocketTransport(conn *websocket.Conn) *WebSocketTransport {
	return &WebSocketTransport{conn: conn}
}

// Send sends data over WebSocket
func (t *WebSocketTransport) Send(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.conn.WriteMessage(websocket.TextMessage, data)
}

// Receive receives data from WebSocket
func (t *WebSocketTransport) Receive() ([]byte, error) {
	_, data, err := t.conn.ReadMessage()
	return data, err
}

// Close closes the WebSocket connection
func (t *WebSocketTransport) Close() error {
	return t.conn.Close()
}

// ProcessTransport implements transport over a subprocess
type ProcessTransport struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.Reader
	encoder *json.Encoder
	decoder *json.Decoder
}

// NewProcessTransport creates a new process transport
func NewProcessTransport(command string, args ...string) (*ProcessTransport, error) {
	cmd := exec.Command(command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	return &ProcessTransport{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		encoder: json.NewEncoder(stdin),
		decoder: json.NewDecoder(stdout),
	}, nil
}

// Send sends data to the process
func (t *ProcessTransport) Send(data []byte) error {
	return t.encoder.Encode(json.RawMessage(data))
}

// Receive receives data from the process
func (t *ProcessTransport) Receive() ([]byte, error) {
	var raw json.RawMessage
	if err := t.decoder.Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// Close closes the process
func (t *ProcessTransport) Close() error {
	t.stdin.Close()
	return t.cmd.Process.Kill()
}

// TCPTransport implements transport over TCP
type TCPTransport struct {
	conn net.Conn
	mu   sync.Mutex
}

// NewTCPTransport creates a new TCP transport
func NewTCPTransport(conn net.Conn) *TCPTransport {
	return &TCPTransport{conn: conn}
}

// Send sends data over TCP
func (t *TCPTransport) Send(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err := t.conn.Write(append(data, '\n'))
	return err
}

// Receive receives data from TCP
func (t *TCPTransport) Receive() ([]byte, error) {
	buf := make([]byte, 4096)
	n, err := t.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// Close closes the TCP connection
func (t *TCPTransport) Close() error {
	return t.conn.Close()
}

// TransportHandler handles transport events
type TransportHandler interface {
	HandleRequest(ctx context.Context, req *Request) (*Response, error)
	HandleNotification(ctx context.Context, notif *Notification) error
}

// ProcessTransportMessages processes messages from a transport
func ProcessTransportMessages(ctx context.Context, transport Transport, handler TransportHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			data, err := transport.Receive()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("receive error: %w", err)
			}

			// Try to parse as request or notification
			var req Request
			if err := json.Unmarshal(data, &req); err == nil && req.Method != "" {
				if req.ID != nil {
					// It's a request
					resp, err := handler.HandleRequest(ctx, &req)
					if err != nil {
						resp = NewErrorResponse(req.ID, InternalError, err.Error(), nil)
					}
					respData, _ := json.Marshal(resp)
					if err := transport.Send(respData); err != nil {
						return fmt.Errorf("send error: %w", err)
					}
				} else {
					// It's a notification
					var notif Notification
					json.Unmarshal(data, &notif)
					if err := handler.HandleNotification(ctx, &notif); err != nil {
						// Log error but continue
						fmt.Printf("notification error: %v\n", err)
					}
				}
			}
		}
	}
}
