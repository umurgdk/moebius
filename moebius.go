package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

type ProjectDraft struct {
	Path        string
	Build       []string
	BeforeBuild []string
	AfterBuild  []string
	Test        []string
	Deploy      []string
}

type Project struct {
	Name        string
	Path        string
	Build       CmdSet
	BeforeBuild CmdSet
	AfterBuild  CmdSet
	Test        CmdSet
	Deploy      CmdSet
}

type CmdSet struct {
	Name string
	Cmds []string
}

type MoebiusDraft struct {
	Cache       string
	OnlyChanged bool
	Projects    map[string]ProjectDraft
	Vars        map[string]string
}

type Moebius struct {
	CachePath   string
	OnlyChanged bool
	Projects    []Project
	Vars        map[string]string
}

func (m MoebiusDraft) Validate(globalVars map[string]string) (*Moebius, error) {
	moebius := &Moebius{
		CachePath:   m.Cache,
		OnlyChanged: m.OnlyChanged,
		Projects:    make([]Project, len(m.Projects)),
		Vars:        make(map[string]string),
	}

	if moebius.CachePath == "" {
		moebius.CachePath = ".moebius"
	}

	for key, value := range m.Vars {
		renderedValue, err := renderText(value, globalVars)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("[ERROR] Configuration has invalid variable: %s, %s", key, err))
		}

		moebius.Vars[key] = renderedValue
	}

	for key, value := range globalVars {
		moebius.Vars[key] = value
	}

	i := 0
	for projectName, project := range m.Projects {
		if err := renderCommands(project.Build, moebius.Vars); err != nil {
			return nil, errors.New(fmt.Sprintf("[ERROR] Project %s has invalid build commands: %s", projectName, err))
		}

		if err := renderCommands(project.BeforeBuild, moebius.Vars); err != nil {
			return nil, errors.New(fmt.Sprintf("[ERROR] Project %s has invalid beforeBuild commands: %s", projectName, err))
		}

		if err := renderCommands(project.AfterBuild, moebius.Vars); err != nil {
			return nil, errors.New(fmt.Sprintf("[ERROR] Project %s has invalid afterBuild commands: %s", projectName, err))
		}

		if err := renderCommands(project.Test, moebius.Vars); err != nil {
			return nil, errors.New(fmt.Sprintf("[ERROR] Project %s has invalid test commands: %s", projectName, err))
		}

		if err := renderCommands(project.Deploy, moebius.Vars); err != nil {
			return nil, errors.New(fmt.Sprintf("[ERROR] Project %s has invalid deploy commands: %s", projectName, err))
		}

		path := project.Path
		if strings.HasPrefix(path, "./") {
			path = path[2:]
		}

		moebius.Projects[i] = Project{
			Name:        projectName,
			Path:        project.Path,
			Build:       NewCmdSet("build", project.Build),
			BeforeBuild: NewCmdSet("beforeBuild", project.BeforeBuild),
			AfterBuild:  NewCmdSet("afterBuild", project.AfterBuild),
			Test:        NewCmdSet("test", project.Test),
			Deploy:      NewCmdSet("deploy", project.Deploy),
		}

		i += 1
	}

	return moebius, nil
}

func NewCmdSet(name string, cmdLines []string) CmdSet {
	return CmdSet{Name: name, Cmds: cmdLines}
}

func renderCommands(commands []string, vars map[string]string) error {
	for i := 0; i < len(commands); i++ {
		cmdLine, err := renderText(commands[i], vars)
		if err != nil {
			return err
		}

		commands[i] = cmdLine
	}

	return nil
}

func renderText(text string, vars map[string]string) (string, error) {
	t, err := template.New("").Parse(text)
	if err != nil {
		return "", err
	}

	r := new(bytes.Buffer)
	if err = t.ExecuteTemplate(r, "", vars); err != nil {
		return "", err
	}

	return r.String(), nil
}
