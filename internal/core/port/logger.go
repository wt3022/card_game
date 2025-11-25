package port

// ========================================
// ロガーインターフェース
// ロギング機能を定義
// 設計方針:
// - シンプルな3レベル（Debug, Info, Error）
// - 実装はinfrastructure層に配置
// ========================================

// Logger ロギング機能
type Logger interface {
	// デバッグレベルのログ
	Debug(format string, args ...any)

	// 情報レベルのログ
	Info(format string, args ...any)

	// エラーレベルのログ
	Error(format string, args ...any)
}
