package contract

type statelessCommandSupport struct {
	requiresAccessToken bool
}

var statelessCommands = map[string]statelessCommandSupport{
	"apply":                 {requiresAccessToken: true},
	"add":                   {requiresAccessToken: true},
	"conditions.list":       {requiresAccessToken: true},
	"conditions.add":        {requiresAccessToken: true},
	"conditions.delete":     {requiresAccessToken: true},
	"conditions.edit":       {requiresAccessToken: true},
	"conditions.move":       {requiresAccessToken: true},
	"conditions.rename":     {requiresAccessToken: true},
	"conditions.show":       {requiresAccessToken: true},
	"conditions.validate":   {requiresAccessToken: true},
	"delete":                {requiresAccessToken: true},
	"duplicate":             {requiresAccessToken: true},
	"experiments.list":      {requiresAccessToken: true},
	"experiments.delete":    {requiresAccessToken: true},
	"experiments.show":      {requiresAccessToken: true},
	"get":                   {requiresAccessToken: true},
	"groups.add":            {requiresAccessToken: true},
	"groups.delete":         {requiresAccessToken: true},
	"groups.edit":           {requiresAccessToken: true},
	"groups.list":           {requiresAccessToken: true},
	"groups.rename":         {requiresAccessToken: true},
	"personalizations.list": {requiresAccessToken: true},
	"personalizations.show": {requiresAccessToken: true},
	"project.defaults":      {requiresAccessToken: true},
	"project.export":        {requiresAccessToken: true},
	"project.import":        {requiresAccessToken: true},
	"project.open":          {},
	"project.show":          {requiresAccessToken: true},
	"projects.diff":         {requiresAccessToken: true},
	"projects.list":         {requiresAccessToken: true},
	"projects.promote":      {requiresAccessToken: true},
	"rollouts.list":         {requiresAccessToken: true},
	"rollouts.delete":       {requiresAccessToken: true},
	"rollouts.show":         {requiresAccessToken: true},
	"update":                {requiresAccessToken: true},
	"versions.diff":         {requiresAccessToken: true},
	"versions.export":       {requiresAccessToken: true},
	"versions.list":         {requiresAccessToken: true},
	"versions.rollback":     {requiresAccessToken: true},
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
