package station

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/vonaka/smc_station/config"
	"github.com/vonaka/smc_station/hls"
	"github.com/vonaka/smc_station/viewer"
)

type Station struct {
	c         *config.Config
	vs        map[*viewer.Viewer]struct{}
	clock     *Clock
	leave     chan *viewer.Viewer
	shutdown  chan struct{}
	newViewer chan *viewer.Viewer
}

func New(c *config.Config) *Station {
	return &Station{
		c:         c,
		vs:        make(map[*viewer.Viewer]struct{}),
		clock:     NewClock(),
		leave:     make(chan *viewer.Viewer, 10),
		shutdown:  make(chan struct{}, 1),
		newViewer: make(chan *viewer.Viewer, 10),
	}
}

func greetViewer(v *viewer.Viewer, wait bool, startTime *time.Time) {
	if !wait {
		v.Record(&viewer.Action{
			Type: "start",
		})
	} else {
		v.Record(&viewer.Action{
			Type: "wait",
			Wait: startTime.Format(time.RFC3339),
		})
	}
}

func (s *Station) Start() {
	go func() {
		ready := make(chan struct{}, 0)
		var program *hls.Program
		for {
			// create a new program
			go func() {
				fmt.Println("preparing a new program")
				if program != nil {
					if err := program.Next(); err == hls.ErrTankIndex {
						program = nil
					} else if err != nil {
						log.Fatal(err)
					}
				}
				if program == nil {
					p, err := hls.MakeProgram(s.c)
					if err != nil {
						// TODO: maybe do something smarter with this err
						log.Fatal(err)
					}
					program = p
				}
				f := filepath.Join(s.c.StaticDir(), "program")
				if _, err := os.Stat(f); os.IsNotExist(err) {
					if err = os.Mkdir(f, 0775); err != nil {
						log.Fatal(err)
					}
				} else {
					if err = cleanProgramDir(f); err != nil {
						log.Fatal(err)
					}
				}
				f = filepath.Join(f, "program.m3u8")
				if err := program.Write(f); err != nil {
					log.Fatal(err)
				}
				fmt.Println("the program is ready")
				ready <- struct{}{}
			}()

			ok, waitDuration, playDuration := s.c.ReadyToPlay()
			if !ok {
				startTime := time.Now().Add(waitDuration)
				wait := time.After(waitDuration)
				for v := range s.vs {
					greetViewer(v, true, &startTime)
				}
				pdone := false
				wdone := false
			loop:
				for {
					select {
					case v := <-s.newViewer:
						s.vs[v] = struct{}{}
						greetViewer(v, true, &startTime)
					case v := <-s.leave:
						delete(s.vs, v)
					case <-wait:
						wdone = true
						if pdone {
							break loop
						}
					case <-ready:
						pdone = true
						if wdone {
							break loop
						}
					}
				}
			} else {
			loopOk:
				for {
					select {
					case v := <-s.newViewer:
						s.vs[v] = struct{}{}
					case v := <-s.leave:
						delete(s.vs, v)
					case <-ready:
						break loopOk
					}
				}
			}
			err := s.play(playDuration)
			if err != nil {
				if errors.Is(err, ErrShut) {
					fmt.Println(err)
					return
				}
				log.Fatal(err)
			}
		}
	}()
}

func cleanProgramDir(dir string) error {
	d, err := os.Open(dir)
    if err != nil {
        return err
    }
    defer d.Close()

	fs, err := d.Readdirnames(-1)
    if err != nil {
        return err
    }

	for _, f := range fs {
		if ext := filepath.Ext(f); ext == ".ts" || ext == ".m3u8" {
			err = os.RemoveAll(filepath.Join(dir, f))
			if err != nil {
				return err
			}
		}
    }
    return nil
}

func (s *Station) play(d time.Duration) error {
	for v := range s.vs {
		greetViewer(v, false, nil)
	}

	wait := time.After(d)
	s.clock.Reset()
	for {
		select {
		case v := <-s.newViewer:
			s.vs[v] = struct{}{}
			greetViewer(v, false, nil)
		case v := <-s.leave:
			delete(s.vs, v)
		case <-s.shutdown:
			return ErrShut
		case <-wait:
			// TODO: notify the viewers that the show is over
			return nil
		}
	}
	return nil
}

func (s *Station) Time() int {
	return s.clock.Time()
}

func (s *Station) Shutdown() {
	s.shutdown <- struct{}{}
}

func (s *Station) AddViewer(v *viewer.Viewer) {
	s.newViewer <- v
}

func (s *Station) Leave(v *viewer.Viewer) {
	s.leave <- v
}

var (
	ErrShut error = errors.New("station is shut")
	defaultStation *Station
)

func Initialize(c *config.Config) {
	defaultStation = New(c)
}

func Start() {
	if defaultStation != nil {
		defaultStation.Start()
	}
}

func Shutdown() {
	if defaultStation != nil {
		defaultStation.Shutdown()
	}
}

func AddViewer(v *viewer.Viewer) {
	if defaultStation != nil {
		defaultStation.AddViewer(v)
	}
}

func Leave(v *viewer.Viewer) {
	if defaultStation != nil {
		defaultStation.Leave(v)
	}
}

func Time() int {
	if defaultStation != nil {
		return defaultStation.Time()
	}
	return 0
}
