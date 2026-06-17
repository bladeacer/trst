# trst

CLI for local LLMs to roast your music taste.

## Why the name?

Short for `track-roast`.

Another way you can think of it is: "trust me bro my music taste is totally good".

## Features

- Read `.lrc` when found in the same directory as audio file(s).
- Semantically infer genre/sub-genre (best effort basis) via LLM
- Local metadata cache for parsed songs
- Read file codec, title, description, metadata
- Choose between personas, options include:
    - Elitist: a snobby record-store clerk from the 90s
    - Therapist: a concerned therapist reading into the user's psychological state
    - Sarcastic: a witty, extremely sarcastic AI
    - Posh: an aristocratic, deeply passive-aggressive British aristocrat
    - Investigator: a brilliant, hyper-observant detective like Sherlock Holmes
    - Spitter: a ruthless underground battle rapper

- Set how much of a jerk the LLM persona can be. On a scale from 1 to 5, defaults to 3.
- Set whether profanities are allowed in the outputs. Defaults to false.
> Most models have safety nets baked into their system prompts, if you do want this
you are very much on your own.


## Installation

### Via Go

```bash
go install github.com/bladeacer/trst@latest
```

### Via binary release (TBC)

Download the latest binary for your platform from the
[releases page](https://github.com/bladeacer/mns/releases), extract it, and
place it in your `$PATH`.

## Usage

```
trst
```

## Dependencies

`ollama`: Local models used to roast ^_^
`ffmpeg/ffprobe`: Getting audio file metadata
`yt-dlp` (Optional): Getting details on YouTube/YouTube Music URL.
`playerctl` (Optional): Getting details on currently playing media
