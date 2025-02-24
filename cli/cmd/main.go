package main

import (
	"fmt"
	"path/filepath"

	"github.com/alecthomas/kong"

	cc "wingmen.io/git-clone-all/pkg/cc"
	"wingmen.io/git-clone-all/pkg/constants"
	r "wingmen.io/git-clone-all/pkg/repository"
)

type CLI struct {
	// Global flags
	Clean   bool `help:"Perform a clean operation." short:"c"` // TODO implement removing local repositories that are removed from server
	Debug   bool `help:"print Debug info" short:"d"`
	Noop    bool `help:"No operation mode." short:"n"`
	Version bool `flag:"" help:"version info"`

	// Argments
	Mode    string     `arg:"" help:"Mode of operation (usr, org)." enum:"usr,org"`
	Filter  string     `arg:"" help:"Filter criteria."`
	Verbose cc.Counter `short:"v" help:"Increase verbosity by repeating. -v, -vv, -vvv, etc."`
}

func main() {
	cli := CLI{}

	options := []kong.Option{
		kong.Name(constants.AppName),
		kong.Description(constants.AppDescription),
		kong.UsageOnError(),
		kong.Vars{
			"version": "0.1",
		},
	}
	kongCtx := kong.Parse(&cli, options...)

	ctx := cc.NewCommonContext(cli.Verbose, cli.Clean, cli.Noop)
	if err := r.CloneMultiple(*ctx, cli.Mode, cli.Filter); err != nil {
		kongCtx.FatalIfErrorf(err)
	}
}

func ExcludedRepositories() {
	excludedRepos := r.LoadExcludedRepos(".gitexclude")

	directories, err := filepath.Glob("*")
	if err != nil {
		fmt.Println("Error reading directories:", err)
		return
	}

	for _, dir := range directories {
		if r.IsExcluded(dir, excludedRepos) {
			fmt.Printf("Skipping %s (in .gitexclude)\n", dir)
			continue
		}

		if r.IsGitRepo(dir) {
			fmt.Printf("Running git pull in %s\n", dir)
			err := r.GitPull(dir)
			if err != nil {
				fmt.Printf("Error running git pull in %s: %v\n", dir, err)
			}
		}
	}
}
