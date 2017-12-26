package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"

	"github.com/bytom/blockchain"
	"github.com/bytom/blockchain/txbuilder"
)

func init() {
	buildTransaction.PersistentFlags().StringVarP(&buildType, "type", "t", "", "transaction type, valid types: 'issue', 'spend'")
	buildTransaction.PersistentFlags().StringVarP(&receiverProgram, "receiver", "r", "", "program of receiver")
	buildTransaction.PersistentFlags().StringVarP(&btmGas, "gas", "g", "20000000", "program of receiver")
	buildTransaction.PersistentFlags().BoolVar(&pretty, "pretty", false, "pretty print json result")
	signTransactionCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "password of the account which sign these transaction(s)")
	signTransactionCmd.PersistentFlags().BoolVar(&pretty, "pretty", false, "pretty print json result")
	signSubTransactionCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "password of the account which sign these transaction(s)")
}

var (
	buildType       string
	btmGas          string
	receiverProgram string
	password        string
	pretty          bool
)

var buildIssueReqFmt = `
	{"actions": [
		{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":%s, "account_id": "%s"},
		{"type": "issue", "asset_id": "%s", "amount": %s},
		{"type": "control_account", "asset_id": "%s", "amount": %s, "account_id": "%s"}
	]}`

var buildSpendReqFmt = `
	{"actions": [
		{"type": "spend_account", "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "amount":%s, "account_id": "%s"},
		{"type": "spend_account", "asset_id": "%s","amount": %s,"account_id": "%s"},
		{"type": "control_receiver", "asset_id": "%s", "amount": %s, "receiver":{"control_program": "%s","expires_at":"2017-12-28T12:52:06.78309768+08:00"}}
	]}`

var buildTransaction = &cobra.Command{
	Use:   "build-transaction <accountID> <assetID> <amount>",
	Short: "Build one transaction template",
	Args:  cobra.RangeArgs(3, 4),
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("type")
		if buildType == "spend" {
			cmd.MarkFlagRequired("receiver")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		var buildReqStr string
		accountID := args[0]
		assetID := args[1]
		amount := args[2]
		switch buildType {
		case "issue":
			buildReqStr = fmt.Sprintf(buildIssueReqFmt, btmGas, accountID, assetID, amount, assetID, amount, accountID)
		case "spend":
			buildReqStr = fmt.Sprintf(buildSpendReqFmt, btmGas, accountID, assetID, amount, accountID, assetID, amount, receiverProgram)
		default:
			jww.ERROR.Println("Invalid transaction template type")
			os.Exit(ErrLocalExe)
		}

		var buildReq blockchain.BuildRequest
		if err := json.Unmarshal([]byte(buildReqStr), &buildReq); err != nil {
			jww.ERROR.Println(err)
			os.Exit(ErrLocalExe)
		}

		data, exitCode := clientCall("/build-transaction", &buildReq)
		if exitCode != Success {
			os.Exit(exitCode)
		}

		if pretty {
			printJSON(data)
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if ok != true {
			jww.ERROR.Println("invalid type assertion")
			os.Exit(ErrLocalParse)
		}

		rawTemplate, err := json.Marshal(dataMap)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(ErrLocalParse)
		}

		jww.FEEDBACK.Printf("Template Type: %s\n%s\n", buildType, string(rawTemplate))
	},
}

var signTransactionCmd = &cobra.Command{
	Use:   "sign-transaction  <json templates>",
	Short: "Sign transaction templates with account password",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("password")
	},
	Run: func(cmd *cobra.Command, args []string) {
		template := txbuilder.Template{}

		err := json.Unmarshal([]byte(args[0]), &template)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(ErrLocalExe)
		}

		var req = struct {
			Auth string
			Txs  txbuilder.Template `json:"transaction"`
		}{Auth: password, Txs: template}

		jww.FEEDBACK.Printf("\n\n")
		data, exitCode := clientCall("/sign-transaction", &req)
		if exitCode != Success {
			os.Exit(exitCode)
		}

		if pretty {
			printJSON(data)
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if ok != true {
			jww.ERROR.Println("invalid type assertion")
			os.Exit(ErrLocalParse)
		}

		rawSign, err := json.Marshal(dataMap)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(ErrLocalParse)
		}
		jww.FEEDBACK.Printf("\nSign Template:\n%s\n", string(rawSign))
	},
}

var submitTransactionCmd = &cobra.Command{
	Use:   "submit-transaction  <signed json template>",
	Short: "Submit signed transaction template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		template := txbuilder.Template{}

		err := json.Unmarshal([]byte(args[0]), &template)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(ErrLocalExe)
		}

		jww.FEEDBACK.Printf("\n\n")
		data, exitCode := clientCall("/submit-transaction", &template)
		if exitCode != Success {
			os.Exit(exitCode)
		}

		printJSON(data)
	},
}

var signSubTransactionCmd = &cobra.Command{
	Use:   "sign-submit-transaction  <json templates>",
	Short: "Sign and Submit transaction templates with account password",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("password")
	},
	Run: func(cmd *cobra.Command, args []string) {
		template := txbuilder.Template{}

		err := json.Unmarshal([]byte(args[0]), &template)
		if err != nil {
			jww.ERROR.Println(err)
			os.Exit(ErrLocalExe)
		}

		var req = struct {
			Auth string
			Txs  txbuilder.Template `json:"transaction"`
		}{Auth: password, Txs: template}

		jww.FEEDBACK.Printf("\n\n")
		data, exitCode := clientCall("/sign-submit-transaction", &req)
		if exitCode != Success {
			os.Exit(exitCode)
		}

		printJSON(data)
	},
}

var listTransactions = &cobra.Command{
	Use:   "list-transactions",
	Short: "List the transactions",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		data, exitCode := clientCall("/list-transactions")
		if exitCode != Success {
			os.Exit(exitCode)
		}

		printJSONList(data)
	},
}

var gasRateCmd = &cobra.Command{
	Use:   "gas-rate",
	Short: "Print the current gas rate",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		data, exitCode := clientCall("/gas-rate")
		if exitCode != Success {
			os.Exit(exitCode)
		}
		printJSON(data)
	},
}
