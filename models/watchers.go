package models

type mutateWatchers struct {
	positions chan posUpdateMsg
}

func MutateWatchers() mutateWatchers {
	return mutateWatchers{
		positions: watchPositionUpdates(),
	}
}
