package output

import (
	"time"

	"github.com/MicronGit/Summoner-Analysis/internal/riot"
)

type PlayerStats struct {
	PlayerInfo          riot.Account    `json:"playerInfo"`
	GeneratedAt         time.Time       `json:"generatedAt"`
	MatchType           string          `json:"matchType"`
	TotalMatches        int             `json:"totalMatches"`
	WinRate             float64         `json:"winRate"`
	AverageKDA          KDAStats        `json:"averageKDA"`
	RankPerformance     RankStats       `json:"rankPerformance"`
	MostPlayedChampions []ChampionStats `json:"mostPlayedChampions"`
	PositionStats       map[string]int  `json:"positionStats"`
	RecentForm          RecentFormStats `json:"recentForm"` // 直近の調子
}

type RankStats struct {
	AverageVisionScore float64 `json:"averageVisionScore"`
	AverageGoldEarned  float64 `json:"averageGoldEarned"`
	AverageCSPerMin    float64 `json:"averageCSPerMin"`
	KillParticipation  float64 `json:"killParticipation"` // キル関与率
}

type KDAStats struct {
	Kills   float64 `json:"kills"`
	Deaths  float64 `json:"deaths"`
	Assists float64 `json:"assists"`
	Ratio   float64 `json:"kdaRatio"`
}

type ChampionStats struct {
	ChampionName string   `json:"championName"`
	GamesPlayed  int      `json:"gamesPlayed"`
	WinRate      float64  `json:"winRate"`
	AverageKDA   KDAStats `json:"averageKDA"`
}

type RecentFormStats struct {
	Last10Games struct {
		WinRate    float64  `json:"winRate"`
		AverageKDA KDAStats `json:"averageKDA"`
	} `json:"last10Games"`
	Last5Games struct {
		WinRate    float64  `json:"winRate"`
		AverageKDA KDAStats `json:"averageKDA"`
	} `json:"last5Games"`
}
