package hls

import (
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vonaka/smc_station/config"
)

type Tank struct {
	cs []chunk
	ds []time.Duration
}

type chunk struct {
	filename    string
	duration    time.Duration
	videostream []int
	audiostream []int
}

func NewDataTank(c *config.Config) (*Tank, error) {
	t := &Tank{}
	err := t.fillChunks(c)
	return t, err
}

func (t *Tank) Update(c *config.Config) error {
	return t.fillChunks(c)
}

func (t *Tank) String() (s string) {
	if t == nil || t.cs == nil {
		return "<empty>"
	}
	for _, c := range t.cs {
		s += c.String() + "\n"
	}
	s += "---"
	return
}

func (t *Tank) Shuffle() {
	rand.Shuffle(len(t.cs), func(i, j int) {
		t.cs[i], t.cs[j] = t.cs[j], t.cs[i]
	})
	d := time.Duration(0)
	for _, c := range t.cs {
		d += c.duration
		t.ds = append(t.ds, d)
	}
}

var defaultTank *Tank

func InitializeDataTank(c *config.Config) (err error) {
	if defaultTank == nil {
		rand.Seed(time.Now().UnixNano())
		defaultTank, err = NewDataTank(c)
		return err
	}
	return nil
}

func UpdateTank(c *config.Config) error {
	if defaultTank != nil {
		return defaultTank.Update(c)
	}
	return nil
}

func ShuffleTank() {
	if defaultTank != nil {
		defaultTank.Shuffle()
	}
}

func ReadTank() string {
	return defaultTank.String()
}

func (t *Tank) fillChunks(c *config.Config) error {
	var readDir func(string, []int, []int, map[string]struct{}) error
	readDir = func(dir string, videos, audios []int, skip map[string]struct{}) error {
		vs, err := ioutil.ReadDir(dir)
		if err != nil {
			return err
		}

		for _, v := range vs {
			if e := filepath.Ext(v.Name()); e == ".conf" || e == ".config" {
				vsp, asp, skipp := readConfig(filepath.Join(dir, v.Name()))
				if vsp != nil {
					videos = vsp
				}
				if asp != nil {
					audios = asp
				}
				if skipp != nil && len(skipp) > 0 {
					skip = skipp
				}
			}
		}

		for _, v := range vs {
			_, toSkip := skip[v.Name()]
			toSkip = toSkip || c.Ignore(filepath.Join(dir, v.Name()))
			if v.IsDir() {
				if toSkip {
					continue
				}
				err = readDir(filepath.Join(dir, v.Name()), videos, audios, skip)
				if err != nil {
					return err
				}
			} else {
				if !toSkip && check(v.Name()) {
					filename := filepath.Join(dir, v.Name())
					duration, err := videoDuration(filename)
					// skip if can't get duration
					// TODO: log skipped files
					if err == nil {
						t.cs = append(t.cs, chunk{
							filename:    filename,
							duration:    duration,
							videostream: copySlice(videos),
							audiostream: copySlice(audios),
						})
					}
				}
			}
		}
		return nil
	}
	return readDir(c.DataDir(), []int{0}, []int{0}, make(map[string]struct{}))
}

func readConfig(filename string) ([]int, []int, map[string]struct{}) {
	str, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, nil
	}
	var (
		vs []int
		as []int
	)
	stringsToInts := func(s []string) []int {
		i := make([]int, len(s))
		c := 0
		for j := range s {
			k, err := strconv.Atoi(s[j])
			if err == nil {
				i[j] = k
				c++
			}
		}
		if c == 0 {
			return nil
		}
		return i[0:c]
	}
	skip := make(map[string]struct{})
	lines := strings.Split(string(str), "\n")
	for _, l := range lines {
		words := strings.Fields(l)
		if len(words) < 2 {
			continue
		}
		switch words[0] {
		case "video":
			vs = stringsToInts(words[1:])
		case "audio":
			as = stringsToInts(words[1:])
		case "skip":
			fallthrough
		case "ignore":
			// FIXME: what if filenames contain '"'?
			for _, w := range words[1:] {
				skip[w] = struct{}{}
			}
		}
	}
	return vs, as, skip
}

func copySlice(s []int) []int {
	n := make([]int, len(s))
	copy(n, s)
	return n
}

func check(filename string) bool {
	ok := struct{}{}
	valid := map[string]struct{}{
		".avi":  ok,
		".m4v":  ok,
		".mkv":  ok,
		".mlv":  ok,
		".mov":  ok,
		".mp4":  ok,
		".mpeg": ok,
	}
	_, isOk := valid[filepath.Ext(filename)]
	return isOk
}

func videoDuration(filename string) (time.Duration, error) {
	len, err := exec.Command("ffprobe", "-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filename).Output()
	if err != nil {
		return time.Duration(0), err
	}
	lenStr := string(strings.Split(string(len), "\n")[0]) + "s"
	return time.ParseDuration(lenStr)
}

func (c chunk) String() string {
	return c.filename
}
