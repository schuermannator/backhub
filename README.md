# github-backup
[![Go Report](https://goreportcard.com/badge/github.com/schuermannator/github-backup)](https://goreportcard.com/report/github.com/schuermannator/github-backup)

This aims to be a simple, explainable tool to backup your repositories and/or stars on GitHub. It is a
self-contained binary that is easily scheduled via cron/systemd/etc.

## Quickstart

Backup your GitHub repositories:

1. Generate a personal access token on GitHub (with `repo` permissions)
2. Run `./ghb <backup_directory> <github_token>`
3. Get coffee as your entire github (public + private repos) are backed up to a timestamped directory wherever
   you specified. 

## Installation/Usage

First, you need to generate a [personal access token][github-token] on GitHub with `repo` permissions.
Then, give a path where backups should be saved (the directory at the path is created if it doesn't exist) and
run the program with the personal access token:
```
$ ./ghb <backup_directory> <github_token>
```

Additionally, there are three flags supported: `-s` will save your user's stars, `-a` will save stars in
addition to your own repositories, `-q` will run quietly (no progress bar), and `-h` will print help. In
summary, the usage is:

```
$ ./ghb [-saqh] <backup_directory> <github_token>
```

[github-token]: https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line
