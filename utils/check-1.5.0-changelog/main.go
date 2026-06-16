// check-1.5.0-changelog cross-checks the v1.5.0 changelog in NEWS against the
// actual git history between the 1.4.5 release and origin/master.
//
// It collects every commit in 1.4.5..origin/master, finds the merge commits on
// master's first-parent line (these correspond to the pull requests that make
// up the release), and pairs each one with its NEWS entry. It then reports:
//
//   - merges matched to a NEWS entry, in the same order they appear in NEWS;
//   - PR merges present in git but missing from the v1.5 NEWS (expected to be
//     changes already shipped in a 1.4.x release);
//   - non-PR "topology" merges that are not expected to have a NEWS entry;
//   - NEWS PR entries with no matching merge (stale/incorrect entries);
//   - leftover commits that were not part of any merge (direct-to-master).
//
// To decide whether an in-range PR merge already shipped in a 1.4.x release we
// use git, not the NEWS text. master and maint/1.4 are parallel branches, so a
// PR is merged into each as a different SHA; the maint-side merge is reachable
// from the 1.4.5 tag while the master-side one is not. A PR therefore already
// shipped iff a commit with the identical "Merge pull request #N from
// owner/branch" subject is reachable from 1.4.5. (We match the full subject,
// not just the number, because fork and canonical PR numbers collide, and we do
// not filter on --merges because the maint-side merges are fast-forwarded to a
// single parent.) This verdict is cross-checked against the NEWS membership to
// flag PRs that look wrongly included or wrongly excluded.
//
// NEWS references PRs in two schemes (see the HEAD~1 commit message): canonical
// irssi/irssi PRs as "#NNNN", and ailin-nemui fork PRs as "an#NN" paired with
// the merge commit's short SHA. Both are matched here.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	baseRef = "1.4.5"
	headRef = "origin/master"
)

// repoRoot is the absolute path to the git working tree, resolved once at
// startup so every git invocation and the NEWS read are anchored to it.
var repoRoot string

// Merge is a merge commit on master's first-parent line.
type Merge struct {
	SHA      string   // full hash
	Subject  string   // commit subject
	PRNum    int      // pull request number from the subject, 0 if not a PR merge
	Involved []string // full hashes brought in by this merge, within range, excluding itself

	Shipped    bool   // a commit with this subject is reachable from the 1.4.5 tag
	EntryIndex int    // index into the NEWS entries it matched, -1 if none
	Label      string // how the match is referenced in NEWS ("#1611", "an#48", ...)
}

// NewsEntry is one bullet in the v1.5.0 section of NEWS.
type NewsEntry struct {
	Index     int    // position within the section, 0-based
	Line      int    // 1-based line number where the bullet starts
	Text      string // full bullet text with continuation lines joined
	Canonical []int  // "#NNNN" references
	Fork      []int  // "an#NN" references
	SHAs      []string

	Matched bool // a merge was paired with this entry
}

// report holds everything the run computed, ready to be printed.
type report struct {
	allCommits      []string
	merges          []Merge
	entries         []NewsEntry
	mergesByEntry   map[int][]*Merge
	topology        []*Merge
	missingFromNews []*Merge
	leftover        []string
	leftoverSubject map[string]string

	// Cross-check of the git "already shipped" verdict against NEWS membership.
	wronglyIncluded []*Merge // in v1.5 NEWS, yet already shipped in 1.4.x
	wronglyExcluded []*Merge // new in 1.5 (not in 1.4.x), yet missing from NEWS
}

