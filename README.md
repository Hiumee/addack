# Addack

Addack is a tool to run exploits to multiple targets.

## Instalation

### Download

Download the binary from the [release page](https://github.com/Hiumee/addack/releases)

Run the binary. A `database.db` file will be created in the same directory.

Edit the `flagger.py` file to send the flag to the server. The command to run the exploit is in the `FlaggerCommand` variable. The directory used for this command is the `ExploitsPath` variable.

Default configuration:

```
FlaggerCommand: "python3 flagger.py",
ExploitsPath:   "./exploits",
TickTime:       10 * 1000,
FlagRegex:      "FLAG{.*}",
TimeZone:       "Europe/Bucharest",
TimeFormat:     "2006-01-02 15:04:05",
```

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