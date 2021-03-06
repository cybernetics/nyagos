package commands

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/zetamatta/go-texts/mbcs"
)

func cat(ctx context.Context, _r io.Reader, w io.Writer) bool {
	r := mbcs.NewAutoDetectReader(_r, mbcs.ConsoleCP())
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if done := ctx.Done(); done != nil {
			select {
			case <-done:
				return false
			default:

			}
		}
		text := scanner.Text()
		text = strings.Replace(text, "\xEF\xBB\xBF", "", 1)
		fmt.Fprintln(w, text)
	}
	return true
}

func cmdType(ctx context.Context, cmd Param) (int, error) {
	if len(cmd.Args()) <= 1 {
		cat(ctx, cmd.In(), cmd.Out())
	} else {
		for _, arg1 := range cmd.Args()[1:] {
			r, err := os.Open(arg1)
			if err != nil {
				return 1, err
			}
			stat1, err := r.Stat()
			if err != nil {
				r.Close()
				return 2, err
			}
			if stat1.IsDir() {
				r.Close()
				return 3, fmt.Errorf("%s: Permission denied", arg1)
			}
			cont := cat(ctx, r, cmd.Out())
			r.Close()
			if !cont {
				return 0, nil
			}
		}
	}
	return 0, nil
}
