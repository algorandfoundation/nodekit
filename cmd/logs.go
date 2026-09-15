package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	cmdutils "github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/logs"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// defaultLogLines is how many matching entries are shown when --lines is not
// given. Zero means every match: warnings are well under one percent of a real
// log, so the whole history of them is a readable amount of output, and a
// fixed count would silently hide the older ones.
const defaultLogLines = 0

// defaultFollowLogLines is the smaller default used with --follow, so that a
// long backlog does not scroll past before the stream begins. This mirrors the
// asymmetry between tail and tail -f.
const defaultFollowLogLines = 10

// Flag values for the logs command.
//
// These are deliberately separate package variables. The shared `force` value
// in root.go is rebound to -f by start, stop, install and uninstall; the logs
// command never registers --force, which is precisely what leaves -f free for
// --follow here. Likewise -n is free because root's --no-incentives is local to
// the root command. Do not add --force to this command. --file takes -F rather
// than -f for the same reason.
var (
	logsDataDir string
	logsFile    string
	logsFollow  bool
	logsLines   int
	logsLevel   string
	logsAll     bool
	logsSince   string
	logsFilter  string
	logsJSON    bool
)

var logsShort = "Display the node's logs"

var logsLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(logsShort),
	"",
	style.BoldUnderline("Overview:"),
	"Reads the log file algod writes, wherever config.json puts it.",
	"Only warnings and errors are shown by default; use --all to see every level.",
	"Lines that are not valid log entries, such as crash output, are always shown.",
	"",
	style.BoldUnderline("Notes:"),
	"Every matching entry is shown; --lines N shows only the newest N of them.",
	"--lines counts entries that match the filters, not raw lines of the file.",
	"--follow shows the newest 10 before it starts streaming, the way tail -f",
	"does; --lines N sets that backlog, and --lines 0 shows the whole history.",
	"The rotated archives are read as well, so the history reaches back past the",
	"last rotation. Pass --file to read one file on its own instead.",
	"--filter matches plain text in the message and the fields shown beside it,",
	"such as Round=49291042, with no pattern syntax.",
	"A node only writes entries at or above its configured level, so asking for a",
	"lower level than the node records will find nothing.",
	"",
)

var logsCmd = cmdutils.WithAlgodFlags(&cobra.Command{
	Use:          "logs",
	Short:        logsShort,
	Long:         logsLong,
	SilenceUsage: true,
	Args:         cobra.NoArgs,
	// No PersistentPreRun on purpose: reading the log of a node that is not
	// running is the main reason to reach for this command.
	RunE: func(cmd *cobra.Command, args []string) error {
		filter, err := buildLogFilter()
		if err != nil {
			return err
		}

		source, err := resolveLogSource()
		if err != nil {
			return err
		}

		// Caught before anything is read, so the user is told why rather than
		// watching a compressed container stream past as unreadable entries.
		if logsFollow && logs.IsCompressed(source.Path) {
			return errors.New(explanations.LogsFollowCompressedErrorMsg)
		}

		out := bufio.NewWriter(cmd.OutOrStdout())
		defer func() { _ = out.Flush() }()
		errOut := cmd.ErrOrStderr()

		lines := logsLines
		if !cmd.Flags().Changed("lines") && logsFollow {
			lines = defaultFollowLogLines
		}

		// What is about to be read, before any of it is read. Archives that
		// cannot hold anything as new as --since are dropped first, so that the
		// status line names the files the scan will really open. The unpruned
		// list is kept: whether a history exists is a different question from
		// whether this search has any use for it.
		history := source.LogFiles()
		sources := logs.PruneSources(history, filter.Since)
		writeLogStatus(errOut, sources, filter, lines)

		// A node whose floor sits above the level being asked for never records
		// what the user wants to see, so the view is incomplete in a way that is
		// invisible from the output. Gated on the requested level and not just on
		// warnings: --level error against an error floor is a complete answer,
		// and saying otherwise sends the user looking for entries that a
		// correctly configured node was never going to write.
		//
		// --all names no level at all. It asks for whatever the log holds, and
		// the log holds everything the node wrote, so nothing the user asked for
		// is missing from it and there is no incompleteness to report. Measuring
		// it against the floor anyway fires this on every default node -- --all
		// asks for trace, which no algod writes -- on the one run where the
		// request was for all of it. A node that really is holding levels back
		// is still explained by reportEmptyLogResult, at the point where that
		// leaves nothing to show and the user needs to know why.
		if !logsAll && source.Hides(filter.MinLevel) {
			if value, ok := source.FloorValue(); ok {
				floor, _ := source.Floor()
				fmt.Fprintln(errOut, style.Yellow.Render(fmt.Sprintf(
					"warning: this node logs at %s and above (config.json BaseLoggerDebugLevel: %d)\n"+
						"         lower levels are never written to the log, so this view is incomplete",
					floor, value)))
			}
		}

		// Entries are written as they are found rather than collected first:
		// asking for every match on a gigabyte of log would otherwise hold the
		// whole result in memory before printing a line of it.
		result, err := logs.Scan(sources, lines, filter, func(entry logs.Entry) error {
			writeLogEntry(out, entry)
			return nil
		})
		if err != nil {
			return explainLogError(err, source)
		}
		if err := out.Flush(); err != nil {
			return err
		}

		// An archive that could not be read costs its entries and nothing else,
		// but the user has to know the history has a gap in it.
		for _, skipped := range result.Skipped {
			fmt.Fprintln(errOut, style.Yellow.Render(fmt.Sprintf(
				"warning: could not read %s (%v)\n"+
					"         its entries are not included; a log being rotated is unreadable until that finishes",
				filepath.Base(skipped.Path), skipped.Err)))
		}

		if result.ScanLimited {
			fmt.Fprintln(errOut, style.Yellow.Render(fmt.Sprintf(
				"searched back %d MiB without finding %d entries, and stopped there;\n"+
					"older entries were not read. Use --lines 0 to search all of it",
				logs.ScanLimit()>>20, lines)))
		}

		if !logsFollow {
			reportEmptyLogResult(errOut, source, filter, result, len(history))
			return nil
		}

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		err = logs.Follow(ctx, source.Path, result.Offset, filter, func(entry logs.Entry) error {
			writeLogEntry(out, entry)
			return out.Flush()
		}, logs.FollowOptions{
			OnRotate: func() {
				fmt.Fprintln(errOut, style.Yellow.Render("--- the log was rotated, continuing with the new file ---"))
			},
		})
		if err != nil {
			return explainLogError(err, source)
		}
		return nil
	},
}, &logsDataDir)

