package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/fossology/report-aggregator/internal/diff"
	"github.com/fossology/report-aggregator/internal/export"
	"github.com/fossology/report-aggregator/internal/merge"
	"github.com/fossology/report-aggregator/internal/model"
	"github.com/fossology/report-aggregator/internal/parser"
	"github.com/fossology/report-aggregator/internal/validate"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "report-aggregator",
		Short: "A tool to merge and analyze software compliance reports",
	}

	// Merge command
	var outputFormat string
	var outputFile string
	var mergeCmd = &cobra.Command{
		Use:   "merge [file1] [file2] ...",
		Short: "Merge multiple compliance reports into one",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var parsedReports []*model.ParsedReport
			for _, file := range args {
				report, err := parser.DetectAndParse(file)
				if err != nil {
					return fmt.Errorf("failed to parse %s: %w", file, err)
				}
				parsedReports = append(parsedReports, report)
			}

			result, err := merge.Merge(parsedReports)
			if err != nil {
				return fmt.Errorf("merge failed: %w", err)
			}

			var data []byte
			switch outputFormat {
			case "spdx":
				data, err = export.ToSPDX(result)
			case "cyclonedx":
				data, err = export.ToCycloneDX(result)
			default:
				return fmt.Errorf("unsupported output format: %s", outputFormat)
			}

			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			if outputFile != "" {
				if err := os.WriteFile(outputFile, data, 0644); err != nil {
					return fmt.Errorf("failed to write to file: %w", err)
				}
				fmt.Printf("✅ Successfully merged %d reports to %s\n", len(args), outputFile)
				fmt.Printf("   Components: %d\n", result.Summary.TotalComponents)
				fmt.Printf("   Conflicts: %d\n", result.Summary.TotalConflicts)
			} else {
				fmt.Println(string(data))
			}

			return nil
		},
	}
	mergeCmd.Flags().StringVarP(&outputFormat, "format", "f", "spdx", "Output format (spdx, cyclonedx)")
	mergeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")

	// Validate command
	var validateCmd = &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate a report against internal structural rules",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := args[0]
			report, err := parser.DetectAndParse(file)
			if err != nil {
				return fmt.Errorf("parse failed: %w", err)
			}

			result := validate.Validate(report)
			if result.Valid {
				fmt.Printf("✅ %s is valid\n", file)
			} else {
				fmt.Printf("❌ %s has issues:\n", file)
			}

			for _, e := range result.Errors {
				fmt.Printf("  [ERROR] %s: %s\n", e.Field, e.Message)
			}
			for _, w := range result.Warnings {
				fmt.Printf("  [WARNING] %s: %s\n", w.Field, w.Message)
			}

			if !result.Valid {
				os.Exit(1)
			}
			return nil
		},
	}

	// Diff command
	var diffCmd = &cobra.Command{
		Use:   "diff [old_file] [new_file]",
		Short: "Detect changes between two report versions",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldReport, err := parser.DetectAndParse(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse old file: %w", err)
			}

			newReport, err := parser.DetectAndParse(args[1])
			if err != nil {
				return fmt.Errorf("failed to parse new file: %w", err)
			}

			changes := diff.DetectChanges(oldReport, newReport)

			data, err := json.MarshalIndent(changes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format diff output: %w", err)
			}

			fmt.Println(string(data))
			return nil
		},
	}

	rootCmd.AddCommand(mergeCmd, validateCmd, diffCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
