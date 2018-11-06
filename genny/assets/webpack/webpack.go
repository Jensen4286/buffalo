package webpack

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/buffalo/genny/assets/standard"
	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/movinglater/gotools"
	"github.com/gobuffalo/packr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// BinPath is the path to the local install of webpack
var BinPath = filepath.Join("node_modules", ".bin", "webpack")

// New generator for create webpack asset files
func New(opts *Options) (*genny.Generator, error) {
	g := genny.New()

	if _, err := exec.LookPath("npm"); err != nil {
		logrus.Info("Could not find npm. Skipping webpack generation.")
		return standard.New(&standard.Options{})
	}

	if err := opts.Validate(); err != nil {
		return g, errors.WithStack(err)
	}

	g.Box(packr.NewBox("../webpack/templates"))

	data := map[string]interface{}{
		"opts": opts,
	}
	t := gotools.TemplateTransformer(data, gotools.TemplateHelpers)
	g.Transformer(t)
	g.Transformer(genny.Dot())

	g.RunFn(func(r *genny.Runner) error {
		return installPkgs(r, opts)
	})

	g.RunFn(func(r *genny.Runner) error {
		f, err := r.FindFile("templates/application.html")
		if err != nil {
			return errors.WithStack(err)
		}
		css := bs4
		if opts.Bootstrap == 3 {
			css = bs3
		}
		s := strings.Replace(f.String(), "</title>", "</title>\n"+css, 1)
		return r.File(genny.NewFileS(f.Name(), s))
	})

	return g, nil
}

func installPkgs(r *genny.Runner, opts *Options) error {
	command := "yarnpkg"

	if !opts.App.WithYarn {
		command = "npm"
	} else {
		if err := installYarn(r); err != nil {
			return errors.WithStack(err)
		}
	}
	args := []string{"install", "--no-progress", "--save"}

	c := exec.Command(command, args...)
	c.Stdout = yarnWriter{
		fn: r.Logger.Debug,
	}
	c.Stderr = yarnWriter{
		fn: r.Logger.Debug,
	}
	return r.Exec(c)
}

type yarnWriter struct {
	fn func(...interface{})
}

func (y yarnWriter) Write(p []byte) (int, error) {
	y.fn(string(p))
	return len(p), nil
}

func installYarn(r *genny.Runner) error {
	// if there's no yarn, install it!
	if _, err := exec.LookPath("yarnpkg"); err == nil {
		return nil
	}
	yargs := []string{"install", "-g", "yarn"}
	return r.Exec(exec.Command("npm", yargs...))
}

const bs3 = `<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">`

const bs4 = `<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.0/css/bootstrap.min.css" integrity="sha384-9gVQ4dYFwwWSjIDZnLEWnxCjeSWFphJiwGPXr1jddIhOegiu1FwO5qRGvFXOdJZ4" crossorigin="anonymous">`