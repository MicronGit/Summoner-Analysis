<template>
  <div id="app">
    <div class="app-container">
      <!-- ヘッダー -->
      <header class="app-header">
        <div class="header-content">
          <h1>League of Legends Summoner Analysis</h1>
          <p>プレイヤーの戦績を詳細に分析します</p>
        </div>
      </header>

      <!-- メインコンテンツ -->
      <main class="main-content">
        <!-- 検索フォーム -->
        <SearchForm
          v-if="!analysisResult"
          :is-loading="isLoading"
          :error="error"
          @submit="handleAnalyze"
        />

        <!-- 統計表示 -->
        <div v-if="analysisResult && !isLoading" class="results-section">
          <div class="results-header">
            <button @click="resetAnalysis" class="new-search-button">
              新しい検索
            </button>
          </div>
          <StatsDisplay :stats="analysisResult" />
        </div>

        <!-- ローディング画面 -->
        <LoadingScreen
          v-if="isLoading"
          :title="loadingState.title"
          :message="loadingState.message"
          :progress="loadingState.progress"
          :current-step="loadingState.currentStep"
        />
      </main>

      <!-- フッター -->
      <footer class="app-footer">
        <p>&copy; 2024 Summoner Analysis. Powered by Riot Games API.</p>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import SearchForm from './components/SearchForm.vue'
import StatsDisplay from './components/StatsDisplay.vue'
import LoadingScreen from './components/LoadingScreen.vue'
import ApiService from './services/api'
import type { SearchForm as SearchFormType, PlayerStats } from './types'

// State
const isLoading = ref(false)
const error = ref('')
const analysisResult = ref<PlayerStats | null>(null)

const loadingState = reactive({
  title: '分析中...',
  message: 'プレイヤーデータを取得しています',
  progress: 0,
  currentStep: 1
})

// Methods
const handleAnalyze = async (form: SearchFormType) => {
  isLoading.value = true
  error.value = ''
  analysisResult.value = null

  try {
    // ローディング状態の更新
    updateLoadingState(1, 'アカウント情報を取得中...', 20)

    // APIリクエスト
    const request = {
      gameName: form.gameName,
      tagLine: form.tagLine,
      region: form.region,
      gameType: form.gameType,
      matchCount: form.matchCount
    }

    // 段階的にローディング状態を更新
    setTimeout(() => updateLoadingState(2, 'マッチ履歴を取得中...', 50), 1000)
    setTimeout(() => updateLoadingState(3, '統計を計算中...', 80), 3000)

    const result = await ApiService.analyzePlayer(request)

    updateLoadingState(3, '完了', 100)

    setTimeout(() => {
      analysisResult.value = result
      isLoading.value = false
    }, 500)

  } catch (err) {
    console.error('Analysis error:', err)
    error.value = err instanceof Error ? err.message : '予期しないエラーが発生しました'
    isLoading.value = false
  }
}

const updateLoadingState = (step: number, message: string, progress: number) => {
  loadingState.currentStep = step
  loadingState.message = message
  loadingState.progress = progress
}

const resetAnalysis = () => {
  analysisResult.value = null
  error.value = ''
  loadingState.currentStep = 1
  loadingState.progress = 0
  loadingState.message = 'プレイヤーデータを取得しています'
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  line-height: 1.6;
  color: #1e293b;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  min-height: 100vh;
}

#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.app-header {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  padding: 2rem 0;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  text-align: center;
  padding: 0 2rem;
}

.header-content h1 {
  color: white;
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header-content p {
  color: rgba(255, 255, 255, 0.9);
  font-size: 1.125rem;
  font-weight: 400;
}

.main-content {
  flex: 1;
  padding: 2rem 0;
  position: relative;
}

.results-section {
  animation: fadeInUp 0.6s ease-out;
}

.results-header {
  max-width: 1200px;
  margin: 0 auto 2rem;
  padding: 0 2rem;
  display: flex;
  justify-content: flex-end;
}

.new-search-button {
  background: rgba(255, 255, 255, 0.9);
  color: #1e293b;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  backdrop-filter: blur(10px);
}

.new-search-button:hover {
  background: white;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.app-footer {
  background: rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  padding: 1rem 0;
  text-align: center;
}

.app-footer p {
  color: rgba(255, 255, 255, 0.8);
  font-size: 0.875rem;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 768px) {
  .header-content h1 {
    font-size: 2rem;
  }

  .header-content p {
    font-size: 1rem;
  }

  .main-content {
    padding: 1rem 0;
  }

  .results-header {
    padding: 0 1rem;
  }
}
</style>