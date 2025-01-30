package commands

type ItemCmd struct {
	Table   string
	PartKey []byte
	Value   []byte
}

type GetItemCmd struct {
	ItemCmd
}

func (p GetItemCmd) Run() error {

	return nil
}

type PutItemCmd struct {
	ItemCmd
}

func (p PutItemCmd) Run() error {

	return nil
}

type UpdateItemCmd struct {
	ItemCmd
}

func (p UpdateItemCmd) Run() error {

	return nil
}

type DelItemCmd struct {
	ItemCmd
}

func (p DelItemCmd) Run() error {

	return nil
}
