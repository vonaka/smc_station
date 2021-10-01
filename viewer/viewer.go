package viewer

type Viewer struct {
	actions chan *Action
}

type Action struct {
	Type string `json:"type"`
	Wait string `json:"wait,omitempty"`
}

func New() *Viewer {
	return &Viewer{
		actions: make(chan *Action, 4),
	}
}

func (v *Viewer) Record(a *Action) {
	v.actions <- a
}

func (v *Viewer) GetAction() *Action {
	return <-v.actions
}
