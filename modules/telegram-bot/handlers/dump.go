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
	logger := ctx.Value("logger").(utils.Logger)

	logger.Info("creating dump")

	qb := db.Post.Query().
		WithImageHash().
		WithMessageIds().
		Where(post.TypeEQ(post.TypePhoto))

	totalPosts, err := qb.Count(ctx)
	offset := 0
	dump := make([]utils.ImportItem, 0, totalPosts)
	for offset < totalPosts {
		logger.With("offset", offset).With("total", totalPosts).Info("fetching posts chunk")
		postsChunk, err := qb.
			Offset(offset).
			Limit(500).
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
		offset += len(postsChunk)
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
