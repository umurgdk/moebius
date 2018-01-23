# Moebius
Moebius is a simple task runner built for monorepos to reduce build/test times by understanding which project changed since the last successful build/test/deploy and running tests only for them.

## Install

```
$ go install github.com/umurgdk/moebius
```

## Usage

```
A fast and easy task runner built with love by @umurgdk to helpwith your monorepo builds. Please check https://moebius-build.iofor full documentation.

Usage:
  moebius [command]

Available Commands:
  build       Runs build commands
  deploy      Runs deploy commands
  help        Help about any command
  test        Runs test commands

Flags:
      --all           Run for all projects, without detecting the changes
      --dir string    Working dir for the tasks (default is `pwd`)
      --file string   Path for the build configuration file (default "moebius.yml")
  -h, --help          help for moebius
      --no-run        No run flag runs build, test, and deploy tasks as like a simulation and will not run actual commands

Use "moebius [command] --help" for more information about a command.
```

## Configuring with `moebius.yml`
Assuming you have a monorepo structure like:
```
monorepo
  - project1
    - prebuild.sh
    - main.go
  - project2
    - main.go
  - project3
    - main.go
  - moebius.yml
```

a simple moebius configuration might be like:

```yaml
# If you want to run tasks for all projects for some reason set onlyChanged 
# to false. default value is true
onlyChanged: true

# cache path to keep record of last successful build commits. Most of the
# continus integration systems like travis, circleci supports caching. Please
# refer to your CI systems documentation to set a correct path
# default value is `.moebius`
cache: .moebius

# In vars block you can define some variables to use in command lines.
# There are some predefined variables you can use:
#   dir => monorepo root or specified directory by passing --dir flag to moebius
# currently only string variables are supported
vars:
    # you can use predefined variables in your variable values please check
    # https://golang.org/pkg/text/template/ for the template language
    dist: {{.dir}}/bin

projects:
    # project names can be anything and it is independent from it's path
    project1:
        # project's path relative to monorepo root or specified path given
        # with --dir flag passed on moebius executable
        path: project1

        # There are 5 tasks in total to give commands, all commands are
        # passed to bash as `sh -c "<command line>"`. All the command
        # lines should be relative to project directory and they can use
        # builtin variables as well as user defined variables in `vars:`
        

        # NOTE: moebius will run beforeBuild, build, and afterBuild if you
        # pass build command.

        # beforeBuild is optional, if set of commands given they will run
        # before the build commands. If beforeBuild commands fail somehow
        # moebius will fail and stop running further steps
        beforeBuild:
            - chmod +x prebuild.sh
            - sh prebuild.sh

        # build commands will run after beforeBuild commands only if
        # beforeBuild commands ran successfully
        build:
            - go build -o {{.dist}}/project1

        # afterBuild is optional and it's commands will run after 
        # successful run of build section
        afterBuild:
            - "echo Wow cool!"

        # test commands will only run if test comand is passed to moebius
        test:
            - go test $(go list ./...)

        # deploy commands will only run if deploy command is passed to moebius
        deploy:
            - docker build -t project1:latest .
            - docker push

    # project2 configuration left empty for simplicity
    project2:
        path: project23
```
