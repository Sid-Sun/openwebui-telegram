package contract

type MessageLink struct {
	Parent int
	Text   string
	From   string
}

type CompletionUpdate struct {
	Content   string
	Reasoning string
	IsLast    bool
}
