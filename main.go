package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var timeNow int64 = time.Now().UTC().Unix()

func main() {
	doStars := flag.Bool("s", false, "(stars) save user's stars instead of their own repositories")
	doAll := flag.Bool("a", false, "(all) save owned repositories and starred")
	beQuiet := flag.Bool("q", false, "(quiet) be quiet. no progress bar (better for non-interactive use)")

	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		log.Printf("usage: %s [-qsah] <backup_location_path> <github_acces_token>\n", os.Args[0])
		return
	}
	archive_path := args[0]
	token := args[1]

	if *doStars && *doAll {
		log.Fatal("-s and -a specified is ambiguous. Did you mean to save stars (-s) or stars and owned repositories (-a)?")
	}

	// create backup dir
	backup := archive_path + "/" + "backup" + strconv.FormatInt(timeNow, 10) + "/"
	log.Printf("Backing up to %s\n", backup)
	os.Mkdir(backup, os.ModePerm)

	if !*doStars {
		// get user's repos
		repos, username := getRepos(token)
		log.Printf("Detected %d repositories for %s.\n", len(repos), username)
		saveAll(repos, token, username, archive_path, *beQuiet)
	}
	if *doStars || *doAll {
		// get user's stars
		stars, username := getStars(token)
		log.Printf("Detected %d stars for %s.\n", len(stars), username)
		saveAll(stars, token, username, archive_path, *beQuiet)
	}
}

func saveAll(repos map[string]string, token string, username string, archive_path string, beQuiet bool) {
	maxGo := 20 // avoid having too many files open
	limiter := make(chan struct{}, maxGo)
	if beQuiet {
		// archive user's repos
		var wg sync.WaitGroup
		for name, url := range repos {
			limiter <- struct{}{}
			wg.Add(1)
			url = strings.Replace(url, "https://", "https://"+username+":"+token+"@", 1)
			go archive(name, url, archive_path, &wg, nil, limiter)
		}
		wg.Wait()
	} else {
		// archive user's repos
		bar := pb.Simple.Start(len(repos))
		var wg sync.WaitGroup
		for name, url := range repos {
			limiter <- struct{}{}
			wg.Add(1)
			url = strings.Replace(url, "https://", "https://"+username+":"+token+"@", 1)
			go archive(name, url, archive_path, &wg, bar, limiter)
		}
		wg.Wait()
		bar.Finish()
	}
}

func archive(name string, url string, archive_path string, wg *sync.WaitGroup, bar *pb.ProgressBar, limiter <-chan struct{}) {
	defer wg.Done()
	defer func() {
		<-limiter
	}()
	if bar != nil {
		defer bar.Increment()
	}

	path := archive_path + "/" + "backup" + strconv.FormatInt(timeNow, 10) + "/" + name + "/"

	//create dir at path
	os.Mkdir(path, os.ModePerm)

	//log.Println(path)
	//log.Println("scraping " + url + "...")

	cmd := exec.Command("git", "clone", url, path)
	_, err := cmd.CombinedOutput()

	//log.Println(string(out))
	if err != nil {
		log.Printf("Error: could not save %s\n", url)
		log.Println(err)
	}
}

func getRepos(token string) (map[string]string, string) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	var q struct {
		Viewer struct {
			Login        githubv4.String
			Repositories struct {
				Nodes []struct {
					Name githubv4.String
					//SshUrl githubv4.String
					Url   githubv4.String
					Owner struct {
						//Id    githubv4.ID
						Login githubv4.String
					}
				}
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				TotalCount githubv4.Int
			} `graphql:"repositories(first: 100, after: $cursor)"`
		}
	}

	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil), // Null after argument to get first page.
	}

	repo_set := make(map[string]string)
	for {
		err := client.Query(context.Background(), &q, variables)
		if err != nil {
			panic(err)
		}
		for _, repo := range q.Viewer.Repositories.Nodes {
			id := genId(repo.Name, repo.Owner.Login)
			repo_set[id] = string(repo.Url)
		}
		log.Printf("added %d/%d repos", len(q.Viewer.Repositories.Nodes), q.Viewer.Repositories.TotalCount)
		if !q.Viewer.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(q.Viewer.Repositories.PageInfo.EndCursor)
	}

	return repo_set, string(q.Viewer.Login)
}

func getStars(token string) (map[string]string, string) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	var q struct {
		Viewer struct {
			Login               githubv4.String
			StarredRepositories struct {
				Nodes []struct {
					Name githubv4.String
					//SshUrl githubv4.String
					Url   githubv4.String
					Owner struct {
						//Id    githubv4.ID
						Login githubv4.String
					}
				}
				IsOverLimit githubv4.Boolean
				PageInfo    struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				TotalCount githubv4.Int
			} `graphql:"starredRepositories(first: 100, after: $cursor)"`
		}
	}

	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil), // Null after argument to get first page.
	}

	star_set := make(map[string]string)
	for {
		err := client.Query(context.Background(), &q, variables)
		if err != nil {
			panic(err)
		}
		for _, repo := range q.Viewer.StarredRepositories.Nodes {
			id := genId(repo.Name, repo.Owner.Login)
			star_set[id] = string(repo.Url)
		}
		if q.Viewer.StarredRepositories.IsOverLimit {
			log.Println("[WARNING] Starred repositories is over github limit. Only a truncated list is present.")
		}
		log.Printf("added %d/%d stars", len(q.Viewer.StarredRepositories.Nodes), q.Viewer.StarredRepositories.TotalCount)
		if !q.Viewer.StarredRepositories.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(q.Viewer.StarredRepositories.PageInfo.EndCursor)
	}

	return star_set, string(q.Viewer.Login)
}

func genId(name githubv4.String, login githubv4.String) string {
	return string(login) + "_" + string(name)
}
