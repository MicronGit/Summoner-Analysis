package riot

// League of Legends キューID定数
const (
	// ランク戦
	QueueSoloRanked = 420 // ソロ/デュオランク戦
	QueueFlexRanked = 440 // フレックスランク戦

	// ノーマル
	QueueNormalDraft = 400 // ノーマル（ドラフト）
	QueueNormalBlind = 430 // ノーマル（ブラインド）

	// その他
	QueueARAM = 450 // ARAM
	QueueURF  = 900 // URF（期間限定）
)

// キューID から キュー名への変換
var QueueIDToName = map[int]string{
	420: "ソロ/デュオランク",
	440: "フレックスランク",
	400: "ノーマル（ドラフト）",
	430: "ノーマル（ブラインド）",
	450: "ARAM",
	900: "URF",
}

// ランク戦かどうかを判定
func IsRankedQueue(queueID int) bool {
	return queueID == QueueSoloRanked || queueID == QueueFlexRanked
}

// ノーマル戦かどうかを判定
func IsNormalQueue(queueID int) bool {
	return queueID == QueueNormalDraft || queueID == QueueNormalBlind
}

// ARAMかどうかを判定
func IsARAMQueue(queueID int) bool {
	return queueID == QueueARAM
}
