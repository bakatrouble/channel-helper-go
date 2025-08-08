package scripts

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"channel-helper-go/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/DrSmithFr/go-console"
	"github.com/alitto/pond/v2"
	"github.com/moroz/uuidv7-go"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"time"
)

type ImportItem struct {
	Type       post.Type `json:"type"`
	FileId     string    `json:"file_id"`
	MessageIds []int     `json:"message_ids"`
	Processed  bool      `json:"processed"`
	Datetime   time.Time `json:"datetime"`
}

func ComputeHash(fileId string, directory string, callback func(*string)) {
	bytes, err := os.ReadFile(path.Join(directory, fmt.Sprintf("%s.jpg", fileId)))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to read image file for post %s: %v\n", fileId, err)
		callback(nil)
		return
	}
	hash, err := utils.HashImage(bytes)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to hash image for post %s: %v\n", fileId, err)
		callback(nil)
		return
	}
	callback(&hash)
}

func ImportScript(cmd *go_console.Script) go_console.ExitCode {
	ctx := context.Background()

	config, err := utils.ParseConfig(cmd.Input.Option("config"))
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to parse config file: %v", err)
		return go_console.ExitError
	}

	imageDirectory, _ := filepath.Abs(cmd.Input.Argument("directory"))

	db, err := ent.ConnectToDB(config.DbName, ctx)
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to connect to database: %v", err)
		return go_console.ExitError
	}

	dump, err := os.ReadFile(cmd.Input.Argument("dump"))
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to read dump file: %v", err)
		return go_console.ExitError
	}

	var items []ImportItem
	err = json.Unmarshal(dump, &items)
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to parse dump file: %v", err)
		return go_console.ExitError
	}

	fileIds, err := db.Post.Query().
		Select(post.FieldFileID).
		Strings(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to query existing file IDs: %v", err)
		return go_console.ExitError
	}
	existingFileIds := make(map[string]bool, len(fileIds))
	for _, fileId := range fileIds {
		existingFileIds[fileId] = true
	}

	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond),
	)
	importBar := p.AddBar(
		int64(len(items)),
		mpb.PrependDecorators(
			decor.Name("Importing ", decor.WCSyncSpaceR),
			decor.CountersNoUnit("%d / %d ", decor.WCSyncSpaceR),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
		),
	)
	hashBar := p.AddBar(
		0,
		mpb.PrependDecorators(
			decor.Name("Hashing ", decor.WCSyncSpaceR),
			decor.CountersNoUnit("%d / %d ", decor.WCSyncSpaceR),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
		),
	)

	var posts []*ent.PostCreate
	var postMessageIds []*ent.PostMessageIdCreate
	var hashTotal int64
	pool := pond.NewPool(runtime.NumCPU())
	for _, item := range items {
		importBar.Increment()
		if existingFileIds[item.FileId] {
			continue
		}

		postId := uuidv7.Generate()

		createdPost := db.Post.Create().
			SetID(postId).
			SetType(item.Type).
			SetFileID(item.FileId).
			SetIsSent(item.Processed).
			SetCreatedAt(item.Datetime)
		if item.Processed {
			createdPost = createdPost.SetSentAt(time.Now().UTC())
		}
		posts = append(posts, createdPost)
		if item.Type == post.TypePhoto {
			hashTotal += 1
			hashBar.SetTotal(hashTotal, false)
			i := len(posts) - 1
			pool.Submit(func() {
				ComputeHash(item.FileId, imageDirectory, func(hash *string) {
					hashBar.Increment()
					posts[i] = posts[i].SetNillableImageHash(hash)
				})
			})
		}
		for _, messageId := range item.MessageIds {
			postMessageIds = append(
				postMessageIds,
				db.PostMessageId.Create().
					SetChatID(config.AllowedSenderChats[0]).
					SetMessageID(messageId).
					SetPostID(postId),
			)
		}
	}
	pool.StopAndWait()
	hashBar.SetTotal(hashTotal, true)
	hashBar.Wait()

	_, _ = fmt.Fprintf(cmd, "Inserting into database...\n")
	for chunk := range slices.Chunk(posts, 1000) {
		err = db.Post.CreateBulk(chunk...).Exec(ctx)
		if err != nil {
			_, _ = fmt.Fprintf(cmd, "Failed to insert posts: %v\n", err)
			return go_console.ExitError
		}
	}
	for chunk := range slices.Chunk(postMessageIds, 1000) {
		err = db.PostMessageId.CreateBulk(chunk...).Exec(ctx)
		if err != nil {
			_, _ = fmt.Fprintf(cmd, "Failed to insert post message IDs: %v\n", err)
			return go_console.ExitError
		}
	}

	return go_console.ExitSuccess
}
