# Addack

Addack is a tool to run exploits to multiple targets.

![Addack](https://github.com/Hiumee/addack/assets/42638867/14eb911c-b4b1-41c7-90c9-90b0fb545be4)

## Installation

Use a linux distribution or WSL.

### Download
Download the binary from the [release page](https://github.com/Hiumee/addack/releases)

Make an `exploits` directory in the same directory as the binary. Add a script called `flagger.py` in the `exploits` directory. This script will be called when a flag is found.

The command to run the exploit is in the `FlaggerCommand` variable. The directory used for this command is the `ExploitsPath` variable.

Run the binary. A `database.db` file will be created in the same directory.

Default configuration:

```
database_path = "./database.db"
flagger_command = "python3 flagger.py"
exploits_path = "./exploits"
tick_time = 10000 # in ms
flag_regex = "FLAG{.*?}"
timezone = "Europe/Bucharest"
time_format = "2006-01-02 15:04:05"
listening_addr = "127.0.0.1:8080"
```

You can edit the configuration by editing the `config.toml` file.

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
