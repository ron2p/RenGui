package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rengui/pkg/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/vp8"
)

// -------------------------------------------------------------------------
// [비디오 플레이어]
// -------------------------------------------------------------------------
type VideoPlayer struct {
	file       *os.File
	decoder    *vp8.Decoder
	currentImg *ebiten.Image
	timeBase   time.Duration
	lastFrame  time.Time
	isPlaying  bool
}

func NewVideoPlayer(filename string) (*VideoPlayer, error) {
	path := filepath.Join("assets", "images", filename)
	f, err := os.Open(path)
	if err != nil {
		path = filepath.Join("..", "..", "assets", "images", filename)
		f, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	}

	header := make([]byte, 32)
	if _, err := io.ReadFull(f, header); err != nil {
		return nil, err
	}

	rate := binary.LittleEndian.Uint32(header[16:20])
	scale := binary.LittleEndian.Uint32(header[20:24])
	if scale == 0 {
		scale = 1
	}
	fps := float64(rate) / float64(scale)
	if fps == 0 {
		fps = 30
	}

	return &VideoPlayer{
		file:      f,
		decoder:   vp8.NewDecoder(),
		timeBase:  time.Duration(float64(time.Second) / fps),
		lastFrame: time.Now(),
		isPlaying: true,
	}, nil
}

func (v *VideoPlayer) Update() error {
	if !v.isPlaying {
		return nil
	}
	if time.Since(v.lastFrame) < v.timeBase {
		return nil
	}
	v.lastFrame = time.Now()

	fh := make([]byte, 12)
	if _, err := io.ReadFull(v.file, fh); err != nil {
		v.isPlaying = false
		return nil
	}

	frameSize := binary.LittleEndian.Uint32(fh[:4])
	frameData := make([]byte, frameSize)
	if _, err := io.ReadFull(v.file, frameData); err != nil {
		return err
	}

	v.decoder.Init(bytes.NewReader(frameData), int(frameSize))
	img, err := v.decoder.DecodeFrame()
	if err == nil {
		v.currentImg = ebiten.NewImageFromImage(img)
	}
	return nil
}

func (v *VideoPlayer) Close() {
	if v.file != nil {
		v.file.Close()
	}
}

// -------------------------------------------------------------------------
// [오디오 매니저]
// -------------------------------------------------------------------------
type AudioManager struct {
	ctx    *audio.Context
	bgm    *audio.Player
	curBGM string
}

func (am *AudioManager) PlayBGM(filename string) {
	if filename == "" || am.curBGM == filename {
		return
	}
	if am.bgm != nil {
		am.bgm.Close()
	}

	path := filepath.Join("assets", "sounds", filename)
	f, err := os.Open(path)
	if err != nil {
		path = filepath.Join("..", "..", "assets", "sounds", filename)
		f, err = os.Open(path)
		if err != nil {
			return
		}
	}

	var s io.ReadSeeker
	if strings.HasSuffix(filename, ".mp3") {
		d, _ := mp3.Decode(am.ctx, f)
		s = audio.NewInfiniteLoop(d, d.Length())
	} else {
		d, _ := wav.Decode(am.ctx, f)
		s = audio.NewInfiniteLoop(d, d.Length())
	}

	am.bgm, _ = am.ctx.NewPlayer(s)
	am.bgm.Play()
	am.curBGM = filename
}

// -------------------------------------------------------------------------
// [메인 게임 엔진]
// -------------------------------------------------------------------------
type Game struct {
	Data     model.GameData
	SceneMap map[string]int
	CurScene int
	CurLine  int

	Font  font.Face
	Audio *AudioManager
	Video *VideoPlayer
	BgImg *ebiten.Image

	ImageCache  map[string]*ebiten.Image
	SpriteCache map[string]*ebiten.Image
}

func NewGame() *Game {
	fontPath := filepath.Join("assets", "fonts", "font.ttf")
	fontDat, err := os.ReadFile(fontPath)
	if err != nil {
		fontDat, _ = os.ReadFile(filepath.Join("..", "..", "assets", "fonts", "font.ttf"))
	}

	var face font.Face
	if len(fontDat) > 0 {
		tt, _ := opentype.Parse(fontDat)
		// 폰트 크기도 설정값으로 가져오면 좋지만, 여기서는 기본값 24 사용 (로드 시점 문제)
		face, _ = opentype.NewFace(tt, &opentype.FaceOptions{Size: 24, DPI: 72, Hinting: font.HintingFull})
	}

	jsonPath := "story.json"
	jsonDat, err := os.ReadFile(jsonPath)
	if err != nil {
		jsonDat, _ = os.ReadFile(filepath.Join("..", "..", "story.json"))
	}

	var gData model.GameData
	json.Unmarshal(jsonDat, &gData)

	sMap := make(map[string]int)
	for i, s := range gData.Scenes {
		sMap[s.ID] = i
	}

	g := &Game{
		Data: gData, SceneMap: sMap, Font: face,
		Audio:       &AudioManager{ctx: audio.NewContext(44100)},
		ImageCache:  make(map[string]*ebiten.Image),
		SpriteCache: make(map[string]*ebiten.Image),
	}

	if len(g.Data.Scenes) > 0 {
		g.LoadState()
	}
	return g
}

