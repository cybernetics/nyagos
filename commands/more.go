package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"

	"github.com/mattn/go-isatty"
	"github.com/mattn/go-runewidth"

	"github.com/zetamatta/go-box"
	"github.com/zetamatta/go-getch"
	"github.com/zetamatta/go-texts/mbcs"
)

var ansiStrip = regexp.MustCompile("\x1B[^a-zA-Z]*[A-Za-z]")

var bold = false
var screenWidth int
var screenHeight int

func more(_r io.Reader, cmd Param) bool {
	r := mbcs.NewAutoDetectReader(_r, mbcs.ConsoleCP())
	scanner := bufio.NewScanner(r)
	count := 0

	if f, ok := cmd.Out().(*os.File); !ok || !isatty.IsTerminal(f.Fd()) {
		screenHeight = math.MaxInt32
	}

	for scanner.Scan() {
		text := scanner.Text()
		width := runewidth.StringWidth(ansiStrip.ReplaceAllString(text, ""))
		lines := (width + screenWidth) / screenWidth
		for count+lines >= screenHeight {
			io.WriteString(cmd.Err(), "more>")
			ch := getch.Rune()
			io.WriteString(cmd.Err(), "\r     \b\b\b\b\b")
			if ch == 'q' {
				return false
			} else if ch == '\r' {
				count--
			} else {
				count = 0
			}
		}
		if bold {
			io.WriteString(cmd.Out(), "\x1B[1m")
		}
		fmt.Fprintln(cmd.Out(), text)
		count += lines
	}
	return true
}

func cmdMore(ctx context.Context, cmd Param) (int, error) {
	count := 0
	screenWidth, screenHeight = box.GetScreenBufferInfo().ViewSize()
	for _, arg1 := range cmd.Args()[1:] {
		if arg1 == "-b" {
			bold = true
			continue
		} else if arg1 == "-h" {
			return 1, errors.New("more : Color-Unicoded more")
		}
		r, err := os.Open(arg1)
		if err != nil {
			return 1, err
		}
		if !more(r, cmd) {
			r.Close()
			return 0, nil
		}
		r.Close()
		count++
	}
	if count <= 0 {
		more(cmd.In(), cmd)
	}
	return 0, nil
}
