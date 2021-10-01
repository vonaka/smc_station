package webserver

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vonaka/smc_station/hls"
	"github.com/vonaka/smc_station/station"
	"github.com/vonaka/smc_station/viewer"
)

type fileWrapper struct {
	handler   http.Handler
	staticDir string
}

var (
	serverLock sync.Mutex
	wsUpgrader = websocket.Upgrader{
		HandshakeTimeout: 30 * time.Second,
	}
)

func (f fileWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(filepath.Base(r.URL.String()), "now.m3u8?version=") {
		serverLock.Lock()
		src := filepath.Join(f.staticDir, "program.m3u8")
		dst := filepath.Join(f.staticDir, "now.m3u8")
		time := station.Time()
		err := hls.CopyWithTime(src, dst, time)
		if err != nil {
			log.Fatal(err)
		}
		serverLock.Unlock()
	}
	f.handler.ServeHTTP(w, r)
}

func Serve(address, staticRoot string) error {
	wrapper := fileWrapper{
		handler:   http.FileServer(http.Dir(staticRoot)),
		staticDir: filepath.Join(staticRoot, "program"),
	}
	http.Handle("/", wrapper)
	http.HandleFunc("/ws", wsHandler)

	s := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: 60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	err := s.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}
	v := viewer.New()
	station.AddViewer(v)
	go func() {
		for {
			a := v.GetAction()
			err := conn.WriteJSON(a)
			if err != nil {
				station.Leave(v)
				conn.Close()
				return
			}
		}
	}()
}
