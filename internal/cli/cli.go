package cli

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/qu1queee/cartitas/internal/pick"
	"github.com/qu1queee/cartitas/internal/publish"
	"github.com/qu1queee/cartitas/internal/queuepr"
	"github.com/qu1queee/cartitas/internal/repo"
	"github.com/qu1queee/cartitas/internal/trilingual"
	"github.com/qu1queee/cartitas/internal/validate"
	"github.com/qu1queee/cartitas/internal/webexport"
)

func Run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "drill":
		return cmdDrill(rest)
	case "validate":
		return cmdValidate(rest)
	case "validate-trilingual":
		return cmdValidateTrilingual(rest)
	case "publish":
		return cmdPublish(rest)
	case "pick-generate-target":
		return cmdPick(rest)
	case "queue-pr", "draft-pr":
		return cmdQueuePR(rest)
	case "web-export":
		return cmdWebExport(rest)
	case "-h", "--help", "help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		return 2
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `cartitas — kid flashcard tooling

Usage:
  go run ./cmd/cartitas <command> [flags]

Commands:
  drill                  Drill cards in one language (hashcards)
  validate               Basic Q/A and cloze checks
  validate-trilingual    Require en/es/de siblings in draft/queue (--changed-from for PR diffs)
  publish                Move queue/{lang}/ into cards/{lang}/
  pick-generate-target   Next topic/subtopic/age band
  queue-pr               Commit queue/ and open automation PR
  web-export             Write docs/data/cards.json for GitHub Pages
`)
}

func mustRoot() (string, int) {
	root, err := repo.FindRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", 1
	}
	return root, 0
}

func cmdDrill(args []string) int {
	root, code := mustRoot()
	if code != 0 {
		return code
	}
	cfg, err := repo.LoadLanguages(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	supported := repo.LanguageCodes(cfg)
	fs := flag.NewFlagSet("drill", flag.ContinueOnError)
	lang := fs.String("lang", cfg.Default, "Language code")
	topic := fs.String("topic", "", "Topic folder, e.g. Animals, Geography, Space")
	limit := fs.Int("new-card-limit", 5, "Max new cards per session")
	list := fs.Bool("list", false, "List supported languages and exit")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *list {
		for _, code := range supported {
			meta := cfg.Languages[code]
			mark := ""
			if code == cfg.Default {
				mark = " (default)"
			}
			name := meta.Name
			if name == "" {
				name = code
			}
			fmt.Printf("%s: %s%s\n", code, name, mark)
		}
		return 0
	}
	ok := false
	for _, c := range supported {
		if c == *lang {
			ok = true
			break
		}
	}
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown language: %s\n", *lang)
		return 2
	}
	if _, err := exec.LookPath("hashcards"); err != nil {
		fmt.Fprintln(os.Stderr, "hashcards not found. Install: cargo install hashcards --locked")
		return 1
	}
	cardsDir := filepath.Join(root, "cards", *lang)
	if st, err := os.Stat(cardsDir); err != nil || !st.IsDir() {
		fmt.Fprintf(os.Stderr, "No cards for language '%s' at %s\n", *lang, cardsDir)
		return 1
	}
	drillPath := cardsDir
	if *topic != "" {
		drillPath = filepath.Join(cardsDir, *topic)
	}
	if st, err := os.Stat(drillPath); err != nil || !st.IsDir() {
		fmt.Fprintf(os.Stderr, "Path not found: %s\n", drillPath)
		return 1
	}
	name := cfg.Languages[*lang].Name
	if name == "" {
		name = *lang
	}
	rel, _ := filepath.Rel(root, drillPath)
	fmt.Printf("Drilling %s (%s): %s\n", name, *lang, filepath.ToSlash(rel))
	cmd := exec.Command("hashcards", "drill", drillPath, "--new-card-limit", fmt.Sprintf("%d", *limit), "--answer-controls", "binary")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		return 1
	}
	return 0
}

