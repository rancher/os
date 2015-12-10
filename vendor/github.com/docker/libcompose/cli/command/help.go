package command

import (
	"github.com/codegangsta/cli"
	"os"
	"path"
)

func init() {
	cli.AppHelpTemplate = `Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]

{{.Usage}}

Version: {{.Version}}{{if or .Author .Email}}

Author:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}
{{if .Flags}}
Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`
	cli.CommandHelpTemplate = `Usage: ` + path.Base(os.Args[0]) + ` {{.Name}}{{if .Flags}} [OPTIONS]

{{.Usage}}

Options:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

}
