package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MicronGit/Summoner-Analysis/internal/config"
	"github.com/MicronGit/Summoner-Analysis/internal/output"
	"github.com/MicronGit/Summoner-Analysis/internal/riot"
)

func main() {
	// キャンセル可能なコンテキスト作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ctrl+C での中断処理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n処理を中断しています...")
		cancel()
	}()

	// タイムアウト設定（15分）
	ctx, cancel = context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	// 以下既存の処理にコンテキストを追加
	cfg := config.Load()

	if cfg.RiotAPIKey == "" {
		log.Fatal("RIOT_API_KEY が設定されていません")
	}

	client := riot.NewClient(cfg.RiotAPIKey, cfg.Region)

	gameName := "そっちん"
	tagLine := "JP1"

	fmt.Printf("=== %s#%s のランク戦分析を開始 ===\n", gameName, tagLine)

	// アカウント情報取得
	fmt.Println("1. アカウント情報を取得中...")
	account, err := client.GetAccountByRiotID(gameName, tagLine)
	if err != nil {
		log.Fatalf("アカウント取得エラー: %v", err)
	}

	fmt.Printf("   取得完了: %s#%s\n", account.SummonerName, account.TagLine)

	// ランク戦分析データ取得
	fmt.Println("2. ランク戦マッチ履歴と詳細データを取得中...")
	matchCount := 100
	analysis, err := client.GetPlayerRankedAnalysisWithContext(ctx, account, matchCount)
	if err != nil {
		if ctx.Err() != nil {
			fmt.Printf("処理がキャンセルまたはタイムアウトしました: %v\n", ctx.Err())
			os.Exit(1)
		}
		log.Fatalf("プレイヤー分析エラー: %v", err)
	}

	fmt.Printf("   取得完了: %d試合のランク戦データを分析\n", analysis.TotalMatches)

	if analysis.TotalMatches == 0 {
		fmt.Println("ランク戦の履歴が見つかりませんでした。")
		return
	}

	// JSONファイルに出力
	fmt.Println("3. データをJSONファイルに出力中...")
	outputDir := "./output"

	// 詳細データ出力
	detailPath, err := output.SavePlayerAnalysisToJSON(analysis, outputDir)
	if err != nil {
		log.Fatalf("詳細データ出力エラー: %v", err)
	}

	// 統計データ出力
	statsPath, err := output.SavePlayerStats(analysis, outputDir)
	if err != nil {
		log.Fatalf("統計データ出力エラー: %v", err)
	}

	fmt.Printf("=== ランク戦分析完了 ===\n")
	fmt.Printf("詳細データ: %s\n", detailPath)
	fmt.Printf("統計データ: %s\n", statsPath)
	fmt.Printf("総ランク戦試合数: %d\n", analysis.TotalMatches)
}
