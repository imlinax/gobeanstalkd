package main

import "fmt"

type Tube struct {
	Name      string
	Jobs      []Job
	Reservers []*Client
	MinPri    uint32
}

func (tube *Tube) SaveJob(j *Job) {
	tube.Jobs = append(tube.Jobs, *j)
	if j.Pri < tube.MinPri {
		tube.MinPri = j.Pri
	}

	for _, client := range tube.Reservers {
		client.JobChan <- j
	}
	tube.Reservers = make([]*Client, 0)
}

func (tube *Tube) DeleteJob(id uint64) error {
	for i, j := range tube.Jobs {
		if j.ID == id {
			tube.Jobs = append(tube.Jobs[:i], tube.Jobs[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("not find")
}