// buildLogFilter turns the command's flags into a logs.Filter.
func buildLogFilter() (logs.Filter, error) {
	var filter logs.Filter

	level := logs.LevelWarn
	if logsAll {
		level = logs.LevelTrace
	} else if logsLevel != "" {
		parsed, err := logs.ParseLevel(logsLevel)
		if err != nil {
			return filter, err
		}
		level = parsed
	}
	filter.MinLevel = level

	if logsLines < 0 {
		return filter, errors.New("--lines cannot be negative")
	}

	if logsSince != "" {
		since, err := logs.ParseSince(logsSince, time.Now())
		if err != nil {
			return filter, err
		}
		filter.Since = since
	}

	filter.Text = logsFilter

	return filter, nil
}

// resolveLogSource locates the data directory and the log file within it,
// reporting the problems a user is most likely to hit in the order that gives
// the most accurate message.
func resolveLogSource() (logs.Source, error) {
	var source logs.Source

	// An explicit file replaces discovery entirely: no data directory, no
	// config.json, and no archives. That last part is what makes --file the way
	// to read one file on its own, the live log included, since the archives are
	// otherwise always part of the history.
	if logsFile != "" {
		if _, err := os.Stat(logsFile); err != nil {
			if os.IsPermission(err) {
				return source, fmt.Errorf(explanations.LogsFilePermissionErrorMsg, logsFile)
			}
			return source, err // the error already names the file
		}
		return logs.Source{Path: logsFile}, nil
	}

	dataDir, err := algod.GetDataDir(logsDataDir)
	if err != nil {
		return source, err
	}

	// Stat the directory before validating it. Once IsDataDir declines to panic
	// on a permission error it also returns false for one, and reporting that as
	// "not a data directory" would send the user down the wrong path entirely.
	if _, err := os.Stat(dataDir); err != nil {
		if os.IsPermission(err) {
			return source, errors.New(explanations.LogsPermissionErrorMsg)
		}
		if os.IsNotExist(err) {
			return source, fmt.Errorf("%s\n\n%s", algod.InvalidDataDirMsg, explanations.NodeNotFound)
		}
		return source, err
	}

	if !utils.IsDataDir(dataDir) {
		// IsDataDir looks for genesis.json and cannot say why it failed to find
		// it. A directory the user may stat but not search fails that lookup
		// with a permission error, and reporting it as "not a data directory"
		// tells someone with a perfectly good node to go and install one.
		if _, err := os.Stat(filepath.Join(dataDir, "genesis.json")); os.IsPermission(err) {
			return source, errors.New(explanations.LogsPermissionErrorMsg)
		}
		return source, fmt.Errorf("%s\n\n%s", algod.InvalidDataDirMsg, explanations.NodeNotFound)
	}

	source = logs.ResolveSource(dataDir)

	// A node told to log to stdout never creates a log file, so the missing
	// file below would otherwise be reported as "please run start".
	if source.LogsToStdout() {
		return source, errors.New(explanations.LogsToStdoutErrorMsg)
	}

	return source, nil
}

