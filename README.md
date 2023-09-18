# Addack

Addack is a tool to run exploits to multiple targets.

## Instalation

### Download

Download the binary from the [release page](https://github.com/Hiumee/addack/releases)

Run the binary. A `database.db` file and a `exploits` folder will be created in the same directory.

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

### Build

```bash
# Compile the css
npm install
npx tailwindcss -i ./assets/css/index.css -o ./assets/css/output.css
# Build the binary
go build -o addack
./addack
```