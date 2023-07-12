package template

import (
	"context"
	"html/template"
	"os"
	"strings"
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
func (t Template) TemplateGenerator(ctx context.Context, workdir string) error {
	// Has to implement
	if !strings.HasSuffix(workdir, "/") {
		workdir += "/"
	}

	// fmt.Println("running templateGenerator")
	// fmt.Println("Template is", t)
	// os.ReadFile("to/path")
	// Implementation details:
	// Read all the `tmpl` files from the current directory one by one and
	// and create template.New("name of the template (set it diffrent for diffrent tmpl file)").Parse("Each file string")

	// readfiles, err := os.ReadDir("./")
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("files are: ")
	// for _, file := range readfiles {
	// 	fmt.Println(file.Name())
	// }

	// Parsing all the templates.
	codingStyleTmpl, err := template.New("CondingStyleMd").Parse(CodingStyleTemplateGenerator())
	if err != nil {
		return err
	}

	configUkTmpl, err := template.New("ConfigUk").Parse(ConfigUkTemplateGenerator())
	if err != nil {
		return err
	}

	contributingMdTmpl, err := template.New("ContributingMd").Parse(ContributingMdTemplateGenerator())
	if err != nil {
		return err
	}

	copyingMdTmpl, err := template.New("CopyingMd").Parse(CopyingMdTemplateGenerator())
	if err != nil {
		return err
	}

	mainFileTmpl, err := template.New("Main").Parse(MainTemplateGenerator())
	if err != nil {
		return err
	}

	// maintainerMdTmpl, err := template.New("MaintainerMd").Parse(MaintainersMdTemplateGenerator())
	// if err != nil {
	// 	return err
	// }

	// makefileUkTmpl, err := template.New("MakefileUk").Parse(MakefileUkGenerator())
	// if err != nil {
	// 	return err
	// }

	// manifestYamlTmpl, err := template.New("ManifestYaml").Parse(ManifestYamlTemplateGenerator())
	// if err != nil {
	// 	return err
	// }

	// readmeMdTmpl, err := template.New("ReadmeMd").Parse(ReadmeMdTemplateGenerator())
	// if err != nil {
	// 	return err
	// }

	// Creating projectName directory to store all the template files.
	projectDir := workdir + t.ProjectName + "/"
	err = os.Mkdir(projectDir, os.ModePerm)
	if err != nil {
		return err
	}

	// Creating template files to store template text.
	codingStyleFile, err := os.Create(projectDir + "CODING_STYLE.md")
	if err != nil {
		return err
	}

	configUkFile, err := os.Create(projectDir + "Config.uk")
	if err != nil {
		return err
	}

	contributingMdFile, err := os.Create(projectDir + "CONTRIBUTING.md")
	if err != nil {
		return err
	}

	copyingMdFile, err := os.Create(projectDir + "COPYING.md")
	if err != nil {
		return err
	}

	mainFile, err := os.Create(projectDir + "main.c")
	if err != nil {
		return err
	}

	// maintainerMdFile, err := os.Create(projectDir + "MAINTAINERS.md")
	// if err != nil {
	// 	return err
	// }

	// makefileUkfile, err := os.Create(projectDir + "Makefile.uk")
	// if err != nil {
	// 	return err
	// }

	// manifestYamlFile, err := os.Create(projectDir + "manifest.yaml")
	// if err != nil {
	// 	return err
	// }

	// readmeMdFile, err := os.Create(projectDir + "README.md")
	// if err != nil {
	// 	return err
	// }

	// Executing Templates with Template struct values
	err = codingStyleTmpl.Execute(codingStyleFile, t)
	if err != nil {
		return err
	}

	err = configUkTmpl.Execute(configUkFile, t)
	if err != nil {
		return err
	}

	err = contributingMdTmpl.Execute(contributingMdFile, t)
	if err != nil {
		return err
	}

	err = copyingMdTmpl.Execute(copyingMdFile, t)
	if err != nil {
		return err
	}

	err = mainFileTmpl.Execute(mainFile, t)
	if err != nil {
		return err
	}

	// err = maintainerMdTmpl.Execute(maintainerMdFile, t)
	// if err != nil {
	// 	return err
	// }

	// err = makefileUkTmpl.Execute(makefileUkfile, t)
	// if err != nil {
	// 	return err
	// }

	// err = manifestYamlTmpl.Execute(manifestYamlFile, t)
	// if err != nil {
	// 	return err
	// }

	// err = readmeMdTmpl.Execute(readmeMdFile, t)
	// if err != nil {
	// 	return err
	// }

	return nil
}
