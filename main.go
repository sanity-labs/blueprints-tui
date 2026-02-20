package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/sanity-io/blueprints-tui/internal/api"
	"github.com/sanity-io/blueprints-tui/internal/config"
	"github.com/sanity-io/blueprints-tui/internal/tui"
)

func main() {
	token := flag.String("token", "", "Sanity API auth token")
	org := flag.String("org", "", "Sanity organization ID")
	project := flag.String("project", "", "Sanity project ID")
	apiURL := flag.String("api-url", "", "Blueprints API base URL")
	debug := flag.Bool("debug", false, "print debug info to stderr")
	staging := flag.Bool("staging", false, "use staging environment (sanity.work)")
	flag.Parse()

	cfg, err := config.Load(*token, *org, *project, *apiURL, *staging)
	cfg.Debug = *debug
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	client := api.NewClient(cfg.APIURL, cfg.Token, cfg.ScopeType, cfg.ScopeID, cfg.Debug)
	model := tui.NewModel(client, cfg.ScopeID != "")

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
