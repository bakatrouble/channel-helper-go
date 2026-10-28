package scripts

import (
	"bytes"
	"channel-helper-go/database"
	"channel-helper-go/utils"
	"channel-helper-go/utils/cfg"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	go_console "github.com/DrSmithFr/go-console"
)

// DumpScript writes the project database to a dump file in the same
// Go-schema JSON format the Telegram `/dump` handler sends and the
// `import` command reads. Photo hashes ride along when present; a `.gz`
// output suffix writes the gzipped form instead of plain JSON.
func DumpScript(cmd *go_console.Script) go_console.ExitCode {
	ctx := context.Background()

	config, err := cfg.ParseConfig(cmd.Input.Option("config"))
	if err != nil {
		fmt.Printf("Failed to parse config file: %v\n", err)
		return go_console.ExitError
	}

	output := cmd.Input.Argument("output")

	logger := utils.NewLogger(config.DbName, "dump")
	sqldb, err := database.NewSQLDB(config.DbName)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return go_console.ExitError
	}
	db, err := database.NewDBStruct(sqldb, true, logger)
	if err != nil {
		fmt.Printf("Failed to create database struct: %v\n", err)
		return go_console.ExitError
	}
	defer func(db *database.DBStruct) {
		_ = db.Close()
	}(db)

	posts, err := db.Post.GetAll(ctx)
	if err != nil {
		fmt.Printf("Failed to fetch posts: %v\n", err)
		return go_console.ExitError
	}

	dump := make([]utils.ImportItem, 0, len(posts))
	for _, p := range posts {
		item := utils.ImportItem{
			Type:      p.Type,
			FileID:    p.FileID,
			Messages:  make([]utils.ImportMessage, 0, len(p.MessageIDs)),
			Processed: p.IsSent,
			Datetime:  p.CreatedAt,
		}
		if p.ImageHash != nil {
			item.Hash = p.ImageHash.Hash
		}
		for _, msgID := range p.MessageIDs {
			item.Messages = append(item.Messages, utils.ImportMessage{
				ChatID:    msgID.ChatID,
				MessageID: msgID.MessageID,
			})
		}
		dump = append(dump, item)
	}

	j, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal dump data: %v\n", err)
		return go_console.ExitError
	}

	data := j
	if strings.HasSuffix(output, ".gz") {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(j); err != nil {
			_ = gz.Close()
			fmt.Printf("Failed to compress dump data: %v\n", err)
			return go_console.ExitError
		}
		if err := gz.Close(); err != nil {
			fmt.Printf("Failed to compress dump data: %v\n", err)
			return go_console.ExitError
		}
		data = buf.Bytes()
	}

	if err := os.WriteFile(output, data, 0644); err != nil {
		fmt.Printf("Failed to write dump file: %v\n", err)
		return go_console.ExitError
	}

	fmt.Printf("Dumped %d posts to %s\n", len(dump), output)
	return go_console.ExitSuccess
}
