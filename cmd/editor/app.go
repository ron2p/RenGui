package main

import (
	"context"
	"encoding/json"
	"os"

	"rengui/pkg/model" // ★ 중요: go.mod에 replace 설정이 되어 있어야 인식됨
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ---------------------------------------------------
// 백엔드 함수들 (Javascript에서 호출)
// ---------------------------------------------------

// SaveStory: 작성된 데이터를 받아서 'story.json' 파일로 저장
func (a *App) SaveStory(jsonStr string) string {
	var data model.GameData

	// 1. JSON 파싱 확인
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "Error: JSON 포맷 오류 - " + err.Error()
	}

	// 2. 파일 쓰기용 변환 (들여쓰기 적용)
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "Error: 데이터 변환 실패"
	}

	// 3. 프로젝트 루트의 story.json에 저장 (상위 폴더 2번 이동)
	// cmd/editor -> cmd -> RenGui(Root)
	err = os.WriteFile("../../story.json", bytes, 0644)
	if err != nil {
		return "저장 실패: " + err.Error()
	}

	return "저장 완료! (story.json)"
}

// LoadStory: story.json 파일을 읽어서 반환
func (a *App) LoadStory() string {
	bytes, err := os.ReadFile("../../story.json")
	if err != nil {
		return "" // 파일이 없으면 빈 문자열 반환
	}
	return string(bytes)
}

// GetImageList: assets/images 폴더의 파일 목록 반환
func (a *App) GetImageList() []string {
	return listFiles("../../assets/images")
}

// GetSoundList: assets/sounds 폴더의 파일 목록 반환
func (a *App) GetSoundList() []string {
	return listFiles("../../assets/sounds")
}

func (a *App) GetSpriteList() []string {
	return listFiles("../../assets/sprites")
}

// ---------------------------------------------------
// 내부 유틸리티 함수
// ---------------------------------------------------

// 폴더 내 파일 이름들만 리스트로 반환
func listFiles(dir string) []string {
	// 1. files를 nil이 아니라 빈 슬라이스로 초기화
	files := []string{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return files // 에러 나도 빈 배열 반환
	}

	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	return files
}
