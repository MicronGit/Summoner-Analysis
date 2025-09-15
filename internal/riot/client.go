package riot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	APIKey      string
	Region      string
	HTTPClient  *http.Client
	RateLimiter *RateLimiter
}

func NewClient(apiKey, region string) *Client {
	return &Client{
		APIKey: apiKey,
		Region: region,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30, // タイムアウトを長めに設定
		},
		RateLimiter: NewRateLimiter(),
	}
}

// レート制限対応のHTTPリクエスト実行（改善版）
func (c *Client) doRequestWithRateLimit(ctx context.Context, req *http.Request) (*http.Response, error) {
	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		// レート制限チェック
		if err := c.RateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("レート制限待機エラー: %w", err)
		}

		// リクエスト実行
		resp, err := c.HTTPClient.Do(req.WithContext(ctx))
		if err != nil {
			// ネットワークエラーの場合は短時間待機後にリトライ
			if attempt < maxRetries-1 {
				fmt.Printf("ネットワークエラー (試行 %d/%d): %v - 1秒後にリトライ\n",
					attempt+1, maxRetries, err)
				select {
				case <-time.After(time.Second):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			return nil, fmt.Errorf("HTTPリクエストエラー: %w", err)
		}

		// 成功またはクライアントエラー（4xx）の場合はそのまま返す
		if resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, nil
		}

		// 429 Too Many Requests の場合
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			var waitDuration time.Duration = 2 * time.Second // デフォルト待機時間

			if retryAfter != "" {
				if seconds, err := time.ParseDuration(retryAfter + "s"); err == nil {
					waitDuration = seconds
				}
			}

			fmt.Printf("429エラー: %v後に再試行 (試行 %d/%d)\n",
				waitDuration, attempt+1, maxRetries)

			resp.Body.Close() // レスポンスボディを閉じる

			if attempt < maxRetries-1 {
				select {
				case <-time.After(waitDuration):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}

			return nil, fmt.Errorf("レート制限エラー: 最大試行回数に到達")
		}

		// 5xxサーバーエラーの場合
		if resp.StatusCode >= 500 {
			resp.Body.Close()

			if attempt < maxRetries-1 {
				waitTime := time.Duration(attempt+1) * time.Second
				fmt.Printf("サーバーエラー %d: %v後に再試行 (試行 %d/%d)\n",
					resp.StatusCode, waitTime, attempt+1, maxRetries)

				select {
				case <-time.After(waitTime):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}

			return nil, fmt.Errorf("サーバーエラー: ステータスコード %d", resp.StatusCode)
		}

		// その他のエラー
		return resp, nil
	}

	return nil, fmt.Errorf("最大試行回数に到達")
}

