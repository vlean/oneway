package netx

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message cap allowed from peer.
	maxMessageSize = 512
)

var (
	Upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func NewWebsocket(conn *websocket.Conn) *Ws {
	ws := &Ws{
		conn: conn,
	}
	go ws.Read()
	go ws.Write()
	return ws
}

type Ws struct {
	conn *websocket.Conn
}

func (w *Ws) Read() {
	defer func() {
		w.conn.Close()
	}()
	w.conn.SetReadLimit(maxMessageSize)
	w.conn.SetReadDeadline(time.Now().Add(pongWait))
	w.conn.SetPongHandler(func(string) error {
		w.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		log.Debugf("receive message %v", string(message))
	}
}

func (w *Ws) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		w.conn.Close()
	}()
	for range ticker.C {
		w.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := w.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
		if err := w.conn.WriteJSON(map[string]string{"hi": "json"}); err != nil {
			return
		}
	}
}

func WrapH(f func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			log.WithError(err).WithContext(r.Context()).Error("handel error")
		}
	}
}
