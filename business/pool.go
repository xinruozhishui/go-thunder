package business

type WorkerPool struct {
	workers map[string]*MonitoredWorker
}

// append a worker to the workerpool
func (wp *WorkerPool) AppendWork(iv *MonitoredWorker) {
	if wp.workers == nil {
		wp.workers = make(map[string]*MonitoredWorker)
	}
	wp.workers[iv.GetId()] = iv
}


// start all workers
func (wp *WorkerPool) StartAll() []error {
	var errs []error
	for _, value := range wp.workers {
		if err := value.Start(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// stop all workers
func (wp *WorkerPool) StopAll() []error {
	var errs []error
	for _, value := range wp.workers {
		if err := value.Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// get all workers'progress
func (wp *WorkerPool) GetAllProgress() interface{} {
	var pr []interface{}
	for _, value := range wp.workers {
		pr = append(pr, value.GetProgress())
	}
	return pr
}
