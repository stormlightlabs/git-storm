/*
USAGE

	storm unreleased <subcommand> [options]

SUBCOMMANDS

	add         Add a new unreleased change entry
	list        List all unreleased changes
	review      Review unreleased changes interactively

USAGE

	storm unreleased add [options]

FLAGS

	--type <type>       Change type (added, changed, fixed, removed, security)
	--scope <scope>     Optional subsystem or module name
	--summary <text>    Short description of the change
	--repo <path>       Path to the repository (default: .)

USAGE

	storm unreleased list [options]

FLAGS

	--json              Output as JSON
	--repo <path>       Path to the repository (default: .)

USAGE

	storm unreleased review [options]

FLAGS

	--repo <path>       Path to the repository (default: .)
	--output <file>     Optional file to export reviewed notes
*/
package main
