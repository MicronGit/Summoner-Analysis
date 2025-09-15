package riot

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// レート制限管理
type RateLimiter struct {
	// 短期間制限: 20 requests per second
	shortTermTokens int
	shortTermLimit  int
	shortTermWindow time.Duration
	shortTermReset  time.Time

	// 長期間制限: 100 requests per 2 minutes
	longTermTokens int
	longTermLimit  int
	longTermWindow time.Duration
	longTermReset  time.Time

	mu sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	now := time.Now()
	return &RateLimiter{
		shortTermTokens: 20,
		shortTermLimit:  20,
		shortTermWindow: time.Second,
		shortTermReset:  now.Add(time.Second),

		longTermTokens: 100,
		longTermLimit:  100,
		longTermWindow: 2 * time.Minute,
		longTermReset:  now.Add(2 * time.Minute),
	}
}

// リクエスト許可を待機（ブロッキング）
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		// mutexをロックして状態をチェック
		rl.mu.Lock()

		now := time.Now()

		// トークンをリセット（時間窓が過ぎた場合）
		if now.After(rl.shortTermReset) {
			rl.shortTermTokens = rl.shortTermLimit
			rl.shortTermReset = now.Add(rl.shortTermWindow)
		}

		if now.After(rl.longTermReset) {
			rl.longTermTokens = rl.longTermLimit
			rl.longTermReset = now.Add(rl.longTermWindow)
		}

		// 利用可能なトークンがある場合
		if rl.shortTermTokens > 0 && rl.longTermTokens > 0 {
			// トークンを消費
			rl.shortTermTokens--
			rl.longTermTokens--
			rl.mu.Unlock() // ここでアンロック
			return nil
		}

		// 利用可能なトークンがない場合は待機時間を計算
		var waitTime time.Duration

		if rl.shortTermTokens <= 0 {
			waitTime = rl.shortTermReset.Sub(now)
		}

		if rl.longTermTokens <= 0 {
			longWaitTime := rl.longTermReset.Sub(now)
			if longWaitTime > waitTime {
				waitTime = longWaitTime
			}
		}

		// 待機時間が0以下の場合は少し待つ
		if waitTime <= 0 {
			waitTime = 100 * time.Millisecond
		}

		rl.mu.Unlock() // 待機前にアンロック

		fmt.Printf("レート制限に達しました。%v 待機中...\n", waitTime.Round(time.Second))

		// ノンブロッキングで待機（コンテキストキャンセル対応）
		select {
		case <-time.After(waitTime):
			// 待機完了、ループを継続して再チェック
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// 現在の制限状況を取得
func (rl *RateLimiter) GetStatus() (shortAvailable, longAvailable int, shortReset, longReset time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// 期限切れチェック
	if now.After(rl.shortTermReset) {
		rl.shortTermTokens = rl.shortTermLimit
		rl.shortTermReset = now.Add(rl.shortTermWindow)
	}

	if now.After(rl.longTermReset) {
		rl.longTermTokens = rl.longTermLimit
		rl.longTermReset = now.Add(rl.longTermWindow)
	}

	return rl.shortTermTokens, rl.longTermTokens, rl.shortTermReset, rl.longTermReset
}

// トークンの強制リセット（テスト用）
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rl.shortTermTokens = rl.shortTermLimit
	rl.shortTermReset = now.Add(rl.shortTermWindow)
	rl.longTermTokens = rl.longTermLimit
	rl.longTermReset = now.Add(rl.longTermWindow)
}
