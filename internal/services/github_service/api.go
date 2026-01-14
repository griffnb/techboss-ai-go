package github_service

import (
	"context"

	"github.com/google/go-github/v66/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// APIService provides GitHub API operations using installation tokens
type APIService struct {
	authService *AuthService
}

// NewAPIService creates a new GitHub API service
func NewAPIService(authService *AuthService) *APIService {
	return &APIService{
		authService: authService,
	}
}

// GetClient creates an authenticated GitHub client for the specified installation
func (s *APIService) GetClient(ctx context.Context, installationID string) (*github.Client, error) {
	if installationID == "" {
		return nil, errors.New("installationID cannot be empty")
	}

	// Get installation token from auth service
	token, err := s.authService.GetInstallationToken(ctx, installationID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get installation token")
	}

	// Create OAuth2 token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Create GitHub client
	client := github.NewClient(tc)

	return client, nil
}

// GetRepository retrieves a repository by owner and name
func (s *APIService) GetRepository(ctx context.Context, installationID, owner, repo string) (*github.Repository, error) {
	if installationID == "" {
		return nil, errors.New("installationID cannot be empty")
	}

	client, err := s.GetClient(ctx, installationID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create GitHub client")
	}

	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get repository %s/%s", owner, repo)
	}

	return repository, nil
}

// ListRepositories lists all repositories accessible to the installation
func (s *APIService) ListRepositories(ctx context.Context, installationID string) ([]*github.Repository, error) {
	if installationID == "" {
		return nil, errors.New("installationID cannot be empty")
	}

	client, err := s.GetClient(ctx, installationID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create GitHub client")
	}

	// List repositories for the installation
	opts := &github.ListOptions{
		PerPage: 100,
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Apps.ListRepos(ctx, opts)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list repositories")
		}

		allRepos = append(allRepos, repos.Repositories...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// GetBranch retrieves a branch by name
func (s *APIService) GetBranch(ctx context.Context, installationID, owner, repo, branch string) (*github.Branch, error) {
	if installationID == "" {
		return nil, errors.New("installationID cannot be empty")
	}

	client, err := s.GetClient(ctx, installationID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create GitHub client")
	}

	branchInfo, _, err := client.Repositories.GetBranch(ctx, owner, repo, branch, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get branch %s", branch)
	}

	return branchInfo, nil
}

// CreateBranch creates a new branch from a SHA
func (s *APIService) CreateBranch(ctx context.Context, installationID, owner, repo, branch, sha string) error {
	if installationID == "" {
		return errors.New("installationID cannot be empty")
	}

	client, err := s.GetClient(ctx, installationID)
	if err != nil {
		return errors.Wrapf(err, "failed to create GitHub client")
	}

	// Create reference for the new branch
	ref := &github.Reference{
		Ref: github.String("refs/heads/" + branch),
		Object: &github.GitObject{
			SHA: github.String(sha),
		},
	}

	_, _, err = client.Git.CreateRef(ctx, owner, repo, ref)
	if err != nil {
		return errors.Wrapf(err, "failed to create branch %s", branch)
	}

	return nil
}

// CreatePullRequest creates a new pull request
func (s *APIService) CreatePullRequest(
	ctx context.Context,
	installationID, owner, repo string,
	pr *github.NewPullRequest,
) (*github.PullRequest, error) {
	if installationID == "" {
		return nil, errors.New("installationID cannot be empty")
	}

	client, err := s.GetClient(ctx, installationID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create GitHub client")
	}

	pullRequest, _, err := client.PullRequests.Create(ctx, owner, repo, pr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create pull request")
	}

	return pullRequest, nil
}