func cmdValidate(args []string) int {
	root, code := mustRoot()
	if code != 0 {
		return code
	}
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	errs := validate.All(root)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		return 1
	}
	fmt.Println("All card files passed basic validation.")
	return 0
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func cmdValidateTrilingual(args []string) int {
	root, code := mustRoot()
	if code != 0 {
		return code
	}
	fs := flag.NewFlagSet("validate-trilingual", flag.ContinueOnError)
	langCfg := fs.String("languages-config", filepath.Join(root, repo.LanguagesFile), "languages.yaml path")
	var stages stringList
	fs.Var(&stages, "stage", "Stage to check (default: draft and queue)")
	only := fs.String("only", "", "Filter to paths containing this string")
	changedFrom := fs.String("changed-from", "", "Git ref: only check sets touched since this commit (PR base)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	cfg, err := repo.LoadLanguagesFile(*langCfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	required := repo.LanguageCodes(cfg)
	check := []string(stages)
	if len(check) == 0 {
		check = append([]string{}, trilingual.Stages...)
	}
	var changedPaths []string
	if *changedFrom != "" {
		changedPaths, err = gitChangedPaths(root, *changedFrom)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	var errs []string
	checked := 0
	for _, stage := range check {
		var onlyKeys []string
		if *changedFrom != "" {
			onlyKeys = trilingual.KeysFromPaths(stage, changedPaths)
			checked += len(onlyKeys)
		}
		var stageErrs []string
		if *changedFrom != "" {
			stageErrs = trilingual.ValidateStageFilter(root, stage, required, onlyKeys)
		} else {
			stageErrs = trilingual.ValidateStage(root, stage, required)
		}
		errs = append(errs, stageErrs...)
	}
	if *only != "" {
		filtered := errs[:0]
		for _, e := range errs {
			if strings.Contains(e, *only) {
				filtered = append(filtered, e)
			}
		}
		errs = filtered
	}
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		fmt.Fprintln(os.Stderr, "\nFix: add matching files under {stage}/{en,es,de}/ for each incomplete set.")
		return 1
	}
	if *changedFrom != "" && checked == 0 {
		fmt.Printf("No changed .md files under %s since %s; trilingual check skipped.\n", strings.Join(check, ", "), *changedFrom)
		return 0
	}
	if *changedFrom != "" {
		fmt.Printf("Changed %s sets since %s are complete for: %s\n", strings.Join(check, ", "), *changedFrom, strings.Join(required, ", "))
		return 0
	}
	fmt.Printf("All %s sets complete for: %s\n", strings.Join(check, ", "), strings.Join(required, ", "))
	return 0
}

func gitChangedPaths(root, from string) ([]string, error) {
	var args []string
	if from == "HEAD" {
		args = []string{"diff", "--name-only", "--diff-filter=ACDMRT", "HEAD"}
	} else {
		args = []string{"diff", "--name-only", "--diff-filter=ACDMRT", from + "...HEAD"}
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if ee, ok := err.(*exec.ExitError); ok {
			msg = strings.TrimSpace(string(ee.Stderr))
		}
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("git diff %s: %s", from, msg)
	}
	var paths []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	if from == "HEAD" {
		u := exec.Command("git", "ls-files", "--others", "--exclude-standard")
		u.Dir = root
		extra, err := u.Output()
		if err != nil {
			return nil, fmt.Errorf("git ls-files: %w", err)
		}
		for _, line := range strings.Split(string(extra), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				paths = append(paths, line)
			}
		}
	}
	return paths, nil
}

