package main

import (
	"channel-helper-go/scripts"
	go_console "github.com/DrSmithFr/go-console"
	"github.com/DrSmithFr/go-console/input/argument"
	"github.com/DrSmithFr/go-console/input/option"
)

var configOption = go_console.Option{
	Name:     "config",
	Shortcut: "c",
	Value:    option.Required,
}

var configOptions = []go_console.Option{
	configOption,
}

func main() {
	script := go_console.Command{
		Description: "A Telegram channel helper",
		Scripts: []*go_console.Script{
			{
				Name:        "start",
				Description: "Start the bot",
				Options:     configOptions,
				Runner:      scripts.StartScript,
			},
			{
				Name:        "import",
				Description: "Import a previous project DB dump",
				Options:     configOptions,
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
			{
				Name:        "db:migrate",
				Description: "Migrate the database to the latest version",
				Options:     configOptions,
				Runner:      scripts.MigrateScript,
			},
			{
				Name:        "db:create-go-migration",
				Description: "Create a Go migration script",
				Options:     configOptions,
				Arguments: []go_console.Argument{
					{
						Name:        "name",
						Description: "Name of the migration script",
						Value:       argument.Required,
					},
				},
				Runner: scripts.CreateGoMigrationScript,
			},
			{
				Name:        "db:init",
				Description: "Initialize the database",
				Options:     configOptions,
				Runner:      scripts.InitDBScript,
			},
		},
	}
	script.Run()
}
