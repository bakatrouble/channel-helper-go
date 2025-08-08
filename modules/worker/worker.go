package worker

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	channels "channel-helper-go/modules"
	"channel-helper-go/utils"
	"context"
	"entgo.io/ent/dialect/sql"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"sync"
	"time"
)

func sendPost(ctx context.Context, postObj *ent.Post) error {
	bot := ctx.Value("bot").(*telego.Bot)
	config := ctx.Value("config").(*utils.Config)
	hub := ctx.Value("hub").(*channels.Hub)
	var err error

	println("Sending post:", post.ID)

	switch postObj.Type {
	case post.TypePhoto:
		_, err = bot.SendPhoto(ctx, tu.Photo(
			telego.ChatID{ID: config.TargetChatId},
			telego.InputFile{
				FileID: postObj.FileID,
			},
		))
	case post.TypeVideo:
		_, err = bot.SendVideo(ctx, tu.Video(
			telego.ChatID{ID: config.TargetChatId},
			telego.InputFile{
				FileID: postObj.FileID,
			},
		))
	case post.TypeAnimation:
		_, err = bot.SendAnimation(ctx, tu.Animation(
			telego.ChatID{ID: config.TargetChatId},
			telego.InputFile{
				FileID: postObj.FileID,
			},
		))
	}

	if err != nil {
		println("Error sending post:", err.Error())
		return err
	}

	err = postObj.Update().
		SetIsSent(true).
		SetSentAt(time.Now()).
		Exec(ctx)
	if err != nil {
		println("Error updating post as sent:", err.Error())
		return err
	}

	hub.PostSent <- postObj

	return nil
}

func StartWorker(ctx context.Context) {
	db := ctx.Value("db").(*ent.Client)
	config := ctx.Value("config").(*utils.Config)
	wg := ctx.Value("wg").(*sync.WaitGroup)

	defer wg.Done()

	ticker := time.NewTimer(config.Interval)
	for {
		select {
		case <-ticker.C:
			postObj, err := db.Post.Query().
				Where(post.IsSent(false)).
				Order(sql.OrderByRand()).
				First(ctx)
			if ent.IsNotFound(err) {
				continue
			} else if err != nil {
				println("Error fetching posts:", err.Error())
				continue
			}
			_ = sendPost(ctx, postObj)

		case <-ctx.Done():
			return
		}
	}
}