// アカウント情報取得（レート制限対応）
func (c *Client) GetAccountByRiotID(gameName, tagLine string) (*Account, error) {
	ctx := context.Background()

	baseURL := fmt.Sprintf("https://%s.api.riotgames.com", c.Region)
	endpoint := fmt.Sprintf("/riot/account/v1/accounts/by-riot-id/%s/%s",
		url.PathEscape(gameName),
		url.PathEscape(tagLine))

	fullURL := baseURL + endpoint

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成エラー: %w", err)
	}

	req.Header.Add("X-Riot-Token", c.APIKey)
	req.Header.Add("Accept", "application/json")

	resp, err := c.doRequestWithRateLimit(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("APIリクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("APIエラー (status: %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取りエラー: %w", err)
	}

	var account Account
	if err := json.Unmarshal(body, &account); err != nil {
		return nil, fmt.Errorf("JSON解析エラー: %w", err)
	}

	return &account, nil
}

// マッチ詳細取得（レート制限対応）
func (c *Client) GetMatchDetailWithContext(ctx context.Context, matchID string) (*MatchDetail, error) {
	matchRegion := c.getMatchRegion()
	baseURL := fmt.Sprintf("https://%s.api.riotgames.com", matchRegion)
	endpoint := fmt.Sprintf("/lol/match/v5/matches/%s", url.PathEscape(matchID))

	fullURL := baseURL + endpoint

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成エラー: %w", err)
	}

	req.Header.Add("X-Riot-Token", c.APIKey)
	req.Header.Add("Accept", "application/json")

	resp, err := c.doRequestWithRateLimit(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("APIリクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("APIエラー (status: %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取りエラー: %w", err)
	}

	var matchDetail MatchDetail
	if err := json.Unmarshal(body, &matchDetail); err != nil {
		return nil, fmt.Errorf("JSON解析エラー: %w", err)
	}

	return &matchDetail, nil
}

// マッチ履歴取得
func (c *Client) GetRankedMatchHistoryWithContext(ctx context.Context, puuid string, count int) (MatchHistory, error) {
	if count <= 0 || count > 100 {
		count = 100
	}

	// リクエストURL作成
	matchRegion := c.getMatchRegion()
	baseURL := fmt.Sprintf("https://%s.api.riotgames.com", matchRegion)
	endpoint := fmt.Sprintf("/lol/match/v5/matches/by-puuid/%s/ids?start=0&count=%d", url.PathEscape(puuid), count)

	fullURL := baseURL + endpoint

	request, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成エラー: %w", err)
	}

	request.Header.Add("X-Riot-Token", c.APIKey)
	request.Header.Add("Accept", "application/json")

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("APIリクエストエラー: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("APIエラー (status: %d): %s", response.StatusCode, string(body))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス抽出エラー: %w", err)
	}

	var matchHistory MatchHistory
	if err := json.Unmarshal(body, &matchHistory); err != nil {
		return nil, fmt.Errorf("JSONパースエラー: %w", err)
	}

	return matchHistory, nil
}

// リージョンをマッチAPIのリージョンに変換
func (c *Client) getMatchRegion() string {
	switch c.Region {
	case "asia":
		return "asia"
	case "kr", "jp1":
		return "asia"
	case "na1", "br1", "la1", "la2":
		return "americas"
	case "euw1", "eun1", "tr1", "ru":
		return "europe"
	default:
		return "asia" // デフォルト
	}
}

// プレイヤーのランク戦分析データを取得（キャンセル対応）
func (c *Client) GetPlayerRankedAnalysisWithContext(ctx context.Context, account *Account, matchCount int) (*PlayerMatchSummary, error) {
	fmt.Printf("ランク戦マッチ履歴を取得中（最大%d試合）...\n", matchCount)

	// ランク戦マッチ履歴取得
	matchIDs, err := c.GetRankedMatchHistoryWithContext(ctx, account.PUUID, matchCount)
	if err != nil {
		return nil, fmt.Errorf("ランク戦マッチ履歴取得エラー: %w", err)
	}

	fmt.Printf("取得したランク戦マッチ数: %d\n", len(matchIDs))

	if len(matchIDs) == 0 {
		return &PlayerMatchSummary{
			Account:      *account,
			MatchHistory: []MatchDetail{},
			GeneratedAt:  time.Now(),
			TotalMatches: 0,
			MatchType:    "ranked",
		}, nil
	}

	// 各マッチの詳細を取得
	var matchDetails []MatchDetail
	startTime := time.Now()

	fmt.Printf("マッチ詳細取得開始...\n")

	for i, matchID := range matchIDs {
		// コンテキストキャンセルチェック
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("処理がキャンセルされました: %w", ctx.Err())
		default:
		}

		// 進捗表示
		if i%5 == 0 {
			elapsed := time.Since(startTime)
			if i > 0 {
				avgTime := elapsed / time.Duration(i)
				remaining := avgTime * time.Duration(len(matchIDs)-i)
				fmt.Printf("進捗: %d/%d (%.1f%%) - 経過: %v, 推定残り: %v\n",
					i, len(matchIDs), float64(i)/float64(len(matchIDs))*100,
					elapsed.Round(time.Second), remaining.Round(time.Second))
			}
		}

		detail, err := c.GetMatchDetailWithContext(ctx, matchID)
		if err != nil {
			fmt.Printf("⚠️  マッチ %s の取得に失敗: %v\n", matchID, err)
			continue
		}

		if IsRankedQueue(detail.Info.QueueID) {
			matchDetails = append(matchDetails, *detail)
		}
	}

	totalTime := time.Since(startTime)
	fmt.Printf("✅ マッチ詳細取得完了: %d試合を%vで処理\n",
		len(matchDetails), totalTime.Round(time.Second))

	return &PlayerMatchSummary{
		Account:      *account,
		MatchHistory: matchDetails,
		GeneratedAt:  time.Now(),
		TotalMatches: len(matchDetails),
		MatchType:    "ranked",
	}, nil
}

