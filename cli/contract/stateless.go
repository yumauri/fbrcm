package contract

type statelessCommandSupport struct {
	requiresAccessToken bool
}

var statelessCommands = map[string]statelessCommandSupport{
	"conditions.list":       {requiresAccessToken: true},
	"conditions.show":       {requiresAccessToken: true},
	"conditions.validate":   {requiresAccessToken: true},
	"experiments.list":      {requiresAccessToken: true},
	"experiments.show":      {requiresAccessToken: true},
	"get":                   {requiresAccessToken: true},
	"groups.list":           {requiresAccessToken: true},
	"personalizations.list": {requiresAccessToken: true},
	"personalizations.show": {requiresAccessToken: true},
	"project.defaults":      {requiresAccessToken: true},
	"project.export":        {requiresAccessToken: true},
	"project.open":          {},
	"project.show":          {requiresAccessToken: true},
	"projects.diff":         {requiresAccessToken: true},
	"projects.list":         {requiresAccessToken: true},
	"rollouts.list":         {requiresAccessToken: true},
	"rollouts.show":         {requiresAccessToken: true},
	"versions.diff":         {requiresAccessToken: true},
	"versions.export":       {requiresAccessToken: true},
	"versions.list":         {requiresAccessToken: true},
	"versions.show":         {requiresAccessToken: true},
}

// SupportsStatelessCommand reports whether a command has a complete
// profileless execution path and matching machine contract.
func SupportsStatelessCommand(commandID string) bool {
	_, ok := statelessCommands[commandID]
	return ok
}

// StatelessCommandRequiresAccessToken reports whether a supported stateless
// command contacts a Google API with the one-shot access token.
func StatelessCommandRequiresAccessToken(commandID string) bool {
	support, ok := statelessCommands[commandID]
	return ok && support.requiresAccessToken
}
