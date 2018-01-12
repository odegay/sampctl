package rook

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"

	"github.com/Southclaws/sampctl/compiler"
	"github.com/Southclaws/sampctl/print"
	"github.com/Southclaws/sampctl/types"
	"github.com/Southclaws/sampctl/util"
)

// Build compiles a package, dependencies are ensured and a list of paths are sent to the compiler.
func Build(pkg *types.Package, build, cacheDir, platform string, ensure bool) (problems []types.BuildProblem, result types.BuildResult, err error) {
	config := GetBuildConfig(*pkg, build)
	if config == nil {
		err = errors.Errorf("no build config named '%s'", build)
		return
	}

	config.WorkingDir = filepath.Dir(util.FullPath(pkg.Entry))
	config.Input = filepath.Join(pkg.Local, pkg.Entry)
	config.Output = filepath.Join(pkg.Local, pkg.Output)

	if ensure {
		err = EnsureDependencies(pkg)
		if err != nil {
			err = errors.Wrap(err, "failed to ensure dependencies before build")
			return
		}
	}

	err = ResolveDependencies(pkg)
	if err != nil {
		err = errors.Wrap(err, "failed to resolve dependencies before build")
		return
	}

	for _, depMeta := range pkg.AllDependencies {
		depDir := filepath.Join(pkg.Local, "dependencies", depMeta.Repo)
		incPath := depMeta.Path

		// check if local package has a definition, if so, check if it has an IncludePath field
		pkg, err := types.PackageFromDir(depDir)
		if err == nil {
			if pkg.IncludePath != "" {
				incPath = pkg.IncludePath
			}
		}

		config.Includes = append(config.Includes, filepath.Join(depDir, incPath))
	}

	print.Verb("building", pkg, "with", config.Version)

	problems, result, err = compiler.CompileSource(pkg.Local, cacheDir, platform, *config)
	if err != nil {
		err = errors.Wrap(err, "failed to compile package entry")
		return
	}

	return
}

// BuildWatch runs the Build code on file changes
func BuildWatch(pkg *types.Package, build, cacheDir, platform string, ensure bool) (err error) {
	config := GetBuildConfig(*pkg, build)
	if config == nil {
		err = errors.Errorf("no build config named '%s'", build)
		return
	}

	config.WorkingDir = filepath.Dir(util.FullPath(pkg.Entry))
	config.Input = filepath.Join(pkg.Local, pkg.Entry)
	config.Output = filepath.Join(pkg.Local, pkg.Output)

	if ensure {
		err = EnsureDependencies(pkg)
		if err != nil {
			err = errors.Wrap(err, "failed to ensure dependencies before build")
			return
		}
	}

	err = ResolveDependencies(pkg)
	if err != nil {
		err = errors.Wrap(err, "failed to resolve dependencies before build")
		return
	}

	for _, depMeta := range pkg.AllDependencies {
		depDir := filepath.Join(pkg.Local, "dependencies", depMeta.Repo)
		incPath := depMeta.Path

		// check if local package has a definition, if so, check if it has an IncludePath field
		pkg, err := types.PackageFromDir(depDir)
		if err == nil {
			if pkg.IncludePath != "" {
				incPath = pkg.IncludePath
			}
		}

		config.Includes = append(config.Includes, filepath.Join(depDir, incPath))
	}

	print.Verb("watching", pkg)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "failed to create new filesystem watcher")
	}

	err = watcher.Add(pkg.Local)
	if err != nil {
		return errors.Wrap(err, "failed to add package directory to filesystem watcher")
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

loop:
	for {
		select {
		case sig := <-signals:
			fmt.Println("") // insert newline after the ^C
			print.Info("signal received", sig, "stopping watcher...")
			break loop

		case event := <-watcher.Events:
			ext := filepath.Ext(event.Name)
			if ext != ".pwn" && ext != ".inc" {
				continue
			}
			if event.Op != fsnotify.Write && event.Op != fsnotify.Create {
				continue
			}
			print.Verb("changed:", event.Name)

			_, _, err = compiler.CompileSource(pkg.Local, cacheDir, platform, *config)
			if err != nil {
				err = errors.Wrap(err, "failed to compile package entry")
				return
			}
		}
	}

	print.Info("finished running build watcher")

	return
}

// GetBuildConfig returns a matching build by name from the package build list. If no name is
// specified, the first build is returned. If the package has no build definitions, a default
// configuration is returned.
func GetBuildConfig(pkg types.Package, name string) (config *types.BuildConfig) {
	def := types.GetBuildConfigDefault()

	if len(pkg.Builds) == 0 {
		config = def
	} else {
		if name == "" {
			config = &pkg.Builds[0]
		} else {
			for _, cfg := range pkg.Builds {
				if cfg.Name == name {
					config = &cfg
				}
			}
		}
		if config.Version == "" {
			config.Version = def.Version
		}
	}

	return
}