// プレイヤーのノーマル戦分析データを取得
func (c *Client) GetPlayerNormalAnalysisWithContext(ctx context.Context, account *Account, matchCount int) (*PlayerMatchSummary, error) {
	fmt.Printf("ノーマル戦マッチ履歴を取得中（最大%d試合）...\n", matchCount)

	matchIDs, err := c.GetRankedMatchHistoryWithContext(ctx, account.PUUID, matchCount)
	if err != nil {
		return nil, fmt.Errorf("マッチ履歴取得エラー: %w", err)
	}

	fmt.Printf("取得したマッチ数: %d\n", len(matchIDs))

	if len(matchIDs) == 0 {
		return &PlayerMatchSummary{
			Account:      *account,
			MatchHistory: []MatchDetail{},
			GeneratedAt:  time.Now(),
			TotalMatches: 0,
			MatchType:    "normal",
		}, nil
	}

	var matchDetails []MatchDetail
	startTime := time.Now()

	fmt.Printf("マッチ詳細取得開始...\n")

	for i, matchID := range matchIDs {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("処理がキャンセルされました: %w", ctx.Err())
		default:
		}

		if i%5 == 0 {
			elapsed := time.Since(startTime)
			if i > 0 {
				avgTime := elapsed / time.Duration(i)
				remaining := avgTime * time.Duration(len(matchIDs)-i)
				fmt.Printf("進捗: %d/%d (%.1f%%) - 経過: %v, 推定残り: %v\n",
					i, len(matchIDs), float64(i)/float64(len(matchIDs))*100,
					elapsed.Round(time.Second), remaining.Round(time.Second))
			}
		}

		detail, err := c.GetMatchDetailWithContext(ctx, matchID)
		if err != nil {
			fmt.Printf("⚠️  マッチ %s の取得に失敗: %v\n", matchID, err)
			continue
		}

		if IsNormalQueue(detail.Info.QueueID) {
			matchDetails = append(matchDetails, *detail)
		}
	}

	totalTime := time.Since(startTime)
	fmt.Printf("✅ マッチ詳細取得完了: %d試合を%vで処理\n",
		len(matchDetails), totalTime.Round(time.Second))

	return &PlayerMatchSummary{
		Account:      *account,
		MatchHistory: matchDetails,
		GeneratedAt:  time.Now(),
		TotalMatches: len(matchDetails),
		MatchType:    "normal",
	}, nil
}

