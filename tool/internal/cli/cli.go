// CRC: crc-CLI.md | R79, R80, R81, R82, R83, R89
package cli

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zot/minispec/internal/alarm"
	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/phase"
	"github.com/zot/minispec/internal/project"
	"github.com/zot/minispec/internal/query"
	"github.com/zot/minispec/internal/update"
	"github.com/zot/minispec/internal/validate"
	"github.com/zot/simple-dom/minispecsdom"
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

	// An unrecognised command is reported as such before the bootstrap gate runs: a
	// typo deserves "unknown command", not "this isn't a minispec project".
	if !knownCommands[cmd] {
		return c.unknownCommand(cmd)
	}

	// CRC: crc-CLI.md | Seq: seq-bootstrap.md#1 | R152, R154
	// Nothing beyond a version report runs until the repository has a configuration and
	// a track value that agrees with what git reports.
	if code, stop := c.gate(cmd); stop {
		return code
	}

	switch cmd {
	case "init":
		return c.runInit(cmdArgs)
	case "check-version":
		return c.runCheckVersion()
	case "query":
		return c.runQuery(cmdArgs)
	case "update":
		return c.runUpdate(cmdArgs)
	case "validate":
		return c.runValidate(cmdArgs)
	case "phase":
		return c.runPhase(cmdArgs)
	case "pending":
		return c.runPending(cmdArgs)
	case "help", "-h", "--help":
		c.printUsage()
		return 0
	default:
		// Unreachable while knownCommands and this switch agree, which
		// TestKnownCommandsCoversEveryDispatchedCommand is what keeps true.
		return c.unknownCommand(cmd)
	}
}

// unknownCommand reports a command the dispatcher does not accept. Shared by the
// pre-gate check and the dispatch switch so the two cannot drift into wording the same
// refusal differently.
func (c *CLI) unknownCommand(cmd string) int {
	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
	c.printUsage()
	return 1
}

func (c *CLI) printUsage() {
	fmt.Printf("minispec %s - structural operations on mini-spec design files\n", Version)
	fmt.Println(`
Usage: minispec [flags] <command> [args]

Commands:
  init --track-<style>  Create the repository config (--track-none|-private-trajectory|-all)
  check-version         Verify tool and skill versions match
  query <subcommand>    Query design files
  update <subcommand>   Update design files
  validate              Run structural validations
  validate trajectory   Check the queue files and carves against each other
  phase <phase-name>    Run phase-specific validation
  pending <subcommand>  Trajectory-item operations and the backup slot

Query subcommands:
  project               Show resolved paths (repo root, design root, design, src, specs)
  config                Show effective settings with the file each came from
  requirements          List all requirements
  coverage              Show requirement coverage by design files
  uncovered             List requirements with no design coverage
  orphan-designs        List CRC cards missing Requirements field
  artifacts             List artifacts with checkbox states
  gaps                  List gap items
  migrations            List in-flight migration specs
  unindexed-specs       List specs not referenced in the root index (specs/index.md)
  alarms [--unverified] [--brief]  Fire alarms with their freshness state; --unverified lists only what carries a decision (the count still covers every alarm); --brief prints the spawn prompt for a delegated re-pull
  carves [--open]       Open and landed parts per carve, from the status block only; --open lists the open parts and stateless lines
  next-id <class>       Next free ID for item|gap|req, with the files it counted
  traceability <file>   Check file for traceability comments
  traceability --all    Check all code files
  comment-patterns      Show recognized comment patterns per file extension

Update subcommands:
  check <file> <item>           Check a checkbox
  uncheck <file> <item>         Uncheck a checkbox
  add-ref <crc> <Rn>            Add requirement to CRC card
  remove-ref <crc> <Rn>         Remove requirement from CRC card
  add-gap <type> <desc>         Add new gap (type: S/R/D/C/I/O/A/T)
  resolve-gap <id>              Mark gap as resolved (S/R/D/C/I/O only)
  approve-gap <id>              Convert gap to approved (A) type
  retire <Rold> <Rnew|-> <reason>  Retire a requirement, append a Tn gap
  migration-complete <name>     Move migration spec to complete/ with NNN- prefix

Pending subcommands (the trajectory files at the repository root):
  add-item <title> --from <doc>#<part>|<gap-id> --status <text> [--skill S] [--next-action T] [--next|--nth N|--after N|--last]
  start <N> [--context T|--context-file F]
  finish <N> --commit <hash> [--body T|--body-file F] [--discharged T|--discharged-file F] (--resolve|--no-resolve)
  revert                Undo the most recent trajectory change (one level)
  replay                Redo what revert undid

Phase subcommands:
  spec                  Validate spec files exist
  requirements          Validate requirements.md format
  design                Validate design files and coverage
  implementation        Validate code files and traceability
  gaps                  Validate gaps section

Init flags:
  --track-none                 No version control
  --track-private-trajectory   Git; trajectory files ignored
  --track-all                  Git; trajectory files tracked
  --repair                     Reconcile an existing config and .gitignore

Flags:
  --design-dir PATH    Override design directory
  --src-dir PATH       Override source directory
  --quiet              Minimal output
  --json               Output as JSON
  --version            Display version and exit`)
}

