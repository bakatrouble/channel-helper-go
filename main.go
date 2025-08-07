package main

import (
	"channel-helper-go/scripts"
	"github.com/DrSmithFr/go-console"
	"github.com/DrSmithFr/go-console/input/argument"
	"github.com/DrSmithFr/go-console/input/option"
)

func main() {
	script := go_console.Command{
		Description: "A Telegram channel helper",
		Scripts: []*go_console.Script{
			{
				Name:        "start",
				Description: "Start the bot",
				Options: []go_console.Option{
					{
						Name:        "config",
						Description: "Configuration YAML file",
						Shortcut:    "c",
						Value:       option.Required,
					},
				},
				Runner: scripts.StartScript,
			},
			{
				Name:        "import",
				Description: "Import a previous project DB dump",
				Options: []go_console.Option{
					{
						Name:     "config",
						Shortcut: "c",
						Value:    option.Required,
					},
				},
				Arguments: []go_console.Argument{
					{
						Name:        "dump",
						Description: "Path to the DB dump file",
						Value:       argument.Required,
					},
					{
						Name:        "directory",
						Description: "Image directory",
						Value:       argument.Required,
					},
				},
				Runner: scripts.ImportScript,
			},
		},
	}
	script.Run()
}
