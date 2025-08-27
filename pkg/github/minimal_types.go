package github

// MinimalUser is the output type for user and organization search results.
type MinimalUser struct {
	Login      string       `json:"login"`
	ID         int64        `json:"id,omitempty"`
	ProfileURL string       `json:"profile_url,omitempty"`
	AvatarURL  string       `json:"avatar_url,omitempty"`
	Details    *UserDetails `json:"details,omitempty"` // Optional field for additional user details
}

// MinimalSearchUsersResult is the trimmed output type for user search results.
type MinimalSearchUsersResult struct {
	TotalCount        int           `json:"total_count"`
	IncompleteResults bool          `json:"incomplete_results"`
	Items             []MinimalUser `json:"items"`
}

// MinimalRepository is the trimmed output type for repository objects to reduce verbosity.
type MinimalRepository struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	FullName      string   `json:"full_name"`
	Description   string   `json:"description,omitempty"`
	HTMLURL       string   `json:"html_url"`
	CloneURL      string   `json:"clone_url,omitempty"`
	Language      string   `json:"language,omitempty"`
	Stars         int      `json:"stargazers_count"`
	Forks         int      `json:"forks_count"`
	OpenIssues    int      `json:"open_issues_count"`
	UpdatedAt     string   `json:"updated_at,omitempty"`
	CreatedAt     string   `json:"created_at,omitempty"`
	Topics        []string `json:"topics,omitempty"`
	Private       bool     `json:"private"`
	Fork          bool     `json:"fork"`
	Archived      bool     `json:"archived"`
	DefaultBranch string   `json:"default_branch,omitempty"`
}

// MinimalSearchRepositoriesResult is the trimmed output type for repository search results.
type MinimalSearchRepositoriesResult struct {
	TotalCount        int                 `json:"total_count"`
	IncompleteResults bool                `json:"incomplete_results"`
	Items             []MinimalRepository `json:"items"`
}

// MinimalCommitAuthor represents commit author information.
type MinimalCommitAuthor struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Date  string `json:"date,omitempty"`
}

// MinimalCommitInfo represents core commit information.
type MinimalCommitInfo struct {
	Message   string               `json:"message"`
	Author    *MinimalCommitAuthor `json:"author,omitempty"`
	Committer *MinimalCommitAuthor `json:"committer,omitempty"`
}

// MinimalCommitStats represents commit statistics.
type MinimalCommitStats struct {
	Additions int `json:"additions,omitempty"`
	Deletions int `json:"deletions,omitempty"`
	Total     int `json:"total,omitempty"`
}

// MinimalCommitFile represents a file changed in a commit.
type MinimalCommitFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status,omitempty"`
	Additions int    `json:"additions,omitempty"`
	Deletions int    `json:"deletions,omitempty"`
	Changes   int    `json:"changes,omitempty"`
}

// MinimalCommit is the trimmed output type for commit objects.
type MinimalCommit struct {
	SHA       string              `json:"sha"`
	HTMLURL   string              `json:"html_url"`
	Commit    *MinimalCommitInfo  `json:"commit,omitempty"`
	Author    *MinimalUser        `json:"author,omitempty"`
	Committer *MinimalUser        `json:"committer,omitempty"`
	Stats     *MinimalCommitStats `json:"stats,omitempty"`
	Files     []MinimalCommitFile `json:"files,omitempty"`
}

// MinimalRelease is the trimmed output type for release objects.
type MinimalRelease struct {
	ID          int64        `json:"id"`
	TagName     string       `json:"tag_name"`
	Name        string       `json:"name,omitempty"`
	Body        string       `json:"body,omitempty"`
	HTMLURL     string       `json:"html_url"`
	PublishedAt string       `json:"published_at,omitempty"`
	Prerelease  bool         `json:"prerelease"`
	Draft       bool         `json:"draft"`
	Author      *MinimalUser `json:"author,omitempty"`
}

// MinimalBranch is the trimmed output type for branch objects.
type MinimalBranch struct {
	Name      string `json:"name"`
	SHA       string `json:"sha"`
	Protected bool   `json:"protected"`
}

// Minimal response types for create/edit operations

// MinimalCreateResponse represents a minimal response for resource creation operations.
type MinimalCreateResponse struct {
	URL    string `json:"url"`
	ID     int64  `json:"id,omitempty"`
	Number int    `json:"number,omitempty"`
	Name   string `json:"name,omitempty"`
	State  string `json:"state,omitempty"`
}

// MinimalPullRequestResponse represents a minimal response for pull request operations.
type MinimalPullRequestResponse struct {
	URL    string `json:"url"`
	Number int    `json:"number"`
	State  string `json:"state"`
	Title  string `json:"title,omitempty"`
}

// MinimalRepositoryResponse represents a minimal response for repository operations.
type MinimalRepositoryResponse struct {
	URL      string `json:"url"`
	CloneURL string `json:"clone_url"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

// MinimalIssueResponse represents a minimal response for issue operations.
type MinimalIssueResponse struct {
	URL    string `json:"url"`
	Number int    `json:"number"`
	State  string `json:"state"`
	Title  string `json:"title,omitempty"`
}

// MinimalUpdateResponse represents a minimal response for update operations.
type MinimalUpdateResponse struct {
	URL     string `json:"url"`
	Updated bool   `json:"updated"`
	Message string `json:"message,omitempty"`
}

// MinimalGistResponse represents a minimal response for gist operations.
type MinimalGistResponse struct {
	URL         string `json:"url"`
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	Public      bool   `json:"public"`
}
