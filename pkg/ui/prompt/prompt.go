package prompt

type Prompt interface {
	Choose(title string, options []string, defaultOpt string) (answer string, err error)
	Confirm(title string) (confirm bool, err error)
	Select(title string, options []string, defaultOpt string) (answers []string, err error)
}
