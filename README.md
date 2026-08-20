# Cartitas

## What

Kid-friendly flashcards for geography, space, animals, sports, and science. Cards are in English, Español, and Deutsch. [hashcards](https://github.com/eudoxia0/hashcards) runs spaced repetition on your machine.

Published decks live in `cards/{lang}/{Topic}/`.

## Why

Same idea as [cartas](https://github.com/qu1queee/cartas), aimed at kids from toddler years through adolescence. Short facts, one idea per card, in the language each child drills.

## How

Needs [Go](https://go.dev/dl/) 1.23+ and hashcards:

```sh
cargo install hashcards --locked

go run ./cmd/cartitas drill --lang en
go run ./cmd/cartitas drill --lang es
go run ./cmd/cartitas drill --lang de --topic Animals
```

The drill UI is at `http://127.0.0.1:8000`.

## GitHub Pages

A static deck runner lives in `docs/`. After Pages is on: [https://qu1queee.github.io/cartitas/](https://qu1queee.github.io/cartitas/). Pick language, topic, and age, then start. Preferences stay on that browser; there are no users or saved scores.

Refresh the card catalog:

```sh
go run ./cmd/cartitas web-export
```

Local preview:

```sh
go run ./cmd/cartitas web-export
python3 -m http.server -d docs 8080
```

Then open `http://127.0.0.1:8080`.

Adding or publishing cards: [docs/contributing.md](docs/contributing.md).
