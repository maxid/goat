package context

// Job represents a job.
type Job struct {
	Watcher *Watcher
	Message string
	Path    string
	Name    string
}
