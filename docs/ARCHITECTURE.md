# アーキテクチャ設計書

## 概要

このプロジェクトはクリーンアーキテクチャを採用しており、ビジネスロジックを中心に据えた保守性と拡張性の高い設計を実現しています。

## クリーンアーキテクチャの原則

### 1. 依存関係逆転の原則

依存関係は常に外側から内側（具象から抽象へ）に向かいます。内側の層は外側の層について何も知りません。

```
Infrastructure → Adapter → Application → Core (Entity + UseCase)
                                              ↑
                                            Port (Interface)
```

### 2. レイヤー構成

#### Core層 (内側)

**Entity** (`internal/core/entity/`)
- ビジネスルールを持つドメインオブジェクト
- 他のどの層にも依存しない
- 例: `Player`, `Card`, `Unit`, `Deck`

**UseCase** (`internal/core/usecase/`)
- アプリケーション固有のビジネスルール
- Entityを操作してビジネスロジックを実現
- 例: ゲームエンジン、戦闘処理、効果処理

**Port** (`internal/core/port/`)
- 外部層とのインターフェース定義
- Repository、Logger、AuthProviderなどのインターフェース
- 依存性逆転の原則を実現

#### Application層

**Service** (`internal/application/service/`)
- ユースケースを組み合わせてアプリケーション機能を提供
- トランザクション管理や複数ユースケースの調整
- 例: `AuthService`, `GameService`, `CardService`

#### Adapter層

**Connect** (`internal/adapter/connect/`)
- Connect-RPCのハンドラーとインターセプター
- 外部からのリクエストをアプリケーション層に橋渡し
- データ変換（Proto ↔ Domain）を担当

**Converter** (`internal/adapter/converter/`)
- データ構造の変換ロジック
- Protobufとドメインモデルのマッピング

#### Infrastructure層 (外側)

**Repository** (`internal/infrastructure/repository/`)
- Portで定義されたインターフェースの実装
- データベースアクセスの具体的な実装（GORM使用）

**Persistence** (`internal/infrastructure/persistence/`)
- データベース接続の管理
- マイグレーション処理

**Auth** (`internal/infrastructure/auth/`)
- 認証の具体的な実装
- JWT生成・検証、パスワードハッシュ化

**Logger** (`internal/infrastructure/logger/`)
- ロギングの具体的な実装

## データフロー

### リクエスト処理の流れ

1. **外部からのリクエスト受信** (Connect Handler)
2. **認証・認可** (Interceptor)
3. **データ変換** (Proto → Domain Model)
4. **アプリケーションサービス呼び出し**
5. **ユースケース実行** (Core)
6. **リポジトリ経由でのデータ永続化**
7. **結果の変換** (Domain Model → Proto)
8. **レスポンス返却**

### 依存性の注入

依存性は `cmd/connect-server/main.go` で解決され、外側から内側に注入されます：

```
Repository実装 → Service → Handler
     ↑              ↑         ↑
   (Port実装)  (Portを要求) (Serviceを要求)
```

## 設計の利点

### 1. テスタビリティ
- Portインターフェースのモック実装で単体テストが容易
- 各層が独立しているため、層ごとのテストが可能

### 2. 保守性
- 関心の分離により、変更の影響範囲が限定される
- ビジネスロジックが外部技術から独立

### 3. 拡張性
- 新しい機能は適切な層に追加するだけ
- インターフェースを維持すれば実装の切り替えが容易

### 4. 技術的負債の軽減
- フレームワークやライブラリへの依存が外側の層に限定
- 技術スタックの変更が容易

## 開発ガイドライン

### 新機能追加時の手順

1. **Entity定義**: ドメインモデルを `internal/core/entity/` に定義
2. **Port定義**: 必要なインターフェースを `internal/core/port/` に定義
3. **UseCase実装**: ビジネスロジックを `internal/core/usecase/` に実装
4. **Repository実装**: データアクセスを `internal/infrastructure/repository/` に実装
5. **Service実装**: 調整ロジックを `internal/application/service/` に実装
6. **Handler追加**: APIエンドポイントを `internal/adapter/connect/` に追加
7. **Proto定義更新**: 必要に応じて `api/proto/` を更新

### レイヤー間のルール

- **Core層は外部に依存しない**: 標準ライブラリとドメイン内のパッケージのみ使用可
- **Portを経由する**: 外部システムへのアクセスは必ずPortインターフェースを通す
- **データ変換はAdapter層**: ProtobufとDomainモデルの変換はConverter層で実施
- **ビジネスロジックはCore層**: 条件分岐やアルゴリズムはCore層に集約

## まとめ

クリーンアーキテクチャの採用により、このプロジェクトは以下を実現しています：

- ビジネスロジックの独立性
- 高いテスタビリティ
- 保守性と拡張性の両立
- 技術的負債の管理

新機能の追加や既存機能の変更時は、常にこのアーキテクチャの原則に従って実装してください。
