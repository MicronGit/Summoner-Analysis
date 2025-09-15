import axios, { AxiosResponse } from 'axios'
import type {
  AnalysisRequest,
  AnalysisResponse,
  PlayerStats
} from '../types'

const api = axios.create({
  baseURL: '/api',
  timeout: 15 * 60 * 1000, // 15分タイムアウト
})

export class ApiService {
  // プレイヤー分析実行
  static async analyzePlayer(request: AnalysisRequest): Promise<PlayerStats> {
    try {
      const response: AxiosResponse<AnalysisResponse> = await api.post('/analyze', request)

      if (!response.data.success) {
        throw new Error(response.data.error || '分析に失敗しました')
      }

      if (!response.data.data) {
        throw new Error('データが取得できませんでした')
      }

      return response.data.data
    } catch (error) {
      if (axios.isAxiosError(error)) {
        if (error.code === 'ECONNABORTED') {
          throw new Error('リクエストがタイムアウトしました。時間をおいて再試行してください。')
        }
        if (error.response?.status === 404) {
          throw new Error('プレイヤーが見つかりませんでした。名前とタグラインを確認してください。')
        }
        if (error.response?.status === 403) {
          throw new Error('API キーが無効です。設定を確認してください。')
        }
        if (error.response?.status === 429) {
          throw new Error('レート制限に達しました。しばらく待ってから再試行してください。')
        }
        throw new Error(error.response?.data?.error || 'サーバーエラーが発生しました')
      }
      throw error
    }
  }

  // ヘルスチェック
  static async healthCheck(): Promise<boolean> {
    try {
      const response = await api.get('/health')
      return response.status === 200
    } catch {
      return false
    }
  }
}

export default ApiService