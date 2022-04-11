package db

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"log"
	"regexp"
	"strconv"
	"strings"
)

const (
	//test
	testTrigger   = "!test"
	testTrigger2  = "!test2"
	testResponse  = "This is a statement you might disagree with."
	testResponse2 = "This is another statement you might disagree with."

	//for sqlite
	dbname = "scoreboardDB"

	//bot commands
	commandChallenge  = "!challenge"
	commandCheckScore = "!checkscore"

	//bot messages
	challengeMessage1 = " has challenged "
	challengeMessage2 = "Vote below to decide who's right!"
	challengeMessage3 = "\n\n🟦 = "
	challengeMessage4 = "\n🟨 = "
	challengeMessage5 = "\n🟥 = Abstain"
	challengeMessage6 = "\n✋  = Close Voting"

	//values
	maxIDLength = 18
)

var RegexUserPatternID = regexp.MustCompile(fmt.Sprintf(`^(<@!(\d{%d,})>)$`, maxIDLength))

func oops(e error, n string) {
	log.Printf("Error %s in %s", e, n)
}

func alreadyVoted() {
	log.Println("User has voted already")
}

// MessageCreate trigger>response for messagecreate events
func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	var messageContent = m.Content
	var messageType = m.Type

	//to send a message when m.Content == <whatever trigger you want>
	//follow this format (EqualFold compares strings, ignores case and returns True if they are equal):
	if strings.EqualFold(messageContent, testTrigger) {
		_, err := s.ChannelMessageSend(m.ChannelID, testResponse)
		if err != nil {
			oops(err, "ChannelMessageSend")
			return
		}
	}

	if strings.EqualFold(messageContent, testTrigger2) {
		_, err := s.ChannelMessageSend(m.ChannelID, testResponse2)
		if err != nil {
			oops(err, "ChannelMessageSend")
			return
		}
	}

	//!challenge
	if strings.EqualFold(messageContent, commandChallenge) && messageType == discordgo.MessageTypeReply {
		//connect to challengeDB
		db, err := ConnectToDB()
		if err != nil {
			oops(err, "connectToDB()")
			return
		}
		defer func(db *sqlx.DB) {
			err := db.Close()
			if err != nil {
				oops(err, "Close()")
			}
		}(db)

		authorUsername := m.Message.Author.Username
		authorUserID := m.Message.Author.ID
		referencedAuthorUsername := m.ReferencedMessage.Author.Username
		referencedAuthorID := m.ReferencedMessage.Author.ID

		challengerInfo := "<@" + authorUserID + ">" + challengeMessage1 + "<@" + referencedAuthorID + ">" + "!"
		debate := "\n\n<@" + referencedAuthorID + ">" + " says: `" + m.ReferencedMessage.Content + "`\n\n<@" + authorUserID + "> disagrees!\n"
		votingInfo := "\n" + challengeMessage2 + challengeMessage3 + "<@" + authorUserID + ">" + challengeMessage4 + "<@" + referencedAuthorID + ">" + challengeMessage5 + challengeMessage6
		fullChallengeMessage := challengerInfo + debate + votingInfo
		announcementMessage, err := s.ChannelMessageSend(m.ChannelID, fullChallengeMessage)
		if err != nil {
			oops(err, "ChannelMessageSend")
			return
		}
		err = s.MessageReactionAdd(m.ChannelID, announcementMessage.ID, "🟦")
		if err != nil {
			oops(err, "MessageReactionAdd")
			return
		}
		err = s.MessageReactionAdd(m.ChannelID, announcementMessage.ID, "🟨")
		if err != nil {
			oops(err, "MessageReactionAdd")
			return
		}
		err = s.MessageReactionAdd(m.ChannelID, announcementMessage.ID, "🟥")
		if err != nil {
			oops(err, "MessageReactionAdd")
			return
		}
		err = s.MessageReactionAdd(m.ChannelID, announcementMessage.ID, "✋")
		if err != nil {
			oops(err, "MessageReactionAdd")
			return
		}
		announcementMessageID := announcementMessage.ID

		//create ChallengeTableEntry
		challengeTableEntry := initChallengeTableEntry(announcementMessageID, authorUserID, authorUsername, referencedAuthorID, referencedAuthorUsername)
		insertChallengeRow(db, challengeTableEntry)
		_, err = selectChallengeRow(db, announcementMessageID)
		if err != nil {
			oops(err, "selectChallengeRow")
		}

		//createScoreboardTableEntry x2 (one for challenger, one for defender)
		if !userInScoreboard(db, authorUserID) {
			scoreboardRow := initScoreBoardRow(authorUserID, authorUsername)
			insertScoreboardRow(db, scoreboardRow)
		}

		if !userInScoreboard(db, referencedAuthorID) {
			scoreboardRow := initScoreBoardRow(referencedAuthorID, referencedAuthorUsername)
			insertScoreboardRow(db, scoreboardRow)
		}
	}

	//!checkscore @username
	parameters := strings.Split(messageContent, " ")
	if strings.EqualFold(parameters[0], commandCheckScore) && RegexUserPatternID.MatchString(parameters[1]) {
		//connect to challengeDB
		db, err := ConnectToDB()
		if err != nil {
			oops(err, "connectToDB")
			return
		}
		defer func(db *sqlx.DB) {
			err := db.Close()
			if err != nil {
				oops(err, "Close()")
			}
		}(db)
		mentionedUser := parameters[1]
		re, err := regexp.Compile(`[^\w]`)
		if err != nil {
			oops(err, "regexp.Compile()")
		}
		mentionedUser = re.ReplaceAllString(mentionedUser, "")
		mentionedScoreboard, err := selectScoreboardRow(db, mentionedUser)
		output := "<@" + mentionedUser + "> has the following challenge record:\n" + scoreboardToString(mentionedScoreboard)
		_, err = s.ChannelMessageSend(m.ChannelID, output)
		if err != nil {
			oops(err, "channelMessageSend")
			return
		}
	}

}

