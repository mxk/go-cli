// Package bash generates CLI auto-complete scripts.
package bash

import (
	"bytes"
	"flag"
	"sort"
	"strings"
	"text/template"

	"github.com/mxk/go-cli"
)

const tpl = `_{{.Bin}}() {
{{- range $id, $cmd := .Cmds}}
	local _cmd{{$id}}=({{$cmd.Spec}})
	{{- range $cmd.Refs}} \
	      _ref{{.}}={{$cmd.Name}}
	{{- end}}
	{{- range $arg, $spec := $cmd.Args}} \
	      _arg{{$id}}_{{$arg}}=({{$spec}})
	{{- end}}
{{- end}}

	# Find current command
	local comp=_cmd_ cur help
	for (( i=1; i<COMP_CWORD; i++ )); do
		cur="${COMP_WORDS[i]}"
		case "$cur" in (help|-help|--help|-h|/?)
			help=1
			continue;;
		esac
		[[ "$cur" =~ ^[a-z_-]+$ ]] || return 0
		comp=${comp}${cur//-/_}

		# Alias check
		cur=_ref${comp#_cmd}
		[[ ${!cur+ref} ]] && comp=${comp%_*}_${!cur}

		# Final command check (no '_' suffix)
		[[ ${!comp+last} ]] && break
		comp=${comp}_
		[[ ${!comp+more} ]] || return 0
	done

	# If final command (no '_' suffix), complete current argument
	cur="${COMP_WORDS[COMP_CWORD]}"
	case $comp in
	*_) ;;
	*)
		[[ $help ]] && return
		local prev="${COMP_WORDS[COMP_CWORD-1]}" strip
		if [[ ! "$prev" =~ ^-([a-z_-]+)$ && "$cur" =~ ^-([a-z_-]+)= ]]; then
			strip=1
		fi
		local arg=_arg${comp#_cmd}_${BASH_REMATCH[1]//-/_}
		if [[ ${!arg+special} ]]; then
			comp=$arg
			[[ $strip ]] && cur="${cur#-${BASH_REMATCH[1]}=}"
		fi
		;;
	esac

	comp=$comp[@]
	COMPREPLY=($(compgen "${!comp}" -- "$cur"))
}

complete -F _{{.Bin}} {{.Bin}}
`

// Compgen returns a bash auto-complete script for command hierarchy rooted at
// c. It assumes that c.Name is "", as is the case for cli.Main.
func Compgen(c *cli.Cfg) ([]byte, error) {
	cmds := make(map[string]*cmdSpec)
	newCmdSpec(cmds, "", c)
	var b bytes.Buffer
	t, err := template.New("").Parse(tpl)
	if err == nil {
		err = t.Execute(&b, struct {
			Bin  string
			Cmds map[string]*cmdSpec
		}{cli.Bin, cmds})
	}
	return b.Bytes(), err
}

// boolFlag is copied from flag package to identify bool-style flags.
type boolFlag interface {
	flag.Value
	IsBoolFlag() bool
}

// cmdSpec contains bash completion data for one command.
type cmdSpec struct {
	Name string
	Spec string
	Refs []string
	Args map[string]string
}

// newCmdSpec adds a cmdSpec entry for c to m.
func newCmdSpec(m map[string]*cmdSpec, root string, c *cli.Cfg) {
	names := strings.Split(c.Name, "|")
	cs := &cmdSpec{Name: safeName(names[0])}
	if len(names) > 1 {
		cs.Refs = names[1:]
		for i := range cs.Refs {
			cs.Refs[i] = root + safeName(cs.Refs[i])
		}
	}
	var spec strings.Builder
	spec.WriteString("-W '")
	if cmds := c.Children(); len(cmds) > 0 {
		root += cs.Name + "_"
		names = append(make([]string, 0, 1+len(cmds)), "help")
		for _, c := range cmds {
			if !c.Hide {
				newCmdSpec(m, root, c)
				names = append(names, cli.Name(c))
			}
		}
		sort.Strings(names)
		spec.WriteString(names[0])
		for _, name := range names[1:] {
			spec.WriteByte(' ')
			spec.WriteString(name)
		}
	} else {
		root += cs.Name
		cli.NewFlagSet(cli.New(c)).VisitAll(func(f *flag.Flag) {
			if cs.Args == nil {
				cs.Args = make(map[string]string)
			} else {
				spec.WriteByte(' ')
			}
			spec.WriteByte('-')
			spec.WriteString(f.Name)
			if b, ok := f.Value.(boolFlag); ok && b.IsBoolFlag() {
				return
			}
			var argSpec string
			switch arg, _ := flag.UnquoteUsage(f); arg {
			case "file":
				argSpec = "-f"
			case "dir":
				argSpec = "-d"
			default:
				argSpec = "-W ''"
			}
			cs.Args[safeName(f.Name)] = argSpec
		})
	}
	spec.WriteByte('\'')
	if c.MaxArgs > 0 || c.MaxArgs < c.MinArgs {
		spec.WriteString(" -o bashdefault")
	}
	cs.Spec = spec.String()
	m[root] = cs
}

// safeName replaces all characters outside of [0-9A-Za-z_] class in s with '_'.
func safeName(s string) string {
	var b []byte
	for i := range s {
		switch c := s[i]; {
		case '0' <= c && c <= '9':
		case 'A' <= c && c <= 'Z':
		case 'a' <= c && c <= 'z':
		case c == '_':
		default:
			if b == nil {
				b = make([]byte, len(s))
				copy(b, s)
			}
			b[i] = '_'
		}
	}
	if b == nil {
		return s
	}
	return string(b)
}
