# trst

CLI for local LLMs to roast your music taste.

## Why the name?

Short for `track-roast`.

Another way you can think of it is: "trust me bro my music taste is totally good".

## Who is this not for?

Those who cannot tolerate humour ranging from light-hearted banter to absolute emotional damage.

## Features

- Read `.lrc` when found in the same directory as audio file(s).
- Semantically infer genre/sub-genre (best effort basis) via LLM
- Local metadata cache for parsed songs
> No CGO Sqlite database, defaults to where it makes sense on your platform

- Read file codec, title, description, metadata

- Choose between distinct personas, including:
    - Elitist: A snobby 90s record-store clerk who despises everything mainstream.
    - Therapist: A concerned professional diagnosing your fragile psyche via your tragic music choices.
    - Sarcastic: A hyper-sassy, cynical lady dispensing raw, unfiltered mockery and heavy side-eye.
    - Posh: A passive-aggressive British aristocrat dispensing polite, devastating insults over tea.
    - Detective: A Sherlock Holmes-inspired sleuth treating audio file tags as an active crime scene.
    - Spitter: A ferocious American battle rapper delivering technical, multi-syllabic rhyming bars line by line.
    - Influencer: A cringe-inducing content creator plagued by toxic enthusiasm and desperate vocal fry.
    - Brainrot: A chaotic stream of internet culture, emoji tags, and non-repeating Gen Z slang.
    - Pianist: A conservative classical concert pianist appalled by modern production shortcuts.
    - Hater: A relentlessly toxic internet troll providing pure, bad-faith hostility regardless of configurations.
    - Normie: A delusional, middle-school pop consumer convinced they possess elite musical knowledge.
    - Parent: A traditional, deeply disappointed Asian parent wielding the threat of the flying house slipper.
    - Nerd: A pedantic, socially awkward pedant starting every sentence with 'Erm, actually...' to obsess over technical metadata flaws.

- Set how much of a jerk the LLM persona can be. On a scale from 1 to 5, defaults to 3. 
1 is lowest, 5 is highest.
> Note: Hater will still be a hater even you set jerk value to 1

- Set whether profanities are allowed in the outputs. Defaults to false.
> Most models have safety nets baked into their system prompts, if you do want this
you are very much on your own.


## Installation

### Via Go

```bash
go install github.com/bladeacer/trst/cmd/trst@latest
```

### Via binary release

Download the latest binary for your platform from the
[releases page](https://github.com/bladeacer/trst/releases), extract it, and
place it in your `$PATH`.

Pre-built binaries are available for:

| OS | Architectures |
| --- | --- |
| Linux / WSL | `amd64` (x86-64), `arm64` (ARM 64-bit) |
| macOS | `amd64` (Intel), `arm64` (Apple Silicon) |
| Windows | `amd64` (x86-64), `arm64` (ARM 64-bit) |

All binaries are fully static (compiled with `CGO_ENABLED=0`) with no
C runtime dependencies - the Linux archive works on both native Linux
and WSL without extra setup.

## Usage

Install and start `ollama` first.
> See [ollama's website](https://ollama.com/) for more details.

```
trst
```

## Dependencies

- `ollama`: Powers the local LLM inference engine for generating roasts. ^_^
- `ffmpeg/ffprobe`: Handles deep container queries to extract core audio file metadata.
- `yt-dlp` (Optional): Pulls live details and metadata directly from YouTube/YouTube Music URLs.
- `playerctl` (Optional): Queries MPRIS D-Bus interfaces to roast currently playing media.

## LLM Usage Disclosure

I used AI assistance for writing the code.

## License

This Golang CLI app, "trst" is released under the GNU Affero General Public
License version 3 (AGPLv3) License.

### License Notice

This file is part of trst. trst is a CLI for local LLMs to roast your music taste.
Copyright (c) 2026 bladeacer

trst is free software: you can redistribute it and/or modify it under the
terms of the GNU Affero General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later version.

trst is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
PURPOSE. See the GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License along with trst.
If not, see <https://www.gnu.org/licenses/>.

### License file

You can find the [license file here](./LICENSE).

## Credits

This CLI was made possible by the following open-source libraries

- [`modernc.org/sqlite`](https://gitlab.com/cznic/sqlite): A pure Go, zero-dependency SQLite driver that requires no CGO implementation.
- [`dhowden/tag`](https://github.com/dhowden/tag): A clean audio metadata parsing engine used for reading track container properties.
