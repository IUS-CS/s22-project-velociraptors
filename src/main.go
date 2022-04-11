package main

import (
	"flag"
	bot "github.com/IUS-CS/s22-project-velociraptors/src/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func oops(e error, n string) {
	log.Printf("Error %s in %s", e, n)
}

// Token variable used for command line parameters
var Token string

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	//create a new Discord session using the provided bot token
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		oops(err, "New(Bot + Token")
		return
	}
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentGuildMessageReactions
	//open a websocket connection to Discord and begin listening
	err = dg.Open()
	if err != nil {
		oops(err, "Open()")
		return
	}
	//connect to scoreboardDB
	db, err := bot.ConnectToDB()
	if err != nil {
		oops(err, "ConnectToDB")
		return
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			oops(err, "Close()")
		}
	}(db)
	log.Printf("Successfully connected to database")
	//create tables
	err = bot.CreateChallengeTable(db)
	if err != nil {
		oops(err, "CreateChallengeTable")
		return
	}
	err = bot.CreateScoreboardTable(db)
	if err != nil {
		oops(err, "CreateScoreboardTable")
		return
	}
	err = bot.CreateVotingRecord(db)
	if err != nil {
		oops(err, "CreateVotingRecord")
		return
	}

	//register messageCreate function as a callback for MessageCreate events
	dg.AddHandler(bot.MessageCreate)
	dg.AddHandler(bot.MessageReactionCreate)
	dg.AddHandler(bot.MessageReactionDelete)

	//everything runs here until one of the term signals is received
	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	//close the Discord session
	err = dg.Close()
	if err != nil {
		oops(err, "Close()")
		return
	}
}