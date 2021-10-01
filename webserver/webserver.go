package webserver

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"path/filepath"
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
	};
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

func Serve(address, staticRoot, dataDir string) error {
	wrapper := fileWrapper{
		handler:   http.FileServer(http.Dir(staticRoot)),
		staticDir: filepath.Join(staticRoot, "program"),
	}
	http.Handle("/", wrapper)

	//http.Handle("/", http.FileServer(http.Dir(staticRoot)))

	//http.Handle("/data", http.FileServer(http.Dir(dataDir)))
	//http.HandleFunc("/time.json", timeHandler)
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

// type Time struct {
// 	Time        string `json:"time"`
// 	DisplayName string `json:"displayName,omitempty"`
// }

// func timeHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("content-type", "application/json")
// 	w.Header().Set("cache-control", "no-cache")

// 	fmt.Println("time handler")

// 	if r.Method == "HEAD" {
// 		return
// 	}

// 	e := json.NewEncoder(w)
// 	e.Encode(Time{
// 		Time: time.Now().String(),
// 	})
// }

// type Program struct {
// 	Type        string `json:"type"`
// 	DisplayName string `json:"displayName,omitempty"`
// }

// Old one
// func wsHandler(w http.ResponseWriter, r *http.Request) {
// 	conn, err := wsUpgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Printf("Websocket upgrade: %v", err)
// 		return
// 	}
// 	go func() {
// 		conn.WriteJSON(Program{
// 			Type: "program",
// 		})
// 		fmt.Println("new client")
// 		fmt.Println("loop...")
// 		c := make(chan int, 2)
// 		<-c
// 		// err := rtpconn.StartClient(conn)
// 		// if err != nil {
// 		// 	log.Printf("client: %v", err)
// 		// }
// 	}()
// }

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
