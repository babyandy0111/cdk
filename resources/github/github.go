package github

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v37/github"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type GithubClient struct {
	Client        *github.Client
	Organization  string
	OrgPublicKey  string
	EnvCollection map[string]string
}

func New(username, userPassword, organization string) *GithubClient {
	auth := github.BasicAuthTransport{
		Username:  username,
		Password:  userPassword,
		Transport: nil,
	}
	client := github.NewClient(auth.Client())

	outputClient := &GithubClient{
		Client:        client,
		Organization:  organization,
		EnvCollection: make(map[string]string, 0),
	}

	publicKey, err := outputClient.GetOrgPublicKey()
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	outputClient.OrgPublicKey = *publicKey

	if err != nil {
		return nil
	}
	return outputClient
}

// 拉出/設定即將寫入到 github 的環境變數
func (client *GithubClient) SetEnvironmentVariable(key, value string) error {
	if client.EnvCollection[key] != "" {
		return errors.New(fmt.Sprintf("Variable %s is existed."))
	}
	client.EnvCollection[key] = value
	return nil
}
func (client *GithubClient) GetEnvironmentVariable(key string) string {
	return client.EnvCollection[key]
}
func (client *GithubClient) GetEnvironmentVariables() map[string]string {
	return client.EnvCollection
}

func encryptedValue(key, value string) (string, error) {
	e := exec.Command("/usr/local/bin/node", []string{os.Getenv("PWD") + "/tools/sodium/index.js", key, value}...)
	out, err := e.CombinedOutput()
	if err != nil {
		return "", err
	}
	output := strings.Trim(string(out), "\n\r")
	fmt.Println("encrypt output length: " + strconv.Itoa(len([]rune(output))))
	return strings.Trim(string(out), "\n\r"), nil
}

func (client *GithubClient) AddEnvironmentVariablesToRepository(repoId int, repoName, envName string) error {
	//publicKey, err := client.GetEnvPublicKey(repoId, envName)
	publicKey, err := client.GetRepoPublicKey(repoName)
	if err != nil {
		fmt.Println("Get Repository Public Key error: " + err.Error())
		return err
	}
	fmt.Println("public key for " + repoName + "; content: " + *publicKey)
	for k, v := range client.EnvCollection {
		fmt.Println("proceeding key: " + k)
		encrypted, err := encryptedValue(*publicKey, v)
		if err != nil {
			fmt.Printf("Encrypted value for %s failed: %s", k)
			fmt.Println("Encrypted failed: " + err.Error())
			continue
		}
		fmt.Print("result: " + strings.Trim(encrypted, "\n\r"))
		fmt.Println("count length: " + strconv.Itoa(len([]rune(encrypted))))
		if _, err := client.Client.Actions.CreateOrUpdateRepoSecret(context.TODO(), client.Organization, repoName, &github.EncryptedSecret{
			Name:           k,
			EncryptedValue: encrypted,
			Visibility:     "all",
		}); err != nil {
			return err
		}
	}
	return nil
}

func (client *GithubClient) GetOrgPublicKey() (*string, error) {
	publicKey, _, err := client.Client.Actions.GetOrgPublicKey(context.TODO(), client.Organization)
	if err != nil {
		return nil, err
	}
	return publicKey.Key, nil
}
func (client *GithubClient) GetRepoPublicKey(repoName string) (*string, error) {
	publicKey, _, err := client.Client.Actions.GetRepoPublicKey(context.TODO(), client.Organization, repoName)
	if err != nil {
		return nil, err
	}
	return publicKey.Key, nil
}
func (client *GithubClient) GetEnvPublicKey(repoId int, envName string) (*string, error) {
	publicKey, _, err := client.Client.Actions.GetEnvPublicKey(context.TODO(), repoId, envName)
	if err != nil {
		return nil, err
	}
	return publicKey.Key, nil
}

func (client *GithubClient) CreateEnvironmentVariableToRepository(repoName, key, value string) {
	client.Client.Actions.CreateOrUpdateRepoSecret(context.TODO(), client.Organization, repoName, &github.EncryptedSecret{
		Name:                  key,
		KeyID:                 "",
		EncryptedValue:        "",
		Visibility:            "",
		SelectedRepositoryIDs: nil,
	})
}

