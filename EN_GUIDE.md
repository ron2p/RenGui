# 📘 RenGui User Guide

RenGui is a next-generation engine that allows you to create visual novels with simple clicks, requiring no coding.   
This document explains how to create your own game using the RenGui Editor.

---

## 🗂️ 1. Preparation (Folder Structure)
Before creating a game, you must place the image and music files you intend to use into specific folders.

* `assets/images`: Place background images (`.png`, `.jpg`) and video backgrounds (`.ivf`) here.
* `assets/sprites`: Place character standing images (`.png`) here (Background must be transparent).
* `assets/sounds`: Place background music (`.mp3`, `.wav`) and sound effects here.
* `assets/fonts`: Place the font file (`font.ttf`) to be used in the game here.

> [!WARNING]
> Please use **English** for file names whenever possible.
> (e.g., `bg_school.png`, `char_aris.png`)

---

## 🎬 2. Editor Usage
When you launch the editor (`wails dev`), the following screen will appear.

### 2.1. Tab Menu
There are two tabs at the top of the sidebar.

* **🎬 Action**: Where you add and edit scenario cards.
* **⚙️ Config**: Where you set the game title, resolution, and design.

### 2.2. Adding Cards
Click the button on the left to add an action to the timeline.

* **💬 Add Dialogue**: Creates a standard scene where a character speaks.
* **🎬 Video Effect**: Creates a cutscene where only a video plays without dialogue.
* **🌿 Branch**: Creates a moment where the player must choose an option (A/B).

### 2.3. Editing Cards
Click the **pencil icon** on an added card to open the detailed edit window.

| Item | Description |
| :--- | :--- |
| **Actor** | The name of the speaker (e.g., Teacher, Aris). |
| **Dialogue** | The text that will actually be displayed. |
| **Character Placement** | Select character images to appear on the **Left**, **Center**, or **Right**. |
| **Background Image** | Select the background to display behind characters. Selecting an `.ivf` file creates a **moving background**. |
| **Audio** | **BGM** plays in a continuous loop, while **SFX** plays once. |
| **Condition** | Displays this dialogue only when specific variable conditions are met (e.g., `love >= 100`). |

---

## 🎥 3. Using Videos (.ivf)
RenGui uses the **VP8/IVF** format for lightweight performance.
Directly inserting `.mp4` files will not work.

**How to convert using FFmpeg:**
```bash
ffmpeg -i my_video.mp4 -c:v libvpx -f ivf -an assets/images/my_video.ivf
```
> [!TIP]
> Place the converted .ivf file in the assets/images folder for automatic detection in the editor.
> It is recommended to remove audio (-an) and play BGM separately.

---

## 4. Game Customization (In Development)
You can change the game's appearance in the [Config] tab.   
• Resolution: Default is 1280 x 720. Can be changed to FHD (1920 x 1080).   
• UI Design: Freely change the dialog box color, opacity, and text color.   
• Changes are reflected in story.json only after clicking the [Save Project] button.   

---
## 🚀 5. Running and Deploying the Game
Test Play.  
After saving in the editor, run the player to verify.  

```bash
go run ./cmd/player
```

### Build (Deployment)
To distribute the game, you need to compile it into a single file.   

```bash
# Build for Windows (.exe)
cd cmd/player
go build -o MyGame.exe -ldflags "-H=windowsgui"
```

> [!IMPORTANT]
> When sharing the game, you must send the generated MyGame.exe file along with the assets folder and the story.json file.