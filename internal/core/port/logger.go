package port

import (
	"fmt"
	"log"
	"os"
)

// ========================================
// ロガーインターフェース
// ロギング機能を定義
// 設計方針:
// - シンプルな3レベル（Debug, Info, Error）
// - 実装も同パッケージに含む（実用性優先）
// - テスト用のNoOpLoggerを提供
// ========================================

// ========================================
// ロガーインターフェース
// ========================================

// ロギング機能
type Logger interface {
	// デバッグレベルのログ
	Debug(format string, args ...any)

	// 情報レベルのログ
	Info(format string, args ...any)

	// エラーレベルのログ
	Error(format string, args ...any)
}

// ========================================
// ConsoleLogger実装
// ========================================

// 標準出力へのシンプルなロガー
type ConsoleLogger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

// 新しいConsoleLoggerを作成
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:  log.New(os.Stdout, "[INFO]  ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// デバッグレベルのログ
func (l *ConsoleLogger) Debug(format string, args ...any) {
	l.debugLogger.Output(2, fmt.Sprintf(format, args...))
}

// 情報レベルのログ
func (l *ConsoleLogger) Info(format string, args ...any) {
	l.infoLogger.Output(2, fmt.Sprintf(format, args...))
}

// エラーレベルのログ
func (l *ConsoleLogger) Error(format string, args ...any) {
	l.errorLogger.Output(2, fmt.Sprintf(format, args...))
}

// ========================================
// NoOpLogger実装（テスト用）
// ========================================

// ログを出力しないロガー
type NoOpLogger struct{}

// 新しいNoOpLoggerを作成
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

// デバッグレベルのログ（何もしない）
func (l *NoOpLogger) Debug(format string, args ...any) {}

// 情報レベルのログ（何もしない）
func (l *NoOpLogger) Info(format string, args ...any) {}

// エラーレベルのログ（何もしない）
func (l *NoOpLogger) Error(format string, args ...any) {}
