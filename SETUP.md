# GitHub MCP Server Setup Guide

このガイドでは、GitHub MCP ServerをClaude Desktopで使用するための設定方法を説明します。

## セットアップ方法

### 1. GitHub Personal Access Token (PAT) の作成

1. GitHubにログインし、Settings → Developer settings → Personal access tokens → Tokens (classic) に移動
2. "Generate new token" をクリック
3. 必要なスコープを選択:
   - `repo` - リポジトリへのフルアクセス
   - `workflow` - GitHub Actionsワークフローへのアクセス
   - `read:org` - 組織情報の読み取り（必要に応じて）
   - `gist` - Gistへのアクセス（必要に応じて）
4. トークンを生成し、安全な場所に保存

### 2. Claude Desktop設定の更新

1. Claude Desktopを終了
2. 設定ファイルを編集:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`

3. `GITHUB_PERSONAL_ACCESS_TOKEN` の値を実際のトークンに置き換え:
   ```json
   "github": {
     "command": "docker",
     "args": [
       "run",
       "-i",
       "--rm",
       "-e",
       "GITHUB_PERSONAL_ACCESS_TOKEN",
       "ghcr.io/github/github-mcp-server"
     ],
     "env": {
       "GITHUB_PERSONAL_ACCESS_TOKEN": "YOUR_ACTUAL_TOKEN_HERE"
     }
   }
   ```

### 3. Dockerのインストール確認

GitHub MCP ServerはDockerを使用して実行されます。Dockerがインストールされていることを確認してください:

```bash
docker --version
```

Dockerがインストールされていない場合は、[Docker Desktop](https://www.docker.com/products/docker-desktop/)からインストールしてください。

### 4. イメージのプル（オプション）

事前にDockerイメージをプルしておくことで、初回起動時の時間を短縮できます:

```bash
docker pull ghcr.io/github/github-mcp-server
```

### 5. Claude Desktopを再起動

設定を反映させるために、Claude Desktopを再起動してください。

## 使用方法

Claude Desktopで以下のような操作が可能になります:

- リポジトリの閲覧とコード検索
- Issue/PRの作成と管理
- GitHub Actionsワークフローの監視
- コードスキャニングアラートの確認
- Dependabotアラートの管理

## トラブルシューティング

### Dockerが起動しない場合
- Docker Desktopが実行されていることを確認
- `docker run hello-world` でDockerが正常に動作することを確認

### 認証エラーが発生する場合
- PATが正しく設定されているか確認
- PATに必要なスコープが付与されているか確認
- トークンの有効期限が切れていないか確認

### MCPサーバーが表示されない場合
- Claude Desktopを完全に終了して再起動
- 設定ファイルのJSONが正しい形式であることを確認
- ログファイルでエラーを確認:
  - macOS: `~/Library/Logs/Claude/`