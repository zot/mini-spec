// CRC: crc-CLI.md
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zot/minispec/internal/phase"
	"github.com/zot/minispec/internal/project"
	"github.com/zot/minispec/internal/query"
	"github.com/zot/minispec/internal/update"
	"github.com/zot/minispec/internal/validate"
)

// Version is set at build time via -ldflags
var Version = "dev"

// CLI handles command-line interface
type CLI struct {
	DesignDir string
	SrcDir    string
	Quiet     bool
	JSON      bool
}

// Run parses arguments and executes the appropriate command
func (c *CLI) Run(args []string) int {
	// Handle --version early (before other parsing)
	for _, arg := range args {
		if arg == "--version" || arg == "-v" {
			fmt.Printf("minispec %s\n", Version)
			return 0
		}
	}

	if len(args) < 1 {
		c.printUsage()
		return 1
	}

	// Parse global flags
	fs := flag.NewFlagSet("minispec", flag.ContinueOnError)
	fs.StringVar(&c.DesignDir, "design-dir", "", "Override design directory")
	fs.StringVar(&c.SrcDir, "src-dir", "", "Override source directory")
	fs.BoolVar(&c.Quiet, "quiet", false, "Minimal output")
	fs.BoolVar(&c.JSON, "json", false, "Output as JSON")

	// Find command position (first non-flag arg)
	cmdIdx := 0
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			cmdIdx = i
			break
		}
	}

	// Parse flags before command
	if cmdIdx > 0 {
		if err := fs.Parse(args[:cmdIdx]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}

	cmd := args[cmdIdx]
	cmdArgs := args[cmdIdx+1:]

	switch cmd {
	case "query":
		return c.runQuery(cmdArgs)
	case "update":
		return c.runUpdate(cmdArgs)
	case "validate":
		return c.runValidate(cmdArgs)
	case "phase":
		return c.runPhase(cmdArgs)
	case "help", "-h", "--help":
		c.printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		c.printUsage()
		return 1
	}
}

func (c *CLI) printUsage() {
	fmt.Printf("minispec %s - structural operations on mini-spec design files\n", Version)
	fmt.Println(`
Usage: minispec [flags] <command> [args]

Commands:
  query <subcommand>    Query design files
  update <subcommand>   Update design files
  validate              Run structural validations
  phase <phase-name>    Run phase-specific validation

Query subcommands:
  requirements          List all requirements
  coverage              Show requirement coverage by design files
  uncovered             List requirements with no design coverage
  orphan-designs        List CRC cards missing Requirements field
  artifacts             List artifacts with checkbox states
  gaps                  List gap items
  traceability <file>   Check file for traceability comments
  traceability --all    Check all code files

Update subcommands:
  check <file> <item>       Check a checkbox
  uncheck <file> <item>     Uncheck a checkbox
  add-ref <crc> <Rn>        Add requirement to CRC card
  remove-ref <crc> <Rn>     Remove requirement from CRC card
  add-gap <type> <desc>     Add new gap (type: S/R/D/C/O)
  resolve-gap <id>          Mark gap as resolved

Phase subcommands:
  spec                  Validate spec files exist
  requirements          Validate requirements.md format
  design                Validate design files and coverage
  implementation        Validate code files and traceability
  gaps                  Validate gaps section

Flags:
  --design-dir PATH    Override design directory
  --src-dir PATH       Override source directory
  --quiet              Minimal output
  --json               Output as JSON
  --version            Display version and exit`)
}

func (c *CLI) getProject() (*project.Project, error) {
	p, err := project.Detect()
	if err != nil {
		return nil, err
	}

	// Apply overrides
	if c.DesignDir != "" {
		p.DesignDir = c.DesignDir
	}
	if c.SrcDir != "" {
		p.SrcDir = c.SrcDir
	}

	return p, nil
}

func (c *CLI) output(data any) {
	if c.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(data)
	} else {
		fmt.Println(data)
	}
}

