package handlers

import (
	"channel-helper-go/database"
	"channel-helper-go/utils"
	"channel-helper-go/utils/cfg"
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
			Type:       p.Type,
			FileID:     p.FileID,
			MessageIds: make([]int, 0, len(p.MessageIDs)),
			Processed:  p.IsSent,
			Datetime:   p.CreatedAt,
		}
		for _, msgID := range p.MessageIDs {
			item.MessageIds = append(item.MessageIds, msgID.MessageID)
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
	_, err = ctx.Bot().SendDocument(ctx, tu.Document(message.Chat.ChatID(),
		tu.FileFromBytes(j, fmt.Sprintf("%s.json", config.DbName)),
	))
	return err
}
