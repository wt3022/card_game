package security

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ========================================
// 入力バリデーション
// SQLインジェクション、XSS、その他の攻撃を防ぐ
// ========================================

// InputValidator 入力検証
type InputValidator struct{}

// NewInputValidator 新しいInputValidatorを作成
func NewInputValidator() *InputValidator {
	return &InputValidator{}
}

// ValidateUsername ユーザー名を検証
func (v *InputValidator) ValidateUsername(username string) error {
	if len(username) == 0 {
		return fmt.Errorf("username cannot be empty")
	}

	if len(username) > 50 {
		return fmt.Errorf("username too long (max 50 characters)")
	}

	// 英数字、アンダースコア、ハイフンのみ許可
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validUsername.MatchString(username) {
		return fmt.Errorf("username contains invalid characters (only alphanumeric, underscore, and hyphen allowed)")
	}

	// SQLキーワードチェック
	if containsSQLKeywords(username) {
		return fmt.Errorf("username contains prohibited keywords")
	}

	return nil
}

// ValidateCardName カード名を検証
func (v *InputValidator) ValidateCardName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("card name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("card name too long (max 100 characters)")
	}

	// 制御文字をチェック
	for _, r := range name {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return fmt.Errorf("card name contains invalid control characters")
		}
	}

	// SQLキーワードチェック
	if containsSQLKeywords(name) {
		return fmt.Errorf("card name contains prohibited keywords")
	}

	// XSS攻撃パターンチェック
	if containsXSSPatterns(name) {
		return fmt.Errorf("card name contains prohibited patterns")
	}

	return nil
}

// ValidateDeckName デッキ名を検証
func (v *InputValidator) ValidateDeckName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("deck name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("deck name too long (max 100 characters)")
	}

	// 制御文字をチェック
	for _, r := range name {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return fmt.Errorf("deck name contains invalid control characters")
		}
	}

	// SQLキーワードチェック
	if containsSQLKeywords(name) {
		return fmt.Errorf("deck name contains prohibited keywords")
	}

	// XSS攻撃パターンチェック
	if containsXSSPatterns(name) {
		return fmt.Errorf("deck name contains prohibited patterns")
	}

	return nil
}

// ValidateDescription 説明文を検証
func (v *InputValidator) ValidateDescription(description string) error {
	if len(description) > 1000 {
		return fmt.Errorf("description too long (max 1000 characters)")
	}

	// 制御文字をチェック
	for _, r := range description {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return fmt.Errorf("description contains invalid control characters")
		}
	}

	// XSS攻撃パターンチェック
	if containsXSSPatterns(description) {
		return fmt.Errorf("description contains prohibited patterns")
	}

	return nil
}

// ValidateGameID ゲームIDを検証
func (v *InputValidator) ValidateGameID(gameID string) error {
	if len(gameID) == 0 {
		return fmt.Errorf("game ID cannot be empty")
	}

	if len(gameID) > 100 {
		return fmt.Errorf("game ID too long")
	}

	// UUIDまたは英数字とハイフンのみ許可
	validID := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !validID.MatchString(gameID) {
		return fmt.Errorf("game ID contains invalid characters")
	}

	return nil
}

// containsSQLKeywords SQLインジェクション攻撃のキーワードチェック
func containsSQLKeywords(input string) bool {
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
		"EXEC", "EXECUTE", "UNION", "DECLARE", "SCRIPT", "JAVASCRIPT",
		"--", "/*", "*/", "xp_", "sp_",
	}

	upperInput := strings.ToUpper(input)
	for _, keyword := range sqlKeywords {
		if strings.Contains(upperInput, keyword) {
			return true
		}
	}

	return false
}

// containsXSSPatterns XSS攻撃のパターンチェック
func containsXSSPatterns(input string) bool {
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "onerror=", "onload=",
		"onclick=", "onmouseover=", "<iframe", "<object", "<embed",
		"eval(", "expression(", "vbscript:", "data:text/html",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

// SanitizeString 文字列をサニタイズ（危険な文字を除去）
func (v *InputValidator) SanitizeString(input string) string {
	// 制御文字を除去（改行、タブは維持）
	var result strings.Builder
	for _, r := range input {
		if !unicode.IsControl(r) || r == '\n' || r == '\r' || r == '\t' {
			result.WriteRune(r)
		}
	}

	return strings.TrimSpace(result.String())
}
