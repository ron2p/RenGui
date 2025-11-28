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

// ... (VideoPlayer, AudioManager 구조체는 기존과 동일 - 생략) ...
// (이전 코드의 NewVideoPlayer, Update, Close, NewAudioManager, PlayBGM 그대로 유지하세요)

// =================================================================
// 1. 비디오 플레이어 (기존 동일)
// =================================================================
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

// =================================================================
// 2. 오디오 매니저 (기존 동일)
// =================================================================
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

// =================================================================
// 3. 메인 게임 엔진 (캐릭터 그리기 추가됨)
// =================================================================
type Game struct {
	Data     model.GameData
	SceneMap map[string]int
	CurScene int
	CurLine  int

	Font  font.Face
	Audio *AudioManager
	Video *VideoPlayer
	BgImg *ebiten.Image

	// 이미지 캐시 (배경용, 캐릭터용)
	ImageCache  map[string]*ebiten.Image
	SpriteCache map[string]*ebiten.Image
}

func NewGame() *Game {
	// (폰트, JSON 로드 부분 기존과 동일 - 생략)
	fontPath := filepath.Join("assets", "fonts", "font.ttf")
	fontDat, err := os.ReadFile(fontPath)
	if err != nil {
		fontDat, _ = os.ReadFile(filepath.Join("..", "..", "assets", "fonts", "font.ttf"))
	}
	var face font.Face
	if len(fontDat) > 0 {
		tt, _ := opentype.Parse(fontDat)
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
		SpriteCache: make(map[string]*ebiten.Image), // 초기화
	}

	if len(g.Data.Scenes) > 0 {
		g.LoadState()
	}
	return g
}

// LoadState, Update 함수는 기존과 동일하므로 생략합니다. (복붙해주세요)
// ...
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
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.CurLine++
		if g.CurLine < len(scene.Dialogues) {
			g.LoadState()
		}
	}
	return nil
}

// ...

// [신규] 배경 이미지 로더 (assets/images)
func (g *Game) loadImage(name string) *ebiten.Image {
	if img, ok := g.ImageCache[name]; ok {
		return img
	}
	// ... (경로 찾기 로직 기존 동일) ...
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

// ★ [신규] 캐릭터 이미지 로더 (assets/sprites)
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
		} // 파일 없으면 nil 반환
	}
	defer f.Close()

	i, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	eImg := ebiten.NewImageFromImage(i)
	g.SpriteCache[name] = eImg
	return eImg
}

// ★ [신규] 캐릭터 그리기 헬퍼 함수
func (g *Game) drawCharacter(screen *ebiten.Image, filename string, pos string) {
	if filename == "" {
		return
	}
	img := g.loadSprite(filename)
	if img == nil {
		return
	}

	w, h := img.Size()
	screenW, screenH := 1280.0, 720.0

	op := &ebiten.DrawImageOptions{}

	// 1. 크기 조정 (화면 높이의 80% 정도로 맞춤 - 조절 가능)
	scale := (screenH * 0.8) / float64(h)
	op.GeoM.Scale(scale, scale)
	scaledW := float64(w) * scale
	scaledH := float64(h) * scale

	// 2. 위치 계산
	var x float64
	switch pos {
	case "left":
		x = screenW * 0.15 // 왼쪽 15% 지점
	case "center":
		x = (screenW - scaledW) / 2 // 중앙
	case "right":
		x = screenW*0.85 - scaledW // 오른쪽 85% 지점 기준 정렬
	}

	// Y좌표: 바닥에서 대화창 높이(약 200px)만큼 띄움
	y := screenH - scaledH - 180

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

	// [Layer 1] 배경 그리기
	if g.Video != nil && g.Video.currentImg != nil {
		op := &ebiten.DrawImageOptions{}
		w, h := g.Video.currentImg.Size()
		op.GeoM.Scale(1280.0/float64(w), 720.0/float64(h))
		screen.DrawImage(g.Video.currentImg, op)
	} else if g.BgImg != nil {
		op := &ebiten.DrawImageOptions{}
		w, h := g.BgImg.Size()
		op.GeoM.Scale(1280.0/float64(w), 720.0/float64(h))
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

	// ★ [Layer 2] 캐릭터 그리기 (배경 위, UI 아래)
	// 순서대로 그립니다 (뒤에 있는게 앞으로 옴)
	g.drawCharacter(screen, diag.CharCenter, "center") // 중앙을 먼저 그리고
	g.drawCharacter(screen, diag.CharLeft, "left")     // 좌우를 그 위에 (취향차이)
	g.drawCharacter(screen, diag.CharRight, "right")

	// [Layer 3] UI 및 텍스트 (기존 동일)
	if len(diag.Choices) > 0 {
		// ... (선택지 그리기)
		for i, ch := range diag.Choices {
			y := 300 + i*70
			btn := ebiten.NewImage(600, 60)
			btn.Fill(color.RGBA{255, 255, 255, 200})
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(340, float64(y))
			screen.DrawImage(btn, op)
			if g.Font != nil {
				text.Draw(screen, ch.Text, g.Font, 400, y+40, color.Black)
			}
		}
	} else {
		// ... (대화창 그리기)
		box := ebiten.NewImage(1280, 200)
		box.Fill(color.RGBA{0, 0, 0, 150})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, 520)
		screen.DrawImage(box, op)

		if g.Font != nil {
			text.Draw(screen, diag.Actor, g.Font, 50, 560, color.RGBA{255, 255, 0, 255})
			text.Draw(screen, diag.Text, g.Font, 50, 610, color.White)
		}
	}
}

func (g *Game) Layout(w, h int) (int, int) { return 1280, 720 }

func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("RenGui Player")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
