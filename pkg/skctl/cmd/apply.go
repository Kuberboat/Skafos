package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"

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

			switch ruleKind.Kind {
			case string(core.RatioType):
				applyRatioRule(data)
			case string(core.RegexType):
				applyRegexRule(data)
			}
		},
	}
)

func applyRatioRule(data []byte) {
	var rule core.RatioRule
	if err := yaml.Unmarshal(data, &rule); err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}

	// Do some sanity checks
	if rule.Spec.Ratio > 100 {
		log.Fatalf("ratio cannot be more than 100")
	}

	client := client.NewCtlClient()
	resp, err := client.ApplyRatioRule(&rule)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response status: %v ;Ratio rule Applied\n", resp.Status)
}

func applyRegexRule(data []byte) {
	var rule core.RegexRule
	if err := yaml.Unmarshal(data, &rule); err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}

	// Do some sanity checks
	for _, matcher := range rule.Spec.Matchers {
		_, err := regexp.Compile(matcher.Regex)
		if err != nil {
			log.Fatalf("incorrect regex %s", matcher.Regex)
		}
	}

	client := client.NewCtlClient()
	resp, err := client.ApplyRegexRule(&rule)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response status: %v ;Regex rule Applied\n", resp.Status)
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringVarP(&file, "file", "f", "", "specify the configuration file")
	applyCmd.MarkFlagRequired("file")
}