func (g *Game) LoadState() {
	if len(g.Data.Scenes) == 0 || g.CurScene >= len(g.Data.Scenes) {
		return
	}
	scene := g.Data.Scenes[g.CurScene]

	for {
		if g.CurLine >= len(scene.Dialogues) {
			return
		}
		diag := scene.Dialogues[g.CurLine]

		if diag.Condition != "" {
			parts := strings.Split(diag.Condition, " ")
			if len(parts) == 3 {
				key, op, valStr := parts[0], parts[1], parts[2]
				var curVal float64
				if val, ok := g.Data.Variables[key]; ok {
					switch v := val.(type) {
					case float64:
						curVal = v
					case int:
						curVal = float64(v)
					}
				}
				targetVal, _ := strconv.ParseFloat(valStr, 64)

				pass := false
				if op == ">=" {
					pass = curVal >= targetVal
				}
				if op == "==" {
					pass = curVal == targetVal
				}
				if op == "<" {
					pass = curVal < targetVal
				}

				if !pass {
					g.CurLine++
					continue
				}
			}
		}

		g.Audio.PlayBGM(diag.BGM)

		if strings.HasSuffix(diag.Background, ".ivf") {
			if g.Video == nil {
				g.Video, _ = NewVideoPlayer(diag.Background)
			}
			g.BgImg = nil
		} else if diag.Background != "" {
			if g.Video != nil {
				g.Video.Close()
				g.Video = nil
			}
			g.BgImg = g.loadImage(diag.Background)
		}
		break
	}
}

func (g *Game) Update() error {
	if g.Video != nil {
		g.Video.Update()
	}
	if len(g.Data.Scenes) == 0 || g.CurScene >= len(g.Data.Scenes) {
		return nil
	}
	scene := g.Data.Scenes[g.CurScene]
	if g.CurLine >= len(scene.Dialogues) {
		return nil
	}

	diag := scene.Dialogues[g.CurLine]

	// 선택지
	if len(diag.Choices) > 0 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			for i, ch := range diag.Choices {
				y := 300 + i*70
				if mx > 340 && mx < 940 && my > y && my < y+60 {
					if nextIdx, ok := g.SceneMap[ch.NextID]; ok {
						g.CurScene = nextIdx
						g.CurLine = 0
						g.LoadState()
					}
				}
			}
		}
		return nil
	}

	// 대사 넘기기
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.CurLine++
		if g.CurLine < len(scene.Dialogues) {
			g.LoadState()
		}
	}
	return nil
}

// 리소스 로더
func (g *Game) loadImage(name string) *ebiten.Image {
	if img, ok := g.ImageCache[name]; ok {
		return img
	}
	path := filepath.Join("assets", "images", name)
	f, err := os.Open(path)
	if err != nil {
		path = filepath.Join("..", "..", "assets", "images", name)
		f, _ = os.Open(path)
	}
	defer f.Close()
	i, _, _ := image.Decode(f)
	eImg := ebiten.NewImageFromImage(i)
	g.ImageCache[name] = eImg
	return eImg
}

func (g *Game) loadSprite(name string) *ebiten.Image {
	if img, ok := g.SpriteCache[name]; ok {
		return img
	}
	path := filepath.Join("assets", "sprites", name)
	f, err := os.Open(path)
	if err != nil {
		path = filepath.Join("..", "..", "assets", "sprites", name)
		f, err = os.Open(path)
		if err != nil {
			return nil
		}
	}
	defer f.Close()
	i, _, _ := image.Decode(f)
	if err != nil {
		return nil
	}
	eImg := ebiten.NewImageFromImage(i)
	g.SpriteCache[name] = eImg
	return eImg
}

// 색상 변환 헬퍼 (Hex -> RGBA)
func parseHexColor(s string, opacity float64) color.RGBA {
	c := color.RGBA{0, 0, 0, 255}
	if len(s) != 7 || s[0] != '#' {
		return c
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		return 0
	}
	c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
	c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
	c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	c.A = uint8(opacity * 255)
	return c
}

