package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/yaml.v2"
)

var (
	Pwd            string
	File           string
	Config         *Moebius
	RunCache       *Cache
	ActiveProjects []Project
	Dir            string
	NoRun          bool
)

func main() {
	root := &cobra.Command{
		Use:   "moebius",
		Short: "Moebius is a companion application to help with your monorepo builds",
		Long: "A fast and easy task runner built with love by @umurgdk to help" +
			"with your monorepo builds. Please check https://moebius-build.io" +
			"for full documentation.",
		PersistentPreRun:  initMoebius,
		PersistentPostRun: updateCache,
	}

	build := &cobra.Command{
		Use:   "build",
		Short: "Runs build commands",
		Args:  cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				var project Project
				for _, p := range Config.Projects {
					if p.Name == args[0] {
						project = p
					}
				}

				if project.Name != args[0] {
					log.Fatalf("[ERROR] Project %s not found", args[0])
				}

				ActiveProjects = []Project{project}
			}
		},
		Run: Build,
	}

	test := &cobra.Command{
		Use:   "test",
		Short: "Runs test commands",
		Run:   Test,
	}

	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "Runs deploy commands",
		Run:   Deploy,
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	Pwd = pwd

	viper.SetEnvPrefix("moebius")
	viper.AutomaticEnv()

	var file string
	var all bool
	root.PersistentFlags().StringVar(&file, "file", "moebius.yml", "Path for the build configuration file")
	root.PersistentFlags().StringVar(&Dir, "dir", pwd, "Working dir for the tasks")
	root.PersistentFlags().BoolVar(&NoRun, "no-run", false, "No run flag runs build, test, and deploy tasks as like a simulation and will not run actual commands")
	root.PersistentFlags().BoolVar(&all, "all", false, "Run for all projects, without detecting the changes")

	viper.BindPFlags(root.PersistentFlags())

	root.AddCommand(build, test, deploy)

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func initMoebius(_ *cobra.Command, _ []string) {
	file := viper.GetString("file")
	NoRun = viper.GetBool("no-run")
	Dir = viper.GetString("dir")
	all := viper.GetBool("all")
	if !path.IsAbs(Dir) {
		Dir = path.Join(Pwd, Dir)
	}

	if !path.IsAbs(file) {
		file = path.Join(Dir, file)
	}

	if NoRun {
		log.Printf("[INFO] --no-run flag is set, commands will only be printed to the screen.")
	}

	moebius, err := readMoebiusFile(file)
	if err != nil {
		if err == os.ErrNotExist {
			log.Fatalf("[ERROR] Moebius file at %s not found", file)
		}

		log.Fatalf("[ERROR] Failed to parse moebius file! %s", err)
	}
	Config = moebius

	cache, err := NewCache("file://" + path.Join(Dir, moebius.CachePath, "cache.yml"))
	if err != nil {
		log.Fatalf("[ERROR] Failed to create cache! %s", err)
	}
	RunCache = cache

	changedProjects := Config.Projects
	if !all {
		changedProjects, err = FindChangedProjects(moebius, cache)
		if err != nil {
			log.Fatalf("[ERROR] Failed to detect changed projects! %s", err)
		}
	}

	ActiveProjects = changedProjects
}

func readMoebiusFile(file string) (*Moebius, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	var draft MoebiusDraft
	err = yaml.Unmarshal(content, &draft)
	if err != nil {
		log.Fatalf("[ERROR] Failed to parse %s: %s", file, err)
	}

	globalVars := make(map[string]string)
	globalVars["dir"] = Dir

	return draft.Validate(globalVars)
}

func updateCache(_ *cobra.Command, _ []string) {
	repo, err := git.PlainOpen(Dir)
	if err != nil {
		log.Fatal("[ERROR] Failed to open repository")
	}

	head, err := repo.Head()
	if err != nil {
		log.Fatal("[ERROR] Failed to get repo head")
	}

	RunCache.LastSuccessfulCommit = head.Hash().String()
	if err := RunCache.Commit(); err != nil {
		log.Fatal("[ERROR] Failed to update cache")
	}
}
