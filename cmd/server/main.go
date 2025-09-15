package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MicronGit/Summoner-Analysis/internal/config"
	"github.com/MicronGit/Summoner-Analysis/internal/riot"
)

type APIRequest struct {
	GameName   string `json:"gameName"`
	TagLine    string `json:"tagLine"`
	Region     string `json:"region"`
	GameType   string `json:"gameType"`
	MatchCount int    `json:"matchCount"`
}

type APIResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Server struct {
	cfg    *config.Config
	client *riot.Client
}

func NewServer() *Server {
	cfg := config.Load()
	if cfg.RiotAPIKey == "" {
		log.Fatal("RIOT_API_KEY が設定されていません")
	}

	return &Server{
		cfg:    cfg,
		client: riot.NewClient(cfg.RiotAPIKey, cfg.Region),
	}
}

func (s *Server) enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req APIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// バリデーション
	if req.GameName == "" || req.TagLine == "" {
		s.sendError(w, "GameName and TagLine are required", http.StatusBadRequest)
		return
	}

	if req.MatchCount <= 0 || req.MatchCount > 100 {
		req.MatchCount = 50
	}

	// クライアントのリージョンを更新
	if req.Region != "" {
		s.client.Region = req.Region
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	log.Printf("Starting analysis for %s#%s (region: %s, gameType: %s, matches: %d)",
		req.GameName, req.TagLine, req.Region, req.GameType, req.MatchCount)

	// アカウント情報取得
	account, err := s.client.GetAccountByRiotID(req.GameName, req.TagLine)
	if err != nil {
		log.Printf("Account fetch error: %v", err)
		s.sendError(w, fmt.Sprintf("アカウント取得エラー: %v", err), http.StatusNotFound)
		return
	}

	// マッチ分析実行
	var analysis *riot.PlayerMatchSummary

	switch req.GameType {
	case "ranked":
		analysis, err = s.client.GetPlayerRankedAnalysisWithContext(ctx, account, req.MatchCount)
	case "normal":
		analysis, err = s.client.GetPlayerNormalAnalysisWithContext(ctx, account, req.MatchCount)
	case "aram":
		analysis, err = s.client.GetPlayerARAMAnalysisWithContext(ctx, account, req.MatchCount)
	case "all":
		analysis, err = s.client.GetPlayerAllAnalysisWithContext(ctx, account, req.MatchCount)
	default:
		analysis, err = s.client.GetPlayerRankedAnalysisWithContext(ctx, account, req.MatchCount)
	}

	if err != nil {
		log.Printf("Analysis error: %v", err)
		s.sendError(w, fmt.Sprintf("分析エラー: %v", err), http.StatusInternalServerError)
		return
	}

	// 統計計算（簡略版）
	stats := s.calculateStats(analysis)

	log.Printf("Analysis completed for %s#%s: %d matches", req.GameName, req.TagLine, analysis.TotalMatches)

	s.sendSuccess(w, stats)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

func (s *Server) sendSuccess(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

type ChampionStat struct {
	ChampionName string  `json:"championName"`
	GamesPlayed  int     `json:"gamesPlayed"`
	Wins         int     `json:"wins"`
	TotalKills   int     `json:"totalKills"`
	TotalDeaths  int     `json:"totalDeaths"`
	TotalAssists int     `json:"totalAssists"`
	WinRate      float64 `json:"winRate"`
	AverageKDA   struct {
		Kills    float64 `json:"kills"`
		Deaths   float64 `json:"deaths"`
		Assists  float64 `json:"assists"`
		KDARatio float64 `json:"kdaRatio"`
	} `json:"averageKDA"`
}

type RecentFormData struct {
	WinRate    float64 `json:"winRate"`
	AverageKDA struct {
		Kills    float64 `json:"kills"`
		Deaths   float64 `json:"deaths"`
		Assists  float64 `json:"assists"`
		KDARatio float64 `json:"kdaRatio"`
	} `json:"averageKDA"`
}

func (s *Server) calculateStats(analysis *riot.PlayerMatchSummary) map[string]any {
	if len(analysis.MatchHistory) == 0 {
		return map[string]any{
			"playerInfo": map[string]string{
				"gameName": analysis.Account.SummonerName,
				"tagLine":  analysis.Account.TagLine,
			},
			"totalMatches": 0,
			"winRate":      0.0,
			"averageKDA": map[string]float64{
				"kills":    0,
				"deaths":   0,
				"assists":  0,
				"kdaRatio": 0,
			},
		}
	}

	// プレイヤーデータの抽出
	var playerMatches []riot.Participant
	for _, match := range analysis.MatchHistory {
		for _, participant := range match.Info.Participants {
			if participant.PUUID == analysis.Account.PUUID {
				playerMatches = append(playerMatches, participant)
				break
			}
		}
	}

	if len(playerMatches) == 0 {
		return map[string]any{
			"playerInfo": map[string]string{
				"gameName": analysis.Account.SummonerName,
				"tagLine":  analysis.Account.TagLine,
			},
			"totalMatches": 0,
			"winRate":      0.0,
		}
	}

	// 基本統計計算
	totalStats := s.calculateBasicStats(playerMatches)

	// チャンピオン統計
	championStats := s.calculateChampionStats(playerMatches)

	// ポジション統計
	positionStats := s.calculatePositionStats(playerMatches)

	// 直近フォーム
	recentForm := s.calculateRecentForm(playerMatches)

	// CS/分計算
	avgCSPerMin := s.calculateCSPerMin(analysis.MatchHistory, playerMatches)

	return map[string]any{
		"playerInfo": map[string]string{
			"gameName": analysis.Account.SummonerName,
			"tagLine":  analysis.Account.TagLine,
		},
		"generatedAt":  analysis.GeneratedAt.Format(time.RFC3339),
		"matchType":    analysis.MatchType,
		"totalMatches": len(playerMatches),
		"winRate":      totalStats.WinRate,
		"averageKDA": map[string]float64{
			"kills":    totalStats.AvgKills,
			"deaths":   totalStats.AvgDeaths,
			"assists":  totalStats.AvgAssists,
			"kdaRatio": totalStats.KDARatio,
		},
		"rankPerformance": map[string]float64{
			"averageVisionScore": totalStats.AvgVisionScore,
			"averageGoldEarned":  totalStats.AvgGoldEarned,
			"averageCSPerMin":    avgCSPerMin,
		},
		"mostPlayedChampions": championStats,
		"positionStats":       positionStats,
		"recentForm":          recentForm,
	}
}

type BasicStats struct {
	WinRate          float64
	AvgKills         float64
	AvgDeaths        float64
	AvgAssists       float64
	KDARatio         float64
	AvgVisionScore   float64
	AvgGoldEarned    float64
}

func (s *Server) calculateBasicStats(playerMatches []riot.Participant) BasicStats {
	var totalKills, totalDeaths, totalAssists int
	var totalVisionScore, totalGoldEarned int
	var wins int

	for _, match := range playerMatches {
		totalKills += match.Kills
		totalDeaths += match.Deaths
		totalAssists += match.Assists
		totalVisionScore += match.VisionScore
		totalGoldEarned += match.GoldEarned

		if match.Win {
			wins++
		}
	}

	matchCount := float64(len(playerMatches))
	avgKills := float64(totalKills) / matchCount
	avgDeaths := float64(totalDeaths) / matchCount
	avgAssists := float64(totalAssists) / matchCount

	var kdaRatio float64
	if avgDeaths > 0 {
		kdaRatio = (avgKills + avgAssists) / avgDeaths
	} else {
		kdaRatio = avgKills + avgAssists
	}

	return BasicStats{
		WinRate:          float64(wins) / matchCount * 100,
		AvgKills:         avgKills,
		AvgDeaths:        avgDeaths,
		AvgAssists:       avgAssists,
		KDARatio:         kdaRatio,
		AvgVisionScore:   float64(totalVisionScore) / matchCount,
		AvgGoldEarned:    float64(totalGoldEarned) / matchCount,
	}
}

func (s *Server) calculateChampionStats(playerMatches []riot.Participant) []ChampionStat {
	championData := make(map[string]*ChampionStat)

	for _, match := range playerMatches {
		champName := match.ChampionName
		if _, exists := championData[champName]; !exists {
			championData[champName] = &ChampionStat{
				ChampionName: champName,
			}
		}

		champ := championData[champName]
		champ.GamesPlayed++
		champ.TotalKills += match.Kills
		champ.TotalDeaths += match.Deaths
		champ.TotalAssists += match.Assists

		if match.Win {
			champ.Wins++
		}
	}

	// 統計計算
	var championStats []ChampionStat
	for _, champ := range championData {
		games := float64(champ.GamesPlayed)
		champ.WinRate = float64(champ.Wins) / games * 100
		champ.AverageKDA.Kills = float64(champ.TotalKills) / games
		champ.AverageKDA.Deaths = float64(champ.TotalDeaths) / games
		champ.AverageKDA.Assists = float64(champ.TotalAssists) / games

		if champ.AverageKDA.Deaths > 0 {
			champ.AverageKDA.KDARatio = (champ.AverageKDA.Kills + champ.AverageKDA.Assists) / champ.AverageKDA.Deaths
		} else {
			champ.AverageKDA.KDARatio = champ.AverageKDA.Kills + champ.AverageKDA.Assists
		}

		championStats = append(championStats, *champ)
	}

	// ゲーム数でソート（多い順）
	for i := 0; i < len(championStats)-1; i++ {
		for j := i + 1; j < len(championStats); j++ {
			if championStats[i].GamesPlayed < championStats[j].GamesPlayed {
				championStats[i], championStats[j] = championStats[j], championStats[i]
			}
		}
	}

	return championStats
}

func (s *Server) calculatePositionStats(playerMatches []riot.Participant) map[string]int {
	positionStats := make(map[string]int)

	for _, match := range playerMatches {
		position := match.TeamPosition
		if position != "" {
			positionStats[position]++
		}
	}

	return positionStats
}

func (s *Server) calculateRecentForm(playerMatches []riot.Participant) map[string]RecentFormData {
	// 最新順にソート（既に最新順になっているはず）
	last10 := playerMatches
	if len(playerMatches) > 10 {
		last10 = playerMatches[:10]
	}

	last5 := playerMatches
	if len(playerMatches) > 5 {
		last5 = playerMatches[:5]
	}

	return map[string]RecentFormData{
		"last10Games": s.calculateFormData(last10),
		"last5Games":  s.calculateFormData(last5),
	}
}

func (s *Server) calculateFormData(matches []riot.Participant) RecentFormData {
	if len(matches) == 0 {
		return RecentFormData{}
	}

	var totalKills, totalDeaths, totalAssists int
	var wins int

	for _, match := range matches {
		totalKills += match.Kills
		totalDeaths += match.Deaths
		totalAssists += match.Assists

		if match.Win {
			wins++
		}
	}

	matchCount := float64(len(matches))
	avgKills := float64(totalKills) / matchCount
	avgDeaths := float64(totalDeaths) / matchCount
	avgAssists := float64(totalAssists) / matchCount

	var kdaRatio float64
	if avgDeaths > 0 {
		kdaRatio = (avgKills + avgAssists) / avgDeaths
	} else {
		kdaRatio = avgKills + avgAssists
	}

	form := RecentFormData{
		WinRate: float64(wins) / matchCount * 100,
	}
	form.AverageKDA.Kills = avgKills
	form.AverageKDA.Deaths = avgDeaths
	form.AverageKDA.Assists = avgAssists
	form.AverageKDA.KDARatio = kdaRatio

	return form
}

func (s *Server) calculateCSPerMin(matches []riot.MatchDetail, playerMatches []riot.Participant) float64 {
	if len(matches) != len(playerMatches) {
		return 0
	}

	var totalCS int
	var totalMinutes float64

	for i, match := range matches {
		if i < len(playerMatches) {
			totalCS += playerMatches[i].TotalMinionsKilled + playerMatches[i].NeutralMinionsKilled
			totalMinutes += float64(match.Info.GameDuration) / 60.0
		}
	}

	if totalMinutes > 0 {
		return float64(totalCS) / totalMinutes
	}
	return 0
}

func main() {
	server := NewServer()

	http.HandleFunc("/api/analyze", server.handleAnalyze)
	http.HandleFunc("/api/health", server.handleHealth)

	// 静的ファイル配信（本番用）
	fs := http.FileServer(http.Dir("./dist"))
	http.Handle("/", fs)

	port := "8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}