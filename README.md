# Lucy

Lucy is a Discord bot built using **Golang**. This project aims to provide a foundation for creating bots with custom commands and events.

---

## Features
- Written in Go for performance and scalability.
- Modular structure with organized packages and event handling.
- Example commands and events to get started quickly.

---

## Getting Started

### Prerequisites
Ensure you have the following installed:
- **Go** (Golang) - [Install Go](https://go.dev/doc/install)
- **Git** - [Install Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- A Discord bot token - [Create a bot](https://discord.com/developers/applications)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/Ariffansyah/Lucy.git
   cd Lucy
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up the environment variables:
   - Create a `.env` file in the root directory.
   - Use the provided `.envExample` file as a template.
   - Fill in your **Discord Bot Token** and other required variables.

4. Run the bot:
   ```bash
   go run main.go
   ```

---

## Directory Structure
- `commands/runs/`: Contains bot commands.
- `events/jtc/`: Handles Discord events such as voice state updates.
- `pkg/`: Additional utility packages.
- `.envExample`: Example environment variables file.
- `main.go`: Entry point of the bot.

---

## Contributing

Contributions are welcome! Please feel free to open issues or submit pull requests for new features or bug fixes.

---

## License
This project is licensed under the [MIT License](LICENSE).

---
