# --- SETTINGS ---
Output demo.mp4

# Global playback speed up (Compensates for the 15-20s LLM latency)
Set PlaybackSpeed 4.0

# Slow down the character typing speed significantly to compensate for the 4x global warp speed
Set TypingSpeed 250ms

# Visual styling
Set Width 1920
Set Height 1080
Set FontSize 22
Set Padding 40
Set FontFamily "DepartureMono Nerd Font Mono"
Set Theme { "background": "#161821", "foreground": "#c6c8d1", "black": "#161821", "red": "#e27878", "green": "#b4be82", "yellow": "#e2a478", "blue": "#84a0c6", "magenta": "#a093c0", "cyan": "#89b8c2", "white": "#c6c8d1", "brightBlack": "#6b7089", "brightRed": "#e98989", "brightGreen": "#c0ca95", "brightYellow": "#e9b189", "brightBlue": "#91acd1", "brightMagenta": "#ada0d3", "brightCyan": "#95c4ce", "brightWhite": "#d2d4de" }

# --- COMMANDS ---

# 1. Help Menu
Type "./trst --help"
Sleep 2s          # Multiplied (Saves as 500ms in video time)
Enter
Sleep 12s         # Multiplied (Saves as 3s in video time for reading the help)

Type "clear"
Enter
Sleep 2s

# 2. Showcase the Parent Persona
Type "./trst --persona parent --jerk 4 '~/Music/NCS/Heroes Tonight [yP32JFCwxB0].opus'"
Sleep 4s          # Short pause before hitting enter
Enter

# --- THE LLM GAP ---
# While Ollama works in the background for 20 real-world seconds,
# the video will zip through this dead air frame window in exactly 5 seconds.
Sleep 20s          

# --- READING PAUSE ---
# We want the viewer to have 7 full seconds to read the output text.
# Because the video is running at 4x speed, we multiply the sleep time by 4 (7 * 4 = 28s).
Sleep 28s          

# Clear the screen
Type "clear"
Enter
Sleep 2s

# 3. Showcase the Brainrot Persona
Type "./trst --persona brainrot --jerk 4 '~/Music/NCS/Heroes Tonight [yP32JFCwxB0].opus'"
Sleep 4s
Enter

# Zips through the processing window
Sleep 20s          

# Gives plenty of time to read the punchline at 4x speed
Sleep 28s
