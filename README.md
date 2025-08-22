# watch-node

**A real-time TUI monitor for Node.js processes (CPU & RAM usage)**

`watch-node` is a lightweight command-line tool built in Go that monitors **Node.js processes** in real-time.  
It displays CPU usage (%) and memory consumption (RAM) using a clean terminal interface powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea).

---

## Features

- 📊 Real-time monitoring of **CPU** and **RAM** usage.
- 🎯 Focused specifically on **Node.js processes**.
- 🖥️ Terminal-based UI (TUI) with refresh every second.
- 🔍 Automatically scans `/proc` for active Node.js processes.
- ⚡ Built in Go for efficiency and portability.
- 🐧 Works on **Linux** and should also work on other UNIX-like systems (BSD).

---

## Requirements

- **Go 1.20+**
- A **UNIX-like operating system** with `/proc` available (Linux recommended).

---

## Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/BenBerg02/watch-node.git
cd watch-node
go build -o watch-node main.go
