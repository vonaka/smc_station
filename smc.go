package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/vonaka/smc_station/config"
	"github.com/vonaka/smc_station/hls"
	"github.com/vonaka/smc_station/station"
	"github.com/vonaka/smc_station/webserver"
)

func main() {
	if runtime.GOOS != "linux" {
		log.Fatal("unsupported operating system")
	}

	home := flag.String("home", "$HOME", "smc home `directory`")
	conf := flag.String("config", "smc.conf", "smc configuration `file`")
	logger := flag.String("log", "<stderr>", "logger destination `file`")
	httpAddr := flag.String("http", ":8080", "web server `address`")

	flag.Parse()

	if *home == "$HOME" {
		dir, err := os.UserHomeDir()
		check(err)
		*home = dir
	}
	if *logger != "<stderr>" {
		l, err := os.OpenFile(*logger, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
        check(err)
		log.SetOutput(l)
	}

	c, err := config.Open(filepath.Join(*home, *conf))
	check(err)
	station.Initialize(c)
	err = hls.InitializeDataTank(c)
	check(err)
	station.Start()
	fmt.Println("starting server at", *httpAddr)
	err = webserver.Serve(*httpAddr, c.StaticDir(), c.DataDir())
	check(err)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
