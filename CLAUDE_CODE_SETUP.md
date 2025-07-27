# Claude Code でのGitHub MCP Server設定ガイド

## 🎯 設定概要

Claude Code CLI で GitHub MCP Server を使用するための完全設定ガイド

## 📂 現在の設定状況

✅ **MCP設定ファイル作成済み**: `/Users/shunsuke/Dev/.claude/mcp.json`
✅ **GitHub MCP Server バイナリ**: `/Users/shunsuke/Dev/organized/mcp-servers/github-mcp-server/github-mcp-server`

## 🔐 GitHub Personal Access Token の準備

### 1. 新しいトークンの作成
⚠️ **重要**: 以前公開されたトークンは無効化してください

1. [GitHub Personal Access Tokens](https://github.com/settings/personal-access-tokens/new) にアクセス
2. 新しいトークンを作成
3. 必要なスコープを選択:
   ```
   ✅ repo - リポジトリへのフルアクセス
   ✅ read:packages - パッケージ読み取り
   ✅ read:org - 組織情報読み取り (オプション)
   ✅ workflow - GitHub Actions (オプション)
   ```

### 2. 安全な環境変数設定

```bash
# ~/.bashrc または ~/.zshrc に追加
export GITHUB_PAT="your_new_token_here"

# 設定を反映
source ~/.bashrc  # または source ~/.zshrc

# 確認
echo $GITHUB_PAT
```

### 3. Server-Sent Events (SSE) トランスポートを使用する場合

リモートサーバーとしてSSE経由で接続する場合：

```bash
# GitHub MCP Server (リモート)
claude mcp add --transport sse github https://api.githubcopilot.com/mcp/ --header "Authorization: Bearer your_github_pat_here"
```

### 4. 設定されたサーバーの確認

```bash
# MCP サーバーのリスト表示
claude mcp list

# GitHub MCP サーバーの詳細確認
claude mcp get github

# MCPサーバーのステータス確認
/mcp
```

### 5. スコープ設定の選択

#### ローカルスコープ（デフォルト - プロジェクト固有）
```bash
claude mcp add github -s local ...
```

#### プロジェクトスコープ（チーム共有用）
```bash
claude mcp add github -s project ...
```

#### ユーザースコープ（全プロジェクトで利用）
```bash
claude mcp add github -s user ...
```

## 使用方法

### GitHub リソースの参照

Claude Codeでは @ メンションを使用してGitHubリソースを参照できます：

```
> @github:issue://123 を分析して修正案を提案してください
> @github:pr://456 のコードレビューをしてください
> @github:repo://owner/repo-name の構造を説明してください
```

### GitHub MCP プロンプトの使用

スラッシュコマンドとしてGitHub MCPの機能を利用：

```
> /mcp__github__list_prs
> /mcp__github__pr_review 123
> /mcp__github__create_issue "バグ修正" high
```

## 認証とセキュリティ

### OAuth認証（推奨）

OAuth認証を使用する場合：

1. SSEサーバーとして追加
2. `/mcp` コマンドで認証メニューを開く
3. ブラウザでOAuth認証を完了

### 環境変数での認証

`.mcp.json` ファイルで環境変数を使用：

```json
{
  "mcpServers": {
    "github": {
      "type": "stdio",
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "GITHUB_PERSONAL_ACCESS_TOKEN", "ghcr.io/github/github-mcp-server"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PAT:-default_token}"
      }
    }
  }
}
```

## トラブルシューティング

### よくある問題

1. **Docker エラー**
   - Docker Desktopが起動していることを確認
   - `docker run hello-world` でDockerをテスト

2. **認証エラー**
   - PATの有効期限を確認
   - 必要なスコープが付与されているか確認

3. **MCP サーバーが認識されない**
   - `claude mcp list` で設定を確認
   - Claude Codeを再起動

### ログの確認

```bash
# Claude Code のログを確認
claude --verbose mcp list
```

## 参考

- [MCP Protocol Documentation](https://modelcontextprotocol.io/docs)
- [GitHub MCP Server Documentation](https://github.com/github/github-mcp-server)
- [Claude Code MCP Configuration Guide](/Users/shunsuke/Dev/claudecode_mcp_config.md)