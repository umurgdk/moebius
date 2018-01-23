package main

import (
	"errors"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func FindChangedProjects(moebius *Moebius, cache *Cache) ([]Project, error) {
	// If cache is not exist we want to build all the projects
	if cache == nil {
		return moebius.Projects, nil
	}

	projectDirs := make(map[string]*Project, len(moebius.Projects))
	for _, p := range moebius.Projects {
		projectDirs[p.Path] = &p
	}

	repo, err := git.PlainOpen(Dir)
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	headCommit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	iter := object.NewCommitPreorderIter(headCommit, nil, nil)
	if err != nil {
		return nil, err
	}

	successHash := plumbing.NewHash(cache.LastSuccessfulCommit)
	terminator := errors.New("")

	// Collect commits from head to last successful one
	var commitsToSuccess []*object.Commit
	err = iter.ForEach(func(c *object.Commit) error {
		if c.Hash != successHash {
			commitsToSuccess = append(commitsToSuccess, c)
			return nil
		}

		commitsToSuccess = append(commitsToSuccess, c)
		return terminator
	})

	if err != nil && err != terminator {
		return nil, err
	}

	var changedFiles []string
	for i := len(commitsToSuccess) - 2; i >= 0; i-- {
		pcom := commitsToSuccess[i+1]
		ncom := commitsToSuccess[i]

		patch, err := pcom.Patch(ncom)
		if err != nil {
			return nil, err
		}

		files := patch.FilePatches()
		for _, f := range files {
			from, to := f.Files()

			var path string
			if from == nil {
				path = to.Path()
			} else {
				path = from.Path()
			}

			changedFiles = append(changedFiles, path)
		}
	}

	var projects []Project
	projectsSet := make(map[string]interface{})
	for _, f := range changedFiles {
		for _, p := range moebius.Projects {
			_, added := projectsSet[p.Name]
			if !added && strings.HasPrefix(f, p.Path) {
				projectsSet[p.Name] = struct{}{}
				projects = append(projects, p)
			}
		}
	}

	return projects, nil
}
