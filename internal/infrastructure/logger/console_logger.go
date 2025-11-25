package logger

import (
	"fmt"
	"log"
	"os"

	"card_game/internal/core/port"
)

// ========================================
// ConsoleLogger実装
// ========================================

// ConsoleLogger 標準出力へのシンプルなロガー
type ConsoleLogger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

// NewConsoleLogger 新しいConsoleLoggerを作成
func NewConsoleLogger() port.Logger {
	return &ConsoleLogger{
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:  log.New(os.Stdout, "[INFO]  ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Debug デバッグレベルのログ
func (l *ConsoleLogger) Debug(format string, args ...any) {
	l.debugLogger.Output(2, fmt.Sprintf(format, args...))
}

// Info 情報レベルのログ
func (l *ConsoleLogger) Info(format string, args ...any) {
	l.infoLogger.Output(2, fmt.Sprintf(format, args...))
}

// Error エラーレベルのログ
func (l *ConsoleLogger) Error(format string, args ...any) {
	l.errorLogger.Output(2, fmt.Sprintf(format, args...))
}

// ========================================
// NoOpLogger実装（テスト用）
// ========================================

// NoOpLogger ログを出力しないロガー
type NoOpLogger struct{}

// NewNoOpLogger 新しいNoOpLoggerを作成
func NewNoOpLogger() port.Logger {
	return &NoOpLogger{}
}

// Debug デバッグレベルのログ（何もしない）
func (l *NoOpLogger) Debug(format string, args ...any) {}

// Info 情報レベルのログ（何もしない）
func (l *NoOpLogger) Info(format string, args ...any) {}

// Error エラーレベルのログ（何もしない）
func (l *NoOpLogger) Error(format string, args ...any) {}
