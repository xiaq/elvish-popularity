package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

const (
	backoff         = time.Minute
	perPage         = 10
	attemptsPerPage = 10
)

func main() {
	ctx := context.Background()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN is required")
	}
	httpClient := oauth2.NewClient(
		ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))

	client := github.NewClient(httpClient)
	opt := &github.SearchOptions{
		Sort: "indexed", ListOptions: github.ListOptions{PerPage: perPage, Page: 1}}
	fileURLs := make(map[string]bool)
	repoURLs := make(map[string]int)
	attempts := make(map[int]int)

	for {
		results, resp, err := client.Search.Code(ctx, "extension:.elv NOT Elvira", opt)
		if err != nil {
			if resp.StatusCode == 403 {
				log.Printf("  error: %v\n", err)
				log.Printf("  sleeping for %v\n", backoff)
				time.Sleep(backoff)
				continue
			} else {
				log.Fatal(err)
			}
		}
		for _, result := range results.CodeResults {
			fileURLs[*result.HTMLURL] = true
			repoURLs[*result.Repository.HTMLURL]++
		}
		log.Printf("page %v: got %v results (so far: %v files, %v repos)\n",
			opt.Page, len(results.CodeResults), len(fileURLs), len(repoURLs))
		if resp.NextPage == 0 {
			break
		}
		if len(results.CodeResults) < perPage {
			attempts[opt.Page]++
			if attempts[opt.Page] < attemptsPerPage {
				log.Printf("  too few results, will attempt again")
				continue
			} else {
				log.Printf("  too few results, giving up since already attempted %v times\n",
					attemptsPerPage)
			}
		}
		opt.Page = resp.NextPage
	}

	fmt.Println("Popularity as of", time.Now().Format("2006-01-02"))
	fmt.Println()

	fmt.Printf("## %v files\n\n", len(fileURLs))
	for fileURL := range fileURLs {
		fmt.Printf("  %v\n", fileURL)
	}

	fmt.Println()
	fmt.Printf("## %v repos\n\n", len(repoURLs))
	for repoURL, files := range repoURLs {
		if files > 1 {
			fmt.Printf("  %v # %v files\n", repoURL, files)
		} else {
			fmt.Printf("  %v\n", repoURL)
		}
	}
}
