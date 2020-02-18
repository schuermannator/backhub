# BackHub
Backup your whole GitHub

## Installation/Usage
The easiest way is to run using docker:
```
docker run -it -e GITHUB_TOKEN=<insert-oauth-token> -v <local-backup-directory>:/archive schuermannator/backhub /archive
```

## Todo
- log coloring
- log file creation
- cron/scheduling setup
- only scraping first/last 100 - change to all
- fix dockerfile (multi-stage) and pass values better