func main() {
	if err := run(context.Background()); err != nil {
		//nolint:errcheck // best-effort error report to stderr before exiting.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run does all the work and returns an error rather than exiting, so the only
// process exit lives in main.
func run(ctx context.Context) error {
	root, err := repoRootPath(ctx)
	if err != nil {
		return err
	}
	repoRoot = root

	allCommits, allSet, err := collectCommits(ctx)
	if err != nil {
		return err
	}
	merges, err := collectMerges(ctx, allSet)
	if err != nil {
		return err
	}
	entries, err := parseNews(filepath.Join(repoRoot, "NEWS"))
	if err != nil {
		return err
	}
	released, err := releasedPRSubjects(ctx)
	if err != nil {
		return err
	}

	canon := map[int]int{} // canonical PR number -> entry index
	for i := range entries {
		for _, n := range entries[i].Canonical {
			canon[n] = i
		}
	}

	// Pair each first-parent merge with a NEWS entry and decide whether it
	// already shipped in 1.4.x.
	r := &report{
		allCommits:      allCommits,
		merges:          merges,
		entries:         entries,
		mergesByEntry:   map[int][]*Merge{},
		leftoverSubject: map[string]string{},
	}
	for i := range r.merges {
		m := &r.merges[i]
		m.Shipped = released[m.Subject]
		idx, label := matchMerge(m, r.entries, canon)
		m.EntryIndex = idx
		m.Label = label
		switch {
		case idx >= 0:
			r.entries[idx].Matched = true
			r.mergesByEntry[idx] = append(r.mergesByEntry[idx], m)
		case m.PRNum == 0:
			r.topology = append(r.topology, m)
		default:
			r.missingFromNews = append(r.missingFromNews, m)
		}
		// Cross-check the git verdict against NEWS membership (PR merges only).
		switch {
		case m.PRNum == 0:
		case m.Shipped && idx >= 0:
			r.wronglyIncluded = append(r.wronglyIncluded, m)
		case !m.Shipped && idx < 0:
			r.wronglyExcluded = append(r.wronglyExcluded, m)
		}
	}

	// Account for which commits belong to a merge so we can find the leftovers.
	accounted := map[string]bool{}
	for i := range r.merges {
		accounted[r.merges[i].SHA] = true
		for _, c := range r.merges[i].Involved {
			accounted[c] = true
		}
	}
	for _, c := range r.allCommits {
		if accounted[c] {
			continue
		}
		r.leftover = append(r.leftover, c)
		subject, err := git(ctx, "log", "-1", "--format=%s", c)
		if err != nil {
			return err
		}
		r.leftoverSubject[c] = strings.TrimSpace(subject)
	}

	r.print()
	return nil
}

// repoRootPath returns the git working tree root.
func repoRootPath(ctx context.Context) (string, error) {
	//nolint:gosec // G204: the argument list is fixed and not user-controlled.
	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", &failure{msg: "determine git repo root", cause: err}
	}
	return strings.TrimSpace(string(out)), nil
}

// collectCommits returns every commit in the release range, newest first, plus
// a set for membership tests.
func collectCommits(ctx context.Context) ([]string, map[string]bool, error) {
	commits, err := gitLines(ctx, "rev-list", baseRef+".."+headRef)
	if err != nil {
		return nil, nil, err
	}
	set := make(map[string]bool, len(commits))
	for _, c := range commits {
		set[c] = true
	}
	return commits, set, nil
}

// collectMerges returns the merge commits on the first-parent line of the
// release range, each populated with the commits it brought in (restricted to
// the range).
func collectMerges(ctx context.Context, allSet map[string]bool) ([]Merge, error) {
	prRe := regexp.MustCompile(`^Merge pull request #(\d+) `)
	// %H, %s separated by a unit separator so subjects can contain anything.
	lines, err := gitLines(
		ctx,
		"log",
		"--first-parent",
		"--merges",
		"--format=%H%x1f%s",
		baseRef+".."+headRef,
	)
	if err != nil {
		return nil, err
	}

	merges := make([]Merge, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, "\x1f", 2)
		if len(parts) != 2 {
			continue
		}
		m := Merge{SHA: parts[0], Subject: parts[1], EntryIndex: -1}
		if g := prRe.FindStringSubmatch(m.Subject); g != nil {
			if n, convErr := strconv.Atoi(g[1]); convErr == nil {
				m.PRNum = n
			}
		}
		// Commits reachable from the merge but not its first parent, i.e. the
		// branch that was merged in. Drop the merge itself and anything outside
		// the release range (e.g. the already-released side of a tag merge).
		branch, err := gitLines(ctx, "rev-list", m.SHA+"^1.."+m.SHA)
		if err != nil {
			return nil, err
		}
		for _, c := range branch {
			if c != m.SHA && allSet[c] {
				m.Involved = append(m.Involved, c)
			}
		}
		merges = append(merges, m)
	}
	return merges, nil
}

