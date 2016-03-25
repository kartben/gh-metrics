package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	if len(os.Args) < 6 {
		fmt.Println("usage : ./gh-metrics [token] [year] [month idx] [owner] [repo list]")
		fmt.Println(" ex: for March 2016 ./gh-metrics xxxxxxxxxx eclipse 2016 3 leshan leshan.osgi")
		os.Exit(-1)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Args[1]},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	year, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	sMonth, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}

	loc, err := time.LoadLocation("UTC")
	if err != nil {
		log.Fatal(err)
	}

	from := time.Date(year, time.Month(sMonth), 1, 0, 0, 0, 0, loc)
	fmt.Println("From:", from)

	to := time.Date(year, time.Month(sMonth)+1, 1, 0, 0, 0, 0, loc)
	fmt.Println("To:", to)

	owner := os.Args[4]
	fmt.Println("owner:", owner)

	for i := 5; i < len(os.Args); i++ {
		fmt.Println("repo", os.Args[i])
		issuesCount, prCount, issuesCommentCount, prCommentCount := getStats(owner, os.Args[i], client, from, to)
		fmt.Println("issues", issuesCount, "PR", prCount, "issue comments", issuesCommentCount, "pr comments", prCommentCount, "\n")

	}

}

func getStats(owner string, repo string, client *github.Client, from time.Time, to time.Time) (int, int, int, int) {

	prCount := 0
	issuesCount := 0
	prCommentCount := 0
	issuesCommentCount := 0

	opt := &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       "all",
		Since:       from}

	for {
		issues, r, err := client.Issues.ListByRepo(owner, repo, opt)
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Println("count:", len(issues))
		for _, v := range issues {
			// skip issues not in the range
			if v.CreatedAt.After(to) {
				continue
			}
			//fmt.Println(*v.Title, v.PullRequestLinks != nil)
			if v.PullRequestLinks != nil {
				prCount++
			} else {
				issuesCount++
			}

			optComments := &github.IssueListCommentsOptions{
				Since: from}
			for {
				comments, rc, err := client.Issues.ListComments(owner, repo, *v.Number, optComments)
				if err != nil {
					log.Fatal(err)
				}

				for _, comm := range comments {
					if to.After(*comm.CreatedAt) {
						if v.PullRequestLinks != nil {
							prCommentCount++
						} else {
							issuesCommentCount++
						}

					}
				}
				if rc.NextPage == 0 {
					break
				}
				optComments.ListOptions.Page = rc.NextPage

			}

		}
		if r.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = r.NextPage
	}

	return issuesCount, prCount, issuesCommentCount, prCommentCount
}
