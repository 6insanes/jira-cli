package root

import (
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var helpFlagRef = regexp.MustCompile(`--[a-z][a-z0-9-]*`)

// TestHelpMentionsOnlyKnownFlags guards against examples and descriptions that reference
// flags a command does not accept.
func TestHelpMentionsOnlyKnownFlags(t *testing.T) {
	root := NewCmdRoot()

	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		known := map[string]bool{}
		collect := func(f *pflag.Flag) {
			known[f.Name] = true
			if f.Shorthand != "" {
				known[f.Shorthand] = true
			}
		}
		c.Flags().VisitAll(collect)
		c.InheritedFlags().VisitAll(collect)

		refs := helpFlagRef.FindAllString(c.Example, -1)
		refs = append(refs, helpFlagRef.FindAllString(c.Long, -1)...)
		c.Flags().VisitAll(func(f *pflag.Flag) {
			refs = append(refs, helpFlagRef.FindAllString(f.Usage, -1)...)
		})

		for _, ref := range refs {
			if name := strings.TrimPrefix(ref, "--"); !known[name] {
				t.Errorf("%s: help references unknown flag %s", c.CommandPath(), ref)
			}
		}

		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)
}
