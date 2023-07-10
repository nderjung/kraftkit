package init

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/packmanager"
)

type Init struct {
	ProjectName     string `long:"project-name" short:"pn" usage:"Set the project name to the template"`
	LibraryName     string `long:"library-name" short:"ln" usage:"Set the library name to the template"`
	LibraryKName    string `long:"library-kname" short:"lkn" usage:"Set the library kname to the template"`
	Version         string `long:"version" short:"v" usage:"Set the packge version to the template"`
	Description     string `long:"description" short:"desc" usage:"Set the description to the template"`
	AuthorName      string `long:"author-name" short:"an" usage:"Set the author name to the template"`
	AuthorEmail     string `long:"author-email" short:"ae" usage:"Set the author email to the template"`
	InitialBranch   string `long:"initial-branch" short:"ib" usage:"Set the initial branch name to the template"`
	CopyrightHolder string `long:"copyright-holder" short:"ch" usage:"Set the copyright holder name to the template"`
	ProvideMain     bool   `long:"provide-main" short:"pm" usage:"provide provide-main to the template"`
	WithGitignore   bool   `long:"git-ignore" short:"pm" usage:"provide git-ignore to the template"`
	WithDocs        bool   `long:"docs" short:"pm" usage:"provide docs to the template"`
	WithPatchedir   bool   `long:"patch-dir" short:"pm" usage:"provide patch directory to the template"`
}

func New() *cobra.Command {
	cmd, err := cmdfactory.New(&Init{}, cobra.Command{
		Short:   "Initialise a package template",
		Use:     "init [FLAGS] [DIR]",
		Aliases: []string{"l", "list"},
		Args:    cmdfactory.MaxDirArgs(1),
		Long: heredoc.Doc(`
		Initialise a package template
		`),
		Example: heredoc.Doc(`
			$ kraft pkg init`),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "pkg",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (*Init) Pre(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	pm, err := packmanager.NewUmbrellaManager(ctx)
	if err != nil {
		return err
	}

	cmd.SetContext(packmanager.WithPackageManager(ctx, pm))

	return nil
}

func (opts *Init) Run(cmd *cobra.Command, args []string) error {
	var err error

	// ctx := cmd.Context()
	workdir := ""
	if len(args) > 0 {
		workdir = args[0]
	}

	if workdir == "." || workdir == "" {
		workdir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	fmt.Println("Current location is ", workdir)

	return nil
}
