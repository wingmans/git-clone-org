package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	execute "github.com/alexellis/go-execute/pkg/v1"
	cc "wingmen.io/git-clone-all/pkg/cc"
)

type Context struct {
	Clean   bool
	Debug   bool
	Noop    bool
	Verbose bool
}

func NewContext(clean, debug, noop, verbose bool) Context {

	return Context{
		Clean:   clean,
		Debug:   debug,
		Noop:    noop,
		Verbose: verbose,
	}
}

type Repositories []Repository

type Repository struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"`
	HTMLURL  string `json:"html_url"`
	Archived bool   `json:"archived"`
}

func (r Repository) ToString() string {
	return fmt.Sprintf("[%d] %s %s", r.ID, r.FullName, r.CloneURL)
}

func (r Repository) existsIn(repos Repositories) bool {
	for _, repo := range repos {
		if repo.ID == r.ID && repo.FullName == r.FullName {
			return true
		}
	}
	return false
}

func (repos Repositories) In(otherRepos Repositories) Repositories {
	var result Repositories
	for _, repo := range repos {
		if repo.existsIn(otherRepos) {
			result = append(result, repo)
		}
	}
	return result
}

func (repos Repositories) NotIn(otherRepos Repositories) Repositories {
	var result Repositories
	for _, repo := range repos {
		if !repo.existsIn(otherRepos) {
			result = append(result, repo)
		}
	}
	return result
}

// builds a list of github repositories based on the filter.
func getRepositoryList(ctx cc.CommonCtx, baseURL string) (Repositories, error) {

	// the PAT to query the GITHUB repos
	tokkie := os.Getenv("GIT_TOKEN")
	if tokkie == "" {
		log.Fatal("the environment variable GIT_TOKEN is not set.")
	}

	ctx.LogInfo("Fetching repository list.")
	var repositories Repositories
	var page []Repository

	u, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("Failed to parse URL: %v", err)
	}

	query := u.Query()
	query.Set("per_page", "40")
	pageNumber := 1
	for {
		query.Set("page", fmt.Sprintf("%d", pageNumber))
		u.RawQuery = query.Encode()
		req, err := http.NewRequestWithContext(context.Background(), "GET", u.String(), http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		req.Header.Add("Accept", "application/vnd.github.inertia-preview+json")
		req.Header.Add("Authorization", fmt.Sprintf("token %s", tokkie))
		ctx.LogTrace(u.String())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %v", err)
		}
		defer resp.Body.Close() // TRY TO FIX THIS ONE TOO

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}

		ctx.LogTrace(string(body))
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}

		if len(page) == 0 {
			break
		}

		repositories = append(repositories, page...)
		pageNumber++
	}

	return repositories, nil
}

func getRepositoryFromOrgs(ctx cc.CommonCtx, orgName string) (Repositories, error) {
	ep := fmt.Sprintf("https://api.github.com/orgs/%s/repos", orgName)
	ctx.LogTrace(fmt.Sprintf("Fetching repositories from: %s", ep))
	return getRepositoryList(ctx, ep)
}

func getRepositoryFromUser(ctx cc.CommonCtx, userName string) (Repositories, error) {
	ep := fmt.Sprintf("https://api.github.com/users/%s/repos", userName)
	ctx.LogTrace(fmt.Sprintf("Fetching repositories from: %s", ep))
	return getRepositoryList(ctx, ep)
}

func CloneMultiple(ctx cc.CommonCtx, mode, filter string) error {

	var repositories Repositories
	if mode == "usr" {
		repositories, _ = getRepositoryFromUser(ctx, filter)
		ctx.LogTrace(fmt.Sprintf("cloning %d repositories...\n", len(repositories)))
	} else if mode == "org" {
		repositories, _ = getRepositoryFromOrgs(ctx, filter)
		ctx.LogTrace(fmt.Sprintf("cloning %d repositories...\n", len(repositories)))
	} else {
		return fmt.Errorf("unknown mode (%s)", mode)
	}

	for _, r := range repositories {
		err := cloneRepository(ctx, filter, r)
		if err != nil {
			return fmt.Errorf("error cloning repository %s ", r.FullName)
		}
	}

	return nil
}

func cloneRepository(ctx cc.CommonCtx, orgName string, repository Repository) error {

	baseFolder := orgName
	if !FolderExists(baseFolder) {
		if err := os.MkdirAll(baseFolder, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create base folder %s: %w", baseFolder, err)
		}
	}

	repositoryPath := repository.FullName
	if !PathExists(repositoryPath) {
		if ctx.Noop {
			ctx.LogInfo(fmt.Sprintf("would clone %s", repository.FullName))
			return nil
		}

		ctx.LogInfo(fmt.Sprintf("cloning %s", repository.FullName))
		cmd := execute.ExecTask{
			Command: "git",
			Args:    []string{"-C", baseFolder, "clone", repository.CloneURL},
			Shell:   false,
		}

		_, err := cmd.Execute()
		if err != nil {
			return fmt.Errorf("failed to clone %s: %w", repository.FullName, err)
		}
	} else {
		if ctx.Noop {
			ctx.LogInfo(fmt.Sprintf("would skip %s (already exists).", repository.FullName))
			return nil
		}

		ctx.LogInfo(fmt.Sprintf("skipping %s (already exists).", repository.FullName))
	}

	return nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func FolderExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// LoadExcludedRepos reads the .gitexclude file and returns a list of directories to exclude.
func LoadExcludedRepos(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", filename, err)
		return []string{}
	}
	defer file.Close()

	var repos []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			repos = append(repos, line)
		}
	}
	return repos
}

func IsExcluded(dir string, excludedRepos []string) bool {
	for _, repo := range excludedRepos {
		if dir == repo {
			return true
		}
	}
	return false
}

func IsGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

func GitPull(dir string) error {
	cmd := exec.Command("git", "-C", dir, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
