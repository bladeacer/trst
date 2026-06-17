# trst

CLI for local LLMs to roast your music taste.

## Why the name?

Short for `track-roast`.

Also abbreviates to "trust me bro my music taste is good".

## Features

- Read `.lrc` when found in the same directory as audio file(s).

## Installation

### Via Go

```bash
go install github.com/bladeacer/trst@latest
```

### Via binary release (TBC)

Download the latest binary for your platform from the
[releases page](https://github.com/bladeacer/mns/releases), extract it, and
place it in your `$PATH`.


## Dependencies

`ollama`: Local models used to roast ^_^
`ffmpeg/ffprobe`: Getting audio file metadata
`yt-dlp` (Optional): Getting details on YouTube/YouTube Music URL.
`playerctl` (Optional): Getting details on currently playing media
