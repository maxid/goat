package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/maxid/goat/config"
	"github.com/maxid/goat/consts"
	"github.com/maxid/goat/context"
)

// main executes main processes.
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	version := flag.Bool("v", false, "Show Goat version")
	configFile := flag.String("c", "", "Show Goat version")
	interval := flag.Int("i", consts.DefaultInterval, "An interval(ms) of a watchers' file check loop")
	flag.Parse()

	if *configFile != "" {
		if strings.HasSuffix(*configFile, "yml") {
			config.YAMLConfigFile = *configFile
		} else if strings.HasSuffix(*configFile, "json") {
			config.JSONConfigFile = *configFile
		}
	}

	if *version {
		fmt.Printf("Goat %s\n", consts.Version)
		os.Exit(0)
	}

	ctx, err := context.NewContext(*interval)
	if err != nil {
		log.Fatal(err)
	}

	initTasks := ctx.Config.InitTasks
	if initTasks != nil && len(initTasks) > 0 {
		executeTasks(initTasks, nil, "/nil", "nil")
	}

	jobsC := make(chan context.Job, consts.JobsChannelBuffer)

	launchWatchers(ctx, jobsC)

	handleJobs(jobsC)
}

// launchWatchers launches watchers.
func launchWatchers(ctx *context.Context, jobsC chan<- context.Job) {
	for _, watcher := range ctx.Config.Watchers {
		go watcher.Launch(ctx, jobsC)
	}
}

// handleJobs handle jobs.
func handleJobs(jobsC <-chan context.Job) {
	for job := range jobsC {
		watcher := job.Watcher
		watcher.Printf("%s", job.Message)
		executeTasks(watcher.Tasks, watcher, job.Path, job.Name)
	}
}

// executeTasks executes tasks.
func executeTasks(tasks []*context.Task, watcher *context.Watcher, filePath string, fileName string) {
	for _, task := range tasks {
		command := task.Command
		command = strings.Replace(command, "{file}", fileName, -1)
		command = strings.Replace(command, "{path}", filePath, -1)
		tokens := strings.Split(command, " ")
		name := tokens[0]
		var cmdArg []string
		if len(tokens) > 1 {
			cmdArg = tokens[1:]
		}
		cmd := exec.Command(name, cmdArg...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if task.Nowait {
			printf(watcher, "execute(nowait): %s", command)
			if err := cmd.Start(); err != nil {
				printf(watcher, "An error occurred: %s \n\n", err.Error())
			} else {
				printf(watcher, "end(nowait): %s \n\n", command)
			}
		} else {
			printf(watcher, "execute: %s", command)
			err := cmd.Run()
			if err != nil {
				printf(watcher, "An error occurred: %s \n\n", err.Error())
			} else {
				printf(watcher, "end: %s \n\n", command)
			}
		}
	}
}

func printf(watcher *context.Watcher, format string, v ...interface{}) {
	if watcher != nil {
		watcher.Printf(format, v)
	} else {
		log.Printf(format, v)
	}
}
