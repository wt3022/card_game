package entity

// ========================================
// 認証関連の値オブジェクト
// ========================================

// LoginRequest ログインリクエスト
type LoginRequest struct {
	Username string
	Password string
}

// LoginResponse ログインレスポンス
type LoginResponse struct {
	AccessToken string
	UserID      string
	Username    string
	Role        string
}