func (c *CLI) runCheckVersion() int {
	// Directories that may own a `.claude/`, in precedence order. The layout beneath
	// each one is identical, so it is spelled out once at the point of use below.
	var bases []string

	// CRC: crc-CLI.md | R117
	// `.claude/` is repository-scoped, so the lookup anchors at the repository root
	// rather than the current directory. Anchoring at cwd meant that running from
	// anywhere below the top found no local .claude/ and fell silently through to the
	// user-level skill — while still printing a verdict. Measured on a project that
	// vendors its own copy of the skill to pin a version: from a subdirectory the
	// check reported "ok" against the user-level skill and ignored the pin entirely.
	if root, err := project.RepoRoot(); err == nil {
		bases = append(bases, root)
	}

	// User-level
	if home, err := os.UserHomeDir(); err == nil {
		bases = append(bases, home)
	}

	for _, base := range bases {
		path := filepath.Join(base, ".claude", "skills", "mini-spec", "README.md")
		ver, err := readSkillVersion(path)
		if err != nil {
			continue
		}
		if ver == Version {
			fmt.Printf("ok: tool and skill both at %s\n", Version)
			return 0
		}
		fmt.Fprintf(os.Stderr, "version mismatch: tool=%s skill=%s (%s)\n", Version, ver, path)
		return 1
	}

	fmt.Fprintln(os.Stderr, "skill README.md not found in .claude/skills/mini-spec/")
	return 1
}