// releasedPRSubjects returns the set of "Merge pull request ..." commit
// subjects reachable from the 1.4.5 tag, i.e. the pull requests that already
// shipped in a 1.4.x release. We scan all commits (not just --merges) because
// on maint/1.4 these merges were fast-forwarded to a single parent.
func releasedPRSubjects(ctx context.Context) (map[string]bool, error) {
	lines, err := gitLines(ctx, "log", "--format=%s", baseRef)
	if err != nil {
		return nil, err
	}
	subjects := map[string]bool{}
	for _, s := range lines {
		if strings.HasPrefix(s, "Merge pull request #") {
			subjects[s] = true
		}
	}
	return subjects, nil
}

// parseNews reads the first (newest) version section of NEWS and returns its
// bullet entries in file order.
func parseNews(path string) ([]NewsEntry, error) {
	//nolint:gosec // G304: path is the NEWS file under the trusted repo root.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &failure{msg: "read NEWS", cause: err}
	}

	headerRe := regexp.MustCompile(`^v?\d+\.\d+`)
	var entries []NewsEntry
	var cur *NewsEntry
	seenHeader := false

	flush := func() {
		if cur != nil {
			parseRefs(cur)
			entries = append(entries, *cur)
			cur = nil
		}
	}

	lineNo := 0
	for line := range strings.SplitSeq(string(data), "\n") {
		lineNo++
		if headerRe.MatchString(line) {
			if !seenHeader {
				seenHeader = true // this is the v1.5.0 header; start collecting
				continue
			}
			break // reached the previous release's header; stop
		}
		if !seenHeader {
			continue
		}
		switch {
		case strings.HasPrefix(line, "* "):
			flush()
			cur = &NewsEntry{Index: len(entries), Line: lineNo, Text: strings.TrimSpace(line[2:])}
		case strings.TrimSpace(line) == "":
			flush()
		case cur != nil:
			cur.Text += " " + strings.TrimSpace(line)
		}
	}
	flush()
	return entries, nil
}

// parseRefs fills an entry's PR and SHA references from the parenthetical
// groups in its text. References live inside "(...)"; each comma-separated
// token is classified as a canonical PR (#NNNN), a fork PR (an#NN), or a SHA.
func parseRefs(e *NewsEntry) {
	parens := regexp.MustCompile(`\(([^)]*)\)`).FindAllStringSubmatch(e.Text, -1)
	forkRe := regexp.MustCompile(`^an#(\d+)$`)
	canonRe := regexp.MustCompile(`^#(\d+)$`)
	shaRe := regexp.MustCompile(`^[0-9a-f]{7,40}$`)
	for _, p := range parens {
		for tok := range strings.SplitSeq(p[1], ",") {
			tok = strings.TrimSpace(tok)
			switch {
			case forkRe.MatchString(tok):
				if n, err := strconv.Atoi(tok[3:]); err == nil {
					e.Fork = append(e.Fork, n)
				}
			case canonRe.MatchString(tok):
				if n, err := strconv.Atoi(tok[1:]); err == nil {
					e.Canonical = append(e.Canonical, n)
				}
			case shaRe.MatchString(tok):
				e.SHAs = append(e.SHAs, tok)
			}
		}
	}
}

// matchMerge pairs a merge with a NEWS entry, returning the entry index and the
// reference label used in NEWS. Fork PRs and direct merges are matched on the
// merge SHA (NEWS records it); canonical PRs are matched on the PR number.
func matchMerge(m *Merge, entries []NewsEntry, canon map[int]int) (int, string) {
	if idx := findEntryBySHA(m.SHA, entries); idx >= 0 {
		e := &entries[idx]
		switch {
		case len(e.Fork) > 0:
			return idx, "an#" + strconv.Itoa(e.Fork[0])
		case len(e.Canonical) > 0:
			return idx, "#" + strconv.Itoa(e.Canonical[0])
		default:
			return idx, shortSHA(m.SHA)
		}
	}
	if m.PRNum > 0 {
		if idx, ok := canon[m.PRNum]; ok {
			return idx, "#" + strconv.Itoa(m.PRNum)
		}
	}
	return -1, ""
}

// findEntryBySHA returns the index of the NEWS entry whose recorded SHA is a
// prefix of the given full hash, or -1.
func findEntryBySHA(full string, entries []NewsEntry) int {
	for i := range entries {
		for _, s := range entries[i].SHAs {
			if strings.HasPrefix(full, s) {
				return i
			}
		}
	}
	return -1
}