func cmdPublish(args []string) int {
	root, code := mustRoot()
	if code != 0 {
		return code
	}
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	cfgPath := fs.String("config", filepath.Join(root, repo.PublishFile), "publish.yaml path")
	langCfgPath := fs.String("languages-config", filepath.Join(root, repo.LanguagesFile), "languages.yaml path")
	dryRun := fs.Bool("dry-run", false, "Print actions without writing")
	var langs, topics stringList
	fs.Var(&langs, "lang", "Language code (en, es, de). Repeatable.")
	fs.Var(&topics, "topic", "Topic key. Repeatable.")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	cfg, err := repo.LoadPublishFile(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	langCfg, err := repo.LoadLanguagesFile(*langCfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	supported := repo.LanguageCodes(langCfg)
	selectedLangs := []string(langs)
	if len(selectedLangs) == 0 {
		selectedLangs = supported
	}
	supportedSet := map[string]struct{}{}
	for _, l := range supported {
		supportedSet[l] = struct{}{}
	}
	for _, lang := range selectedLangs {
		if _, ok := supportedSet[lang]; !ok {
			fmt.Fprintf(os.Stderr, "Unknown language: %s\n", lang)
			return 1
		}
	}
	if len(cfg.Topics) == 0 {
		fmt.Fprintln(os.Stderr, "No topics defined in publish.yaml")
		return 1
	}
	selectedTopics := []string(topics)
	if len(selectedTopics) == 0 {
		for _, k := range repo.TopicOrder {
			if _, ok := cfg.Topics[k]; ok {
				selectedTopics = append(selectedTopics, k)
			}
		}
		for k := range cfg.Topics {
			found := false
			for _, t := range selectedTopics {
				if t == k {
					found = true
					break
				}
			}
			if !found {
				selectedTopics = append(selectedTopics, k)
			}
		}
	}
	defaultsRate := repo.DefaultsCardsPerPublish(cfg)
	anyPublished := false
	for _, lang := range selectedLangs {
		for _, topicKey := range selectedTopics {
			topicCfg, ok := cfg.Topics[topicKey]
			if !ok {
				fmt.Fprintf(os.Stderr, "Unknown topic: %s\n", topicKey)
				return 1
			}
			if !topicCfg.IsEnabled() {
				fmt.Printf("%s/%s: disabled, skipping\n", lang, topicKey)
				continue
			}
			for _, msg := range publish.Topic(root, topicKey, topicCfg, defaultsRate, lang, *dryRun) {
				fmt.Println(msg)
				if strings.Contains(msg, "published") || strings.Contains(msg, "would publish") {
					anyPublished = true
				}
			}
		}
	}
	if !*dryRun && anyPublished {
		fmt.Printf("\nPublish complete (%s). Run: hashcards check cards/\n", time.Now().Format("2006-01-02"))
	}
	return 0
}

func cmdPick(args []string) int {
	root, code := mustRoot()
	if code != 0 {
		return code
	}
	fs := flag.NewFlagSet("pick-generate-target", flag.ContinueOnError)
	ghOut := fs.String("github-output", "", "Append topic/subtopic/age_band/count to GitHub Actions output file")
	export := fs.Bool("export", false, "Print shell exports: topic=... subtopic=...")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	cfg, err := repo.LoadPublish(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	target, err := pick.TargetFrom(root, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if *ghOut != "" {
		f, err := os.OpenFile(*ghOut, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		_, _ = fmt.Fprintf(f, "topic=%s\nsubtopic=%s\nage_band=%s\ncount=%d\n",
			target.Topic, target.Subtopic, target.AgeBand, target.Count)
		_ = f.Close()
	}
	if *export {
		fmt.Printf("topic=%s\nsubtopic=%s\nage_band=%s\ncount=%d\ncoverage=%d\n",
			target.Topic, target.Subtopic, target.AgeBand, target.Count, target.Coverage)
	} else {
		fmt.Printf("topic=%s subtopic=%s age_band=%s count=%d (lang coverage %d/3)\n",
			target.Topic, target.Subtopic, target.AgeBand, target.Count, target.Coverage)
	}
	return 0
}

func cmdWebExport(args []string) int {
	root, code := mustRoot()
	if code != 0 {
		return code
	}
	fs := flag.NewFlagSet("web-export", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	path, n, err := webexport.WriteCards(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("Wrote %s (%d cards)\n", path, n)
	return 0
}

func cmdQueuePR(args []string) int {
	root, code := mustRoot()
	if code != 0 {
		return code
	}
	fs := flag.NewFlagSet("queue-pr", flag.ContinueOnError)
	message := fs.String("message", "", "Commit message")
	prNote := fs.String("pr-note", "", "Extra line for PR body")
	dryRun := fs.Bool("dry-run", false, "Show status only")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *message == "" {
		fmt.Fprintln(os.Stderr, "--message is required")
		return 2
	}
	has, err := queuepr.HasQueueChanges(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if !has {
		fmt.Println("No changes under queue/; nothing to commit.")
		queuepr.ListOpenPR(root)
		return 0
	}
	if *dryRun {
		cmd := exec.Command("git", "status", "--short", "queue/")
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
		fmt.Printf("Would commit on branch %s and update open PR.\n", queuepr.QueueBranch)
		return 0
	}
	if code := cmdValidate(nil); code != 0 {
		return code
	}
	if code := cmdValidateTrilingual([]string{"-stage", "queue", "-changed-from", "HEAD"}); code != 0 {
		fmt.Fprintln(os.Stderr, "Trilingual validation failed. Add en, es, and de for each new queue set.")
		return code
	}
	if err := queuepr.CheckoutQueueBranch(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if out, _, code, err := queuepr.Run(root, "git", "add", "queue/"); err != nil || code != 0 {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(out))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return 1
	}
	if out, _, code, err := queuepr.Run(root, "git", "commit", "-m", *message); err != nil || code != 0 {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(out))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return 1
	}
	if out, _, code, err := queuepr.Run(root, "git", "push", "-u", "origin", queuepr.QueueBranch); err != nil || code != 0 {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(out))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return 1
	}
	if _, err := queuepr.OpenPR(root, *message, *prNote); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
