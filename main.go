package main

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "download-manager"
	app.Version = "1.0.0"
	app.Compiled = time.Now()
	app.Commands = []cli.Command{}
	app.Commands = append(app.Commands, command)

	app.Run(os.Args)
}

var command = cli.Command{
	Name:        "fetch",
	Usage:       "download big files in chunks from remote server",
	Description: "this command will allow to download a big file in chunks and store it to a defined path, use `-u` to pass the url to fetch from, `-o` to specify the output path and `-f` to specify the filename ",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "u",
			Usage: "input url from where to fetch the file",
		},
		cli.StringFlag{
			Name:  "o",
			Usage: "output path where to save the file",
		},
		cli.StringFlag{
			Name:  "f",
			Usage: "the filename to be used",
		},
	},
	Action: FetchFile(),
}

func init() {
	cli.AppHelpTemplate =
		`{{ "\n" }}` +
			CyanBoldBrightColor(`Download Manager`) +
			`{{ "\n" }}` + `
	 {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}` +
			`{{ "\n" }}` +
			`{{ "\n" }}` +
			GreenBoldBrightColor(`commands:`) + `
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{if .VisibleFlags}}` +
			`{{ "\n" }}` +
			GreenBoldBrightColor(`version:`) + ` {{.Version}}
	 {{end}}{{ "\n" }}
`

	cli.CommandHelpTemplate =
		`{{ "\n" }}` +
			GreenBoldBrightColor(`name`) + `:
   {{.HelpName}} - {{.Usage}}` +
			`{{ "\n" }}` +
			`{{ "\n" }}` +
			GreenBoldBrightColor(`usage`) + `:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}
{{if .VisibleFlags}}` +
			`{{ "\n" }}` +
			GreenBoldBrightColor(`description`) + `:
    {{.Description}}{{ "\n" }}{{end}}
`

	cli.SubcommandHelpTemplate =
		`{{ "\n" }}` +
			GreenBoldBrightColor(`name`) + `:
 {{.HelpName}} - {{.Usage}}` +
			`{{ "\n" }}` +
			`{{ "\n" }}` +
			GreenBoldBrightColor(`usage`) + `:
 {{if .UsageText}}
 {{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}` +
			`{{ "\n" }}` +
			`{{ "\n" }}` +
			GreenBoldBrightColor(`commands`) + `:
{{range .VisibleCategories}}{{if .Name}}{{.Name}}:{{end}}{{range .VisibleCommands}} {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}` +
			`{{if .VisibleFlags}}` +
			GreenBoldBrightColor(`description`) + `:
	{{.Description}}{{ "\n" }}{{end}}
`

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(c.App.Writer, fmt.Sprintf("%s", c.App.Version))
	}

	cli.FlagStringer = func(fl cli.Flag) string {
		return fmt.Sprintf("\t\t%s", fl.GetName())
	}
}
