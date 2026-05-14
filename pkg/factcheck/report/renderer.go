package report

import (
	"encoding/json"
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/pkg/factcheck/core"
)

// Renderer handles the output formatting of a VerdictReport.
type Renderer struct{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render outputs the VerdictReport in the specified format (json or text).
func (r *Renderer) Render(report *core.VerdictReport, format string) error {
	if format == "json" {
		return r.renderJSON(report)
	}
	return r.renderText(report)
}

func (r *Renderer) renderJSON(report *core.VerdictReport) error {
	out, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func (r *Renderer) renderText(report *core.VerdictReport) error {
	fmt.Printf("\n=======================================================\n")
	fmt.Printf(" CHAOS MESH FACT-CHECK REPORT\n")
	fmt.Printf("=======================================================\n\n")
	
	fmt.Printf("Target Experiment: %s (%s)\n", report.ChaosName, report.ChaosKind)
	
	if report.Verdict == core.Matched {
		fmt.Printf("Verdict: %s\n", report.Verdict)
	} else if report.Verdict == core.Mismatched {
		fmt.Printf("Verdict: %s\n", report.Verdict)
	} else {
		fmt.Printf("Verdict: %s\n", report.Verdict)
	}
	
	fmt.Printf("Reason:  %s\n\n", report.Reason)
	fmt.Printf("--- Evidence Collected ---\n")
	
	if len(report.Evidence) == 0 {
		fmt.Println("  (No evidence found)")
	}
	
	for i, ev := range report.Evidence {
		fmt.Printf("\n[%d] Source: %s\n", i+1, ev.Source)
		fmt.Printf("    Target: %s\n", ev.Target)
		fmt.Printf("    Detail: %s\n", ev.Description)
	}
	fmt.Printf("\n=======================================================\n\n")
	return nil
}
