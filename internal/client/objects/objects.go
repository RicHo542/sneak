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
