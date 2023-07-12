package remove

import (
	"fmt"

	"github.com/spf13/cobra"
	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/pack"
	"kraftkit.sh/packmanager"
	"kraftkit.sh/unikraft"
	"kraftkit.sh/unikraft/app"
	"kraftkit.sh/unikraft/lib"
)

type Add struct{}

func New() *cobra.Command {
	cmd, err := cmdfactory.New(&Add{}, cobra.Command{
		Short: "add unikraft package to a project",
		Use:   "add [FLAGS] [PACKAGE|DIR]",
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "pkg",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *Add) Pre(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	pm, err := packmanager.NewUmbrellaManager(ctx)
	if err != nil {
		return err
	}

	cmd.SetContext(packmanager.WithPackageManager(ctx, pm))

	return nil
}

func (opts *Add) Run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	project, err := app.NewApplicationFromOptions()

	packs, err := packmanager.G(ctx).Catalog(ctx,
		packmanager.WithName(args[0]),
		packmanager.WithTypes(unikraft.ComponentTypeLib),
	)
	if err != nil {
		return err
	}

	if len(packs) != 1 {
		return fmt.Errorf("!")
	}

	// pull package

	packs[0].Pull(ctx,
		pack.WithPullWorkdir(workdir),
	)

	library, err := lib.NewLibraryFromPackage(ctx, packs[0])
	if err != nil {
		return err
	}

	project.AddLibrary(ctx, library)

	return project.Save()
}
