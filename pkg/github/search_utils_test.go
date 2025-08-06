package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_hasFilter(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		filterType string
		expected   bool
	}{
		{
			name:       "query has is:issue filter",
			query:      "is:issue bug report",
			filterType: "is",
			expected:   true,
		},
		{
			name:       "query has repo: filter",
			query:      "repo:github/github-mcp-server critical bug",
			filterType: "repo",
			expected:   true,
		},
		{
			name:       "query has multiple is: filters",
			query:      "is:issue is:open bug",
			filterType: "is",
			expected:   true,
		},
		{
			name:       "query has filter at the beginning",
			query:      "is:issue some text",
			filterType: "is",
			expected:   true,
		},
		{
			name:       "query has filter in the middle",
			query:      "some text is:issue more text",
			filterType: "is",
			expected:   true,
		},
		{
			name:       "query has filter at the end",
			query:      "some text is:issue",
			filterType: "is",
			expected:   true,
		},
		{
			name:       "query does not have the filter",
			query:      "bug report critical",
			filterType: "is",
			expected:   false,
		},
		{
			name:       "query has similar text but not the filter",
			query:      "this issue is important",
			filterType: "is",
			expected:   false,
		},
		{
			name:       "empty query",
			query:      "",
			filterType: "is",
			expected:   false,
		},
		{
			name:       "query has label: filter but looking for is:",
			query:      "label:bug critical",
			filterType: "is",
			expected:   false,
		},
		{
			name:       "query has author: filter",
			query:      "author:octocat bug",
			filterType: "author",
			expected:   true,
		},
		{
			name:       "query with complex OR expression",
			query:      "repo:github/github-mcp-server is:issue (label:critical OR label:urgent)",
			filterType: "is",
			expected:   true,
		},
		{
			name:       "query with complex OR expression checking repo",
			query:      "repo:github/github-mcp-server is:issue (label:critical OR label:urgent)",
			filterType: "repo",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasFilter(tt.query, tt.filterType)
			assert.Equal(t, tt.expected, result, "hasFilter(%q, %q) = %v, expected %v", tt.query, tt.filterType, result, tt.expected)
		})
	}
}

func Test_extractRepoFilter(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedOwner string
		expectedRepo  string
		expectedFound bool
	}{
		{
			name:          "query with repo: filter at beginning",
			query:         "repo:github/github-mcp-server is:issue",
			expectedOwner: "github",
			expectedRepo:  "github-mcp-server",
			expectedFound: true,
		},
		{
			name:          "query with repo: filter in middle",
			query:         "is:issue repo:octocat/Hello-World bug",
			expectedOwner: "octocat",
			expectedRepo:  "Hello-World",
			expectedFound: true,
		},
		{
			name:          "query with repo: filter at end",
			query:         "is:issue critical repo:owner/repo-name",
			expectedOwner: "owner",
			expectedRepo:  "repo-name",
			expectedFound: true,
		},
		{
			name:          "query with complex repo name",
			query:         "repo:microsoft/vscode-extension-samples bug",
			expectedOwner: "microsoft",
			expectedRepo:  "vscode-extension-samples",
			expectedFound: true,
		},
		{
			name:          "query without repo: filter",
			query:         "is:issue bug critical",
			expectedOwner: "",
			expectedRepo:  "",
			expectedFound: false,
		},
		{
			name:          "query with malformed repo: filter (no slash)",
			query:         "repo:github bug",
			expectedOwner: "",
			expectedRepo:  "",
			expectedFound: false,
		},
		{
			name:          "empty query",
			query:         "",
			expectedOwner: "",
			expectedRepo:  "",
			expectedFound: false,
		},
		{
			name:          "query with multiple repo: filters (should match first)",
			query:         "repo:github/first repo:octocat/second",
			expectedOwner: "github",
			expectedRepo:  "first",
			expectedFound: true,
		},
		{
			name:          "query with repo: in text but not as filter",
			query:         "this repo: is important",
			expectedOwner: "",
			expectedRepo:  "",
			expectedFound: false,
		},
		{
			name:          "query with complex OR expression",
			query:         "repo:github/github-mcp-server is:issue (label:critical OR label:urgent)",
			expectedOwner: "github",
			expectedRepo:  "github-mcp-server",
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, found := extractRepoFilter(tt.query)
			assert.Equal(t, tt.expectedOwner, owner, "extractRepoFilter(%q) owner = %q, expected %q", tt.query, owner, tt.expectedOwner)
			assert.Equal(t, tt.expectedRepo, repo, "extractRepoFilter(%q) repo = %q, expected %q", tt.query, repo, tt.expectedRepo)
			assert.Equal(t, tt.expectedFound, found, "extractRepoFilter(%q) found = %v, expected %v", tt.query, found, tt.expectedFound)
		})
	}
}

func Test_hasSpecificFilter(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		filterType  string
		filterValue string
		expected    bool
	}{
		{
			name:        "query has exact is:issue filter",
			query:       "is:issue bug report",
			filterType:  "is",
			filterValue: "issue",
			expected:    true,
		},
		{
			name:        "query has is:open but looking for is:issue",
			query:       "is:open bug report",
			filterType:  "is",
			filterValue: "issue",
			expected:    false,
		},
		{
			name:        "query has both is:issue and is:open, looking for is:issue",
			query:       "is:issue is:open bug",
			filterType:  "is",
			filterValue: "issue",
			expected:    true,
		},
		{
			name:        "query has both is:issue and is:open, looking for is:open",
			query:       "is:issue is:open bug",
			filterType:  "is",
			filterValue: "open",
			expected:    true,
		},
		{
			name:        "query has is:issue at the beginning",
			query:       "is:issue some text",
			filterType:  "is",
			filterValue: "issue",
			expected:    true,
		},
		{
			name:        "query has is:issue in the middle",
			query:       "some text is:issue more text",
			filterType:  "is",
			filterValue: "issue",
			expected:    true,
		},
		{
			name:        "query has is:issue at the end",
			query:       "some text is:issue",
			filterType:  "is",
			filterValue: "issue",
			expected:    true,
		},
		{
			name:        "query does not have is:issue",
			query:       "bug report critical",
			filterType:  "is",
			filterValue: "issue",
			expected:    false,
		},
		{
			name:        "query has similar text but not the exact filter",
			query:       "this issue is important",
			filterType:  "is",
			filterValue: "issue",
			expected:    false,
		},
		{
			name:        "empty query",
			query:       "",
			filterType:  "is",
			filterValue: "issue",
			expected:    false,
		},
		{
			name:        "partial match should not count",
			query:       "is:issues bug", // "issues" vs "issue"
			filterType:  "is",
			filterValue: "issue",
			expected:    false,
		},
		{
			name:        "complex query with parentheses",
			query:       "repo:github/github-mcp-server is:issue (label:critical OR label:urgent)",
			filterType:  "is",
			filterValue: "issue",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasSpecificFilter(tt.query, tt.filterType, tt.filterValue)
			assert.Equal(t, tt.expected, result, "hasSpecificFilter(%q, %q, %q) = %v, expected %v", tt.query, tt.filterType, tt.filterValue, result, tt.expected)
		})
	}
}