// MessageReactionCreate trigger>response for messagereactionadd events
func MessageReactionCreate(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	reactionEmoji := r.Emoji.Name
	messageID := r.MessageID
	reactionAuthorID := r.UserID

	//ignore all reactions created by the bot itself
	if r.UserID == s.State.User.ID {
		return
	}

	if reactionEmoji == "🛹" {
		log.Println("Skateboard detected")
	}

	if reactionEmoji == "🟦" {
		//connect to challengeDB
		db, err := ConnectToDB()
		if err != nil {
			oops(err, "connectToDB")
			return
		}
		defer func(db *sqlx.DB) {
			err := db.Close()
			if err != nil {
				oops(err, "Close()")
			}
		}(db)
		if checkStopVotes(db, messageID) == 2 {
			return
		}
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if hasVotedBlue(db, votingRecordEntry) || hasVotedYellow(db, votingRecordEntry) || hasVotedRed(db, votingRecordEntry) {
			alreadyVoted()
			return
		}
		if !hasVotedBlue(db, votingRecordEntry) && !hasVotedRed(db, votingRecordEntry) && !hasVotedYellow(db, votingRecordEntry) {
			votingRecordEntry.UserID = reactionAuthorID
			votingRecordEntry.MessageID = messageID
			insertVotingRecordRow(db, votingRecordEntry)
		}
		votingRecordEntry.ChallengerVotes = 1
		updateVotingRecord(db, votingRecordEntry)
		votes, err := selectVotes(db, messageID)
		if err != nil {
			oops(err, "selectVotes")
			return
		}
		ChallengerVotes := votes.ChallengerVotes + 1
		DefenderVotes := votes.DefenderVotes
		AbstainVotes := votes.AbstainVotes
		StopVotes := votes.StopVotes
		updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes, StopVotes}
		updateVotes(db, messageID, updatedVotes)
		votes, err = selectVotes(db, messageID)
		if err != nil {
			oops(err, "selectVotes")
			return
		}
		updateOutcome(db, messageID, votes)
		_, err = selectChallengeRow(db, messageID)
	}

	if reactionEmoji == "🟨" {
		//connect to challengeDB
		db, err := ConnectToDB()
		if err != nil {
			oops(err, "connectToDB")
			return
		}
		defer func(db *sqlx.DB) {
			err := db.Close()
			if err != nil {
				oops(err, "Close()")
			}
		}(db)
		if checkStopVotes(db, messageID) == 2 {
			return
		}
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if hasVotedYellow(db, votingRecordEntry) || hasVotedBlue(db, votingRecordEntry) || hasVotedRed(db, votingRecordEntry) {
			alreadyVoted()
			return
		}
		if !hasVotedBlue(db, votingRecordEntry) && !hasVotedRed(db, votingRecordEntry) && !hasVotedYellow(db, votingRecordEntry) {
			votingRecordEntry.UserID = reactionAuthorID
			votingRecordEntry.MessageID = messageID
			insertVotingRecordRow(db, votingRecordEntry)
		}
		votingRecordEntry.DefenderVotes = 1
		updateVotingRecord(db, votingRecordEntry)
		votes, err := selectVotes(db, messageID)
		if err != nil {
			oops(err, "selectVotes")
			return
		}
		ChallengerVotes := votes.ChallengerVotes
		DefenderVotes := votes.DefenderVotes + 1
		AbstainVotes := votes.AbstainVotes
		StopVotes := votes.StopVotes
		updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes, StopVotes}
		updateVotes(db, messageID, updatedVotes)
		votes, err = selectVotes(db, messageID)
		if err != nil {
			oops(err, "selectVotes")
			return
		}
		updateOutcome(db, messageID, votes)
		_, err = selectChallengeRow(db, messageID)
		if err != nil {
			oops(err, "selectChallengeRow")
			return
		}
	}

	if reactionEmoji == "🟥" {
		//connect to challengeDB
		db, err := ConnectToDB()
		if err != nil {
			oops(err, "connectToDB")
			return
		}
		defer func(db *sqlx.DB) {
			err := db.Close()
			if err != nil {
				oops(err, "Close()")
			}
		}(db)
		if checkStopVotes(db, messageID) == 2 {
			return
		}
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if err != nil {
			oops(err, "selectVotingRecordRow")
		}
		if hasVotedRed(db, votingRecordEntry) || hasVotedBlue(db, votingRecordEntry) || hasVotedYellow(db, votingRecordEntry) {
			alreadyVoted()
			return
		}
		if !hasVotedBlue(db, votingRecordEntry) && !hasVotedRed(db, votingRecordEntry) && !hasVotedYellow(db, votingRecordEntry) {
			votingRecordEntry.UserID = reactionAuthorID
			votingRecordEntry.MessageID = messageID
			insertVotingRecordRow(db, votingRecordEntry)
		}
		votingRecordEntry.AbstainVotes = 1
		updateVotingRecord(db, votingRecordEntry)
		votes, err := selectVotes(db, messageID)
		if err != nil {
			oops(err, "selectVotes")
			return
		}
		ChallengerVotes := votes.ChallengerVotes
		DefenderVotes := votes.DefenderVotes
		AbstainVotes := votes.AbstainVotes + 1
		StopVotes := votes.StopVotes
		updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes, StopVotes}
		updateVotes(db, messageID, updatedVotes)
		votes, err = selectVotes(db, messageID)
		if err != nil {
			oops(err, "selectVotes")
			return
		}
		updateOutcome(db, messageID, votes)
		_, err = selectChallengeRow(db, messageID)
		if err != nil {
			oops(err, "selectChallengeRow")
			return
		}
	}

	if reactionEmoji == "✋" {
		db, err := ConnectToDB()
		if err != nil {
			oops(err, "connectToDB")
			return
		}
		defer func(db *sqlx.DB) {
			err := db.Close()
			if err != nil {
				oops(err, "Close()")
			}
		}(db)
		if checkStopVotes(db, messageID) == 2 {
			return
		}
		challengeEntry, err := selectChallengeRow(db, messageID)
		if err != nil {
			oops(err, "selectChallengeRow")
			return
		}
		stopVotesTotal := checkStopVotes(db, messageID)
		userVotingRecord, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		challengeVotes, err := selectVotes(db, messageID)
		if !hasVotedStop(db, userVotingRecord) {
			userVotingRecord.StopVotes = 1
			updateVotingRecord(db, userVotingRecord)
			challengeVotes.StopVotes += 1
			updateVotes(db, messageID, challengeVotes)
		}
		stopVotesTotal = checkStopVotes(db, messageID)
		if stopVotesTotal != 2 {
			return
		}
		if stopVotesTotal == 2 {
			pushScore(db, challengeEntry)
			winnerIsChallenger := "\n<@" + challengeEntry.ChallengerID + "> has won the challenge!\n\nThe score was: " + strconv.Itoa(challengeEntry.ChallengerVotes) + " to " + strconv.Itoa(challengeEntry.DefenderVotes)
			winnerIsDefender := "\n<@" + challengeEntry.DefenderID + "> has won the challenge!\n\nThe score was: " + strconv.Itoa(challengeEntry.DefenderVotes) + " to " + strconv.Itoa(challengeEntry.ChallengerVotes)
			tie := "\nThe challenge between <@" + challengeEntry.ChallengerID + "> and <@" + challengeEntry.DefenderID + "> was a tie!"
			_, err = selectScoreboardRow(db, challengeEntry.ChallengerID)
			if err != nil {
				oops(err, "selectScoreboardRow")
			}
			_, err = selectScoreboardRow(db, challengeEntry.DefenderID)
			if winnerID(challengeEntry) == "tie" {
				_, err := s.ChannelMessageSend(r.ChannelID, tie)
				if err != nil {
					oops(err, "ChannelMessageSend")
					return
				}
			}
			if winnerID(challengeEntry) == challengeEntry.ChallengerID {
				_, err := s.ChannelMessageSend(r.ChannelID, winnerIsChallenger)
				if err != nil {
					oops(err, "ChannelMessageSend")
					return
				}
			}
			if winnerID(challengeEntry) == challengeEntry.DefenderID {
				_, err := s.ChannelMessageSend(r.ChannelID, winnerIsDefender)
				if err != nil {
					oops(err, "ChannelMessageSend")
					return
				}
			}
		}
	}
}

