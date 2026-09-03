// Package invocation describes application workflows independently of either
// frontend. Definitions own option defaults and handlers, not process startup,
// command-line parsing, protocol transport, or terminal lifecycle.
package invocation

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/pflag"
)

// Call is the invocation state used by a workflow. Cobra commands satisfy this
// interface, while protocol callers provide an independent request instance.
type Call interface {
	Context() context.Context
	SetContext(context.Context)
	Flags() *pflag.FlagSet
	InheritedFlags() *pflag.FlagSet
	CommandPath() string
	InOrStdin() io.Reader
	OutOrStdout() io.Writer
	ErrOrStderr() io.Writer
}

type Flags interface{ Flags() *pflag.FlagSet }
type FlagGroups interface {
	Flags
	MarkFlagsMutuallyExclusive(...string)
}

type Handler func(Call, []string) error

// Bounds are declarative positional argument constraints. The CLI adapter uses
// Cobra's existing validators; structured callers validate their input schema.
type Bounds struct {
	Min, Max int
	Kind     string
}

var NoArgs = Bounds{Kind: "none"}
var ArbitraryArgs = Bounds{Max: -1, Kind: "any"}

func ExactArgs(n int) Bounds        { return Bounds{Min: n, Max: n, Kind: "exact"} }
func MaximumNArgs(n int) Bounds     { return Bounds{Max: n, Kind: "maximum"} }
func RangeArgs(min, max int) Bounds { return Bounds{Min: min, Max: max, Kind: "range"} }

// Definition contains shared operation metadata, defaults, and implementation.
// Use retains the published argument spelling for help and schema generation.
type Definition struct {
	Use, Short, Long string
	Args             Bounds
	RunE             Handler
	Responses        []any
	NoData           bool
	Required         []string
	Exclusive        [][]string
	Children         []*Definition
	flags            *pflag.FlagSet
}

func (d *Definition) Name() string { return strings.Fields(d.Use)[0] }
func (d *Definition) Flags() *pflag.FlagSet {
	if d.flags == nil {
		d.flags = pflag.NewFlagSet(d.Name(), pflag.ContinueOnError)
	}
	return d.flags
}
func (d *Definition) AddCommand(children ...*Definition) {
	d.Children = append(d.Children, children...)
}
func (d *Definition) MarkFlagRequired(name string) error {
	if d.Flags().Lookup(name) == nil {
		return fmt.Errorf("unknown option %q", name)
	}
	d.Required = append(d.Required, name)
	return nil
}
func (d *Definition) MarkFlagsMutuallyExclusive(names ...string) {
	d.Exclusive = append(d.Exclusive, append([]string(nil), names...))
}
func (d *Definition) Find(path []string) (*Definition, error) {
	if len(path) == 0 {
		return d, nil
	}
	for _, child := range d.Children {
		if child.Name() == path[0] {
			return child.Find(path[1:])
		}
	}
	return nil, fmt.Errorf("unknown operation %q", strings.Join(path, "."))
}
func RegisterResponse(d *Definition, variants ...any) { d.Responses = variants }
func RegisterNoData(d *Definition)                    { d.NoData = true }
func MustRegisterResponsePath(d *Definition, path string, variants ...any) {
	child, err := d.Find(strings.Fields(path))
	if err != nil {
		panic(err)
	}
	RegisterResponse(child, variants...)
}

type versionKey struct{}

func WithVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, versionKey{}, version)
}
func Version(call Call) string {
	if call == nil || call.Context() == nil {
		return ""
	}
	version, _ := call.Context().Value(versionKey{}).(string)
	return version
}
