package create

import (
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/spf13/cobra"

	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/packmanager"
	"kraftkit.sh/unikraft/lib/template"
)

type Create struct {
	ProjectName     string `long:"project-name" usage:"Set the project name to the template"`
	LibraryName     string `long:"library-name" usage:"Set the library name to the template" default:"lib-template"`
	LibraryKName    string `long:"library-kname" usage:"Set the library kname to the template" default:"LIBTEMPLATE"`
	Version         string `long:"version" short:"v" usage:"Set the packge version to the template"`
	Description     string `long:"description"  usage:"Set the description to the template"`
	AuthorName      string `long:"author-name" usage:"Set the author name to the template"`
	AuthorEmail     string `long:"author-email" usage:"Set the author email to the template"`
	InitialBranch   string `long:"initial-branch" usage:"Set the initial branch name to the template" default:"staging"`
	CopyrightHolder string `long:"copyright-holder" usage:"Set the copyright holder name to the template"`
	NoProvideMain   bool   `long:"no-provide-main" usage:"Do not provide provide-main to the template"`
	NoWithGitignore bool   `long:"no-git-ignore" usage:"Do not provide git-ignore to the template"`
	NoWithDocs      bool   `long:"no-docs" usage:"Do not provide docs to the template"`
	WithPatchedir   bool   `long:"patch-dir" usage:"provide patch directory to the template"`
}

func New() *cobra.Command {
	cmd, err := cmdfactory.New(&Create{}, cobra.Command{
		Short:   "Initialise a package template",
		Use:     "create [FLAGS] [DIR]",
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

func (*Create) Pre(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	pm, err := packmanager.NewUmbrellaManager(ctx)
	if err != nil {
		return err
	}

	cmd.SetContext(packmanager.WithPackageManager(ctx, pm))

	return nil
}

func (opts *Create) Run(cmd *cobra.Command, args []string) error {
	var err error

	ctx := cmd.Context()
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

	if len(opts.AuthorName) == 0 {
		input := textinput.New("Author name:")
		input.InitialValue = os.Getenv("USER")
		input.Placeholder = "Author name cannot be empty"

		opts.AuthorName, err = input.RunPrompt()
		if err != nil {
			return err
		}
	}

	templ, err := template.NewTemplate(ctx,
		template.WithProjectName(opts.ProjectName),
		template.WithLibName(opts.LibraryName),
		template.WithLibKName(opts.LibraryKName),
		template.WithVersion(opts.Version),
		template.WithDescription(opts.Description),
		template.WithAuthorName(opts.AuthorName),
		template.WithAuthorEmail(opts.AuthorEmail),
		template.WithInitialBranch(opts.InitialBranch),
		template.WithCopyrightHolder(opts.CopyrightHolder),
		template.WithProvideMain(!opts.NoProvideMain),
		template.WithGitignore(!opts.NoWithGitignore),
		template.WithDocs(!opts.NoWithDocs),
		template.WithPatchedir(opts.WithPatchedir),
	)
	if err != nil {
		return err
	}

	if err = templ.Generate(ctx, workdir); err != nil {
		return err
	}

	return nil
}
