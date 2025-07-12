package websocket_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gwebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/sticks/websocket"
)

func TestWSHandler(t *testing.T) {

	testBytes := []byte("testing")

	upgrader := gwebsocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		HandshakeTimeout:  0,
		WriteBufferPool:   nil,
		Subprotocols:      nil,
		Error:             nil,
		CheckOrigin:       nil,
		EnableCompression: false,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// synchronization helpers
	doneReg := make(chan websocket.Client)
	doneUnreg := make(chan websocket.Client)

	var c websocket.Client
	var cf context.CancelFunc

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	manager := websocket.NewManager()
	go manager.Run(ctx)

	h := websocket.ServeWS(
		upgrader,
		websocket.DefaultSetupConn,
		websocket.NewClient,
		func(ctx context.Context, _cf context.CancelFunc, _c websocket.Client) {
			cf = _cf
			c = _c
			manager.RegisterClient(ctx, cf, c)
			doneReg <- c
		},
		func(_c websocket.Client) {
			manager.UnregisterClient(_c)
			_c.Wait()
			doneUnreg <- _c
		},
		50*time.Second,
		[]websocket.MessageHandler{func(c websocket.Client, b []byte) { _, _ = c.Write(b) }},
	)

	// setup and connect to the the test server using a basic websocket
	s := httptest.NewServer(h)

	defer s.Close()
	rawWS, _, err := gwebsocket.DefaultDialer.Dial(
		"ws"+strings.TrimPrefix(s.URL, "http"), nil)
	assert.NoError(t, err)
	defer func() {
		_ = rawWS.Close()
	}()

	// once registration is done, the manager should have one client
	<-doneReg
	assert.Equal(t, len(manager.Clients()), 1)

	// write a message to the server; this will be echoed back
	err = rawWS.WriteMessage(gwebsocket.TextMessage, testBytes)
	assert.NoError(t, err)
	_, msg, err := rawWS.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, msg, testBytes)

	// close the connection, which should trigger the server to cleanup and
	// unregister the client connection
	_ = rawWS.WriteControl(gwebsocket.CloseMessage, nil, time.Now().Add(1*time.Second))
	_p := <-doneUnreg
	assert.Equal(t, len(manager.Clients()), 0)
	assert.Equal(t, _p, c)
	time.Sleep(1 * time.Second)
	//FIXME: seems to be leaking goroutines
}
