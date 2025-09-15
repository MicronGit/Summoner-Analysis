<template>
  <div class="loading-screen">
    <div class="loading-container">
      <div class="spinner"></div>
      <h3>{{ title }}</h3>
      <p class="loading-message">{{ message }}</p>
      <div v-if="showProgress" class="progress-container">
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: progress + '%' }"></div>
        </div>
        <span class="progress-text">{{ progress }}%</span>
      </div>
      <div class="loading-steps">
        <div class="step" :class="{ active: currentStep >= 1, completed: currentStep > 1 }">
          <div class="step-icon">1</div>
          <span>アカウント情報取得</span>
        </div>
        <div class="step" :class="{ active: currentStep >= 2, completed: currentStep > 2 }">
          <div class="step-icon">2</div>
          <span>マッチ履歴取得</span>
        </div>
        <div class="step" :class="{ active: currentStep >= 3, completed: currentStep > 3 }">
          <div class="step-icon">3</div>
          <span>統計計算</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

// Props
interface Props {
  title?: string
  message?: string
  progress?: number
  currentStep?: number
}

const props = withDefaults(defineProps<Props>(), {
  title: '分析中...',
  message: 'プレイヤーデータを取得しています',
  progress: 0,
  currentStep: 1
})

// Computed
const showProgress = computed(() => props.progress > 0)
</script>

<style scoped>
.loading-screen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.loading-container {
  background: white;
  border-radius: 16px;
  padding: 3rem;
  text-align: center;
  max-width: 400px;
  width: 90%;
  box-shadow: 0 20px 25px rgba(0, 0, 0, 0.2);
}

.spinner {
  width: 60px;
  height: 60px;
  border: 4px solid #e2e8f0;
  border-top: 4px solid #3b82f6;
  border-radius: 50%;
  margin: 0 auto 1.5rem;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

h3 {
  color: #1e293b;
  margin: 0 0 0.5rem 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.loading-message {
  color: #64748b;
  margin: 0 0 1.5rem 0;
  font-size: 1rem;
}

.progress-container {
  margin-bottom: 2rem;
}

.progress-bar {
  background: #e2e8f0;
  height: 8px;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 0.5rem;
}

.progress-fill {
  background: linear-gradient(90deg, #3b82f6, #1d4ed8);
  height: 100%;
  transition: width 0.3s ease;
}

.progress-text {
  color: #64748b;
  font-size: 0.875rem;
  font-weight: 500;
}

.loading-steps {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  text-align: left;
}

.step {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.step.active {
  background: #eff6ff;
  color: #1d4ed8;
}

.step.completed {
  background: #f0fdf4;
  color: #059669;
}

.step-icon {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #e2e8f0;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 600;
  transition: all 0.3s ease;
}

.step.active .step-icon {
  background: #3b82f6;
  color: white;
}

.step.completed .step-icon {
  background: #10b981;
  color: white;
}

.step span {
  font-size: 0.875rem;
  font-weight: 500;
}

@media (max-width: 480px) {
  .loading-container {
    padding: 2rem;
  }

  .loading-steps {
    gap: 0.5rem;
  }
}
</style>