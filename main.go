package main

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

type App struct {
	input_files []string
	output_file string
	rename      bool
	rename_pad  int

	rename_counter   int
	rename_max_files int

	output_writer *zip.Writer
}

func ShowHelp() {
	help := `sqcbz is a tool to squash multiple CBZ files into one.

Usage:
  - sqcbz [FILE]...
  - sqcbz -o [OUTPUT_FILE] [FILE]...

Description:
  Given a list of input FILEs, sqcbz squashes all of them together by
  input order. The files inside the CBZ zip are renamed to their
  position relative to the total amount of files to be processed.

  (todo) If FILE is not specified, or FILE is -, sqcbz will read standard
  input.

  If OUTPUT_FILE is not specified, sqcbz will output to standard output.
`

	fmt.Println(help)
	flag.PrintDefaults()
}

func NewApp() App {
	app := App{}

	flag.Usage = ShowHelp

	var help1, help2 *bool
	help1 = flag.Bool("h", false, "alias for -help")
	help2 = flag.Bool("help", false, "prints the help menu and exits")

	flag.StringVar(&app.output_file, "out", "", "write the squashed CBZs to the given path")
	flag.BoolVar(&app.rename, "rename", false, "renames the files sequentially by order of insertion in the CBZ file")
	flag.IntVar(&app.rename_pad, "rename-pad", 8, "amount of padding used when renaming each file inside the CBZ files.\nA padding of 4 on filename '1.png' results in '0001.png'")
	flag.Parse()

	if *help1 || *help2 {
		ShowHelp()
		os.Exit(1)
	}

	app.input_files = flag.Args()

	return app
}

func (app *App) copyFiles(file string) error {
	r, err := zip.OpenReader(file)
	if err != nil {
		return err
	}

	for _, fi := range r.File {
		if !app.rename {
			app.output_writer.Copy(fi)
			continue
		}

		if app.rename_counter > app.rename_max_files {
			return errors.Join(
				r.Close(),
				errors.New(fmt.Sprintf(
					"reached maximum amount of files with %d padding, try increasing it",
					app.rename_pad,
				)))
		}

		fi.Name = fmt.Sprintf("%0*d%s", app.rename_pad, app.rename_counter, filepath.Ext(fi.Name))
		app.output_writer.Copy(fi)
		app.rename_counter += 1
	}

	return r.Close()
}

func (app *App) squash() error {
	fo := os.Stdout
	if app.output_file != "" {
		var err error
		fo, err = os.Create(app.output_file)
		if err != nil {
			return err
		}
	}

	app.rename_counter = 0
	app.rename_max_files = int(math.Pow10(app.rename_pad)) - 1
	app.output_writer = zip.NewWriter(fo)

	for _, file := range app.input_files {
		if err := app.copyFiles(file); err != nil {
			return err
		}
	}

	return app.output_writer.Close()
}

func main() {
	app := NewApp()

	if err := app.squash(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
