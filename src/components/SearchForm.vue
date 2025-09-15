<template>
  <div class="search-form">
    <div class="form-container">
      <h2>Summoner Analysis</h2>
      <form @submit.prevent="handleSubmit" class="form">
        <div class="form-row">
          <div class="form-group">
            <label for="gameName">ゲーム名</label>
            <input
              id="gameName"
              v-model="form.gameName"
              type="text"
              placeholder="例: そっちん"
              required
              :disabled="isLoading"
            />
          </div>
          <div class="form-group">
            <label for="tagLine">タグライン</label>
            <input
              id="tagLine"
              v-model="form.tagLine"
              type="text"
              placeholder="例: JP1"
              required
              :disabled="isLoading"
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="region">リージョン</label>
            <select id="region" v-model="form.region" :disabled="isLoading">
              <option value="asia">Asia (日本・韓国)</option>
              <option value="americas">Americas (北米・南米)</option>
              <option value="europe">Europe (ヨーロッパ)</option>
            </select>
          </div>
          <div class="form-group">
            <label for="gameType">ゲーム種別</label>
            <select id="gameType" v-model="form.gameType" :disabled="isLoading">
              <option value="ranked">ランク戦のみ</option>
              <option value="normal">ノーマルのみ</option>
              <option value="aram">ARAMのみ</option>
              <option value="all">すべて</option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="matchCount">取得試合数</label>
            <input
              id="matchCount"
              v-model.number="form.matchCount"
              type="number"
              min="1"
              max="100"
              :disabled="isLoading"
            />
            <small>最大100試合まで</small>
          </div>
        </div>

        <button type="submit" class="submit-button" :disabled="isLoading">
          <span v-if="isLoading">分析中...</span>
          <span v-else>分析開始</span>
        </button>

        <div v-if="error" class="error-message">
          {{ error }}
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { SearchForm, Region, GameType } from '../types'

// Props
interface Props {
  isLoading?: boolean
  error?: string
}

defineProps<Props>()

// Emits
interface Emits {
  (e: 'submit', form: SearchForm): void
}

const emit = defineEmits<Emits>()

// Form state
const form = reactive<SearchForm>({
  gameName: '',
  tagLine: '',
  region: 'asia' as Region,
  gameType: 'ranked' as GameType,
  matchCount: 50
})

// Form submission
const handleSubmit = () => {
  if (!form.gameName.trim() || !form.tagLine.trim()) {
    return
  }

  emit('submit', { ...form })
}
</script>

<style scoped>
.search-form {
  max-width: 600px;
  margin: 0 auto;
  padding: 2rem;
}

.form-container {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

h2 {
  color: #1e293b;
  margin-bottom: 1.5rem;
  text-align: center;
  font-size: 1.875rem;
  font-weight: 700;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group:last-child {
  grid-column: 1 / -1;
}

label {
  font-weight: 600;
  color: #374151;
  font-size: 0.875rem;
}

input,
select {
  padding: 0.75rem;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s;
}

input:focus,
select:focus {
  outline: none;
  border-color: #3b82f6;
}

input:disabled,
select:disabled {
  background-color: #f9fafb;
  cursor: not-allowed;
}

small {
  color: #6b7280;
  font-size: 0.75rem;
}

.submit-button {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: white;
  padding: 0.875rem 2rem;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  margin-top: 1rem;
}

.submit-button:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 8px 25px rgba(59, 130, 246, 0.3);
}

.submit-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}

.error-message {
  background: #fef2f2;
  color: #dc2626;
  padding: 0.75rem;
  border-radius: 8px;
  border: 1px solid #fecaca;
  font-size: 0.875rem;
  margin-top: 1rem;
}

@media (max-width: 640px) {
  .form-row {
    grid-template-columns: 1fr;
  }

  .search-form {
    padding: 1rem;
  }

  .form-container {
    padding: 1.5rem;
  }
}
</style>