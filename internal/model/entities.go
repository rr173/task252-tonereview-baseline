package model

// 田野批次状态机：整理中 → 待复核 → 已发布 → 封存（封存为终态）。
const (
	BatchCollecting = "collecting" // 整理中
	BatchReviewing  = "reviewing"  // 待复核
	BatchPublished  = "published"  // 已发布
	BatchSealed     = "sealed"     // 封存
)

// 发音片段状态机：待校验 → 可用 | 噪声 | 排除（排除可恢复为可用）。
const (
	SegPending  = "pending"  // 待校验
	SegUsable   = "usable"   // 可用
	SegNoise    = "noise"    // 噪声
	SegExcluded = "excluded" // 排除（研究者否决）
)

// 声调对立状态机：候选 → 确认 | 否决 | 证据不足（确认/否决为终态）。
const (
	OppCandidate    = "candidate"    // 候选
	OppInsufficient = "insufficient" // 证据不足
	OppConfirmed    = "confirmed"    // 确认
	OppRejected     = "rejected"     // 否决
)

// 分析版本状态机：草稿 → 共享 → 冻结（冻结可被新版本替代）。
const (
	VerDraft      = "draft"      // 草稿
	VerShared     = "shared"     // 共享
	VerFrozen     = "frozen"     // 冻结
	VerSuperseded = "superseded" // 替代
)

// 调型分类（基于归一化半音轮廓斜率）。
const (
	ToneLevel   = "level"   // 平
	ToneRising  = "rising"  // 升
	ToneFalling = "falling" // 降
	ToneContour = "contour" // 曲折
	ToneUnknown = "unknown" // 未知（缺基线或无样本）
)

// Speaker 说话人：提供基线归一化所需的个体声学参考。
type Speaker struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Dialect     string  `json:"dialect"`
	Gender      string  `json:"gender"`
	BirthYear   int     `json:"birth_year"`
	BaselineLog float64 `json:"baseline_log"`
	HasBaseline bool    `json:"has_baseline"`
	CreatedAt   int64   `json:"created_at"`
}

// FieldBatch 田野批次：一组同源录音词条的采集与复核单元。
type FieldBatch struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// Segment 发音片段：单个词条在某说话人、某音段下的录音与基频证据。
type Segment struct {
	ID          string `json:"id"`
	BatchID     string `json:"batch_id"`
	LexicalItem string `json:"lexical_item"`
	PhoneticSeg string `json:"phonetic_seg"`
	SpeakerID   string `json:"speaker_id"`
	AudioFP     string `json:"audio_fp"`
	Status      string `json:"status"`
	DurationMs  int64  `json:"duration_ms"`
	RecordedAt  int64  `json:"recorded_at"`
	ToneType    string `json:"tone_type"`
	CreatedAt   int64  `json:"created_at"`
}

// F0Sample 基频采样：时间（毫秒）与频率（赫兹），时间必须严格递增。
type F0Sample struct {
	ID        string  `json:"id"`
	SegmentID string  `json:"segment_id"`
	TMs       int64   `json:"t_ms"`
	F0Hz      float64 `json:"f0_hz"`
	Reliable  bool    `json:"reliable"`
}

// ToneOpposition 声调对立：同一音段上的两个词条是否构成最小对立。
type ToneOpposition struct {
	ID              string  `json:"id"`
	BatchID         string  `json:"batch_id"`
	LexicalA        string  `json:"lexical_a"`
	PhoneticSeg     string  `json:"phonetic_seg"`
	LexicalB        string  `json:"lexical_b"`
	Status          string  `json:"status"`
	OppositionScore float64 `json:"opposition_score"`
	DecisionReason  string  `json:"decision_reason"`
	DecidedAt       int64   `json:"decided_at"`
	CreatedAt       int64   `json:"created_at"`
}

// Evidence 证据簇条目：某片段在某一侧（a/b）的归一化轮廓快照。
type Evidence struct {
	ID           string `json:"id"`
	OppositionID string `json:"opposition_id"`
	SegmentID    string `json:"segment_id"`
	Side         string `json:"side"`
	Normalized   string `json:"normalized"`
	ToneType     string `json:"tone_type"`
	CreatedAt    int64  `json:"created_at"`
}

// AnalysisVersion 分析版本：冻结为不可变快照的声调证据结论。
type AnalysisVersion struct {
	ID        string `json:"id"`
	BatchID   string `json:"batch_id"`
	Version   int    `json:"version"`
	Status    string `json:"status"`
	Snapshot  string `json:"snapshot"`
	Note      string `json:"note"`
	CreatedAt int64  `json:"created_at"`
}
