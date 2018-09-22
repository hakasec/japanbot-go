# japanbot-go

A Discord bot with some useful features.

## Requirements

This bot requires Go version 1.10 or higher to run.

## Features

- Analyse and breakdown a Japanese sentence or phrase into smaller definitions,
similar to the [Rikaichan](https://addons.mozilla.org/en-US/firefox/addon/rikaichan/)
or [Rikaikun](https://chrome.google.com/webstore/detail/rikaikun/jipdnfibhldikgcjhfnomkfpcebammhp) browser plugins.
- More soon!

## Configuration

To configure the bot, rename the `config.json.example` file to `config.json` and specify your API secret key
and path to your JMDict file.

You can obtain the latest JMDict file from [here](ftp://ftp.monash.edu.au/pub/nihongo/JMdict.gz); 
unzip this file and you're ready to go!

## Using the bot

You can interact with the bot using commands preceded by `jpn!`.
The main feature of the bot is `jpn!analyse` (or `jpn!analyze`, if you prefer the OUP/American spelling).

![analyse_example](docs/images/analyse_example1.png)
