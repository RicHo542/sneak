package objects

type WorkItem struct {
	ID       string
	Key      string
	Summary  string
	Status   string
	Type     string
	Assignee string
}

type ListOptions struct {
	Bindings []string
	Types    []string
}

type UserInfo struct {
	UserHandle  string
	DisplayName string
}

type Comment struct {
	Author    string
	Body      string
	CreatedAt string
}

type WorkItemDetail struct {
	ID            string
	Key           string
	URL           string
	Name          string
	Description   string
	CreatedAt     string
	CreatedBy     string
	IterationPath string
	Owner         string
	Comments      []Comment
	TotalComments int
}
