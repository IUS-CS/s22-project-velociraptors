# Challenge Accepted

## Initial setup:

Install DiscordGo

    go get github.com/bwmarrin/discordgo
    
Windows users install C compiler:

    https://jmeubank.github.io/tdm-gcc/
    (might have to reboot after install)
    
## To run:

Open terminal at project src folder

    /s22-project-velociraptors/src

Assign token variable

    For Mac:
    export BOT_TOKEN=<TOKEN GOES HERE!!!!!>

    For Windows (Powershell):
    Set-Variable -Name "BOT_TOKEN" -Value "<TOKEN GOES HERE!!!!!>"

Run main.go with & pass it the token variable

    go run main.go -t $BOT_TOKEN

## What is it?
Challenge Accepted is a Discord bot with a scoreboard to keep track of who in the server is right/wrong most often.

## How does it work? 
You reply to a message in the channel with !challenge to start a challenge. Users then use emojis reactions to vote for the winner of the challenge.

The winner is chosen/updated in real time once at least 2 votes have been cast.

Other commands include !leaderboard to display the leaderboard and !score '@user' to display the mentioned user's score.
