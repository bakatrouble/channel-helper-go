package scripts

import (
	"channel-helper-go/database"
	"channel-helper-go/database/database_utils"
	"channel-helper-go/utils"
	"channel-helper-go/utils/cfg"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"time"

	"github.com/DrSmithFr/go-console"
	"github.com/alitto/pond/v2"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

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

	config, err := cfg.ParseConfig(cmd.Input.Option("config"))
	if err != nil {
		fmt.Printf("Failed to parse config file: %v\n", err)
		return go_console.ExitError
	}

	imageDirectory, _ := filepath.Abs(cmd.Input.Argument("directory"))

	logger := utils.NewLogger(config.DbName, "import")
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

	dump, err := os.ReadFile(cmd.Input.Argument("dump"))
	if err != nil {
		fmt.Printf("Failed to read dump file: %v\n", err)
		return go_console.ExitError
	}

	var items []utils.ImportItem
	if err = json.Unmarshal(dump, &items); err != nil {
		fmt.Printf("Failed to parse dump file: %v\n", err)
		return go_console.ExitError
	}

	existingFileIds, err := db.Post.GetFileIDs(ctx)
	if err != nil {
		fmt.Printf("Failed to query existing file IDs: %v\n", err)
		return go_console.ExitError
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

	var posts []*database.Post
	var hashTotal int64
	pool := pond.NewPool(runtime.NumCPU())
	now := time.Now().UTC()
	for _, item := range items {
		importBar.Increment()
		if existingFileIds[item.FileID] {
			continue
		}

		postId := database_utils.GenerateID()

		post := database.Post{
			ID:        postId,
			Type:      item.Type,
			FileID:    item.FileID,
			IsSent:    item.Processed,
			CreatedAt: item.Datetime,
		}
		if item.Processed {
			post.SentAt = &now
		}
		for _, messageId := range item.MessageIds {
			post.MessageIDs = append(post.MessageIDs, &database.MessageID{
				ChatID:    config.AllowedSenderChats[0],
				MessageID: messageId,
			})
		}
		posts = append(posts, &post)
		if item.Type == database.MediaTypePhoto {
			hashTotal += 1
			hashBar.SetTotal(hashTotal, false)
			closurePost := posts[len(posts)-1]
			pool.Submit(func() {
				ComputeHash(item.FileID, imageDirectory, func(hash *string) {
					hashBar.Increment()
					closurePost.ImageHash = &database.ImageHash{
						Hash: *hash,
					}
				})
			})
		}
	}
	pool.StopAndWait()
	hashBar.SetTotal(hashTotal, true)
	hashBar.Wait()

	existingImageHashesMap := make(map[string]bool, len(posts))
	filteredPosts := slices.Collect(func(yield func(post *database.Post) bool) {
		for _, post := range posts {
			if post.ImageHash != nil {
				if !existingImageHashesMap[post.ImageHash.Hash] {
					yield(post)
					existingImageHashesMap[post.ImageHash.Hash] = true
				} else {
					fmt.Printf("Duplicate image hash found: %s\n", post.ImageHash.Hash)
				}
			}
		}
	})

	_, _ = p.Write([]byte("Inserting into database..."))
	if err = db.Post.CreateBulk(ctx, filteredPosts, 1000); err != nil {
		fmt.Printf("Failed to insert posts into database: %v", err)
		return go_console.ExitError
	}

	return go_console.ExitSuccess
}
