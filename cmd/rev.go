package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	messages "github.com/cucumber/messages/go/v28"
	"github.com/experimental-software/gherkin/core"
	"github.com/spf13/cobra"
)

var revCmd = &cobra.Command{
	Use:   "rev [flags]",
	Short: "Reverse engineer feature files from source code",

	Run: func(cmd *cobra.Command, args []string) {
		docs, err := parseSpecSources(SourceParameter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = writeFeatureFiles(docs, TargetParameter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Wrote %d feature files to %s\n", len(docs), TargetParameter)
	},
}

func parseSpecSources(source string) ([]*messages.GherkinDocument, error) {
	info, err := os.Stat(source)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return core.ParseSpecFile(source)
	}

	var docs []*messages.GherkinDocument
	err = filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".spec.ts") {
			return nil
		}

		parsedDocs, parseErr := core.ParseSpecFile(path)
		// Use any documents the parser managed to produce even when it
		// reports an error, so partially-invalid spec files still yield
		// their recognisable features instead of aborting the walk.
		docs = append(docs, parsedDocs...)
		if parseErr != nil && len(parsedDocs) == 0 {
			return parseErr
		}
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s: %v\n", path, parseErr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return docs, nil
}

func writeFeatureFiles(docs []*messages.GherkinDocument, targetDir string) error {
	if targetDir == "" {
		return fmt.Errorf("target parameter is required")
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}

	//info, err := os.Stat(targetDir)
	//if err != nil {
	//	if os.IsNotExist(err) {
	//		return fmt.Errorf("target directory does not exist: %s", targetDir)
	//	}
	//	return err
	//}
	//if !info.IsDir() {
	//	return fmt.Errorf("target path is not a directory: %s", targetDir)
	//}

	for _, doc := range docs {
		if doc == nil {
			continue
		}

		uri := core.CalculateFeaturePath(doc.Uri)
		relativeDir, baseName := featureFilePathPartsFromURI(uri)
		baseName = sanitizeFeatureFileBaseName(baseName)

		featurePath := filepath.Join(targetDir, filepath.FromSlash(relativeDir), baseName+".feature")

		if err := os.MkdirAll(filepath.Dir(featurePath), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(featurePath, []byte(renderFeatureDocument(doc)), 0o644); err != nil {
			return err
		}
	}

	return nil
}

// sanitizeFeatureFileBaseName returns a filesystem-safe base name for a feature
// file by replacing whitespace characters with underscores.
func sanitizeFeatureFileBaseName(name string) string {
	return strings.ReplaceAll(name, " ", "_")
}

// featureFilePathPartsFromURI splits a slash-separated feature URI into the
// relative directory and the base name (the last segment).
func featureFilePathPartsFromURI(uri string) (string, string) {
	idx := strings.LastIndex(uri, "/")
	if idx < 0 {
		return "", uri
	}
	return uri[:idx], uri[idx+1:]
}

func renderFeatureDocument(doc *messages.GherkinDocument) string {
	if doc == nil || doc.Feature == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("Feature: ")
	b.WriteString(doc.Feature.Name)
	b.WriteString("\n")

	for _, child := range doc.Feature.Children {
		if child == nil {
			continue
		}

		if child.Background != nil {
			b.WriteString("\n  Background:\n")
			for _, step := range child.Background.Steps {
				if step == nil {
					continue
				}
				b.WriteString("    ")
				b.WriteString(strings.TrimSpace(step.Keyword))
				b.WriteString(" ")
				b.WriteString(step.Text)
				b.WriteString("\n")
			}
		}

		if child.Scenario != nil {
			b.WriteString("\n  Scenario: ")
			b.WriteString(child.Scenario.Name)
			b.WriteString("\n")
			for _, step := range child.Scenario.Steps {
				if step == nil {
					continue
				}
				b.WriteString("    ")
				b.WriteString(strings.TrimSpace(step.Keyword))
				b.WriteString(" ")
				b.WriteString(step.Text)
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	return b.String()
}

func init() {
	revCmd.Flags().StringVarP(&SourceParameter, "source", "s", "", "Path to source file or directory")
	revCmd.Flags().StringVarP(&TargetParameter, "target", "t", "", "Path to target directory")
	_ = revCmd.MarkFlagRequired("source")
	_ = revCmd.MarkFlagRequired("target")

	rootCmd.AddCommand(revCmd)
}
