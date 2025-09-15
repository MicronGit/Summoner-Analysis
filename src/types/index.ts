// リージョン定義
export type Region = 'asia' | 'americas' | 'europe'

// ゲームタイプ定義
export type GameType = 'ranked' | 'normal' | 'aram' | 'all'

// 検索フォーム
export interface SearchForm {
  gameName: string
  tagLine: string
  region: Region
  gameType: GameType
  matchCount: number
}

// アカウント情報
export interface Account {
  puuid: string
  gameName: string
  tagLine: string
}

// KDA統計
export interface KDAStats {
  kills: number
  deaths: number
  assists: number
  kdaRatio: number
}

// チャンピオン統計
export interface ChampionStats {
  championName: string
  gamesPlayed: number
  winRate: number
  averageKDA: KDAStats
}

// ランク成績
export interface RankStats {
  averageVisionScore: number
  averageGoldEarned: number
  averageCSPerMin: number
  killParticipation?: number
}

// 直近フォーム
export interface RecentFormStats {
  last10Games: {
    winRate: number
    averageKDA: KDAStats
  }
  last5Games: {
    winRate: number
    averageKDA: KDAStats
  }
}

// プレイヤー統計
export interface PlayerStats {
  playerInfo: Account
  generatedAt: string
  matchType: string
  totalMatches: number
  winRate: number
  averageKDA: KDAStats
  rankPerformance: RankStats
  mostPlayedChampions: ChampionStats[]
  positionStats: Record<string, number>
  recentForm: RecentFormStats
}

// API レスポンス
export interface AnalysisResponse {
  success: boolean
  data?: PlayerStats
  error?: string
}

// API リクエスト
export interface AnalysisRequest {
  gameName: string
  tagLine: string
  region: Region
  gameType: GameType
  matchCount: number
}

// 選択肢
export interface SelectOption {
  value: string
  label: string
}

// ローディング状態
export interface LoadingState {
  isLoading: boolean
  progress?: number
  message?: string
}