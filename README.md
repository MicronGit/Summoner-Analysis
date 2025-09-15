# Summoner Analysis

League of Legends（LoL）のプレイヤーのランク戦成績を分析し、詳細な統計データをJSON形式で出力するツールです。

## 機能

- **アカウント情報取得**: Riot IDによるプレイヤー検索
- **ランク戦履歴分析**: 最大100試合のランク戦データを取得・分析
- **詳細統計計算**: KDA、勝率、チャンピオン別成績、ポジション統計など
- **JSON出力**: 分析結果を構造化されたJSON形式で保存
- **レート制限対応**: Riot API のレート制限に対応した安全な通信
- **キャンセル対応**: Ctrl+C での処理中断機能
- **タイムアウト設定**: 15分のタイムアウト設定で長時間の処理を制御

## 必要な環境

- Go 1.25.1 以上
- Riot API キー（[Riot Developer Portal](https://developer.riotgames.com/) で取得）

## セットアップ

1. **リポジトリのクローン**
   ```bash
   git clone https://github.com/MicronGit/Summoner-Analysis.git
   cd Summoner-Analysis
   ```

2. **依存関係のインストール**
   ```bash
   go mod download
   ```

3. **環境変数の設定**

   `.env` ファイルを作成し、以下の内容を設定：
   ```env
   RIOT_API_KEY=YOUR_RIOT_API_KEY_HERE
   RIOT_REGION=asia
   ```

   **利用可能なリージョン:**
   - `asia` - アジア（日本、韓国など）
   - `americas` - 北米、南米
   - `europe` - ヨーロッパ

## 使用方法

### 基本的な使用方法

```bash
go run cmd/main/main.go
```

現在のバージョンでは、分析対象のプレイヤーはソースコード内で指定されています（`cmd/main/main.go:45-46`）：

```go
gameName := "そっちん"
tagLine := "JP1"
```

### カスタマイズ

分析対象を変更するには、`cmd/main/main.go` の以下の行を編集してください：

```go
gameName := "YourGameName"
tagLine := "YourTagLine"
```

## 出力データ

### 1. 詳細データ (`*_analysis_*.json`)

プレイヤーの全ランク戦試合の詳細データを含むファイル：

```json
{
  "account": {
    "puuid": "...",
    "gameName": "プレイヤー名",
    "tagLine": "JP1"
  },
  "matchHistory": [
    {
      "metadata": { "matchId": "...", ... },
      "info": {
        "gameCreation": 1234567890,
        "participants": [...],
        ...
      }
    }
  ],
  "generatedAt": "2024-01-01T12:00:00Z",
  "totalMatches": 50,
  "matchType": "ranked"
}
```

### 2. 統計データ (`*_stats_*.json`)

分析結果をまとめた統計データ：

```json
{
  "playerInfo": {
    "gameName": "プレイヤー名",
    "tagLine": "JP1"
  },
  "totalMatches": 50,
  "winRate": 64.5,
  "averageKDA": {
    "kills": 7.2,
    "deaths": 5.1,
    "assists": 8.9,
    "kdaRatio": 3.16
  },
  "rankPerformance": {
    "averageVisionScore": 45.3,
    "averageGoldEarned": 12450.8,
    "averageCSPerMin": 6.7
  },
  "mostPlayedChampions": [
    {
      "championName": "Jinx",
      "gamesPlayed": 15,
      "winRate": 73.3,
      "averageKDA": { ... }
    }
  ],
  "positionStats": {
    "BOTTOM": 35,
    "MIDDLE": 10,
    "TOP": 5
  },
  "recentForm": {
    "last10Games": {
      "winRate": 70.0,
      "averageKDA": { ... }
    },
    "last5Games": {
      "winRate": 80.0,
      "averageKDA": { ... }
    }
  }
}
```

## プロジェクト構成

```
Summoner-Analysis/
├── cmd/
│   └── main/
│       └── main.go              # メインエントリーポイント
├── internal/
│   ├── config/
│   │   └── config.go            # 設定管理
│   ├── riot/
│   │   ├── client.go            # Riot API クライアント
│   │   ├── types.go             # API データ型定義
│   │   ├── constants.go         # キューID等の定数
│   │   ├── ratelimiter.go       # レート制限管理
│   │   └── errors.go            # エラー処理
│   └── output/
│       ├── json.go              # JSON出力処理
│       └── types.go             # 統計データ型定義
├── output/                      # 分析結果出力ディレクトリ
├── .env                         # 環境変数設定
├── go.mod                       # Go モジュール定義
└── README.md                    # このファイル
```

## 技術的特徴

### Riot API 対応

- **レート制限**: 自動的なレート制限管理とリトライ機能
- **エラーハンドリング**: 429エラー、5xxエラーに対する適切な再試行処理
- **コンテキスト対応**: タイムアウトとキャンセレーション機能

### データ処理

- **ランク戦フィルタリング**: ソロ/デュオランク（420）とフレックスランク（440）のみを分析対象
- **統計計算**: KDA、勝率、チャンピオン別成績、直近フォームなどの詳細分析
- **安全なファイル名**: 特殊文字を含むプレイヤー名に対応した安全なファイル名生成

### パフォーマンス

- **並行処理**: レート制限内での効率的なAPI呼び出し
- **進捗表示**: リアルタイムでの処理進捗とETA表示
- **メモリ効率**: 大量のマッチデータを効率的に処理

## ライセンス

このプロジェクトは個人使用を目的としています。Riot Games API利用規約に準拠してください。

## 注意事項

- Riot API キーは絶対に公開しないでください
- API の利用制限に注意してください（開発者キーは1分間に100リクエスト、1日2万リクエスト）
- 大量のデータ取得時はレート制限により処理に時間がかかる場合があります

## トラブルシューティング

### よくあるエラー

1. **"RIOT_API_KEY が設定されていません"**
   - `.env` ファイルに正しいAPIキーが設定されているか確認

2. **"APIエラー (status: 403)"**
   - APIキーが無効または期限切れです。新しいキーを取得してください

3. **"レート制限エラー"**
   - API呼び出し頻度が制限に達しています。しばらく待ってから再実行してください

4. **"アカウント取得エラー"**
   - プレイヤー名とタグライン（例：GameName#JP1）が正しいか確認してください

## 開発者向け情報

### ビルド

```bash
go build -o summoner-analysis cmd/main/main.go
```

### テスト実行

```bash
go test ./...
```

### 新しい統計指標の追加

統計指標を追加したい場合は、以下のファイルを編集してください：

- `internal/output/types.go` - 新しい統計型の定義
- `internal/output/json.go` - 統計計算ロジックの実装