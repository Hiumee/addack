# Addack

Addack is a tool to run exploits to multiple targets.

![Addack](https://github.com/Hiumee/addack/assets/42638867/14eb911c-b4b1-41c7-90c9-90b0fb545be4)

## Installation

### Download
Download the binary from the [release page](https://github.com/Hiumee/addack/releases)

Make an `exploits` directory in the same directory as the binary. Add a script called `flagger.py` in the `exploits` directory. This script will be called when a flag is found.

The command to run the exploit is in the `FlaggerCommand` variable. The directory used for this command is the `ExploitsPath` variable.

Run the binary. A `database.db` file will be created in the same directory.

Default configuration:

```
FlaggerCommand: "python3 flagger.py",
ExploitsPath:   "./exploits",
TickTime:       10 * 1000,
FlagRegex:      "FLAG{.*}",
TimeZone:       "Europe/Bucharest",
TimeFormat:     "2006-01-02 15:04:05",
```

Upload scripts and set the command to run. The IP of the target will be set in the `TARGET` environment variable when the script is run.

## Development

### Requirements

- Go 1.16 or newer
- npm for tailwind

### Development environment

```bash
# Compile the css
npx tailwindcss -i ./assets/css/index.css -o ./assets/css/output.css --watch
# Run the server using air for auto reload
air
```

You can change `main.go` to use the real file system instead of the embedded one.

### Build

```bash
# Compile the css
npm install
npx tailwindcss -i ./assets/css/index.css -o ./assets/css/output.css
# Build the binary
go build -o addack
./addack
```
