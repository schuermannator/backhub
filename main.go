package main

import (
    "context"
    "log"
    "os"
    "os/exec"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/shurcooL/githubv4"
    "golang.org/x/oauth2"
)

var timeNow int64 = time.Now().UTC().Unix()

func main() {
    // parse args
    if len(os.Args) < 2 {
        log.Println("usage: os.Args[0] <backup location path>")
        return
    }

    repos, username := getRepos()	
    backup := os.Args[1] + "/" + "backup"+strconv.FormatInt(timeNow, 10) + "/"
    log.Printf("Backing up to %s\n", backup)
    // create backup dir
    os.Mkdir(backup, os.ModePerm)
    log.Printf("Detected %d repositories.\n", len(repos))
    var wg sync.WaitGroup
    for name, url := range repos {
        wg.Add(1)
        url = strings.Replace(url,"https://", "https://"+username+":"+os.Getenv("GITHUB_TOKEN")+"@", 1)
        go archive(name, url, &wg)
    }
    wg.Wait()
    log.Println("DONE")
}

func archive(name string, url string, wg *sync.WaitGroup) {
    defer wg.Done()

    archive_path := os.Args[1]
    path := archive_path + "/" + "backup"+strconv.FormatInt(timeNow, 10) + "/" + name + "/"

    //create dir at path
    os.Mkdir(path, os.ModePerm)

    log.Println(path)
    log.Println("scraping " + url + "...")

    cmd := exec.Command("git", "clone", url, path)
    out, err := cmd.CombinedOutput()

    log.Println(string(out))
    if err != nil {
        log.Fatal(err)
    }
}

func getRepos() (map[string]string, string) {
    src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)

    var infoquery struct {
	    Viewer struct {
		    Login     githubv4.String
		    CreatedAt githubv4.DateTime
	    }
    }

    err := client.Query(context.Background(), &infoquery, nil)
    if err != nil {
        panic(err)
    }

    username := string(infoquery.Viewer.Login)
    log.Println("    Login:", username)
    log.Println("CreatedAt:", infoquery.Viewer.CreatedAt)


    var repoquery1 struct {
	    Viewer struct {
            Name    githubv4.String
            Repositories struct {
                Nodes []struct {
                    Name githubv4.String
                    //SshUrl githubv4.String
                    Url githubv4.String
                    Owner struct {
                        //Id    githubv4.ID
                        Login githubv4.String
                    }
                }
            } `graphql:"repositories(last: 100)"`
	    }
    }

    var repoquery2 struct {
	    Viewer struct {
            Name    githubv4.String
            Repositories struct {
                Nodes []struct {
                    Name   githubv4.String
                    //SshUrl githubv4.String
                    Url githubv4.String
                    Owner struct {
                        //Id    githubv4.ID
                        Login githubv4.String
                    }
                }
            } `graphql:"repositories(first: 100)"`
	    }
    }


    err = client.Query(context.Background(), &repoquery1, nil)
    if err != nil {
        panic(err)
    }
    err = client.Query(context.Background(), &repoquery2, nil)
    if err != nil {
        panic(err)
    }
    repo_set := make(map[string]string)
    for _, repo := range repoquery1.Viewer.Repositories.Nodes {
        id := genId(repo.Name, repo.Owner.Login)
        repo_set[id] = string(repo.Url)
    }
    dups := false
    for _, repo := range repoquery2.Viewer.Repositories.Nodes {
        id := genId(repo.Name, repo.Owner.Login)
        if _, ok := repo_set[id]; ok {
            dups = true
        } else {
            repo_set[id] = string(repo.Url)
        }
    }
    if !dups {
        log.Println("WARNING: Only first/last 100 scraped.")
    }
    return repo_set, username
}

func genId(name githubv4.String, login githubv4.String) string {
    return string(login) + "_" + string(name)
}
