<div align="center">

# go-anki

**Create and study your flash cards using Anki from the CLI**

  <img alt="Version" src="https://img.shields.io/badge/version-0.0.1-blue.svg?cacheSeconds=2592000" />
  <a href="https://github.com/Aerex/go-anki" target="_blank">
    <img alt="Documentation" src="https://img.shields.io/badge/documentation-yes-brightgreen.svg" />
  </a>
  <a href="#" target="_blank">
    <img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg" />
  </a>

<br/>

`go-anki` is a command line client for managing and studying flash cards on Anki.

# WARNING: `go-anki` is in early alpha phase. Some features are still not working properly


[Installation](#installation) •
[Overview](#overview) •
[Motivation](#motivation) •
[Roadmap](#roadmap) •
[Contributing](#contributing)

</div>
<br/>



## 📖 Key features
- Manage cards, decks and note-types

## Motivation
I use Anki-Droid a lot but making flash cards on the phone can be a pain. There is the desktop app but I am more of a CLI guy so I decided to create one. Intially, I wanted to create an API to interface with the CLI but there is no official API currently. There is [dsnopek/anki-sync-server](https://github.com/dsnopek/anki-sync-server) but it is very limited. My plan is to make the CLI compatible with Anki 2.157+ so that users can sync their decks like any other official client.

## Installation
```sh
go install github.com/Aerex/go-anki@latest
```

## Overview
<div align="center">
  <a href="#deck">Deck</a>&nbsp;·
  <a href="#card">Card</a>&nbsp;·
  <a href="#@card_type">Card Type</a>&nbsp;·
</div>

### 📝 Deck
#### Creating

To create a deck, run `anki deck create`
```bash
# create a deck called Grammar
anki deck create Grammar
```

### 📝 Card
#### Creating
To create a card, run `anki card create`
```bash
# create a new card interactively
anki card create

# set deck name before creating new card interactively
anki card create --deck Grammar
```
`anki card create` with no optional flags will create a card interactively. Use optional flags to skip specific prompt. For instance, use `--deck|-d` to skip the **Deck** prompt

## Roadmap
- [ ] Add translation
- [ ] Add ability to study a deck

## Contributing

Contributions, issues and feature requests are welcome!<br />Feel free to check our [contribution page](./CONTRIBUTING.md)
