package audit

type Probe struct {
	Steps map[string]string
}

type Event struct {
	Meta          map[string]string
	PodsCreated   int
	PodsDestroyed int
	ProbesFailed  int
	Probes        map[string]*Probe
}

// CountPodCreated increments PodsCreated for event
func (e *Event) CountPodCreated() {
	e.PodsCreated = e.PodsCreated + 1
}

// CountPodDestroyed increments PodsDestroyed for event
func (e *Event) CountPodDestroyed() {
	e.PodsDestroyed = e.PodsDestroyed + 1
}

// AuditProbe
func (e *Event) AuditProbe(name string, key string, err error) {
	if e.Probes == nil {
		e.Probes = make(map[string]*Probe)
	}
	if e.Probes[name] == nil {
		e.Probes[name] = new(Probe)
		e.Probes[name].Steps = make(map[string]string)
	}
	if err == nil {
		e.Probes[name].Steps[key] = "Passed"
	} else {
		e.Probes[name].Steps[key] = "Failed"
		e.ProbesFailed = e.ProbesFailed + 1
	}
}
