package model

// ----------------------------------------------------------------
// [RenG Data Protocol v1.0]
// ----------------------------------------------------------------

// Choice: 선택지 버튼 정보
type Choice struct {
	Text   string `json:"text"`   // 버튼 텍스트
	NextID string `json:"nextId"` // 이동할 씬 ID
}

// Dialogue: 대사 및 연출의 최소 단위 (카드 1장)
type Dialogue struct {
	Type  string `json:"type"`
	Actor string `json:"actor"`
	Text  string `json:"text"`

	// 연출 리소스
	Background string `json:"background"`
	BGM        string `json:"bgm"`
	SFX        string `json:"sfx"`

	// ★ [신규] 캐릭터 스탠딩 위치별 파일명
	CharLeft   string `json:"charLeft"`   // 왼쪽
	CharCenter string `json:"charCenter"` // 중앙
	CharRight  string `json:"charRight"`  // 오른쪽

	// 로직
	Condition string   `json:"condition"`
	Choices   []Choice `json:"choices"`
	Video     string   `json:"video"` // (참고: 지난번 에디터 코드에서 video 필드를 따로 쓰는 걸로 구현해서 추가)
}

// Scene: 씬 (대사들의 묶음)
type Scene struct {
	ID        string     `json:"id"`
	Dialogues []Dialogue `json:"dialogues"`
}

// UIConfig: UI 커스터마이징 설정
type UIConfig struct {
	BoxColor   string  `json:"boxColor"`   // Hex Code
	BoxOpacity float64 `json:"boxOpacity"` // 0.0 ~ 1.0
	FontSize   int     `json:"fontSize"`
}

// GameData: 최종 저장 파일 (story.json) 구조
type GameData struct {
	Version   int                    `json:"version"`   // 데이터 버전
	Title     string                 `json:"title"`     // 게임 제목
	Variables map[string]interface{} `json:"variables"` // 전역 변수
	UI        UIConfig               `json:"ui"`        // UI 설정
	Scenes    []Scene                `json:"scenes"`    // 시나리오 데이터
}
