package template

import (
	"context"
	"html/template"
	"os"
	"strings"

	_ "embed"
)

var (
	//go:embed CODING_STYLE.md.tmpl
	CodingStyleTemplate string

	//go:embed Config.uk.tmpl
	ConfigUkTemplate string

	//go:embed Config.uk.tmpl
	ContributingTemplate string

	//go:embed Config.uk.tmpl
	CopyingTemplate string
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

// Generate template using `.tmpl` files and `Template` struct fields.
func (t Template) Generate(ctx context.Context, workdir string) error {
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
	codingStyleTmpl, err := template.New("CondingStyleMd").Parse(CodingStyleTemplate)
	if err != nil {
		return err
	}

	configUkTmpl, err := template.New("ConfigUk").Parse(ConfigUkTemplate)
	if err != nil {
		return err
	}

	contributingMdTmpl, err := template.New("ContributingMd").Parse(ContributingTemplate)
	if err != nil {
		return err
	}

	copyingMdTmpl, err := template.New("CopyingMd").Parse(CopyingTemplate)
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