// MessageReactionDelete trigger>response for messagereactionremove events
func MessageReactionDelete(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	reactionEmoji := r.Emoji.Name
	messageID := r.MessageID
	reactionAuthorID := r.UserID

	if reactionEmoji == "🛹" {
		log.Println("Skateboard removed")
	}

	if reactionEmoji == "🟦" {
		db, err := ConnectToDB()
		if err != nil {
			oops(err, "connectToDB")
			return
		}
		defer func(db *sqlx.DB) {
			err := db.Close()
			if err != nil {
				oops(err, "Close()")
			}
		}(db)
		if checkStopVotes(db, messageID) == 2 {
			return
		}
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if hasVotedBlue(db, votingRecordEntry) {
			removeVotingRecordRow(db, votingRecordEntry)
			votes, err := selectVotes(db, messageID)
			if err != nil {
				oops(err, "selectVotes")
				return
			}
			ChallengerVotes := votes.ChallengerVotes - 1
			DefenderVotes := votes.DefenderVotes
			AbstainVotes := votes.AbstainVotes
			StopVotes := votes.StopVotes
			updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes, StopVotes}
			updateVotes(db, messageID, updatedVotes)
			votes, err = selectVotes(db, messageID)
			if err != nil {
				oops(err, "selectVotes")
				return
			}
			updateOutcome(db, messageID, votes)
			_, err = selectChallengeRow(db, messageID)
			if err != nil {
				oops(err, "selectChallengeRow")
				return
			}
		}
	}

	if reactionEmoji == "🟨" {
		db, err := ConnectToDB()
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
		if checkStopVotes(db, messageID) == 2 {
			return
		}
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if hasVotedYellow(db, votingRecordEntry) {
			removeVotingRecordRow(db, votingRecordEntry)
			votes, err := selectVotes(db, messageID)
			if err != nil {
				oops(err, "selectVotes")
				return
			}
			ChallengerVotes := votes.ChallengerVotes
			DefenderVotes := votes.DefenderVotes - 1
			AbstainVotes := votes.AbstainVotes
			StopVotes := votes.StopVotes
			updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes, StopVotes}
			updateVotes(db, messageID, updatedVotes)
			votes, err = selectVotes(db, messageID)
			if err != nil {
				oops(err, "selectVotes")
				return
			}
			updateOutcome(db, messageID, votes)
			_, err = selectChallengeRow(db, messageID)
		}
	}

	if reactionEmoji == "🟥" {
		db, err := ConnectToDB()
		if err != nil {
			oops(err, "connectToDB")
			return
		}
		defer func(db *sqlx.DB) {
			err := db.Close()
			if err != nil {
				oops(err, "Close()")
			}
		}(db)
		if checkStopVotes(db, messageID) == 2 {
			return
		}
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if hasVotedRed(db, votingRecordEntry) {
			removeVotingRecordRow(db, votingRecordEntry)
			votes, err := selectVotes(db, messageID)
			if err != nil {
				oops(err, "selectVotes")
				return
			}
			ChallengerVotes := votes.ChallengerVotes
			DefenderVotes := votes.DefenderVotes
			AbstainVotes := votes.AbstainVotes - 1
			StopVotes := votes.StopVotes
			updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes, StopVotes}
			updateVotes(db, messageID, updatedVotes)
			votes, err = selectVotes(db, messageID)
			if err != nil {
				oops(err, "selectVotes")
				return
			}
			updateOutcome(db, messageID, votes)
			_, err = selectChallengeRow(db, messageID)
			if err != nil {
				oops(err, "selectChallengeRow")
				return
			}
		}
	}

	if reactionEmoji == "✋" {
		db, err := ConnectToDB()
		if err != nil {
			oops(err, "connectToDB")
			return
		}
		defer func(db *sqlx.DB) {
			err := db.Close()
			if err != nil {
				oops(err, "Close()")
			}
		}(db)
		if checkStopVotes(db, messageID) == 2 {
			return
		}
		challengeEntry, err := selectChallengeRow(db, messageID)
		if err != nil {
			oops(err, "selectChallengeRow")
			return
		}
		userVotingRecord, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		challengeVotes, err := selectVotes(db, messageID)
		stopVotesTotal := checkStopVotes(db, messageID)
		if hasVotedStop(db, userVotingRecord) {
			userVotingRecord.StopVotes = 0
			updateVotingRecord(db, userVotingRecord)
			challengeVotes.StopVotes -= 1
			updateVotes(db, messageID, challengeVotes)
		}
		stopVotesTotal = checkStopVotes(db, messageID)
		if stopVotesTotal != 2 {
			return
		}
		if stopVotesTotal == 2 {
			pushScore(db, challengeEntry)
			winnerIsChallenger := "\n<@" + challengeEntry.ChallengerID + "> has won the challenge!\n\nThe score was: " + strconv.Itoa(challengeEntry.ChallengerVotes) + " to " + strconv.Itoa(challengeEntry.DefenderVotes)
			winnerIsDefender := "\n<@" + challengeEntry.DefenderID + "> has won the challenge!\n\nThe score was: " + strconv.Itoa(challengeEntry.DefenderVotes) + " to " + strconv.Itoa(challengeEntry.ChallengerVotes)
			tie := "\nThe challenge between <@" + challengeEntry.ChallengerID + "> and <@" + challengeEntry.DefenderID + "> was a tie!"
			_, err := selectScoreboardRow(db, challengeEntry.ChallengerID)
			if err != nil {
				oops(err, "selectScoreboardRow")
			}
			_, err = selectScoreboardRow(db, challengeEntry.DefenderID)
			if err != nil {
				oops(err, "selectScoreboardRow")
			}
			if winnerID(challengeEntry) == "tie" {
				_, err := s.ChannelMessageSend(r.ChannelID, tie)
				if err != nil {
					oops(err, "ChannelMessageSend")
					return
				}
			}
			if winnerID(challengeEntry) == challengeEntry.ChallengerID {
				_, err := s.ChannelMessageSend(r.ChannelID, winnerIsChallenger)
				if err != nil {
					oops(err, "ChannelMessageSend")
					return
				}
			}
			if winnerID(challengeEntry) == challengeEntry.DefenderID {
				_, err := s.ChannelMessageSend(r.ChannelID, winnerIsDefender)
				if err != nil {
					oops(err, "ChannelMessageSend")
					return
				}
			}
		}
	}
}