// print writes the cross-check report to stdout.
func (r *report) print() {
	prMerges := 0
	for i := range r.merges {
		if r.merges[i].PRNum > 0 {
			prMerges++
		}
	}

	printf("Range: %s..%s\n", baseRef, headRef)
	printf("Commits in range: %d\n", len(r.allCommits))
	printf(
		"First-parent merges: %d (%d pull-request merges, %d topology)\n\n",
		len(r.merges),
		prMerges,
		len(r.topology),
	)

	r.printDiscrepancies()

	printf("=== MERGES MATCHED TO NEWS (in NEWS order) ===\n")
	for i := range r.entries {
		e := &r.entries[i]
		if len(e.Fork) == 0 && len(e.Canonical) == 0 {
			continue // direct-commit entry, handled under leftovers
		}
		for _, m := range r.mergesByEntry[i] {
			printf(
				"\n%-7s %s  (%d commit(s))  %s%s\n",
				m.Label,
				shortSHA(m.SHA),
				len(m.Involved),
				m.Subject,
				shippedTag(m, false),
			)
			printf("        NEWS L%d: %s\n", e.Line, e.Text)
		}
	}

	printf(
		"\n=== PR MERGES NOT IN THE v1.5 NEWS (verified already shipped in a 1.4.x release) ===\n",
	)
	if len(r.missingFromNews) == 0 {
		printf("(none)\n")
	}
	for _, m := range r.missingFromNews {
		printf(
			"#%-6d %s  (%d commit(s))  %s%s\n",
			m.PRNum,
			shortSHA(m.SHA),
			len(m.Involved),
			m.Subject,
			shippedTag(m, true),
		)
	}

	printf("\n=== NON-PR (topology) MERGES (no NEWS entry expected) ===\n")
	if len(r.topology) == 0 {
		printf("(none)\n")
	}
	for _, m := range r.topology {
		printf("%s  (%d commit(s))  %s\n", shortSHA(m.SHA), len(m.Involved), m.Subject)
	}

	printf("\n=== NEWS PR ENTRIES WITH NO MATCHING MERGE (possible stale entries) ===\n")
	none := true
	for i := range r.entries {
		e := &r.entries[i]
		if (len(e.Fork) > 0 || len(e.Canonical) > 0) && !e.Matched {
			none = false
			printf("NEWS L%d: %s\n", e.Line, e.Text)
		}
	}
	if none {
		printf("(none)\n")
	}

	printf("\n=== LEFTOVER COMMITS (not part of any merge) ===\n")
	if len(r.leftover) == 0 {
		printf("(none)\n")
	}
	for _, c := range r.leftover {
		printf("%s  %s\n", shortSHA(c), r.leftoverSubject[c])
		if idx := findEntryBySHA(c, r.entries); idx >= 0 {
			printf("        NEWS L%d: %s\n", r.entries[idx].Line, r.entries[idx].Text)
		}
	}

	r.printSummary()
}

// printDiscrepancies lists PR merges whose git "already shipped" verdict
// disagrees with their v1.5 NEWS membership -- the entries worth a human look.
func (r *report) printDiscrepancies() {
	printf("=== POTENTIAL CHANGELOG DISCREPANCIES (git verdict vs NEWS) ===\n")
	if len(r.wronglyExcluded) == 0 && len(r.wronglyIncluded) == 0 {
		printf("(none -- every PR merge's shipped/new verdict matches its NEWS membership)\n\n")
		return
	}
	if len(r.wronglyExcluded) > 0 {
		printf(
			"\nNew in 1.5 (no matching PR in 1.4.x) but MISSING from v1.5 NEWS -- consider adding:\n",
		)
		for _, m := range r.wronglyExcluded {
			printf("  #%d %s  %s\n", m.PRNum, shortSHA(m.SHA), m.Subject)
		}
	}
	if len(r.wronglyIncluded) > 0 {
		printf("\nIn v1.5 NEWS but the same PR already shipped in 1.4.x -- review:\n")
		for _, m := range r.wronglyIncluded {
			printf("  %s %s  %s\n", m.Label, shortSHA(m.SHA), m.Subject)
			if m.EntryIndex >= 0 {
				printf(
					"        NEWS L%d: %s\n",
					r.entries[m.EntryIndex].Line,
					r.entries[m.EntryIndex].Text,
				)
			}
		}
	}
	printf("\n")
}