func readSkillVersion(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if ver, ok := strings.CutPrefix(line, "Version:"); ok {
			return strings.TrimSpace(ver), nil
		}
	}
	// Distinguish "read failed" from "read fine, no Version line". Without this a
	// mid-scan I/O error reports absence, and the caller moves on to the next
	// candidate as though this one had simply been silent. R117
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no Version line found")
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

	// CRC: crc-CLI.md | R191, R198
	// The one subcommand that must answer before a design root is resolved. Item IDs are
	// repository-scoped and live outside any design root, so demanding a project here
	// made the command fail in exactly the layout it exists for — this repository, whose
	// design roots are tool/ and example/ while the queue sits above both. Found by
	// running it, not by review.
	if len(args) > 1 && args[0] == "next-id" && args[1] == "item" {
		return c.emitNextID(query.NextItemID())
	}
	// CRC: crc-CLI.md | Seq: seq-carve-status.md#1 | R213
	// Carves are repository-scoped too, and this repository's two design roots sit below
	// the queue and the carves alike.
	if args[0] == "carves" {
		return c.queryCarves(args[1:])
	}

	p, err := c.getProject()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	q := query.New(p)
	subcmd := args[0]

	switch subcmd {
	// CRC: crc-CLI.md | R89, R115, R116
	case "project":
		repoRoot, repoErr := project.RepoRoot()
		info := map[string]string{
			"root":   p.RootPath,
			"design": p.DesignDir,
			"src":    p.SrcDir,
			"specs":  p.SpecsDir(),
		}
		if repoErr == nil {
			info["repo_root"] = repoRoot
		}
		if c.JSON {
			c.output(info)
		} else {
			// The same-directory case is stated rather than printed twice unlabeled:
			// the two roots coincide in most projects, which is exactly why the
			// distinction stayed invisible for so long. R116
			switch {
			case repoErr != nil:
				fmt.Printf("repo root:   (not found: %v)\n", repoErr)
			case repoRoot == p.RootPath:
				fmt.Printf("repo root:   %s (same as design root)\n", repoRoot)
			default:
				fmt.Printf("repo root:   %s\n", repoRoot)
			}
			fmt.Printf("design root: %s\n", p.RootPath)
			fmt.Printf("design:      %s\n", p.DesignDir)
			fmt.Printf("src:         %s\n", p.SrcDir)
			fmt.Printf("specs:       %s\n", p.SpecsDir())
		}

	// CRC: crc-CLI.md | R129
	case "config":
		settings := p.EffectiveSettings()
		if c.JSON {
			c.output(settings)
		} else {
			width := 0
			for _, s := range settings {
				width = max(width, len(s.Name))
			}
			for _, s := range settings {
				fmt.Printf("%-*s  %-26s  %s\n", width, s.Name, s.Value, s.Origin)
			}
		}

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
				if !g.HasCheckbox {
					fmt.Printf("    %s: %s\n", g.ID, g.Description)
					continue
				}
				mark := " "
				if g.Resolved {
					mark = "x"
				}
				fmt.Printf("[%s] %s: %s\n", mark, g.ID, g.Description)
			}
		}

	case "migrations":
		migs, err := q.Migrations()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if c.JSON {
			c.output(migs)
		} else {
			for _, m := range migs {
				fmt.Println(m)
			}
		}

	case "alarms":
		return c.queryAlarms(p, args[1:])

	case "unindexed-specs":
		specs, err := q.UnindexedSpecs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if c.JSON {
			c.output(specs)
		} else {
			for _, s := range specs {
				fmt.Println(s)
			}
		}

	// CRC: crc-CLI.md | R189, R196
	case "next-id":
		// The class is required rather than defaulted: it selects the root as well as
		// the count, so guessing it would answer a question nobody asked. R189
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: minispec query next-id <item|gap|req>")
			return 1
		}
		return c.emitNextID(q.NextID(args[1]))

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

	case "comment-patterns":
		patterns := q.CommentPatterns()
		closers := q.CommentClosers()
		if c.JSON {
			c.output(map[string]any{
				"patterns": patterns,
				"closers":  closers,
			})
		} else {
			fmt.Println("Recognized comment patterns:")
			for ext, pattern := range patterns {
				fmt.Printf("  %s: %s\n", ext, pattern)
			}
			if len(closers) > 0 {
				fmt.Println()
				fmt.Println("Comment closers (MUST appear at end of traceability comments):")
				for ext, closer := range closers {
					fmt.Printf("  %s: %q\n", ext, closer)
				}
				fmt.Println()
				fmt.Println("WARNING: Extensions with closers use block comments.")
				fmt.Println("An unclosed comment will silently swallow all subsequent code.")
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
		if !parser.IsGapType(gapType) {
			fmt.Fprintf(os.Stderr, "Gap type must be one of: %s\n", strings.Join(parser.GapTypes, ", "))
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

	case "approve-gap":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: minispec update approve-gap <id>")
			return 1
		}
		newID, err := u.ApproveGap(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if !c.Quiet {
			fmt.Printf("Approved %s -> %s\n", args[1], newID)
		}

	case "retire":
		if len(args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: minispec update retire <Rold> <Rnew|-> <reason>")
			return 1
		}
		reason := strings.Join(args[3:], " ")
		// CRC: crc-Update.md, crc-CLI.md | Seq: seq-update.md | R80, R103
		tn, sources, err := u.Retire(args[1], args[2], reason)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Println(tn)
		if !c.Quiet {
			fmt.Fprint(os.Stderr, retireReminder(args[1], sources))
		}

	case "migration-complete":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: minispec update migration-complete <name>")
			return 1
		}
		newPath, err := u.MigrationComplete(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Println(newPath)

	default:
		fmt.Fprintf(os.Stderr, "Unknown update subcommand: %s\n", subcmd)
		return 1
	}

	return 0
}

// retireReminder builds the supersede-at-source reminder printed to stderr
// after a retirement: striking out an Rn does not remove the spec/design prose
// that described its old behavior, and that stale prose can lead a future agent
// to revert the change. It names the retired requirement's **Source:** spec(s)
// so the originating prose is one pointer away.
// CRC: crc-CLI.md | Seq: seq-update.md | R103
func retireReminder(oldReq string, sources []string) string {
	specLine := "(no **Source:** recorded — locate its originating spec by hand)"
	if len(sources) > 0 {
		specLine = strings.Join(sources, ", ") + "  (" + oldReq + "'s **Source:**)"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "⚠ supersede at the source: %s is retired, but directives describing its\n", oldReq)
	b.WriteString("  old behavior remain and can lead a future agent to revert the change.\n")
	b.WriteString("  Reconcile every place the old behavior is still described:\n")
	fmt.Fprintf(&b, "    • originating spec prose: %s\n", specLine)
	fmt.Fprintf(&b, "    • design prose: grep design/ for %s and the old names; fix CRC bullets\n", oldReq)
	b.WriteString("      and seq steps that still describe the old state\n")
	b.WriteString("  Completion test: could an agent reading only specs + design be led to undo\n")
	b.WriteString("  this change? If yes, a trap remains.\n")
	return b.String()
}

func (c *CLI) runValidate(args []string) int {
	// CRC: crc-CLI.md | Seq: seq-validate-trajectory.md#1.1 | R285
	// Repository-scoped, and dispatched before any design root is resolved: a repository
	// may hold several design roots — this one holds tool/ and example/ — so the trajectory
	// layer above them is a subcommand rather than a section.
	if len(args) > 0 && args[0] == "trajectory" {
		return c.runValidateTrajectory()
	}
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

// CRC: crc-CLI.md | Seq: seq-validate-trajectory.md#1.4 | R285, R299
// runValidateTrajectory checks the queue files and carves against each other.
func (c *CLI) runValidateTrajectory() int {
	repoRoot, err := project.RepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	issues, err := validate.RunTrajectory(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if c.JSON {
		c.output(issues)
	} else {
		fmt.Print(issues.FormatText())
	}
	if issues.HasIssues() {
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

// CRC: crc-Query.md | Seq: seq-alarm-freshness.md#2.4 | R185, R186, R187
// queryAlarms prints the census of recorded fault injections.
//
// Asked rather than emitted. `validate` reports only the closable states; the two that
// stay non-zero for months while a project adopts the convention live here, where
// looking at them is a decision rather than a line you learn to scroll past.
func (c *CLI) queryAlarms(p *project.Project, args []string) int {
	unverified, brief, err := c.parseAlarmArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	docs, err := filepath.Glob(filepath.Join(p.DesignDir, "test-*.md"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	var alarms []parser.Alarm
	for _, d := range docs {
		found, ferr := parser.ParseTestDoc(d)
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "Error: %s: %v\n", d, ferr)
			return 1
		}
		alarms = append(alarms, found...)
	}
	assessments := alarm.Assess(alarms, project.NewGit(p.RootPath))

	// R199 — the census is computed over the **whole** population, before the selection
	// narrows the list, and printed whichever way the list was filtered. A count filtered
	// along with its list answers *how many are wrong* and silently drops *out of how many*.
	census := alarmCensusLine(assessments)

	// R199, R204 — the selection precedes the output form. Filtering inside the text branch
	// renders everything under `--json` while every markdown test stays green, which is O28's
	// silent ignore one level in — the same defect `query gaps` already met.
	selected := query.SelectAlarms(assessments, unverified)

	if brief {
		artifacts, aerr := parser.ParseArtifacts(p.DesignMdPath())
		if aerr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", aerr)
			return 1
		}
		briefs := query.AlarmBriefs(selected, designRootRel(p), artifacts)
		if c.JSON {
			c.output(briefs)
			return 0
		}
		for _, b := range briefs {
			fmt.Println(b.Brief)
		}
		fmt.Print(census)
		return 0
	}

	if c.JSON {
		c.output(selected)
		return 0
	}
	lastDoc := ""
	for _, a := range selected {
		if a.Alarm.Doc != lastDoc {
			lastDoc = a.Alarm.Doc
			fmt.Printf("%s:\n", lastDoc)
		}
		// R186 — the label says the repository does not record a verification. It does
		// not say the injection was never run, because the documents cannot answer that
		// and a tool asserting the stronger claim would be manufacturing a finding.
		detail := ""
		switch a.State {
		case alarm.Stale:
			detail = fmt.Sprintf(" — pulled %s, %s changed %s",
				a.Alarm.Pulled.Format("2006-01-02"), a.Site, a.Changed)
		case alarm.Unresolvable:
			detail = fmt.Sprintf(" — git cannot resolve %s", a.Site)
		case alarm.Unrecorded:
			detail = " — no verification recorded (not a claim it was never run)"
		case alarm.Unanchored:
			detail = " — no Inject:, so nothing can check it"
		}
		fmt.Printf("  %-13s %s%s\n", a.State, a.Alarm.Test, detail)
	}

	fmt.Print(census)
	return 0
}

// CRC: crc-Query.md | Seq: seq-alarm-freshness.md#2.5 | R185, R199
// alarmCensusLine renders the closing count, always over the whole population.
//
// Separated from the listing so that `--unverified` cannot narrow it by accident: the one
// call site that could filter it is the one that computes it, and it takes the full slice.
func alarmCensusLine(assessments []alarm.Assessment) string {
	census := alarm.Census(assessments)
	parts := make([]string, 0, len(alarm.CensusOrder))
	for _, st := range alarm.CensusOrder {
		if census[st] > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", census[st], st))
		}
	}
	return fmt.Sprintf("\n%d alarms: %s\n", len(assessments), strings.Join(parts, ", "))
}

// CRC: crc-Query.md | Seq: seq-alarm-freshness.md#3.1 | R200
// designRootRel expresses the design root as a path relative to the repository root.
//
// A delegated puller runs in its own git worktree, where this checkout's absolute path
// resolves to nothing — or worse, resolves to the very tree the worktree exists to protect,
// so an injection would land in the working copy. When the repository root cannot be
// resolved at all the answer is the empty string, which the brief states as unknown rather
// than papering over with an absolute path or a confident `.`.
func designRootRel(p *project.Project) string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	repo, err := project.RepoRootFrom(p.RootPath, home)
	if err != nil {
		return ""
	}
	rel, err := filepath.Rel(repo, p.RootPath)
	if err != nil || escapesRoot(rel) {
		return ""
	}
	return rel
}

// CRC: crc-CLI.md | R200
// escapesRoot reports whether a relative path climbs out of the directory it is relative to.
//
// The obvious `strings.HasPrefix(rel, "..")` also fires on a path whose first component
// merely *begins* with two dots — `..foo/design` is a legitimate descendant — and answering
// "unknown" there would tell a puller nothing can be located when everything can. Absurd as
// the input is, the wrong answer is silent and the right test is one line.
func escapesRoot(rel string) bool {
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// parseFlagsAnywhere parses fs from args and returns the positional arguments, **wherever the
// flags sit among them**.
//
// Go's flag package stops at the first non-flag argument. So
// `add-item --from x "a title" --skill mini-spec` silently folded `--skill mini-spec` into the
// title — measured 2026-08-18, in the exact order the command's own usage line advertises, and
// the corrupted title then rode into the done entry on completion.
//
// **The same trap had already been found and fixed in `finish` an hour earlier**, by lifting
// its item number out before parsing, and the fix was not carried across. That is this
// project's own recorded lesson arriving inside the session that recorded it: knowing the
// failure mode is not a substitute for a mechanism that holds it. Lifting positionals out one
// at a time and re-parsing the remainder makes the caller's ordering irrelevant for every verb
// at once, which is why this is shared rather than repeated.
func parseFlagsAnywhere(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		args = fs.Args()
		if len(args) == 0 {
			return positional, nil
		}
		positional = append(positional, args[0])
		args = args[1:]
	}
}

// CRC: crc-CLI.md | R199, R204
// parseAlarmArgs reads `query alarms`'s own flags.
//
// Subcommand flags rather than global ones, for the reason `query gaps` records: a global
// flag written after the subcommand is silently dropped (O28). `--json` is accepted here
// too, so `query alarms --brief --json` is not rejected as unknown by a flag set that owns
// the tail — the same accommodation, and the same admission that O28's general repair is
// still owed.
func (c *CLI) parseAlarmArgs(args []string) (unverified, brief bool, err error) {
	fs := flag.NewFlagSet("alarms", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	unv := fs.Bool("unverified", false, "only alarms whose state carries a decision")
	br := fs.Bool("brief", false, "print the spawn prompt for a delegated re-pull")
	asJSON := fs.Bool("json", false, "machine-readable output")
	rest, perr := parseFlagsAnywhere(fs, args)
	if perr != nil {
		return false, false, perr
	}
	if len(rest) > 0 {
		return false, false, fmt.Errorf("unexpected argument %q; usage: minispec query alarms [--unverified] [--brief]", rest[0])
	}
	c.JSON = c.JSON || *asJSON
	return *unv, *br, nil
}

// CRC: crc-CLI.md | R191, R196, R198
// emitNextID reports one next-ID answer, in whichever form the flags asked for.
//
// Shared by the repository-scoped early dispatch and the design-scoped subcommand case
// so the two cannot drift into presenting the same answer differently — the same reason
// unknownCommand is shared. The (result, error) pair is taken straight from the call so
// neither site can forget half of it.
func (c *CLI) emitNextID(res *query.NextIDResult, err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if c.JSON {
		c.output(res)
	} else {
		c.printNextID(res)
	}
	return 0
}

// CRC: crc-CLI.md | R196, R197
// printNextID writes the markdown form: the answer, then the evidence behind it.
//
// The sources are not decoration. A number alone cannot be checked -- a regex that
// matched nothing and a genuinely empty queue both answer 1 -- so the per-file counts
// are what let a reader tell a correct answer from a broken parser. R197
func (c *CLI) printNextID(res *query.NextIDResult) {
	if res.ByType != nil {
		fmt.Printf("next free %s IDs: %s\n", res.Class, strings.Join(gapIDList(res), ", "))
	} else {
		fmt.Printf("next free %s ID: %s\n", res.Class, res.Next)
	}
	// The evidence block is introduced rather than merely indented: in the gap form the
	// per-type answers and the source lines would otherwise share a margin and read as
	// one list. Width 16 fits `requirements.md`, the longest name any class reports.
	fmt.Println("read:")
	for _, s := range res.Sources {
		if !s.Present {
			fmt.Printf("  %-16s not found\n", s.Name)
			continue
		}
		fmt.Printf("  %-16s %d\n", s.Name, s.Count)
	}
}

// gapIDList renders the per-type gap answers in the canonical order.
func gapIDList(res *query.NextIDResult) []string {
	out := make([]string, 0, len(parser.GapTypes))
	for _, t := range parser.GapTypes {
		out = append(out, res.ByType[t])
	}
	return out
}

// CRC: crc-CLI.md | Seq: seq-carve-status.md#1 | R207, R213, R214, R215
// queryCarves is the cross-document census over carves/ and .carves/ at the repository root.
func (c *CLI) queryCarves(args []string) int {
	fs := flag.NewFlagSet("carves", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	open := fs.Bool("open", false, "list each carve's open parts, stateless lines and unread lines")
	asJSON := fs.Bool("json", false, "machine-readable output")
	rest, err := parseFlagsAnywhere(fs, args)
	if err != nil {
		return 1
	}
	if len(rest) > 0 {
		fmt.Fprintf(os.Stderr, "Error: unexpected argument %q; usage: minispec query carves [--open]\n", rest[0])
		return 1
	}
	c.JSON = c.JSON || *asJSON
	repoRoot, err := project.RepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	scan, err := parser.ScanCarves(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if c.JSON {
		c.output(carveReport(&scan, *open))
	} else {
		printCarves(os.Stdout, &scan, *open)
	}
	return 0
}

// CRC: crc-CLI.md | Seq: seq-carve-status.md#1.4 | R207, R211, R212, R216, R217, R218
// printCarves renders one line per carve, the listed rows beneath it, and the census.
func printCarves(w io.Writer, scan *parser.CarveScan, open bool) {
	width := 0
	for _, cv := range scan.Carves {
		width = max(width, len(cv.Path))
	}
	for _, cv := range scan.Carves {
		// R211 — a document with no status block prints that rather than counts, and prints
		// at all rather than being dropped.
		if !cv.HasStatus {
			fmt.Fprintf(w, "%-*s  no status block\n", width, cv.Path)
			continue
		}
		fmt.Fprintf(w, "%-*s  %3d open  %3d landed", width, cv.Path, cv.Open(), cv.Landed())
		if n := cv.NonConforming(); n > 0 {
			fmt.Fprintf(w, "  %d non-conforming", n)
		}
		if n := len(cv.Stateless); n > 0 {
			fmt.Fprintf(w, "  %d stateless", n) // R217 — the word names the line
		}
		if n := len(cv.Unread); n > 0 {
			fmt.Fprintf(w, "  %d unread", n) // R301
		}
		fmt.Fprintln(w)
		for _, p := range cv.Parts {
			if !listed(p, open) {
				continue
			}
			fmt.Fprintf(w, "  %-9s %-11s %s\n", partKeyLabel(p), queueLabel(p), p.Title())
			printDeviations(w, p.Deviations())
		}
		// R216 — a clean stateless line lists under --open; one carrying a deviation lists
		// always, since a defect behind a flag is hidden exactly when nobody looks again.
		for _, d := range cv.Stateless {
			if !listedStateless(d, open) {
				continue
			}
			fmt.Fprintf(w, "  %-11s L%-8d %s: %s\n", "(stateless)", d.Line, d.Reason, ellipsis(d.Text, 58))
			printDeviations(w, d.Deviations)
		}
		// R301 — what the reader could not read lists under --open, like a clean stateless
		// line: the count on the carve's line is always there, and it is the count that says
		// the rest of the file may be missing from every number above it.
		if open {
			for _, u := range cv.Unread {
				fmt.Fprintf(w, "  %-11s L%-8d %s\n", "(unread)", u.Line, u.Text)
			}
		}
	}
	fmt.Fprintln(w, carveCensus(scan))
}

// listed decides whether a part appears under its carve: any deviation lists it, and --open
// lists the open ones. One rule rather than two, so a non-conforming open part appears once. R212
func listed(p parser.Part, open bool) bool {
	return !p.Conforms() || (open && p.State() == parser.PartOpen)
}

// listedStateless is the same rule for a status-block line with no checkbox, so the text and
// JSON forms cannot drift apart on which lines they show. R216
func listedStateless(s parser.Stateless, open bool) bool {
	return len(s.Deviations) > 0 || open
}

func printDeviations(w io.Writer, devs []minispecsdom.Deviation) {
	for _, d := range devs {
		fmt.Fprintf(w, "      %s: %s\n", d.Rule, d.Target)
	}
}

func partKeyLabel(p parser.Part) string {
	if !p.Keyed() {
		return "(unkeyed)"
	}
	return p.Key()
}

func queueLabel(p parser.Part) string {
	if id := p.QueueID(); id != 0 {
		return fmt.Sprintf("#%d", id)
	}
	return "not queued"
}

func ellipsis(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// carveCensus states every count, zeros included — a zero is evidence the check ran. R218
func carveCensus(scan *parser.CarveScan) string {
	return fmt.Sprintf("%s: %d open, %d landed, %d stateless, %d non-conforming, %d unread; %s with no status block",
		countLabel(scan.WithStatus(), "carve"), scan.Open(), scan.Landed(), scan.Stateless(),
		scan.NonConforming(), scan.Unread(), countLabel(scan.NoStatus(), "document"))
}

func countLabel(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// carveJSON resolves the counts, which the markdown form computes, into fields. R215
type carveJSON struct {
	Path           string                `json:"path"`
	HasStatus      bool                  `json:"has_status"`
	Open           int                   `json:"open"`
	Landed         int                   `json:"landed"`
	Unkeyed        int                   `json:"unkeyed"`
	NonConforming  int                   `json:"non_conforming"`
	Stateless      int                   `json:"stateless"`
	StatelessLines []parser.Stateless    `json:"stateless_lines,omitempty"`
	Unread         int                   `json:"unread"`
	UnreadLines    []minispecsdom.Unread `json:"unread_lines,omitempty"`
	Parts          []parser.Part         `json:"parts,omitempty"`
}

type carveReportJSON struct {
	Dirs   []parser.CarveDir `json:"dirs"`
	Carves []carveJSON       `json:"carves"`
	Census struct {
		Carves        int `json:"carves"`
		Open          int `json:"open"`
		Landed        int `json:"landed"`
		Unkeyed       int `json:"unkeyed"`
		NonConforming int `json:"non_conforming"`
		Stateless     int `json:"stateless"`
		Unread        int `json:"unread"`
		NoStatus      int `json:"no_status"`
	} `json:"census"`
}

func carveReport(scan *parser.CarveScan, open bool) carveReportJSON {
	rep := carveReportJSON{Dirs: scan.Dirs}
	for _, cv := range scan.Carves {
		entry := carveJSON{
			Path:          cv.Path,
			HasStatus:     cv.HasStatus,
			Open:          cv.Open(),
			Landed:        cv.Landed(),
			Unkeyed:       cv.Unkeyed(),
			NonConforming: cv.NonConforming(),
			Stateless:     len(cv.Stateless),
			Unread:        len(cv.Unread),
		}
		if open {
			entry.UnreadLines = cv.Unread
		}
		for _, d := range cv.Stateless {
			if listedStateless(d, open) {
				entry.StatelessLines = append(entry.StatelessLines, d)
			}
		}
		for _, p := range cv.Parts {
			if listed(p, open) {
				entry.Parts = append(entry.Parts, p)
			}
		}
		rep.Carves = append(rep.Carves, entry)
	}
	rep.Census.Carves = scan.WithStatus()
	rep.Census.Open = scan.Open()
	rep.Census.Landed = scan.Landed()
	rep.Census.Unread = scan.Unread()
	rep.Census.Unkeyed = scan.Unkeyed()
	rep.Census.NonConforming = scan.NonConforming()
	rep.Census.Stateless = scan.Stateless()
	rep.Census.NoStatus = scan.NoStatus()
	return rep
}