// explainLogError replaces filesystem errors with guidance the user can act on.
func explainLogError(err error, source logs.Source) error {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf(explanations.LogsNotFoundErrorMsg, source.Path)
	case errors.Is(err, fs.ErrPermission):
		// A file the user named is their own to explain; the guidance about the
		// 'algorand' user only applies to the node's own files.
		if logsFile != "" {
			return fmt.Errorf(explanations.LogsFilePermissionErrorMsg, source.Path)
		}
		return errors.New(explanations.LogsPermissionErrorMsg)
	default:
		return err
	}
}

// writeLogStatus says, before any entry is printed, which files are being read
// and what is being hidden.
//
// The default view is a filtered one, and a filtered view that does not say so
// is how a user concludes their node logged nothing at all. It goes to stderr,
// so a pipe still sees only log entries.
func writeLogStatus(errOut io.Writer, sources []string, filter logs.Filter, lines int) {
	names := make([]string, 0, len(sources))
	for _, path := range sources {
		names = append(names, filepath.Base(path))
	}
	parts := []string{strings.Join(names, ", ")}

	parts = append(parts, "showing: "+shownLevels(filter.MinLevel))
	if filter.Text != "" {
		parts = append(parts, fmt.Sprintf("containing %q", filter.Text))
	}
	if logsSince != "" {
		// What the user typed, not filter.Since: "15m" is the form they will
		// recognise as their own.
		parts = append(parts, "since "+logsSince)
	}
	if lines > 0 {
		parts = append(parts, fmt.Sprintf("newest %d", lines))
	}
	if logsFollow {
		parts = append(parts, "following")
	}

	fmt.Fprintln(errOut, style.Cyan.Render("reading "+strings.Join(parts, " · ")))
}

// shownLevels names the levels a floor lets through, so that the status line
// says what is being shown rather than leaving the user to work out what "warn
// and above" covers.
func shownLevels(floor logs.Level) string {
	if floor <= logs.LevelTrace {
		return "every level"
	}

	var names []string
	for level := floor; level <= logs.LevelPanic; level++ {
		names = append(names, level.String())
	}
	return strings.Join(names, ", ") + " (--all to show everything)"
}

// writeLogEntry emits one entry, either as the original JSON or as a formatted line.
func writeLogEntry(out io.Writer, entry logs.Entry) {
	if !logsJSON {
		fmt.Fprintln(out, logs.Render(entry))
		return
	}
	fmt.Fprintln(out, string(logs.JSONLine(entry)))
}

// reportEmptyLogResult explains an empty result, distinguishing a log that has
// nothing in it from one whose entries were all filtered out, and from a node
// that never records the level that was asked for.
//
// history is the number of log files the node has, counted before --since
// pruned any of them away. Counting what the scan opened instead would let a
// --since newer than every archive turn a node with a year of rotated logs
// into one that has never written a line.
func reportEmptyLogResult(errOut io.Writer, source logs.Source, filter logs.Filter, result logs.ScanResult, history int) {
	if result.Count > 0 {
		return
	}

	// Offset is the size of the live log, so this is only an empty log when
	// there was no archive behind it to hold anything either.
	if result.Offset == 0 && history == 1 {
		fmt.Fprintln(errOut, style.Yellow.Render(explanations.LogsEmptyMsg))
		return
	}

	message := explanations.LogsNoMatchMsg
	if source.Hides(filter.MinLevel) {
		if value, ok := source.FloorValue(); ok {
			floor, _ := source.Floor()
			message += fmt.Sprintf(
				"\nthe node is configured to log at %s and above (config.json BaseLoggerDebugLevel: %d);"+
					"\nentries below that level are never written to the log", floor, value)
		}
	}
	fmt.Fprintln(errOut, style.Yellow.Render(message))
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, style.LightBlue("Stream new entries as they are written"))
	logsCmd.Flags().IntVarP(&logsLines, "lines", "n", defaultLogLines, style.LightBlue("Number of newest matching entries to show, 0 for all (10 with --follow)"))
	logsCmd.Flags().StringVar(&logsLevel, "level", "", style.LightBlue("Minimum level to show: "+strings.Join(logs.LevelNames, ", ")))
	logsCmd.Flags().BoolVarP(&logsAll, "all", "a", false, style.LightBlue("Show entries at every level"))
	logsCmd.Flags().StringVar(&logsSince, "since", "", style.LightBlue("Only entries newer than a duration (15m, 2h) or a timestamp"))
	logsCmd.Flags().StringVar(&logsFilter, "filter", "", style.LightBlue("Only entries whose message or shown fields contain this text"))
	logsCmd.Flags().BoolVar(&logsJSON, "json", false, style.LightBlue("Emit the raw JSON log entries"))
	logsCmd.Flags().StringVarP(&logsFile, "file", "F", "", style.LightBlue("Read this log file instead of the node's own"))

	logsCmd.MarkFlagsMutuallyExclusive("all", "level")
	logsCmd.MarkFlagsMutuallyExclusive("file", "datadir")
}
