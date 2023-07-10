package template

import (
	"context"
)

type Template struct {
	ProjectName     string
	LibName         string
	LibKName        string
	Version         string
	Description     string
	AuthorName      string
	AuthorEmail     string
	ProvideMain     bool
	WithGitignore   bool
	WithDocs        bool
	WithPatchedir   bool
	InitialBranch   string
	CopyrightHolder string
}

type TemplateOption func(*Template)

func NewTemplate(ctx context.Context, topts ...TemplateOption) (Template, error) {
	// Has to implement
	var templ Template

	// Initialising default values to the template struct.
	templ.LibName = "lib-template"
	templ.LibKName = "LIBTEMPLATE"
	templ.InitialBranch = "staging"
	templ.ProvideMain = true
	templ.WithGitignore = true
	templ.WithDocs = true
	templ.WithPatchedir = false

	// Initialising custom values to the template struct.
	for _, topt := range topts {
		topt(&templ)
	}

	return templ, nil
}

func WithProjectName(projectName string) TemplateOption {
	return func(t *Template) {
		t.ProjectName = projectName
	}
}

func WithLibName(libName string) TemplateOption {
	return func(t *Template) {
		t.LibName = libName
	}
}

func WithLibKName(libKName string) TemplateOption {
	return func(t *Template) {
		t.LibKName = libKName
	}
}

func WithVersion(version string) TemplateOption {
	return func(t *Template) {
		t.Version = version
	}
}

func WithDescription(description string) TemplateOption {
	return func(t *Template) {
		t.Description = description
	}
}

func WithAuthorName(authorName string) TemplateOption {
	return func(t *Template) {
		t.AuthorName = authorName
	}
}

func WithAuthorEmail(authorEmail string) TemplateOption {
	return func(t *Template) {
		t.AuthorEmail = authorEmail
	}
}

func WithProvideMain(provideMain bool) TemplateOption {
	return func(t *Template) {
		t.ProvideMain = provideMain
	}
}

func WithGitignore(gitIgnore bool) TemplateOption {
	return func(t *Template) {
		t.WithGitignore = gitIgnore
	}
}

func WithDocs(docs bool) TemplateOption {
	return func(t *Template) {
		t.WithDocs = docs
	}
}

func WithPatchedir(patchedir bool) TemplateOption {
	return func(t *Template) {
		t.WithPatchedir = patchedir
	}
}

func WithInitialBranch(initialBranch string) TemplateOption {
	return func(t *Template) {
		t.InitialBranch = initialBranch
	}
}

func WithCopyrightHolder(copyrightHolder string) TemplateOption {
	return func(t *Template) {
		t.CopyrightHolder = copyrightHolder
	}
}

// Generate template using `.tmpl` files and `Template` struct fields.
func (t Template) templateGenerator(ctx context.Context) error {
	// Has to implement

	// os.ReadFile("to/path")
	// Implementation details:
	// Read all the `tmpl` files from the current directory one by one
	// and create template.New("name of the template (set it diffrent for diffrent tmpl file)").Parse("Each file string")
	return nil
}