func (c *CLI) runQuery(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: minispec query <subcommand>")
		return 1
	}

	p, err := c.getProject()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	q := query.New(p)
	subcmd := args[0]

	switch subcmd {
	case "requirements":
		reqs, err := q.Requirements()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if c.JSON {
			c.output(reqs)
		} else {
			for _, r := range reqs {
				inferred := ""
				if r.Inferred {
					inferred = " (inferred)"
				}
				fmt.Printf("%s: %s%s\n", r.ID, r.Text, inferred)
			}
		}

	case "coverage":
		cov, err := q.Coverage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if c.JSON {
			c.output(cov)
		} else {
			for id, files := range cov.Coverage {
				if len(files) > 0 {
					fmt.Printf("%s: %s\n", id, strings.Join(files, ", "))
				} else {
					fmt.Printf("%s: (none)\n", id)
				}
			}
		}

	case "uncovered":
		uncovered, err := q.Uncovered()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if c.JSON {
			c.output(uncovered)
		} else {
			for _, id := range uncovered {
				fmt.Println(id)
			}
		}

	case "orphan-designs":
		orphans, err := q.OrphanDesigns()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if c.JSON {
			c.output(orphans)
		} else {
			for _, path := range orphans {
				fmt.Println(path)
			}
		}

	case "artifacts":
		artifacts, err := q.Artifacts()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if c.JSON {
			c.output(artifacts)
		} else {
			for _, art := range artifacts {
				fmt.Println(art.DesignFile)
				for _, cf := range art.CodeFiles {
					mark := " "
					if cf.Checked {
						mark = "x"
					}
					fmt.Printf("  [%s] %s\n", mark, cf.Path)
				}
			}
		}

	case "gaps":
		gaps, err := q.Gaps()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if c.JSON {
			c.output(gaps)
		} else {
			for _, g := range gaps {
				mark := " "
				if g.Resolved {
					mark = "x"
				}
				fmt.Printf("[%s] %s: %s\n", mark, g.ID, g.Description)
			}
		}

	case "traceability":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: minispec query traceability <file> or --all")
			return 1
		}
		if args[1] == "--all" {
			traces, err := q.TraceabilityAll()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return 1
			}
			if c.JSON {
				c.output(traces)
			} else {
				for path, trace := range traces {
					if len(trace.CRCRefs) > 0 {
						fmt.Printf("%s: CRC=%s", path, strings.Join(trace.CRCRefs, ","))
						if len(trace.SeqRefs) > 0 {
							fmt.Printf(" Seq=%s", strings.Join(trace.SeqRefs, ","))
						}
						fmt.Println()
					} else {
						fmt.Printf("%s: (missing)\n", path)
					}
				}
			}
		} else {
			trace, err := q.Traceability(args[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return 1
			}
			if c.JSON {
				c.output(trace)
			} else {
				if len(trace.CRCRefs) > 0 {
					fmt.Printf("CRC: %s\n", strings.Join(trace.CRCRefs, ", "))
				}
				if len(trace.SeqRefs) > 0 {
					fmt.Printf("Seq: %s\n", strings.Join(trace.SeqRefs, ", "))
				}
				if len(trace.CRCRefs) == 0 && len(trace.SeqRefs) == 0 {
					fmt.Println("(no traceability comments found)")
				}
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown query subcommand: %s\n", subcmd)
		return 1
	}

	return 0
}

func (c *CLI) runUpdate(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: minispec update <subcommand>")
		return 1
	}

	p, err := c.getProject()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	u := update.New(p)
	subcmd := args[0]

	switch subcmd {
	case "check":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: minispec update check <file> <item>")
			return 1
		}
		if err := u.Check(args[1], args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if !c.Quiet {
			fmt.Printf("Checked %s in %s\n", args[2], args[1])
		}

	case "uncheck":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: minispec update uncheck <file> <item>")
			return 1
		}
		if err := u.Uncheck(args[1], args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if !c.Quiet {
			fmt.Printf("Unchecked %s in %s\n", args[2], args[1])
		}

	case "add-ref":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: minispec update add-ref <crc-file> <Rn>")
			return 1
		}
		if err := u.AddRef(args[1], args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if !c.Quiet {
			fmt.Printf("Added %s to %s\n", args[2], args[1])
		}

	case "remove-ref":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: minispec update remove-ref <crc-file> <Rn>")
			return 1
		}
		if err := u.RemoveRef(args[1], args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if !c.Quiet {
			fmt.Printf("Removed %s from %s\n", args[2], args[1])
		}

	case "add-gap":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: minispec update add-gap <type> <description>")
			return 1
		}
		gapType := strings.ToUpper(args[1])
		if !strings.Contains("SRDCO", gapType) || len(gapType) != 1 {
			fmt.Fprintln(os.Stderr, "Gap type must be one of: S, R, D, C, O")
			return 1
		}
		desc := strings.Join(args[2:], " ")
		id, err := u.AddGap(gapType, desc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if !c.Quiet {
			fmt.Printf("Added %s: %s\n", id, desc)
		}

	case "resolve-gap":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: minispec update resolve-gap <id>")
			return 1
		}
		if err := u.ResolveGap(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if !c.Quiet {
			fmt.Printf("Resolved %s\n", args[1])
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown update subcommand: %s\n", subcmd)
		return 1
	}

	return 0
}

func (c *CLI) runValidate(_ []string) int {
	p, err := c.getProject()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	v := validate.New(p)
	result, err := v.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if c.JSON {
		c.output(result)
	} else {
		fmt.Print(result.FormatText())
	}

	if result.HasIssues() {
		return 1
	}
	return 0
}

func (c *CLI) runPhase(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: minispec phase <spec|requirements|design|implementation|gaps>")
		return 1
	}

	p, err := c.getProject()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	ph := phase.New(p)
	phaseName := args[0]

	var result *phase.Result

	switch phaseName {
	case "spec":
		result = ph.RunSpec()
	case "requirements":
		result = ph.RunRequirements()
	case "design":
		result = ph.RunDesign()
	case "implementation":
		result = ph.RunImplementation()
	case "gaps":
		result = ph.RunGaps()
	default:
		fmt.Fprintf(os.Stderr, "Unknown phase: %s\n", phaseName)
		fmt.Fprintln(os.Stderr, "Valid phases: spec, requirements, design, implementation, gaps")
		return 1
	}

	if c.JSON {
		c.output(result)
	} else {
		fmt.Print(result.FormatText())
	}

	if !result.Passed {
		return 1
	}
	return 0
}
