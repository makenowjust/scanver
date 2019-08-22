package scanver

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/go-github/github"
)

// Repository holds repository owner name and its name.
type Repository struct {
	Owner string
	Name  string
}

// SearchRepositories search the given owner's repositories using pkgName package.
func (c *Client) SearchRepositories(ctx context.Context, owner string, pkgName string) ([]*Repository, error) {
	log.Printf("Start Searching Repositories...")

	repoSet := make(map[Repository]struct{})

	query := fmt.Sprintf("user:%s %s", owner, pkgName)
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100, Page: 1},
	}

	for {
		res, resp, err := c.Search.Code(ctx, query, opt)
		if err != nil {
			return nil, err
		}
		log.Printf("Total: %d, RateLimit: %d/%d, Page: %d/%d", res.GetTotal(), resp.Rate.Remaining, resp.Rate.Limit, opt.Page, resp.LastPage)

		for _, code := range res.CodeResults {
			repo := Repository{
				Owner: code.GetRepository().GetOwner().GetLogin(),
				Name:  code.GetRepository().GetName(),
			}
			repoSet[repo] = struct{}{}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage

		time.Sleep(5 * time.Second)
	}

	repos := make([]*Repository, 0, len(repoSet))
	for repo := range repoSet {
		// Don't `append(repos, &repo)` because `repo` is reused over iteration.
		repos = append(repos, &Repository{Owner: repo.Owner, Name: repo.Name})
	}

	// Sort `repos` by names.
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Owner < repos[j].Owner || repos[i].Owner == repos[j].Owner && repos[i].Name < repos[j].Name
	})

	return repos, nil
}
