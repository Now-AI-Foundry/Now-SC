# Now-SC CLI Tool

A CLI tool for bootstrapping presales projects for solution consultants, written in Go for easy binary distribution.

## Features

- 🚀 Single binary distribution (no runtime dependencies)
- 📁 Automated project structure creation
- 🤖 AI-powered prompt execution via OpenRouter
- 🐙 GitHub repository integration
- 💾 Cross-platform support (Linux, macOS, Windows)

## Installation

### From Binary Release

Download the latest binary for your platform from the [releases page](https://github.com/Now-AI-Foundry/Now-SC/releases):

**Linux (AMD64)**:
```bash
curl -LO https://github.com/Now-AI-Foundry/Now-SC/releases/latest/download/now-sc-linux-amd64
chmod +x now-sc-linux-amd64
sudo mv now-sc-linux-amd64 /usr/local/bin/now-sc
```

**macOS (Intel)**:
```bash
curl -LO https://github.com/Now-AI-Foundry/Now-SC/releases/latest/download/now-sc-darwin-amd64
chmod +x now-sc-darwin-amd64
sudo mv now-sc-darwin-amd64 /usr/local/bin/now-sc
```

**macOS (Apple Silicon)**:
```bash
curl -LO https://github.com/Now-AI-Foundry/Now-SC/releases/latest/download/now-sc-darwin-arm64
chmod +x now-sc-darwin-arm64
sudo mv now-sc-darwin-arm64 /usr/local/bin/now-sc
```

**Windows**:
Download `now-sc-windows-amd64.exe` from the releases page and add it to your PATH.

### From Source

Requires Go 1.21 or later:

```bash
git clone https://github.com/Now-AI-Foundry/Now-SC.git
cd Now-SC
make build
# Binary will be in bin/now-sc
```

## Usage

### Initialize a New Project

```bash
now-sc init
```

With flags:
```bash
now-sc init --name my-project --customer "Acme Corp"
```

Skip GitHub repo creation:
```bash
now-sc init --no-github
```

### Execute Prompts

Navigate to your project directory and run:
```bash
cd my-project
now-sc prompt
```

## Configuration

### Environment Variables

- `OPENROUTER_API_KEY` - Required for prompt execution. Get your key from [OpenRouter](https://openrouter.ai/)
- `GITHUB_PAT` - Optional. Required for automatic GitHub repository creation

Example `.env` file:
```bash
OPENROUTER_API_KEY=your_openrouter_api_key
GITHUB_PAT=your_github_personal_access_token
```

## Project Structure

When you initialize a project, the following structure is created:

```
project-name/
├── 00_Inbox/
│   ├── calls/
│   │   ├── internal/
│   │   └── external/
│   ├── emails/
│   └── notes/
├── 01_Customers/
│   └── [CustomerName]/
├── 10_PromptTemplates/
├── 20_Demo_Library/
├── 30_CommunicationTemplates/
├── 99_Assets/
│   ├── Project_Overview/
│   ├── Communications/
│   └── POC_Documents/
├── README.md
├── .env.example
└── .gitignore
```

## Development

### Build Commands

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean

# Install dependencies
make deps

# Run tests
make test

# Install to $GOPATH/bin
make install
```

### Creating a Release

1. Tag the commit:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. GitHub Actions will automatically:
   - Build binaries for all platforms
   - Create checksums
   - Create a GitHub release with all artifacts

## Advantages Over Node.js Version

The Go implementation provides:

- Single binary distribution (no Node.js installation required)
- Faster startup and execution
- Smaller memory footprint
- Native cross-compilation support
- Easier deployment and distribution

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
