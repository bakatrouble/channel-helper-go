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

	qb := db.Post.Query().
		WithImageHash().
		WithMessageIds().
		Where(post.TypeEQ(post.TypePhoto))

	totalPosts, err := qb.Count(ctx)
	i := 0
	dump := make([]utils.ImportItem, 0, totalPosts)
	for i < totalPosts {
		postsChunk, err := db.Post.Query().
			WithImageHash().
			Where(post.TypeEQ(post.TypePhoto)).
			Offset(i).
			Limit(1000).
			All(ctx)
		if err != nil {
			return err
		}
		for _, p := range postsChunk {
			item := utils.ImportItem{
				Type:       p.Type,
				FileId:     p.FileID,
				MessageIds: make([]int, 0, len(p.Edges.MessageIds)),
				Processed:  p.IsSent,
				Datetime:   p.CreatedAt,
			}
			for _, msgID := range p.Edges.MessageIds {
				item.MessageIds = append(item.MessageIds, msgID.MessageID)
			}
			dump = append(dump, item)
		}
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
