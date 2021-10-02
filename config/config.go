package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	path      string
	each      time.Duration
	start     time.Time
	duration  time.Duration
	dataDir   string
	staticDir string
	ignore    map[string]struct{}
}

func Open(path string) (*Config, error) {
	fmt.Println("searching for config file", path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("config file does not exist, creating a new one...")
		_, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		each, _ := time.ParseDuration("24h")
		start, _ := time.Parse(time.Kitchen, "8:00AM")
		duration, _ := time.ParseDuration("3h")
		dataDir := filepath.Join(filepath.Dir(path), "data")
		staticDir := filepath.Join(filepath.Dir(path), "static")
		config := &Config{
			path:      path,
			each:      each,
			start:     start,
			duration:  duration,
			dataDir:   dataDir,
			staticDir: staticDir,
			ignore:    make(map[string]struct{}),
		}
		config.Write()
		return config, nil
	} else {
		c := &Config{
			path:   path,
			ignore: make(map[string]struct{}),
		}
		return c.read(path)
	}
}

func (c *Config) read(path string) (*Config, error) {
	str, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(str), "\n")
	for _, l := range lines {
		words := strings.Fields(l)
		if len(words) < 1 {
			continue
		}
		switch words[0] {
		case "start":
			check(&words, "start")
			c.start, err = time.Parse(time.Kitchen, words[1])
			if err != nil {
				return nil, err
			}
		case "each":
			check(&words, "each")
			c.each, err = time.ParseDuration(words[1])
			if err != nil {
				return nil, err
			}
		case "duration":
			check(&words, "duration")
			c.duration, err = time.ParseDuration(words[1])
			if err != nil {
				return nil, err
			}
		case "data":
			check(&words, "data")
			if filepath.IsAbs(words[1]) {
				c.dataDir = words[1]
			} else {
				c.dataDir = filepath.Join(filepath.Dir(path), words[1])
			}
		case "static":
			check(&words, "static")
			if filepath.IsAbs(words[1]) {
				c.staticDir = words[1]
			} else {
				c.staticDir = filepath.Join(filepath.Dir(path), words[1])
			}
		case "skip":
			check(&words, "skip")
			fallthrough
		case "ignore":
			f := strings.Join(words[1:], " ")
			if strings.HasPrefix(f, "\"") && strings.HasSuffix(f, "\"") {
				f = strings.Trim(f, "\"")
			}
			if !filepath.IsAbs(f) {
				if c.dataDir != "" {
					f = filepath.Join(c.dataDir, f)
				} else {
					// `dataDirstatic` is not defined yet
					// the path is relative to config location
					f = filepath.Join(filepath.Dir(path), f)
				}
			}
			c.ignore[f] = struct{}{}
			log.Println("config: ignore", f)
		}
	}
	return c, nil
}

func (c *Config) String() string {
	str := "start " + c.start.Format(time.Kitchen) + "\n"
	str += fmt.Sprintln("each", c.each)
	str += fmt.Sprintln("duration", c.duration)
	str += fmt.Sprintln("data", c.dataDir)
	str += fmt.Sprintln("static", c.staticDir)
	return str
}

func (c *Config) Write() {
	os.WriteFile(c.path, []byte(c.String()), 0666)
}

func (c *Config) Update() error {
	_, err := c.read(c.path)
	return err
}

func (c *Config) ReadyToPlay() (bool, time.Duration, time.Duration) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(),
		c.start.Hour(), c.start.Minute(), c.start.Second(),
		c.start.Nanosecond(), now.Location())

	ok := start.Before(now) && start.Add(c.duration).After(now)
	if !ok {
		if now.Before(start) {
			return ok, start.Sub(now), c.duration
		} else {
			return ok, start.Add(c.each).Sub(now), c.duration
		}
	}
	return ok, time.Duration(0), start.Add(c.duration).Sub(now)
}

func (c *Config) Duration() time.Duration {
	return c.duration
}

func (c *Config) DataDir() string {
	return c.dataDir
}

func (c *Config) StaticDir() string {
	return c.staticDir
}

func (c *Config) Ignore(f string) bool {
	_, exist := c.ignore[f]
	return exist
}

func check(line *[]string, attr string) {
	if len(*line) < 2 {
		log.Fatal("invalid config file: '" + attr + "' lacks argument")
	}
}