// プレイヤーのARAM分析データを取得
func (c *Client) GetPlayerARAMAnalysisWithContext(ctx context.Context, account *Account, matchCount int) (*PlayerMatchSummary, error) {
	fmt.Printf("ARAMマッチ履歴を取得中（最大%d試合）...\n", matchCount)

	matchIDs, err := c.GetRankedMatchHistoryWithContext(ctx, account.PUUID, matchCount)
	if err != nil {
		return nil, fmt.Errorf("マッチ履歴取得エラー: %w", err)
	}

	fmt.Printf("取得したマッチ数: %d\n", len(matchIDs))

	if len(matchIDs) == 0 {
		return &PlayerMatchSummary{
			Account:      *account,
			MatchHistory: []MatchDetail{},
			GeneratedAt:  time.Now(),
			TotalMatches: 0,
			MatchType:    "aram",
		}, nil
	}

	var matchDetails []MatchDetail
	startTime := time.Now()

	fmt.Printf("マッチ詳細取得開始...\n")

	for i, matchID := range matchIDs {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("処理がキャンセルされました: %w", ctx.Err())
		default:
		}

		if i%5 == 0 {
			elapsed := time.Since(startTime)
			if i > 0 {
				avgTime := elapsed / time.Duration(i)
				remaining := avgTime * time.Duration(len(matchIDs)-i)
				fmt.Printf("進捗: %d/%d (%.1f%%) - 経過: %v, 推定残り: %v\n",
					i, len(matchIDs), float64(i)/float64(len(matchIDs))*100,
					elapsed.Round(time.Second), remaining.Round(time.Second))
			}
		}

		detail, err := c.GetMatchDetailWithContext(ctx, matchID)
		if err != nil {
			fmt.Printf("⚠️  マッチ %s の取得に失敗: %v\n", matchID, err)
			continue
		}

		if IsARAMQueue(detail.Info.QueueID) {
			matchDetails = append(matchDetails, *detail)
		}
	}

	totalTime := time.Since(startTime)
	fmt.Printf("✅ マッチ詳細取得完了: %d試合を%vで処理\n",
		len(matchDetails), totalTime.Round(time.Second))

	return &PlayerMatchSummary{
		Account:      *account,
		MatchHistory: matchDetails,
		GeneratedAt:  time.Now(),
		TotalMatches: len(matchDetails),
		MatchType:    "aram",
	}, nil
}

// プレイヤーの全ゲーム分析データを取得
func (c *Client) GetPlayerAllAnalysisWithContext(ctx context.Context, account *Account, matchCount int) (*PlayerMatchSummary, error) {
	fmt.Printf("全ゲームマッチ履歴を取得中（最大%d試合）...\n", matchCount)

	matchIDs, err := c.GetRankedMatchHistoryWithContext(ctx, account.PUUID, matchCount)
	if err != nil {
		return nil, fmt.Errorf("マッチ履歴取得エラー: %w", err)
	}

	fmt.Printf("取得したマッチ数: %d\n", len(matchIDs))

	if len(matchIDs) == 0 {
		return &PlayerMatchSummary{
			Account:      *account,
			MatchHistory: []MatchDetail{},
			GeneratedAt:  time.Now(),
			TotalMatches: 0,
			MatchType:    "all",
		}, nil
	}

	var matchDetails []MatchDetail
	startTime := time.Now()

	fmt.Printf("マッチ詳細取得開始...\n")

	for i, matchID := range matchIDs {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("処理がキャンセルされました: %w", ctx.Err())
		default:
		}

		if i%5 == 0 {
			elapsed := time.Since(startTime)
			if i > 0 {
				avgTime := elapsed / time.Duration(i)
				remaining := avgTime * time.Duration(len(matchIDs)-i)
				fmt.Printf("進捗: %d/%d (%.1f%%) - 経過: %v, 推定残り: %v\n",
					i, len(matchIDs), float64(i)/float64(len(matchIDs))*100,
					elapsed.Round(time.Second), remaining.Round(time.Second))
			}
		}

		detail, err := c.GetMatchDetailWithContext(ctx, matchID)
		if err != nil {
			fmt.Printf("⚠️  マッチ %s の取得に失敗: %v\n", matchID, err)
			continue
		}

		matchDetails = append(matchDetails, *detail)
	}

	totalTime := time.Since(startTime)
	fmt.Printf("✅ マッチ詳細取得完了: %d試合を%vで処理\n",
		len(matchDetails), totalTime.Round(time.Second))

	return &PlayerMatchSummary{
		Account:      *account,
		MatchHistory: matchDetails,
		GeneratedAt:  time.Now(),
		TotalMatches: len(matchDetails),
		MatchType:    "all",
	}, nil
}
