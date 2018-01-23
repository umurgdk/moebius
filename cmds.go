package main

import (
	"log"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
)

func Build(_ *cobra.Command, _ []string) {
	seq := NewAsyncSequence()
	seq.Add(len(ActiveProjects))
	for _, p := range ActiveProjects {
		go func(p Project) {
			defer seq.Done()
			log.Printf("[INFO] Building project %s at %s", p.Name, p.Path)

			runCmdSet(seq, p, p.BeforeBuild)
			runCmdSet(seq, p, p.Build)
			runCmdSet(seq, p, p.AfterBuild)
		}(p)
	}

	if err := seq.Wait(); err != nil {
		log.Fatal(err)
	}
}

func Test(_ *cobra.Command, _ []string) {
	seq := NewAsyncSequence()
	seq.Add(len(ActiveProjects))
	for _, p := range ActiveProjects {
		if len(p.Test.Cmds) == 0 {
			continue
		}

		go func(p Project) {
			defer seq.Done()
			log.Printf("[INFO][%s] Running test commands:", p.Name)
			runCmdSet(seq, p, p.Test)
		}(p)
	}

	if err := seq.Wait(); err != nil {
		log.Fatal(err)
	}
}

func Deploy(_ *cobra.Command, _ []string) {
	seq := NewAsyncSequence()
	seq.Add(len(ActiveProjects))
	for _, p := range ActiveProjects {
		if len(p.Test.Cmds) == 0 {
			continue
		}

		go func(p Project) {
			defer seq.Done()
			log.Printf("[INFO][%s] Running deploy commands:", p.Name)
			runCmdSet(seq, p, p.Deploy)
		}(p)
	}

	if err := seq.Wait(); err != nil {
		log.Fatal(err)
	}
}

func runCmdSet(seq *AsyncSequence, p Project, set CmdSet) {
	if len(set.Cmds) > 0 {
		log.Printf("[INFO][%s] Running %s commands:", p.Name, set.Name)
	}

	for _, cmdLine := range set.Cmds {
		if seq.Err() != nil {
			return
		}

		log.Printf("[INFO][%s][%s] $ %s", p.Name, set.Name, cmdLine)

		if NoRun {
			continue
		}

		stepCmd := exec.Command("bash", "-c", cmdLine)

		if path.IsAbs(p.Path) {
			stepCmd.Dir = p.Path
		} else {
			stepCmd.Dir = path.Join(Dir, p.Path)
		}

		output, err := stepCmd.CombinedOutput()
		println(string(output))

		seq.Fail(err)
	}
}
