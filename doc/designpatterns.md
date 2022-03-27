# Design Patterns

## Introduction
Design patterns currently in use:
#### Singleton
Singleton is a creational design pattern that says a class can only be instantiated once. The Discord bot will create and make use of two different databases. There should be only one instance of these databases. When the bot is instructed to create a database, it checks if there already is an instance of a database.
#### Chain of Responsibility
Chain of responsibility is a behavioral design pattern that uses an abstract handler to delegate client requests to a concrete handler. The Discord user is the client making requests. The Discord bot is the abstract handler and delegates requests. The structs and databases will actually handle the requests and if it can't, pass the request on. Though our chain is short and the database can't handle it, the user would get a generic message.
#### Facade
Facade is structural design pattern which implements an interface to hide subsystems. In our model, the Discord bot is the facade, the Discord user is the client, and the databases are the subsystems.

## Patterns That May Fit
#### Observer
Observer is a behavioral design pattern which uses a publish/subcribe model. This could be used to create subcribers that may want to watch for challenges to vote or just watch.

## Going Forward
Given that the model includes a Discord bot, the facade design pattern will be implemented throughout. The chain of responsibility pattern may be deepened if we implement more objects for expanded use of the bot. The singleton pattern will continue to be a strong part of the model so that user data isn't lost.
