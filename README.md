# BackHub
[![Go Report](https://goreportcard.com/badge/github.com/schuermannator/backhub)](https://goreportcard.com/report/github.com/schuermannator/backhub)  
Backup your whole GitHub:
1. Generate personal access token
2. Pass token + backup directory to BackHub
3. Get coffee as your entire github (public + private repos) are backed up to a timestamped directory wherever you specified. 

## Installation/Usage
First, you need to generate a [personal access token][github-token] on GitHub.
The easiest way is to run using docker:
```
docker run -it -e GITHUB_TOKEN=<insert-oauth-token> \
    -v <local-backup-directory>:/archive schuermannator/backhub /archive
```

## Todo
- [ ] log coloring
- [ ] log file creation
- [ ] cron/scheduling setup
- [ ] only scraping first/last 100 - change to all
- [ ] fix dockerfile (multi-stage) and pass values better

[github-token]: https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line
