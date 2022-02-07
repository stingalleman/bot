package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "modernc.org/sqlite"
)

var (
	db *sql.DB
)

func main() {
	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	dg.AddHandler(messageCreate)

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./db.sqlite"
	}

	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `karma` (name VARCHAR(256), count INT DEFAULT 0, PRIMARY KEY (name));")
	if err != nil {
		panic(err)
	}

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == ".karma" {
		res, err := db.Query("SELECT `count`, `name` FROM `karma`;")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("sql is stukkie wukkie (1): %s", err))
			return
		}

		var msg string

		for res.Next() {
			var count int
			var name string
			err = res.Scan(&count, &name)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("sql is stukkie wukkie (2): %s", err))
				return
			}

			if count == 0 {
				continue
			}

			if name == "" {
				continue
			}

			msg += fmt.Sprintf("%s: %d\n", name, count)
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\n%s```", msg))

		res.Close()
	} else if strings.HasSuffix(m.Content, "++") {
		// ++
		msg := strings.Split(m.Content, "++")
		name := strings.TrimSpace(msg[0])
		name = strings.ReplaceAll(name, "@", "")
		name = strings.ReplaceAll(name, "`", "")
		name = strings.ReplaceAll(name, "#", "")

		if name == "" {
			return
		}

		_, err := db.Exec("INSERT OR IGNORE INTO `karma` (name) VALUES(?);", name)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("sql is stukkie wukkie (1): %s", err))
			return
		}
		_, err = db.Exec("UPDATE `karma` SET `count`=`count`+1 WHERE `name`=?;", name)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("sql is stukkie wukkie (2): %s", err))
			return
		}

		res := db.QueryRow("SELECT `count`, `name` FROM `karma` WHERE `name`=?;", name)

		var resCount int
		var resName string
		err = res.Scan(&resCount, &resName)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("sql is stukkie wukkie (4): %s", err))
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s: %v", resName, resCount))
	} else if strings.HasSuffix(m.Content, "--") {
		// --
		msg := strings.Split(m.Content, "--")
		name := strings.TrimSpace(msg[0])
		name = strings.ReplaceAll(name, "@", "")
		name = strings.ReplaceAll(name, "`", "")
		name = strings.ReplaceAll(name, "#", "")

		if name == "" {
			return
		}

		_, err := db.Exec("INSERT OR IGNORE INTO `karma` (name) VALUES(?);", name)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("sql is stukkie wukkie (1): %s", err))
			return
		}
		_, err = db.Exec("UPDATE `karma` SET `count`=`count`-1 WHERE `name`=?;", name)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("sql is stukkie wukkie (2): %s", err))
			return
		}

		res := db.QueryRow("SELECT `count`, `name` FROM `karma` WHERE `name`=?;", name)

		var resCount int
		var resName string
		err = res.Scan(&resCount, &resName)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("sql is stukkie wukkie (4): %s", err))
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s: %v", resName, resCount))
	} else if strings.HasPrefix(m.Content, ".sql") && m.Author.ID == "125916793817530368" {
		res, err := db.Exec(strings.TrimPrefix(m.Content, ".sql"))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("sql is stukkie wukkie: %s", err))
			return
		}

		rows, _ := res.RowsAffected()
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%d rows affected", rows))
	}
}
