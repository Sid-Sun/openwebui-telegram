package contract

type MessageLink struct {
	Parent   int
	Children []int
	Text     string
	From     string
}

type CompletionUpdate struct {
	Message string
	IsLast  bool
}
