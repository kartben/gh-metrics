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

import s "strings"

func main() {
	if len(os.Args) < 6 {
		fmt.Println("usage : ./gh-metrics [token] [year] [month idx] [owner] [repo list]")
		fmt.Println(" ex: for March 2016 ./gh-metrics xxxxxxxxxx 2016 3 eclipse leshan leshan.osgi")
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
	//	fmt.Println("owner:", owner)

	fmt.Println("project", "\t",
		"repository", "\t",
		"issues opened", "\t",
		"issues closed", "\t",
		"PR opened", "\t",
		"PR closed", "\t",
		"issuesComments", "\t",
		"PRComments")

	for i := 5; i < len(os.Args); i++ {
		repo := os.Args[i]
		project := repo
		if s.Contains(os.Args[i], ":") {
			arr := s.Split(os.Args[i], ":")
			project = arr[0]
			repo = arr[1]
		}

		issuesOpenedCount, issuesClosedCount, prOpenedCount, prClosedCount, issuesCommentCount, prCommentCount := getStats(owner, repo, client, from, to)
		fmt.Println(
			project, "\t",
			repo, "\t",
			issuesOpenedCount, "\t",
			issuesClosedCount, "\t",
			prOpenedCount, "\t",
			prClosedCount, "\t",
			issuesCommentCount, "\t",
			prCommentCount)

	}

}

func getStats(owner string, repo string, client *github.Client, from time.Time, to time.Time) (int, int, int, int, int, int) {

	prOpenedCount := 0
	issuesOpenedCount := 0
	prCommentCount := 0
	issuesCommentCount := 0
	prClosedCount := 0
	issuesClosedCount := 0

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

			if v.CreatedAt.After(from) {
				// the issue/PR was *opened* during the period

				//	fmt.Println("created at: ", v.CreatedAt)

				if v.PullRequestLinks != nil {
					prOpenedCount++
				} else {
					issuesOpenedCount++
				}
			}

			if v.ClosedAt != nil && v.ClosedAt.After(from) && v.ClosedAt.Before(to) {
				// the issue/PR was *closed* during the period

				//	fmt.Println("closed at: ", v.ClosedAt)

				if v.PullRequestLinks != nil {
					prClosedCount++
				} else {
					issuesClosedCount++
				}
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

	return issuesOpenedCount, issuesClosedCount, prOpenedCount, prClosedCount, issuesCommentCount, prCommentCount
}
