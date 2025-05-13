package entities

type Comment struct {
	id        int64
	userId    int64
	postId    int64
	writeTime string
	text      string
}