func (g *Game) drawCharacter(screen *ebiten.Image, filename string, pos string) {
	if filename == "" {
		return
	}
	img := g.loadSprite(filename)
	if img == nil {
		return
	}

	w, h := img.Size()

	// 설정된 해상도 사용 (없으면 기본값)
	sysW := float64(g.Data.System.ScreenWidth)
	if sysW == 0 {
		sysW = 1280.0
	}
	sysH := float64(g.Data.System.ScreenHeight)
	if sysH == 0 {
		sysH = 720.0
	}

	// 대화창 높이 고려 (가려지지 않게)
	boxH := float64(g.Data.UI.BoxHeight)
	if boxH == 0 {
		boxH = 200.0
	}

	op := &ebiten.DrawImageOptions{}

	// 크기 조정 (화면 높이 80%)
	scale := (sysH * 0.8) / float64(h)
	op.GeoM.Scale(scale, scale)
	scaledW := float64(w) * scale
	scaledH := float64(h) * scale

	var x float64
	switch pos {
	case "left":
		x = sysW * 0.15
	case "center":
		x = (sysW - scaledW) / 2
	case "right":
		x = sysW*0.85 - scaledW
	}

	y := sysH - scaledH - (boxH - 20) // 박스 위로 살짝 겹치게
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func (g *Game) Draw(screen *ebiten.Image) {
	if len(g.Data.Scenes) == 0 {
		if g.Font != nil {
			text.Draw(screen, "No Data", g.Font, 50, 50, color.White)
		}
		return
	}

	// 설정값 가져오기
	sysW := float64(g.Data.System.ScreenWidth)
	if sysW == 0 {
		sysW = 1280.0
	}
	sysH := float64(g.Data.System.ScreenHeight)
	if sysH == 0 {
		sysH = 720.0
	}

	uiH := float64(g.Data.UI.BoxHeight)
	if uiH == 0 {
		uiH = 200.0
	}

	// [Layer 1] 배경
	if g.Video != nil && g.Video.currentImg != nil {
		op := &ebiten.DrawImageOptions{}
		w, h := g.Video.currentImg.Size()
		op.GeoM.Scale(sysW/float64(w), sysH/float64(h))
		screen.DrawImage(g.Video.currentImg, op)
	} else if g.BgImg != nil {
		op := &ebiten.DrawImageOptions{}
		w, h := g.BgImg.Size()
		op.GeoM.Scale(sysW/float64(w), sysH/float64(h))
		screen.DrawImage(g.BgImg, op)
	} else {
		screen.Fill(color.Black)
	}

	if g.CurScene >= len(g.Data.Scenes) {
		return
	}
	scene := g.Data.Scenes[g.CurScene]
	if len(scene.Dialogues) == 0 || g.CurLine >= len(scene.Dialogues) {
		return
	}

	diag := scene.Dialogues[g.CurLine]

	// [Layer 2] 캐릭터
	g.drawCharacter(screen, diag.CharCenter, "center")
	g.drawCharacter(screen, diag.CharLeft, "left")
	g.drawCharacter(screen, diag.CharRight, "right")

	// [Layer 3] UI
	if len(diag.Choices) > 0 {
		for i, ch := range diag.Choices {
			y := 300 + i*70
			btn := ebiten.NewImage(600, 60)
			btn.Fill(color.RGBA{255, 255, 255, 200})
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate((sysW-600)/2, float64(y)) // 중앙 정렬
			screen.DrawImage(btn, op)
			if g.Font != nil {
				text.Draw(screen, ch.Text, g.Font, int((sysW-600)/2)+50, y+40, color.Black)
			}
		}
	} else {
		// 대화창 커스텀 적용
		boxColor := parseHexColor(g.Data.UI.BoxColor, g.Data.UI.BoxOpacity)
		textColor := parseHexColor(g.Data.UI.TextColor, 1.0)

		box := ebiten.NewImage(int(sysW), int(uiH))
		box.Fill(boxColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, sysH-uiH)
		screen.DrawImage(box, op)

		if g.Font != nil {
			text.Draw(screen, diag.Actor, g.Font, 50, int(sysH-uiH)+50, color.RGBA{255, 255, 0, 255})
			text.Draw(screen, diag.Text, g.Font, 50, int(sysH-uiH)+100, textColor)
		}
	}
}

// 화면 해상도 설정 (커스텀 값 반영)
func (g *Game) Layout(w, h int) (int, int) {
	sw := g.Data.System.ScreenWidth
	sh := g.Data.System.ScreenHeight
	if sw == 0 {
		sw = 1280
	}
	if sh == 0 {
		sh = 720
	}
	return sw, sh
}

func main() {
	// JSON을 먼저 읽어서 윈도우 크기를 설정해야 함
	jsonPath := "story.json"
	jsonDat, err := os.ReadFile(jsonPath)
	if err != nil {
		jsonDat, _ = os.ReadFile(filepath.Join("..", "..", "story.json"))
	}

	var gData model.GameData
	json.Unmarshal(jsonDat, &gData)

	w, h := gData.System.ScreenWidth, gData.System.ScreenHeight
	if w == 0 {
		w = 1280
	}
	if h == 0 {
		h = 720
	}
	title := gData.System.Title
	if title == "" {
		title = "RenGui Game"
	}

	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle(title)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
