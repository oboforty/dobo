package commands

type TableCmd struct {
	Table string
}

type GetTableCmd struct {
	TableCmd
}

func (p GetTableCmd) Run() error {

	return nil
}

type CreateTableCmd struct {
	TableCmd
}

func (p CreateTableCmd) Run() error {

	return nil
}

type EditTableSettingsCmd struct {
	TableCmd
}

func (p EditTableSettingsCmd) Run() error {

	return nil
}

type DropTableCmd struct {
	TableCmd
}

func (p DropTableCmd) Run() error {

	return nil
}
