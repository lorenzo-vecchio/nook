package cmd

import (
	"github.com/lorenzo-vecchio/nook/provider"
	"github.com/lorenzo-vecchio/nook/tui"
	"github.com/spf13/cobra"
)

func NewDetectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "Scan system to detect available provider tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			results := make(map[string]bool)
			for _, name := range provider.List() {
				p, ok := provider.Get(name)
				if !ok {
					continue
				}
				detected, err := p.Detect()
				if err != nil {
					results[name] = false
					continue
				}
				results[name] = detected
			}
			tui.PrintDetectionTable(cmd.OutOrStdout(), results)
			return nil
		},
	}
}
