package steambox

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"unicode"

	//"unicode/utf8"

	"github.com/google/go-github/github"
	steam "github.com/journey-ad/steam-go/pkg"
)

// Box defines the steam box.
type Box struct {
	steam  *steam.Client
	github *github.Client
}

// NewBox creates a new Box with the given API key.
func NewBox(apikey string, ghUsername, ghToken string) *Box {
	box := &Box{}
	box.steam = steam.NewClient(apikey, nil)
	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(ghUsername),
		Password: strings.TrimSpace(ghToken),
	}

	box.github = github.NewClient(tp.Client())

	return box

}

// GetGist gets the gist from github.com.
func (b *Box) GetGist(ctx context.Context, id string) (*github.Gist, error) {
	gist, _, err := b.github.Gists.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return gist, nil
}

// UpdateGist updates the gist.
func (b *Box) UpdateGist(ctx context.Context, id string, gist *github.Gist) error {
	_, _, err := b.github.Gists.Edit(ctx, id, gist)
	return err
}

// GetPlayTime gets the paytime form steam web API.
func (b *Box) GetPlayTime(ctx context.Context, steamID uint64) ([]string, error) {
	params := &steam.GetRecentlyPlayedGamesParams{
		SteamID: steamID,
		Count:   MaximumLines,
	}

	gameRet, err := b.steam.IPlayerService.GetRecentlyPlayedGames(ctx, params)
	if err != nil {
		return nil, err
	}
	var lines []string

	if gameRet.TotalCount == 0 {
		lines = append(lines, "🚫 啊哦，最近好像没有玩过游戏呢")
		return lines, nil
	}

	sort.Slice(gameRet.Games, func(i, j int) bool {
		return gameRet.Games[i].PlaytimeForever > gameRet.Games[j].PlaytimeForever
	})

	games := gameRet.Games[:MaximumLines]
	for _, game := range games {
		line := makeGameLine(game)
		lines = append(lines, line)
	}
	return lines, nil
}

func (b *Box) UpdateMarkdown(ctx context.Context, title, filename string, content []byte) error {
	md, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("steambox.UpdateMarkdown: Error reade a file: %w", err)
	}

	start := []byte("<!-- steam-box start -->")
	before := md[:bytes.Index(md, start)+len(start)]
	end := []byte("<!-- steam-box end -->")
	after := md[bytes.Index(md, end):]

	newMd := bytes.NewBuffer(nil)
	newMd.Write(before)
	newMd.WriteString("\n" + title + "\n")
	newMd.WriteString("```text\n")
	newMd.Write(content)
	newMd.WriteString("\n")
	newMd.WriteString("```\n")
	newMd.WriteString("<!-- Powered by https://github.com/YouEclipse/steam-box . -->\n")
	newMd.Write(after)

	err = ioutil.WriteFile(filename, newMd.Bytes(), os.ModeAppend)
	if err != nil {
		return fmt.Errorf("steambox.UpdateMarkdown: Error write a file: %w", err)
	}

	return nil
}

func getPlayTime(game *steam.Game) string {
	return fmt.Sprintf("🕘 %d hrs %d mins", game.Playtime2Weeks, game.Playtime2Weeks%60)
}

func makeGameLine(game *steam.Game) string {
	before := pad(getGameEmoji(game.Appid)+game.Name, " ", GameLinePrefixWidth)
	after := pad(getPlayTime(game), " ", GameLineSuffixWidth)
	return before + after
}

func getCharSpaceWidth(r rune) int {
	if unicode.Is(unicode.Han, r) {
		return 2
	}
	if unicode.IsSymbol(r) {
		return 2
	}
	return 1
}

func getSentenceSpaceWidth(s string) int {
	var width int
	for _, r := range s {
		width += getCharSpaceWidth(r)
	}
	return width
}

func pad(s, pad string, targetWidth int) string {
	sw := getSentenceSpaceWidth(s)
	padding := targetWidth - sw
	if padding <= 0 {
		return s
	}

	return s + strings.Repeat(pad, padding)
}

func getGameEmoji(id int) string {
	// hard code some game's emoji
	var nameEmojiMap = map[int]string{
		730:    "🔫 ", // CS:GO
		222880: "🔫 ", // Insurgency
		265630: "🔫 ", // Fistful of Frags
		271590: "🚓 ", // GTA 5
		578080: "🍳 ", // PUBG
		431960: "💻 ", // Wallpaper Engine
		8930:   "🌏 ", // Sid Meier's Civilization V
		644560: "🔞 ", // Mirror
		359550: "🔫 ", // Tom Clancy's Rainbow Six Siege
		105600: "👾 ", // Terraria
	}

	if emoji, ok := nameEmojiMap[id]; ok {
		return emoji
	}

	var randomEmojiMap = map[int]string{
		0: "🤡 ",
		1: "👽 ",
		2: "🤖 ",
		3: "💻 ",
		4: "🐉 ",
		5: "🚀 ",
		6: "🦚 ",
		7: "🐝 ",
		8: "🐦 ",
		9: "🐞 ",
	}

	return randomEmojiMap[(id)%len(randomEmojiMap)]
}