// Github 環境（跟 CICD 相關的「環境」，非環境變數）的 CRUD API
func (client *GithubClient) CreateEnvironmentToRepository(repository, envName string) (*github.Environment, *github.Response, error) {
	return client.Client.Repositories.CreateUpdateEnvironment(context.TODO(), client.Organization, repository, envName, &github.CreateUpdateEnvironment{
		WaitTimer:              nil,
		Reviewers:              nil,
		DeploymentBranchPolicy: nil,
	})
}
func (client *GithubClient) GetEnvironmentFromRepository(repository, key string) (*github.Environment, *github.Response, error) {
	return client.Client.Repositories.GetEnvironment(context.TODO(), client.Organization, repository, key)
}
func (client *GithubClient) DeleteEnvironemntFromRepository(repository, key string) (*github.Response, error) {
	return client.Client.Repositories.DeleteEnvironment(context.TODO(), client.Organization, repository, key)
}

// @TODO
func (client *GithubClient) AddRepo() {
	client.Client.Repositories.Create(context.TODO(), "andy-demo", &github.Repository{
		ID:                  nil,
		NodeID:              nil,
		Owner:               nil,
		Name:                nil,
		FullName:            nil,
		Description:         nil,
		Homepage:            nil,
		CodeOfConduct:       nil,
		DefaultBranch:       nil,
		MasterBranch:        nil,
		CreatedAt:           nil,
		PushedAt:            nil,
		UpdatedAt:           nil,
		HTMLURL:             nil,
		CloneURL:            nil,
		GitURL:              nil,
		MirrorURL:           nil,
		SSHURL:              nil,
		SVNURL:              nil,
		Language:            nil,
		Fork:                nil,
		ForksCount:          nil,
		NetworkCount:        nil,
		OpenIssuesCount:     nil,
		OpenIssues:          nil,
		StargazersCount:     nil,
		SubscribersCount:    nil,
		WatchersCount:       nil,
		Watchers:            nil,
		Size:                nil,
		AutoInit:            nil,
		Parent:              nil,
		Source:              nil,
		TemplateRepository:  nil,
		Organization:        nil,
		Permissions:         nil,
		AllowRebaseMerge:    nil,
		AllowSquashMerge:    nil,
		AllowMergeCommit:    nil,
		DeleteBranchOnMerge: nil,
		Topics:              nil,
		Archived:            nil,
		Disabled:            nil,
		License:             nil,
		Private:             nil,
		HasIssues:           nil,
		HasWiki:             nil,
		HasPages:            nil,
		HasProjects:         nil,
		HasDownloads:        nil,
		IsTemplate:          nil,
		LicenseTemplate:     nil,
		GitignoreTemplate:   nil,
		TeamID:              nil,
		URL:                 nil,
		ArchiveURL:          nil,
		AssigneesURL:        nil,
		BlobsURL:            nil,
		BranchesURL:         nil,
		CollaboratorsURL:    nil,
		CommentsURL:         nil,
		CommitsURL:          nil,
		CompareURL:          nil,
		ContentsURL:         nil,
		ContributorsURL:     nil,
		DeploymentsURL:      nil,
		DownloadsURL:        nil,
		EventsURL:           nil,
		ForksURL:            nil,
		GitCommitsURL:       nil,
		GitRefsURL:          nil,
		GitTagsURL:          nil,
		HooksURL:            nil,
		IssueCommentURL:     nil,
		IssueEventsURL:      nil,
		IssuesURL:           nil,
		KeysURL:             nil,
		LabelsURL:           nil,
		LanguagesURL:        nil,
		MergesURL:           nil,
		MilestonesURL:       nil,
		NotificationsURL:    nil,
		PullsURL:            nil,
		ReleasesURL:         nil,
		StargazersURL:       nil,
		StatusesURL:         nil,
		SubscribersURL:      nil,
		SubscriptionURL:     nil,
		TagsURL:             nil,
		TreesURL:            nil,
		TeamsURL:            nil,
		TextMatches:         nil,
		Visibility:          nil,
	})
}
