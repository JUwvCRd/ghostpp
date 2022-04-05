package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const discordToken = ""

const GHOST_STATUS_NOT_RUNNING = 0
const GHOST_STATUS_STARTING = 1
const GHOST_STATUS_RUNNING = 2

type Ghost struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	status int
}

func (g *Ghost) setStatus(status int) {
	g.mu.Lock()
	g.status = status
	g.mu.Unlock()
}

var ghostInstance = Ghost{
	status: GHOST_STATUS_NOT_RUNNING,
}

func main() {
	go discord()
}

func discord() {
	dg, _ := discordgo.New("Bot " + discordToken)

	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err := dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.")

	// Cleanly close down the Discord session.
	defer dg.Close()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)

	<-sigterm
	fmt.Println("Exiting...")
	os.Exit(0)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "/start" {
		if ghostInstance.status == GHOST_STATUS_RUNNING {
			_, err := s.ChannelMessageSend(m.ChannelID, "Terminating existing Ghost instance...")
			if err != nil {
				fmt.Println(err)
			}

			terminateGhost()

			time.Sleep(10 * time.Second)

			if ghostInstance.status != GHOST_STATUS_NOT_RUNNING {
				_, err := s.ChannelMessageSend(m.ChannelID, "Failed to terminate Ghost, try again later.")
				if err != nil {
					fmt.Println(err)
				}
				return
			}
		}

		ghostInstance.mu.Lock()
		if ghostInstance.status == GHOST_STATUS_NOT_RUNNING {
			_, err := s.ChannelMessageSend(m.ChannelID, "Starting Ghost...")
			if err != nil {
				fmt.Println(err)
			}

			go startGhost()
		} else {
			_, err := s.ChannelMessageSend(m.ChannelID, "Something went wrong, try again later.")
			if err != nil {
				fmt.Println(err)
			}
		}
		ghostInstance.mu.Unlock()
	}
}

func startGhost() {
	ghostInstance.setStatus(GHOST_STATUS_STARTING)

	ghostInstance.cmd = exec.Command("./ghost++")
	err := ghostInstance.cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	ghostInstance.setStatus(GHOST_STATUS_RUNNING)

	err = ghostInstance.cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	ghostInstance.setStatus(GHOST_STATUS_NOT_RUNNING)
}

func terminateGhost() {
	ghostInstance.cmd.Process.Kill()
}
