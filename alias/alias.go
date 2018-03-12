package alias

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/zetamatta/nyagos/completion"
	"github.com/zetamatta/nyagos/shell"
	"github.com/zetamatta/nyagos/texts"
)

var dbg = false

type callableT interface {
	String() string
	Call(ctx context.Context, cmd *shell.Cmd) (int, error)
}

type AliasFunc struct {
	BaseStr string
}

func New(baseStr string) *AliasFunc {
	return &AliasFunc{baseStr}
}

func (this *AliasFunc) String() string {
	return this.BaseStr
}

func (this *AliasFunc) Call(ctx context.Context, cmd *shell.Cmd) (next int, err error) {
	isReplaced := false
	if dbg {
		print("AliasFunc.Call('", cmd.Arg(0), "')\n")
	}
	cmdline := paramMatch.ReplaceAllStringFunc(this.BaseStr, func(s string) string {
		if s == "$~*" {
			isReplaced = true
			if cmd.Args() != nil && len(cmd.Args()) >= 2 {
				return strings.Join(cmd.Args()[1:], " ")
			} else {
				return ""
			}
		} else if s == "$*" {
			isReplaced = true
			if cmd.Args != nil && len(cmd.Args()) >= 2 {
				return strings.Join(cmd.RawArgs()[1:], " ")
			} else {
				return ""
			}
		} else if len(s) >= 3 && s[0] == '$' && s[1] == '~' && strings.IndexByte("0123456789", s[2]) >= 0 {
			i, err := strconv.ParseInt(s[2:], 10, 0)
			if err == nil {
				isReplaced = true
				if 0 <= i && cmd.Args() != nil && int(i) < len(cmd.Args()) {
					return cmd.Arg(int(i))
				} else {
					return ""
				}
			}
		}
		i, err := strconv.ParseInt(s[1:], 10, 0)
		if err == nil {
			isReplaced = true
			if 0 <= i && cmd.Args != nil && int(i) < len(cmd.Args()) {
				return cmd.RawArg(int(i))
			} else {
				return ""
			}
		}
		return s
	})

	if !isReplaced {
		var buffer strings.Builder
		buffer.WriteString(this.BaseStr)
		for _, s := range cmd.RawArgs()[1:] {
			fmt.Fprintf(&buffer, " %s", s)
		}
		cmdline = buffer.String()
	}
	if dbg {
		print("replaced cmdline=='", cmdline, "'\n")
		print("cmd.Clone\n")
	}
	it, err := cmd.Clone()
	if err != nil {
		return 255, err
	}
	if dbg {
		print("done cmd.Clone\n")
	}

	arg1 := texts.FirstWord(cmdline)
	if strings.EqualFold(arg1, cmd.Arg(0)) {
		it.SetHookCount(100)
	} else {
		it.SetHookCount(cmd.HookCount() + 1)
	}
	if dbg {
		print("it.Interpret\n")
	}
	next, err = it.InterpretContext(ctx, cmdline)
	if dbg {
		print("done it.Interpret\n")
	}
	return
}

var Table = map[string]callableT{}
var paramMatch = regexp.MustCompile(`\$(\~)?(\*|[0-9]+)`)

func AllNames() []completion.Element {
	names := make([]completion.Element, 0, len(Table))
	for name1 := range Table {
		names = append(names, completion.Element{InsertStr: name1, ListupStr: name1})
	}
	return names
}

var nextHook shell.HookT

func hook(ctx context.Context, cmd *shell.Cmd) (int, bool, error) {
	if cmd.HookCount() > 5 {
		return nextHook(ctx, cmd)
	}
	callee, ok := Table[strings.ToLower(cmd.Arg(0))]
	if !ok {
		return nextHook(ctx, cmd)
	}
	next, err := callee.Call(ctx, cmd)
	return next, true, err
}

func Init() {
	nextHook = shell.SetHook(hook)
}
