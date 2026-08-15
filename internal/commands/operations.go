package commands

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/apicommands"
)

// operationBindings is the handwritten CLI naming layer over the generated
// OpenAPI catalog. More than one command may intentionally expose an operation,
// such as the member create and invite flows.
var operationBindings = map[string]string{
	"checkouts apple-pay-session":              "CreateApplePaySession",
	"checkouts create":                         "CreateCheckout",
	"checkouts deactivate":                     "DeactivateCheckout",
	"checkouts get":                            "GetCheckout",
	"checkouts list":                           "ListCheckouts",
	"checkouts payment-methods":                "GetPaymentMethods",
	"checkouts process":                        "ProcessCheckout",
	"checkouts update":                         "UpdateCheckout",
	"customers create":                         "CreateCustomer",
	"customers get":                            "GetCustomer",
	"customers payment-instruments deactivate": "DeactivatePaymentInstrument",
	"customers payment-instruments list":       "ListPaymentInstruments",
	"customers update":                         "UpdateCustomer",
	"members create":                           "CreateMerchantMember",
	"members delete":                           "DeleteMerchantMember",
	"members get":                              "GetMerchantMember",
	"members invite":                           "CreateMerchantMember",
	"members list":                             "ListMerchantMembers",
	"members update":                           "UpdateMerchantMember",
	"memberships list":                         "ListMemberships",
	"merchants get":                            "GetMerchant",
	"merchants persons get":                    "GetPerson",
	"merchants persons list":                   "ListPersons",
	"payouts list":                             "ListPayoutsV1",
	"readers add":                              "CreateReader",
	"readers checkout":                         "CreateReaderCheckout",
	"readers delete":                           "DeleteReader",
	"readers get":                              "GetReader",
	"readers list":                             "ListReaders",
	"readers status":                           "GetReaderStatus",
	"readers terminate":                        "CreateReaderTerminate",
	"readers update":                           "UpdateReader",
	"receipts get":                             "GetReceipt",
	"roles create":                             "CreateMerchantRole",
	"roles delete":                             "DeleteMerchantRole",
	"roles get":                                "GetMerchantRole",
	"roles list":                               "ListMerchantRoles",
	"roles update":                             "UpdateMerchantRole",
	"transactions get":                         "GetTransactionV2.1",
	"transactions list":                        "ListTransactionsV2.1",
	"transactions refund":                      "RefundTransaction",
}

func bindOpenAPIOperations(commands []*cli.Command) []*cli.Command {
	leaves := make(map[string]*cli.Command)
	var collect func([]string, *cli.Command)
	collect = func(parents []string, command *cli.Command) {
		path := append(parents, command.Name)
		if len(command.Commands) == 0 {
			leaves[strings.Join(path, " ")] = command
			return
		}
		for _, child := range command.Commands {
			collect(path, child)
		}
	}
	for _, command := range commands {
		collect(nil, command)
	}

	for commandPath, operationID := range operationBindings {
		command, ok := leaves[commandPath]
		if !ok {
			panic(fmt.Sprintf("OpenAPI operation binding refers to unknown command %q", commandPath))
		}
		apicommands.Bind(command, operationID)
	}

	return commands
}
