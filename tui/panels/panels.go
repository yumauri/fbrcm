package panels

type ID int

const (
	None ID = iota
	Projects
	Parameters
	Conditions
	History
	ABTests
	Personalizations
	Rollouts
	Promote
	Details
	Logs
)
