package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"p9t.io/skafos/pkg/api/core"
	"p9t.io/skafos/pkg/skctl/client"
)

var (
	file     string
	applyCmd = &cobra.Command{
		Use:   "apply [-f FILENAME]",
		Short: "Apply a routing rule by filename",
		Long: `Apply a routing rule by filename

Examples:
  # Apply the ratio rule in ratio.yaml
  skctl apply -f ./ratio.yaml`,
		Run: func(cmd *cobra.Command, args []string) {
			data, err := os.ReadFile(file)
			if err != nil {
				log.Fatal(err)
			}
			var ruleKind core.RuleKind
			err = yaml.Unmarshal(data, &ruleKind)
			if err != nil {
				log.Fatal("error decoding rule's type")
			}
			if ruleKind.Kind != string(core.RatioType) && ruleKind.Kind != string(core.RegexType) {
				log.Fatalf("type %v is not supported", ruleKind.Kind)
			}
			client := client.NewCtlClient()
			resp, err := client.ApplyRule(data)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Response status: %v ;Rule Applied\n", resp.Status)
		},
	}
)

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringVarP(&file, "file", "f", "", "specify the configuration file")
	applyCmd.MarkFlagRequired("file")
}
