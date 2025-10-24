package task

var stateTransitionMap = map[TaskState][]TaskState{
	Pending:   {Scheduled},
	Scheduled: {Scheduled, Running, Failed},
	Running:   {Running, Completed, Failed},
	Completed: {},
	Failed:    {},
}

func Contains(states []TaskState, state TaskState) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}

	return false
}

func ValidStateTransition(src TaskState, dst TaskState) bool {
	return Contains(stateTransitionMap[src], dst)
}
