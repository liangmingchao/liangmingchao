package main

import (
	"context"
	"os"
	"strconv"
	"strings"

	steambox "github.com/journey-ad/steam-box/pkg"
	"github.com/google/go-github/github"
)

func main() {
	steamAPIKey := os.Getenv("STEAM_API_KEY")
	steamID, _ := strconv.ParseUint(os.Getenv("STEAM_ID"), 10, 64)

	ghToken := os.Getenv("GH_TOKEN")
	ghUsername := os.Getenv("GH_USER")
	gistID := os.Getenv("GIST_ID")

	box := steambox.NewBox(steamAPIKey, ghUsername, ghToken)

	ctx := context.Background()

	lines, err := box.GetPlayTime(ctx, steamID)
	if err != nil {
		panic("GetPlayTime err:" + err.Error())
	}

	filename := "🎮 最近我在玩…"
	gist, err := box.GetGist(ctx, gistID)
	if err != nil {
		panic("GetGist err:" + err.Error())
	}

	f := gist.Files[github.GistFilename(filename)]

	f.Content = github.String(strings.Join(lines, "\n"))
	gist.Files[github.GistFilename(filename)] = f

	err = box.UpdateGist(ctx, gistID, gist)
	if err != nil {
		panic("UpdateGist err:" + err.Error())
	}
}
