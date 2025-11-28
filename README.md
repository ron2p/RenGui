# 🌸 RenGui

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![Engine](https://img.shields.io/badge/Engine-Ebitengine_v2-red?style=for-the-badge&logo=nintendo-switch)
![Framework](https://img.shields.io/badge/Editor-Wails-red?style=for-the-badge&logo=wails)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

**The Modern, High-Performance Visual Novel Engine written in Go.** 가볍고 강력한 차세대 비주얼 노벨 엔진 & 에디터

[Features](#-key-features) • [Getting Started](#-getting-started) • [Architecture](#-architecture) • [Roadmap](#-roadmap)

</div>

---

## 📖 Introduction

**RenGui** is a lightweight visual novel engine designed to replace Python-based legacy engines. It separates the **Editor (Wails)** and the **Runtime (Ebitengine)**, sharing a standardized JSON data structure.

**RenGui**는 기존의 무거운 스크립트 기반 엔진을 대체하기 위해 탄생했습니다. 코딩 없이 직관적인 GUI 에디터로 시나리오를 작성하고, Go 언어의 강력한 성능으로 어디서든 실행되는 게임을 만드세요.

---

## ✨ Key Features

| Feature | Description |
| :--- | :--- |
| ⚡ **Blazing Fast** | Built with **Go** and **Ebitengine**. Compiles to a single native binary. No heavy runtime required. |
| 🌲 **Tree-based Editor** | Visual node/tree editor using **Wails**. Drag & Drop scenarios, branches, and media. No scripting! |
| 🎬 **Cinematic** | Native support for **VP8/IVF Video Backgrounds**. Create dynamic scenes with moving backgrounds. |
| 🔊 **Rich Media** | Full support for **BGM (Looping)**, **SFX**, and **Character Sprites** (Tachie) positioning. |
| 🌐 **Web Ready** | Designed with **WebAssembly (WASM)** in mind. Run your visual novel directly in the browser. |

---

## 📸 Screenshots

### 🎨 The Editor (Wails)
> Modern Dark UI inspired by Gemini. Manage dialogues, branches, and assets visually.

<img src="docs/images/editor_preview.png" alt="Editor Screenshot" width="800">

### 🎮 The Player (Ebitengine)
> High-performance playback with video backgrounds and character sprites.

<img src="docs/images/player_preview.png" alt="Player Screenshot" width="800">

---

## 🚀 Getting Started

### Prerequisites
* **Go** (1.21 or higher)
* **Node.js & npm** (For Editor frontend)
* **Wails CLI** (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### 1. Clone the repository
```bash
git clone [https://github.com/ron2p/RenGui.git](https://github.com/YOUR_GITHUB_ID/RenGui.git)
cd RenGui
```

### 2. Run the Editor
```bash
cd cmd/editor
go mod tidy
wails dev
```

### 3. Run the Player
```bash
# Open a new terminal from the project root
go mod tidy
go run ./cmd/player
```

---

## 📂 Architecture
RenGui follows a Monorepo structure to keep the Editor and Engine in sync.
