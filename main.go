package main

import (
	"bytes"
	"flag"
	"image/gif"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	tokenFile = flag.String("token", "token.txt", "file containing the bot token")
	guildID   = flag.String("guild-id", "", "Guild ID, or empty to register globally")
)

func main() {
	token, err := ioutil.ReadFile("token.txt")
	if err != nil {
		log.Fatal(err)
	}
	session, err := discordgo.New("Bot " + strings.TrimSpace(string(token)))
	if err != nil {
		log.Fatal(err)
	}
	if err := session.Open(); err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "bom",
			Description: "BOM Sydney 128km radar",
			Type:        discordgo.ChatApplicationCommand,
		},
	}
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.ApplicationCommandData().Name {
		case "bom":
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})

			resp := &discordgo.WebhookEdit{}
			defer func() {
				if _, err := s.InteractionResponseEdit(i.Interaction, resp); err != nil {
					log.Println(err)
				}
			}()

			g, err := getRadarGIF("IDR713")
			if err != nil {
				e := "error: " + err.Error()
				resp.Content = &e
				return
			}
			var buf bytes.Buffer
			if err := gif.EncodeAll(&buf, g); err != nil {
				e := "error: " + err.Error()
				resp.Content = &e
				return
			}
			resp.Files = []*discordgo.File{
				{
					Name:        "IDR713.gif",
					ContentType: "image/gif",
					Reader:      &buf,
				},
			}
		}
	})

	_, err = session.ApplicationCommandBulkOverwrite(session.State.User.ID, *guildID, commands)
	if err != nil {
		log.Fatal(err)
	}
	defer session.ApplicationCommandBulkOverwrite(session.State.User.ID, *guildID, nil)

	log.Println("Running.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down.")
}
