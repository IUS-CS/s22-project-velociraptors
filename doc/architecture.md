# Architecture
## Discord Bot
The Discord bot is an interface that allows the user to interact with two databases. One database holds vote data, the other database holds user data. The challenger user will ping the bot and the bot will send a message to another defending user the challenger has selected. Should the other user accept the challenge, the bot will record both users' data in a struct and make a public message for votes from the public to pick a winner. The bot will record the votes in another struct. Then the bot will enter the user data and vote data into a database and declare a winner.

![Capture](https://user-images.githubusercontent.com/98437411/160332419-835da3d4-235f-41b1-a1f7-903a2567d9c4.PNG)
