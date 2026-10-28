package handlers

import (
	"bytes"
	"channel-helper-go/database"
	"channel-helper-go/utils"
	"channel-helper-go/utils/cfg"
	"compress/gzip"
	"encoding/json"
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func DumpDbHandler(ctx *th.Context, message telego.Message) error {
	db := ctx.Value("db").(*database.DBStruct)
	config := ctx.Value("config").(*cfg.Config)
	logger := ctx.Value("logger").(utils.Logger)

	if gtfo(ctx, message) {
		return nil
	}

	logger.Info("creating dump")

	posts, err := db.Post.GetAll(ctx)
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
			hash := p.ImageHash.Hash
			item.Hash = hash
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
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID: message.Chat.ChatID(),
			Text:   "Failed to marshal dump data",
		})
		return err
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(j); err != nil {
		_ = gz.Close()
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID: message.Chat.ChatID(),
			Text:   "Failed to compress dump data",
		})
		return err
	}
	if err := gz.Close(); err != nil {
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID: message.Chat.ChatID(),
			Text:   "Failed to compress dump data",
		})
		return err
	}
	_, err = ctx.Bot().SendDocument(ctx, tu.Document(message.Chat.ChatID(),
		tu.FileFromBytes(buf.Bytes(), fmt.Sprintf("%s.json.gz", config.DbName)),
	))
	return err
}
