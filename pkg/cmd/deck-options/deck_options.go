package options

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/aerex/anki-cli/pkg/anki"
	"github.com/aerex/anki-cli/pkg/template"
	"github.com/spf13/cobra"
)

type DeckOptsOptions struct {
  Edit bool
  Template string
}

func NewOptionsCmd(anki *anki.Anki, overrideF func(*anki.Anki) error) *cobra.Command {

  opts := &DeckOptsOptions{}
  cmd := &cobra.Command {
    Use: "options [deck_name]",
    Short: "Show or edit deck options",
    Example: heredoc.Doc(`
      $ anki deck options Default
      $ anki deck options Default --edit
      $ anki deck options Default group Default
      $ anki deck options Default new.steps 110
    `),
    RunE: func(cmd *cobra.Command, args []string) error {
      if overrideF != nil {
        return overrideF(anki)
      }
      return optionsCmd(anki, args, opts)
    },
  }

  cmd.Flags().BoolVarP(&opts.Edit, "edit", "e", false, "Edit one or multiple deck options in your editor")
  cmd.Flags().StringVarP(&opts.Template, "template", "t", "", "Override template for output")


    // TODO:
    // Add --edit option to open an editor to configure options as yaml (add a limit of 5 if not editing just one)
    // Add the ability to update nested options like new.steps
    // Add the ability to see all options using default template (deck-options) or given template

  return cmd
}

func optionsCmd(anki *anki.Anki, args []string, opts *DeckOptsOptions) error {

  // Handling one deck option
  if len(args) == 1 {
    tmpl := template.DECK_SINGLE_OPTION_LIST
    if opts.Template != "" {
      tmpl = opts.Template
    }
    if err := anki.Templates.Load(tmpl); err != nil {
      return err
    }
    options, err := anki.Api.GetDeckOptions(args[0])
    if err != nil {
      return err
    }
    if err := anki.Templates.Execute(options, anki.IO); err != nil {
      return err
    }
  }
  return nil
}
