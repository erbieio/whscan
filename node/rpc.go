package node

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"server/node/jsonrpc"
)

type RPC struct {
	transport transport
	rawURL    string
}

// NewRPC connects RPC client to the given URL.
func NewRPC(rawurl string) (*RPC, error) {
	client, err := NewClient(context.Background(), rawurl)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewClient(ctx context.Context, rawURL string) (*RPC, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse url")
	}

	var transport transport

	switch parsedURL.Scheme {
	case "http", "https":
		transport, err = newHTTPTransport(ctx, parsedURL)
	case "wss", "ws":
		transport, err = newWebsocketTransport(ctx, parsedURL)
	default:
		transport, err = newIPCTransport(ctx, parsedURL)
	}

	if err != nil {
		return nil, errors.Wrap(err, "could not create client transport")
	}

	return &RPC{
		transport: transport,
		rawURL:    rawURL,
	}, nil
}

func (c *RPC) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {
	return c.transport.Request(ctx, r)
}

func (c *RPC) Call(result interface{}, method string, args ...interface{}) error {
	ctx := context.Background()
	return c.CallContext(ctx, result, method, args...)
}

func (c *RPC) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	if result != nil && reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("call result parameter must be pointer or nil interface: %v", result)
	}

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: method,
		Params: jsonrpc.MustParams(args...),
	}

	response, err := c.Request(ctx, &request)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return errors.New(string(*response.Error))
	}
	return json.Unmarshal(response.Result, &result)
}

type BatchElem struct {
	Method string
	Args   []interface{}
	Result interface{}
	Error  error
}

func (c *RPC) BatchCall(b []BatchElem) error {
	ctx := context.Background()
	return c.BatchCallContext(ctx, b)
}

func (c *RPC) BatchCallContext(ctx context.Context, b []BatchElem) error {
	for _, elem := range b {
		elem.Error = c.CallContext(ctx, elem.Result, elem.Method, elem.Args...)
	}
	return nil
}

type transport interface {
	// Request method can be used to send JSONRPC requests and receive JSONRPC responses
	Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error)
}

func newHTTPTransport(ctx context.Context, parsedURL *url.URL) (transport, error) {
	return &httpTransport{
		rawURL: parsedURL.String(),
	}, nil
}

type httpTransport struct {
	rawURL string
	client *http.Client
	once   sync.Once
}

func (t *httpTransport) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode request json")
	}

	body, err := t.dispatchBytes(ctx, b)
	if err != nil {
		return nil, errors.Wrap(err, "could not dispatch request")
	}

	jr := jsonrpc.RawResponse{}
	err = json.Unmarshal(body, &jr)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode response json")
	}

	return &jr, nil
}

func (t *httpTransport) dispatchBytes(ctx context.Context, input []byte) ([]byte, error) {
	t.once.Do(func() {
		// Since this client is only ever used to access a single endpoint,
		// we allow all the idle connections to point that host
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.MaxIdleConnsPerHost = tr.MaxIdleConns
		t.client = &http.Client{
			Timeout:   120 * time.Second,
			Transport: tr,
		}
	})

	r, err := http.NewRequest(http.MethodPost, t.rawURL, bytes.NewReader(input))
	if err != nil {
		return nil, errors.Wrap(err, "could not create http.Request")
	}

	r = r.WithContext(ctx)
	r.Header.Add("Content-Type", "application/json")

	resp, err := t.client.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error in client.Do")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading body")
	}

	return body, nil
}

type websocketTransport struct {
	*loopingTransport
}

// newWebsocketTransport creates a Connection to the passed in URL.  Use the supplied Context to shutdown the connection by
// cancelling or otherwise aborting the context.
func newWebsocketTransport(ctx context.Context, addr *url.URL) (transport, error) {
	wsConn, _, err := websocket.DefaultDialer.DialContext(ctx, addr.String(), nil)
	if err != nil {
		return nil, err
	}

	readMessage := func() (payload []byte, err error) {
		typ, r, err := wsConn.NextReader()
		if err != nil {
			return nil, errors.Wrap(err, "error reading from backend websocket connection")
		}

		if typ != websocket.TextMessage {
			return nil, nil
		}

		payload, err = io.ReadAll(r)
		if err != nil {
			return nil, errors.Wrap(err, "error reading from backend websocket connection")
		}

		return payload, err
	}

	writeMessage := func(payload []byte) error {
		err := wsConn.WriteMessage(websocket.TextMessage, payload)
		return err
	}

	t := websocketTransport{
		loopingTransport: newLoopingTransport(ctx, wsConn, readMessage, writeMessage),
	}

	return &t, nil
}

func newIPCTransport(ctx context.Context, parsedURL *url.URL) (*ipcTransport, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", parsedURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not connect over IPC")
	}
	scanner := bufio.NewScanner(conn)
	readMessage := func() (payload []byte, err error) {
		if !scanner.Scan() {
			return nil, ctx.Err()
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		payload = []byte(scanner.Text())
		err = nil
		return
	}

	writeMessage := func(payload []byte) error {
		_, err := conn.Write(payload)
		return err
	}

	t := ipcTransport{
		loopingTransport: newLoopingTransport(ctx, conn, readMessage, writeMessage),
	}

	return &t, nil
}

type ipcTransport struct {
	*loopingTransport
}
