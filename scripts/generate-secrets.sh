#!/bin/bash

# ========================================
# セキュアなランダムキー生成スクリプト
# ========================================
# 使用方法:
#   chmod +x scripts/generate-secrets.sh
#   ./scripts/generate-secrets.sh
# ========================================

set -e

echo "=========================================="
echo "セキュアなランダムキー生成"
echo "=========================================="
echo ""

# JWT Secret生成（Base64エンコード、44文字）
echo "JWT_SECRET（32文字以上のランダム文字列）:"
openssl rand -base64 32
echo ""

# MySQL Root Password生成
echo "MYSQL_ROOT_PASSWORD（強力なランダムパスワード）:"
openssl rand -base64 24
echo ""

# Admin Password生成
echo "ADMIN_PASSWORD（初期管理者パスワード）:"
openssl rand -base64 16
echo ""

echo "=========================================="
echo "⚠️  これらの値を.envファイルに設定してください"
echo "⚠️  絶対にGitにコミットしないでください"
echo "=========================================="
