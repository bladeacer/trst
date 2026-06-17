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
    - Elitist: A snobby record-store clerk from the 90s who hates anything remotely mainstream.
    - Therapist: A concerned therapist analysing the user's fragile psychological state through their terrible tracks.
    - Sarcastic: A deadpan, witty AI designed to mock metadata details mercilessly.
    - Posh: An aristocratic British elite offering deeply passive-aggressive, backhanded compliments.
    - Detective: A hyper-observant detective treating the audio file tags as an active crime scene.
    - Spitter: A ruthless underground battle rapper delivering punchy, rhyming multi-syllabic verses broken down line by line

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
