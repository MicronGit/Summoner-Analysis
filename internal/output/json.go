package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MicronGit/Summoner-Analysis/internal/riot"
)

func SavePlayerAnalysisToJSON(analysis *riot.PlayerMatchSummary, outputDir string) (string, error) {
	// 出力ディレクトリ作成
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("ディレクトリ作成エラー: %w", err)
	}

	// ファイル名生成（特殊文字を安全な文字に置換）
	safeGameName := strings.ReplaceAll(analysis.Account.SummonerName, " ", "_")
	safeGameName = strings.ReplaceAll(safeGameName, "#", "_")
	timestamp := analysis.GeneratedAt.Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_analysis_%s.json",
		safeGameName, analysis.Account.TagLine, timestamp)

	filepath := filepath.Join(outputDir, filename)

	// JSON形式でエンコード（インデント付き）
	jsonData, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON変換エラー: %w", err)
	}

	// ファイルに書き込み
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("ファイル書き込みエラー: %w", err)
	}

	return filepath, nil
}

// 簡易的な統計情報も出力
func SavePlayerStats(analysis *riot.PlayerMatchSummary, outputDir string) (string, error) {
	stats := calculateStats(analysis)

	safeGameName := strings.ReplaceAll(analysis.Account.SummonerName, " ", "_")
	timestamp := analysis.GeneratedAt.Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_stats_%s.json",
		safeGameName, analysis.Account.TagLine, timestamp)

	filepath := filepath.Join(outputDir, filename)

	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON変換エラー: %w", err)
	}

	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("ファイル書き込みエラー: %w", err)
	}

	return filepath, nil
}

func calculateStats(analysis *riot.PlayerMatchSummary) *PlayerStats {
	stats := &PlayerStats{
		PlayerInfo:    analysis.Account,
		GeneratedAt:   analysis.GeneratedAt,
		MatchType:     analysis.MatchType,
		TotalMatches:  analysis.TotalMatches,
		PositionStats: make(map[string]int),
	}

	if len(analysis.MatchHistory) == 0 {
		return stats
	}

	var totalKills, totalDeaths, totalAssists int
	var totalVisionScore, totalGoldEarned, totalCS, totalGameDuration int
	var wins int
	championStats := make(map[string]*ChampionStats)

	// 直近の試合用
	var recent10, recent5 []riot.Participant

	for i, match := range analysis.MatchHistory {
		// 該当プレイヤーの参加者データを見つける
		var playerData *riot.Participant
		for _, participant := range match.Info.Participants {
			if participant.PUUID == analysis.Account.PUUID {
				playerData = &participant
				break
			}
		}

		if playerData == nil {
			continue
		}

		// 基本統計
		totalKills += playerData.Kills
		totalDeaths += playerData.Deaths
		totalAssists += playerData.Assists
		totalVisionScore += playerData.VisionScore
		totalGoldEarned += playerData.GoldEarned
		totalCS += playerData.TotalMinionsKilled + playerData.NeutralMinionsKilled
		totalGameDuration += match.Info.GameDuration

		if playerData.Win {
			wins++
		}

		// 直近の試合データ
		if i < 10 {
			recent10 = append(recent10, *playerData)
		}
		if i < 5 {
			recent5 = append(recent5, *playerData)
		}

		// ポジション統計
		stats.PositionStats[playerData.TeamPosition]++

		// チャンピオン統計
		if _, exists := championStats[playerData.ChampionName]; !exists {
			championStats[playerData.ChampionName] = &ChampionStats{
				ChampionName: playerData.ChampionName,
			}
		}
		champStat := championStats[playerData.ChampionName]
		champStat.GamesPlayed++
		champStat.AverageKDA.Kills += float64(playerData.Kills)
		champStat.AverageKDA.Deaths += float64(playerData.Deaths)
		champStat.AverageKDA.Assists += float64(playerData.Assists)
		if playerData.Win {
			champStat.WinRate++
		}
	}

	// 平均値計算
	matchCount := float64(len(analysis.MatchHistory))
	stats.WinRate = float64(wins) / matchCount * 100
	stats.AverageKDA.Kills = float64(totalKills) / matchCount
	stats.AverageKDA.Deaths = float64(totalDeaths) / matchCount
	stats.AverageKDA.Assists = float64(totalAssists) / matchCount

	if stats.AverageKDA.Deaths > 0 {
		stats.AverageKDA.Ratio = (stats.AverageKDA.Kills + stats.AverageKDA.Assists) / stats.AverageKDA.Deaths
	}

	// ランク戦用統計
	stats.RankPerformance.AverageVisionScore = float64(totalVisionScore) / matchCount
	stats.RankPerformance.AverageGoldEarned = float64(totalGoldEarned) / matchCount

	// CS/分計算
	if totalGameDuration > 0 {
		totalMinutes := float64(totalGameDuration) / 60.0 / matchCount
		stats.RankPerformance.AverageCSPerMin = float64(totalCS) / matchCount / totalMinutes
	}

	// 直近の調子計算
	if len(recent10) > 0 {
		stats.RecentForm.Last10Games = calculateRecentForm(recent10)
	}
	if len(recent5) > 0 {
		stats.RecentForm.Last5Games = calculateRecentForm(recent5)
	}

	// チャンピオン統計の最終計算
	for _, champStat := range championStats {
		champStat.WinRate = champStat.WinRate / float64(champStat.GamesPlayed) * 100
		champStat.AverageKDA.Kills /= float64(champStat.GamesPlayed)
		champStat.AverageKDA.Deaths /= float64(champStat.GamesPlayed)
		champStat.AverageKDA.Assists /= float64(champStat.GamesPlayed)
		stats.MostPlayedChampions = append(stats.MostPlayedChampions, *champStat)
	}

	return stats
}

func calculateRecentForm(participants []riot.Participant) struct {
	WinRate    float64  `json:"winRate"`
	AverageKDA KDAStats `json:"averageKDA"`
} {
	var wins, kills, deaths, assists int

	for _, p := range participants {
		if p.Win {
			wins++
		}
		kills += p.Kills
		deaths += p.Deaths
		assists += p.Assists
	}

	count := float64(len(participants))
	result := struct {
		WinRate    float64  `json:"winRate"`
		AverageKDA KDAStats `json:"averageKDA"`
	}{
		WinRate: float64(wins) / count * 100,
		AverageKDA: KDAStats{
			Kills:   float64(kills) / count,
			Deaths:  float64(deaths) / count,
			Assists: float64(assists) / count,
		},
	}

	if result.AverageKDA.Deaths > 0 {
		result.AverageKDA.Ratio = (result.AverageKDA.Kills + result.AverageKDA.Assists) / result.AverageKDA.Deaths
	}

	return result
}
