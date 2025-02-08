package commands

import "io"

type TableCmd struct {
	Table string
}

type ListTablesCmd struct {
}

func (p ListTablesCmd) Run(node Node, writer io.Writer) error {

	return nil
}

type GetTableCmd struct {
	TableCmd
}

func (p GetTableCmd) Run(node Node, writer io.Writer) error {

	return nil
}

type CreateTableCmd struct {
	TableCmd
}

func (p CreateTableCmd) Run(node Node, writer io.Writer) error {

	return nil
}

type EditTableSettingsCmd struct {
	TableCmd
}

func (p EditTableSettingsCmd) Run(node Node, writer io.Writer) error {

	return nil
}

type DropTableCmd struct {
	TableCmd
}

func (p DropTableCmd) Run(node Node, writer io.Writer) error {

	return nil
}
