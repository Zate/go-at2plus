package at2plus

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"
)

// Client represents a connection to an AirTouch 2+ device.
type Client struct {
	conn           net.Conn
	addr           string
	port           int
	requestTimeout time.Duration
	logger         *slog.Logger
	mu             sync.Mutex
	pending        map[uint8]chan *Packet
	pendingMu      sync.Mutex
	nextMsgID      uint8
	closeCh        chan struct{}
	isClosed       bool
}

// NewClient creates a new client and connects to the device.
// The context is used for the connection timeout.
// Options can be provided to configure the client behavior.
func NewClient(ctx context.Context, ip string, opts ...ClientOption) (*Client, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("invalid option: %w", err)
		}
	}

	// Apply connect timeout to context if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.connectTimeout)
		defer cancel()
	}

	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", cfg.port))
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	c := &Client{
		conn:           conn,
		addr:           addr,
		port:           cfg.port,
		requestTimeout: cfg.requestTimeout,
		logger:         cfg.logger,
		pending:        make(map[uint8]chan *Packet),
		closeCh:        make(chan struct{}),
	}

	go c.readLoop()

	return c, nil
}

// Close closes the connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isClosed {
		return nil
	}
	c.isClosed = true
	close(c.closeCh)
	return c.conn.Close()
}

func (c *Client) readLoop() {
	for {
		select {
		case <-c.closeCh:
			return
		default:
			// Read header first? Or just read chunks.
			// Since TCP is a stream, we should read carefully.
			// For simplicity in this proof of concept, we'll read into a buffer and try to decode.
			// A robust implementation would use a scanner or buffer accumulation.
			// Given the packet format has a length, we can read header then length then data.

			// Read Header (2 bytes)
			// Actually, let's just read a chunk and assume packets don't span too weirdly for now,
			// or better, implement a proper reader.

			// Proper reader implementation:
			headerBuf := make([]byte, 8) // Header(2)+Addr(2)+ID(1)+Type(1)+Len(2)
			_, err := io.ReadFull(c.conn, headerBuf)
			if err != nil {
				// Handle error (reconnect or close)
				c.Close()
				return
			}

			// Decode header to get length
			// We can use Decode() but it expects full packet.
			// Let's parse manually just for length.
			// HeaderBytes = 0x5555
			// DataLen is at index 6 (2 bytes)
			// Wait, Decode() checks header.

			// Check header
			if headerBuf[0] != 0x55 || headerBuf[1] != 0x55 {
				// Invalid header, maybe out of sync.
				// For now, just continue/fail.
				continue
			}

			dataLen := int(headerBuf[6])<<8 | int(headerBuf[7])

			// Validate data length to prevent excessive memory allocation
			if dataLen > MaxDataLen {
				// Skip this packet - likely malformed or malicious
				continue
			}

			// Read Data + CRC (2 bytes)
			toRead := dataLen + 2
			dataBuf := make([]byte, toRead)
			_, err = io.ReadFull(c.conn, dataBuf)
			if err != nil {
				c.Close()
				return
			}

			// Combine
			fullPacket := append(headerBuf, dataBuf...)
			packet, err := Decode(fullPacket)
			if err != nil {
				// Invalid packet, skip and continue reading
				continue
			}

			// Dispatch
			c.pendingMu.Lock()
			ch, ok := c.pending[packet.MsgID]
			if ok {
				ch <- packet
				delete(c.pending, packet.MsgID)
			}
			c.pendingMu.Unlock()
		}
	}
}

func (c *Client) sendRequest(ctx context.Context, msgType uint8, data []byte) (*Packet, error) {
	c.mu.Lock()
	msgID := c.nextMsgID
	c.nextMsgID++
	c.mu.Unlock()

	// Determine Address based on MsgType
	addr := AddressSendStandard
	if msgType == MsgTypeExtended {
		addr = AddressSendExtended
	}

	p := NewPacket(uint16(addr), msgID, msgType, data)
	encoded := p.Encode()

	// Register channel
	respCh := make(chan *Packet, 1)
	c.pendingMu.Lock()
	c.pending[msgID] = respCh
	c.pendingMu.Unlock()

	// Send
	_, err := c.conn.Write(encoded)
	if err != nil {
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, err
	}

	// Apply request timeout if context has no deadline
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.requestTimeout)
		defer cancel()
	}

	// Wait for response
	select {
	case resp := <-respCh:
		return resp, nil
	case <-ctx.Done():
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("request canceled: %w", ctx.Err())
	}
}

// GetGroupStatus requests status for all groups.
func (c *Client) GetGroupStatus(ctx context.Context) ([]GroupStatus, error) {
	payload := []byte{SubMsgTypeGroupStatus, 0, 0, 0, 0, 0, 0, 0}

	resp, err := c.sendRequest(ctx, MsgTypeControlStatus, payload)
	if err != nil {
		return nil, err
	}

	return UnmarshalGroupStatus(resp.Data)
}

// GetACStatus requests status for all ACs.
func (c *Client) GetACStatus(ctx context.Context) ([]ACStatus, error) {
	payload := []byte{SubMsgTypeACStatus, 0, 0, 0, 0, 0, 0, 0}

	resp, err := c.sendRequest(ctx, MsgTypeControlStatus, payload)
	if err != nil {
		return nil, err
	}

	return UnmarshalACStatus(resp.Data)
}

// SetGroupControl sends a control command to groups.
func (c *Client) SetGroupControl(ctx context.Context, groups []GroupControl) error {
	data, err := MarshalGroupControl(groups)
	if err != nil {
		return err
	}

	_, err = c.sendRequest(ctx, MsgTypeControlStatus, data)
	return err
}

// SetACControl sends a control command to ACs.
func (c *Client) SetACControl(ctx context.Context, acs []ACControl) error {
	data, err := MarshalACControl(acs)
	if err != nil {
		return err
	}

	_, err = c.sendRequest(ctx, MsgTypeControlStatus, data)
	return err
}

// GetACAbility requests the capabilities of a specific AC unit.
func (c *Client) GetACAbility(ctx context.Context, acNum uint8) ([]ACAbility, error) {
	payload := []byte{0xFF, ExtMsgTypeACAbility, acNum}

	resp, err := c.sendRequest(ctx, MsgTypeExtended, payload)
	if err != nil {
		return nil, err
	}

	return UnmarshalACAbility(resp.Data)
}

// GetGroupNames requests names for all groups.
func (c *Client) GetGroupNames(ctx context.Context) ([]GroupName, error) {
	payload := []byte{0xFF, ExtMsgTypeGroupName}

	resp, err := c.sendRequest(ctx, MsgTypeExtended, payload)
	if err != nil {
		return nil, err
	}

	return UnmarshalGroupName(resp.Data)
}
