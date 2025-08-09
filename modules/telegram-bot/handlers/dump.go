package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"channel-helper-go/utils"
	"encoding/json"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func DumpDbHandler(ctx *th.Context, message telego.Message) error {
	db := ctx.Value("db").(*ent.Client)
	config := ctx.Value("config").(*utils.Config)

	posts, err := db.Post.Query().
		WithImageHash().
		WithMessageIds().
		Where(post.TypeEQ(post.TypePhoto)).
		All(ctx)
	if err != nil {
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID: message.Chat.ChatID(),
			Text:   "Failed to retrieve posts from the database",
		})
		return err
	}
	dump := make([]utils.ImportItem, 0, len(posts))
	for _, post := range posts {
		item := utils.ImportItem{
			Type:       post.Type,
			FileId:     post.FileID,
			MessageIds: make([]int, 0, len(post.Edges.MessageIds)),
			Processed:  post.IsSent,
			Datetime:   post.CreatedAt,
		}
		for _, msgID := range post.Edges.MessageIds {
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