// shippedTag annotates a merge whose shipped verdict is unexpected for the
// section it appears in: expectShipped is true for the "not in NEWS" listing
// (everything there should already have shipped) and false for the "matched to
// NEWS" listing (nothing there should have shipped).
func shippedTag(m *Merge, expectShipped bool) string {
	switch {
	case m.Shipped == expectShipped:
		return ""
	case m.Shipped:
		return "   <- also shipped in 1.4.x?"
	default:
		return "   <- NOT found in 1.4.x; new?"
	}
}

// printSummary writes the reconciling totals so the partition can be checked at
// a glance.
func (r *report) printSummary() {
	matched := 0
	for i := range r.merges {
		if r.merges[i].EntryIndex >= 0 {
			matched++
		}
	}
	leftoverMatched := 0
	for _, c := range r.leftover {
		if findEntryBySHA(c, r.entries) >= 0 {
			leftoverMatched++
		}
	}
	shipped := 0
	for i := range r.merges {
		if r.merges[i].PRNum > 0 && r.merges[i].Shipped {
			shipped++
		}
	}
	prMerges := len(r.merges) - len(r.topology)
	involved := len(r.allCommits) - len(r.merges) - len(r.leftover)

	printf("\n=== SUMMARY ===\n")
	printf("%d commits in %s..%s\n", len(r.allCommits), baseRef, headRef)
	printf("  %d first-parent merges\n", len(r.merges))
	printf("      %d matched to a v1.5 NEWS entry\n", matched)
	printf(
		"      %d PR merges not in v1.5 NEWS (expected already shipped in 1.4.x)\n",
		len(r.missingFromNews),
	)
	printf("      %d topology merges (no entry expected)\n", len(r.topology))
	printf("  %d commits brought in by those merges\n", involved)
	printf("  %d leftover (direct-to-master) commits\n", len(r.leftover))
	printf("      %d matched to a v1.5 NEWS entry\n", leftoverMatched)
	printf("      %d with no entry\n", len(r.leftover)-leftoverMatched)
	printf("v1.5 NEWS entries: %d total\n", len(r.entries))
	printf("\nShipped-in-1.4.x verdict (by merge-subject reachability from %s):\n", baseRef)
	printf(
		"  %d of %d PR merges already shipped in 1.4.x; %d new in 1.5\n",
		shipped,
		prMerges,
		prMerges-shipped,
	)
	printf(
		"  discrepancies vs NEWS: %d wrongly-excluded, %d wrongly-included\n",
		len(r.wronglyExcluded),
		len(r.wronglyIncluded),
	)
}

// shortSHA abbreviates a full hash to the 8 characters NEWS uses.
func shortSHA(full string) string {
	if len(full) > 8 {
		return full[:8]
	}
	return full
}

// printf writes a formatted line to stdout. A failed write to stdout is not
// recoverable and not worth handling for a report-only tool.
func printf(format string, args ...any) {
	//nolint:errcheck // writing the report to stdout; nothing to do on failure.
	fmt.Fprintf(os.Stdout, format, args...)
}

// gitLines runs git and returns its stdout split into non-empty lines.
func gitLines(ctx context.Context, args ...string) ([]string, error) {
	out, err := git(ctx, args...)
	if err != nil {
		return nil, err
	}
	var lines []string
	for l := range strings.SplitSeq(out, "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines, nil
}

// git runs a git command anchored at repoRoot and returns its stdout.
func git(ctx context.Context, args ...string) (string, error) {
	full := append([]string{"-C", repoRoot}, args...)
	//nolint:gosec // G204: git subcommand arguments are program-controlled.
	out, err := exec.CommandContext(ctx, "git", full...).Output()
	if err != nil {
		return "", &failure{msg: "run git " + strings.Join(args, " "), cause: err}
	}
	return string(out), nil
}

// failure is an error carrying a contextual message. It exists because the
// repository linter forbids fmt.Errorf and errors.New (which expect a
// MaxMind-internal helper that is not available in this standalone utility).
type failure struct {
	msg   string
	cause error
}

func (f *failure) Error() string {
	if f.cause == nil {
		return f.msg
	}
	return f.msg + ": " + f.cause.Error()
}

func (f *failure) Unwrap() error { return f.cause }
