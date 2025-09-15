<template>
  <div class="stats-display">
    <div class="player-info">
      <h2>{{ stats.playerInfo.gameName }}#{{ stats.playerInfo.tagLine }}</h2>
      <p class="match-info">{{ stats.matchType === 'ranked' ? 'ランク戦' : stats.matchType }} - {{ stats.totalMatches }}試合</p>
      <p class="generated-at">生成日時: {{ formatDate(stats.generatedAt) }}</p>
    </div>

    <div class="stats-grid">
      <!-- 基本統計 -->
      <div class="stat-card">
        <h3>基本統計</h3>
        <div class="stat-row">
          <span class="stat-label">勝率</span>
          <span class="stat-value win-rate" :class="getWinRateClass(stats.winRate)">
            {{ stats.winRate.toFixed(1) }}%
          </span>
        </div>
        <div class="stat-row">
          <span class="stat-label">平均KDA</span>
          <span class="stat-value">
            {{ stats.averageKDA.kills.toFixed(1) }} /
            {{ stats.averageKDA.deaths.toFixed(1) }} /
            {{ stats.averageKDA.assists.toFixed(1) }}
          </span>
        </div>
        <div class="stat-row">
          <span class="stat-label">KDA比率</span>
          <span class="stat-value kda-ratio" :class="getKDAClass(stats.averageKDA.kdaRatio)">
            {{ stats.averageKDA.kdaRatio.toFixed(2) }}
          </span>
        </div>
      </div>

      <!-- ランク成績 -->
      <div class="stat-card">
        <h3>ランク成績</h3>
        <div class="stat-row">
          <span class="stat-label">平均ビジョンスコア</span>
          <span class="stat-value">{{ stats.rankPerformance.averageVisionScore.toFixed(1) }}</span>
        </div>
        <div class="stat-row">
          <span class="stat-label">平均ゴールド獲得</span>
          <span class="stat-value">{{ formatGold(stats.rankPerformance.averageGoldEarned) }}</span>
        </div>
        <div class="stat-row">
          <span class="stat-label">CS/分</span>
          <span class="stat-value">{{ stats.rankPerformance.averageCSPerMin.toFixed(1) }}</span>
        </div>
      </div>

      <!-- 直近フォーム -->
      <div class="stat-card">
        <h3>直近フォーム</h3>
        <div class="recent-form">
          <div class="form-section">
            <h4>直近10試合</h4>
            <div class="stat-row">
              <span class="stat-label">勝率</span>
              <span class="stat-value win-rate" :class="getWinRateClass(stats.recentForm.last10Games.winRate)">
                {{ stats.recentForm.last10Games.winRate.toFixed(1) }}%
              </span>
            </div>
            <div class="stat-row">
              <span class="stat-label">KDA比率</span>
              <span class="stat-value">{{ stats.recentForm.last10Games.averageKDA.kdaRatio.toFixed(2) }}</span>
            </div>
          </div>
          <div class="form-section">
            <h4>直近5試合</h4>
            <div class="stat-row">
              <span class="stat-label">勝率</span>
              <span class="stat-value win-rate" :class="getWinRateClass(stats.recentForm.last5Games.winRate)">
                {{ stats.recentForm.last5Games.winRate.toFixed(1) }}%
              </span>
            </div>
            <div class="stat-row">
              <span class="stat-label">KDA比率</span>
              <span class="stat-value">{{ stats.recentForm.last5Games.averageKDA.kdaRatio.toFixed(2) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- ポジション統計 -->
      <div class="stat-card">
        <h3>ポジション統計</h3>
        <div class="position-stats">
          <div
            v-for="[position, count] in sortedPositions"
            :key="position"
            class="position-row"
          >
            <span class="position-name">{{ formatPosition(position) }}</span>
            <span class="position-count">{{ count }}試合</span>
            <div class="position-bar">
              <div
                class="position-fill"
                :style="{ width: (count / stats.totalMatches * 100) + '%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 使用チャンピオン -->
    <div class="champion-section">
      <h3>よく使うチャンピオン</h3>
      <div class="champion-grid">
        <div
          v-for="champion in topChampions"
          :key="champion.championName"
          class="champion-card"
        >
          <h4>{{ champion.championName }}</h4>
          <div class="champion-stats">
            <div class="champion-stat">
              <span class="label">試合数</span>
              <span class="value">{{ champion.gamesPlayed }}</span>
            </div>
            <div class="champion-stat">
              <span class="label">勝率</span>
              <span class="value win-rate" :class="getWinRateClass(champion.winRate)">
                {{ champion.winRate.toFixed(1) }}%
              </span>
            </div>
            <div class="champion-stat">
              <span class="label">KDA</span>
              <span class="value">
                {{ champion.averageKDA.kills.toFixed(1) }}/{{ champion.averageKDA.deaths.toFixed(1) }}/{{ champion.averageKDA.assists.toFixed(1) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { PlayerStats } from '../types'

// Props
interface Props {
  stats: PlayerStats
}

const props = defineProps<Props>()

// Computed
const sortedPositions = computed(() => {
  return Object.entries(props.stats.positionStats)
    .sort(([, a], [, b]) => b - a)
    .filter(([, count]) => count > 0)
})

const topChampions = computed(() => {
  return props.stats.mostPlayedChampions
    .sort((a, b) => b.gamesPlayed - a.gamesPlayed)
    .slice(0, 6)
})

// Methods
const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleString('ja-JP')
}

const formatGold = (gold: number): string => {
  return gold.toLocaleString() + 'G'
}

const formatPosition = (position: string): string => {
  const positionMap: Record<string, string> = {
    'TOP': 'トップ',
    'JUNGLE': 'ジャングル',
    'MIDDLE': 'ミッド',
    'BOTTOM': 'ボット',
    'UTILITY': 'サポート'
  }
  return positionMap[position] || position
}

const getWinRateClass = (winRate: number): string => {
  if (winRate >= 60) return 'excellent'
  if (winRate >= 50) return 'good'
  return 'poor'
}

const getKDAClass = (kda: number): string => {
  if (kda >= 3) return 'excellent'
  if (kda >= 2) return 'good'
  return 'poor'
}
</script>

<style scoped>
.stats-display {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.player-info {
  text-align: center;
  margin-bottom: 2rem;
  padding: 1.5rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.player-info h2 {
  color: #1e293b;
  margin: 0 0 0.5rem 0;
  font-size: 2rem;
  font-weight: 700;
}

.match-info {
  color: #64748b;
  font-size: 1.125rem;
  margin: 0.5rem 0;
}

.generated-at {
  color: #94a3b8;
  font-size: 0.875rem;
  margin: 0;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.stat-card h3 {
  color: #1e293b;
  margin: 0 0 1rem 0;
  font-size: 1.25rem;
  font-weight: 600;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 0.5rem;
}

.stat-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 0;
  border-bottom: 1px solid #f1f5f9;
}

.stat-row:last-child {
  border-bottom: none;
}

.stat-label {
  color: #64748b;
  font-weight: 500;
}

.stat-value {
  font-weight: 600;
  color: #1e293b;
}

.win-rate.excellent, .kda-ratio.excellent {
  color: #059669;
}

.win-rate.good, .kda-ratio.good {
  color: #0891b2;
}

.win-rate.poor, .kda-ratio.poor {
  color: #dc2626;
}

.recent-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-section h4 {
  color: #475569;
  margin: 0 0 0.5rem 0;
  font-size: 1rem;
}

.position-stats {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.position-row {
  display: grid;
  grid-template-columns: 1fr auto 2fr;
  gap: 0.75rem;
  align-items: center;
}

.position-name {
  font-weight: 500;
  color: #374151;
}

.position-count {
  font-size: 0.875rem;
  color: #64748b;
}

.position-bar {
  background: #e2e8f0;
  height: 8px;
  border-radius: 4px;
  overflow: hidden;
}

.position-fill {
  background: linear-gradient(90deg, #3b82f6, #1d4ed8);
  height: 100%;
  transition: width 0.3s ease;
}

.champion-section {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.champion-section h3 {
  color: #1e293b;
  margin: 0 0 1.5rem 0;
  font-size: 1.25rem;
  font-weight: 600;
}

.champion-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.champion-card {
  background: #f8fafc;
  border-radius: 8px;
  padding: 1rem;
  border: 1px solid #e2e8f0;
}

.champion-card h4 {
  color: #1e293b;
  margin: 0 0 0.75rem 0;
  font-size: 1rem;
  font-weight: 600;
  text-align: center;
}

.champion-stats {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.champion-stat {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
}

.champion-stat .label {
  color: #64748b;
}

.champion-stat .value {
  font-weight: 600;
  color: #1e293b;
}

@media (max-width: 768px) {
  .stats-display {
    padding: 1rem;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .champion-grid {
    grid-template-columns: 1fr;
  }

  .recent-form {
    gap: 0.75rem;
  }
}
</style>