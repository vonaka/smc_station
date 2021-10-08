package hls

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vonaka/smc_station/config"
)

type Program struct {
	tank  *Tank
	start int
	end   int
}

var (
	ErrTankIndex error = errors.New("tank size exceeded")
)

func MakeProgram(c *config.Config) (*Program, error) {
	p := &Program{
		tank:  defaultTank,
		start: 0,
	}
	ShuffleTank()
	avg := p.tank.ds[len(p.tank.ds)-1].Minutes() / float64(len(p.tank.ds))
	p.end = int(c.Duration().Minutes() / avg)
	if p.end == 0 {
		p.end = 1
	}
	if l := len(p.tank.cs); p.end > l {
		p.end = l
	} else if d := p.tank.ds[p.end-1]; d < c.Duration() {
		for p.end <= l && d < c.Duration() {
			p.end++
			d = p.tank.ds[p.end-1]
		}
	}
	return p, nil
}

func (p *Program) Next() error {
	p.start = p.end
	p.end *= 2
	if l := len(p.tank.cs); p.end >= l {
		return ErrTankIndex
	}
	return nil
}

func CopyWithTime(src, dst string, sec int) error {
	os.RemoveAll(dst)
	srcStr, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	dstStr := ""
	srcLines := strings.Split(string(srcStr), "\n")
	// TODO: optimize
	for _, l := range srcLines {
		exts := strings.Split(l, ":")
		if lgth := len(exts); lgth <= 0 {
			continue
		} else if len(exts) < 2 {
			dstStr += exts[0] + "\n"
			continue
		}
		switch exts[0] {
		case "#EXT-X-VERSION":
			dstStr += exts[0] + ":6\n"
			dstStr += "#EXT-X-START:TIME-OFFSET=" + strconv.Itoa(sec) + ",PRECISE=YES\n"
		default:
			dstStr += exts[0]
			for _, ext := range exts[1:] {
				dstStr += ":" + ext
			}
			dstStr += "\n"
		}
	}
	return os.WriteFile(dst, []byte(dstStr), 0664)
}

func (p *Program) Write(filename string) error {
	cs := p.tank.cs[p.start:p.end]
	alen := -1
	str := ""
	f := strings.TrimSuffix(filename, filepath.Ext(filename))
	for _, c := range cs {
		if alen > len(c.audiostream) || alen == -1 {
			alen = len(c.audiostream)
		}
	}

	for i, c := range cs {
		ffmpeg := "ffmpeg -hide_banner -loglevel error -i "
		ffmpeg += "\"" + c.filename + "\""
		if c.vcodec == "h264" {
			ffmpeg += " -vcodec copy"
		} else {
			ffmpeg += " -crf 17"
			ffmpeg += " -vcodec h264"
		}
		// FIXME: for now assumed only one video stream
		ffmpeg += fmt.Sprintf(" -map 0:v")
		ffmpeg += " -acodec aac"
		// TODO: use all audiostreams ?
		for j := 0; j < alen; j++ {
			ffmpeg += fmt.Sprintf(" -map 0:a:%v", c.audiostream[j])
		}
		ffmpeg += " -metadata service_name='program'"
		ffmpeg += " -pix_fmt yuv420p"
		ffmpeg += " -f hls"
		ffmpeg += " -hls_list_size 0"
		ffmpeg += " -hls_segment_type mpegts"
		part := fmt.Sprintf("%v_%v_part.m3u8", f, i)
		ffmpeg += " " + part
		log.Printf("hls: %v\n%v\n", c.filename, ffmpeg)
		_, err := exec.Command("sh", "-c", ffmpeg).Output()
		if err != nil {
			return err
		}

		partStr, err := os.ReadFile(part)
		if err != nil {
			return err
		}
		lines := strings.Split(string(partStr), "\n")
		for _, l := range lines {
			if l == "" {
				continue
			} else if l == "#EXT-X-ENDLIST" {
				if i != len(cs)-1 {
					str += "#EXT-X-DISCONTINUITY\n"
				} else {
					str += l + "\n"
				}
				continue
			}
			if i != 0 && strings.HasPrefix(l, "#EXT") && !strings.HasPrefix(l, "#EXTINF") {
				continue
			}
			str += l + "\n"
		}
	}

	return os.WriteFile(filename, []byte(str), 0666)
}
