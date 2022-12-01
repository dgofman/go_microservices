package packages

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"gitlab.com/soteapps/packages/v2021/sError"
	"gitlab.com/soteapps/packages/v2021/sHelper"
)

var (
	AssertEqual = sHelper.AssertEqual
)

func createSubscriber(t *testing.T, schema sHelper.Schema) sHelper.Subscriber {
	soteErr := schema.Validate()
	AssertEqual(t, soteErr.FmtErrMsg, "")
	s := sHelper.Subscriber{
		Schema: &schema,
		Run: &sHelper.Run{
			Env: sHelper.Environment{
				TargetEnvironment: "staging",
				TestMode:          true,
			},
		},
	}
	s.PublishMessage = func(header sHelper.RequestHeaderSchema, soteErr sError.SoteError, message interface{}) sError.SoteError {
		var returnValue int64 = 123
		m, ok := message.(map[string]int64)
		AssertEqual(t, ok, true)
		AssertEqual(t, soteErr.FmtErrMsg, "")
		AssertEqual(t, m["transaction-id"], returnValue)
		return sError.SoteError{}
	}
	return s
}

func TestRun(t *testing.T) {
	os.Chdir("..")
	env := sHelper.MockRunHelper(t, natsConsumerName, natsSubject)
	Run(env)
}

func TestRunMessageListener(t *testing.T) {
	var (
		addGuard    *sHelper.PatchGuard
		removeGuard *sHelper.PatchGuard
		listGuard   *sHelper.PatchGuard
	)
	addGuard = sHelper.Patch(addFintrans, func(s *sHelper.Subscriber, msg *sHelper.Msg) sError.SoteError {
		addGuard.Unpatch()
		AssertEqual(t, msg.Subject, "bsl.fin-trans.trip.add")
		return sError.SoteError{}
	})
	messageListener(nil, &sHelper.Msg{Subject: "bsl.fin-trans.trip.add"})

	removeGuard = sHelper.Patch(removeFintrans, func(s *sHelper.Subscriber, msg *sHelper.Msg) sError.SoteError {
		removeGuard.Unpatch()
		AssertEqual(t, msg.Subject, "bsl.fin-trans.trip.remove")
		return sError.SoteError{}
	})
	messageListener(nil, &sHelper.Msg{Subject: "bsl.fin-trans.trip.remove"})

	listGuard = sHelper.Patch(listFintrans, func(s *sHelper.Subscriber, msg *sHelper.Msg) sError.SoteError {
		listGuard.Unpatch()
		AssertEqual(t, msg.Subject, "bsl.fin-trans.trip.list")
		return sError.SoteError{}
	})
	messageListener(nil, &sHelper.Msg{Subject: "bsl.fin-trans.trip.list"})

	soteErr := messageListener(nil, &sHelper.Msg{Subject: "bsl.organization"})
	AssertEqual(t, soteErr.FmtErrMsg, "109999: bsl.organization was/were not found")
}

func TestRunAddFintrans(t *testing.T) {
	var (
		createPatch *sHelper.PatchGuard
	)
	createPatch = sHelper.Patch(createTripFinancialTransactions, func(s *sHelper.Subscriber, body FintransAdd) (int64, sError.SoteError) {
		createPatch.Unpatch()
		return 123, sError.SoteError{}
	})
	data, err := json.Marshal(map[string]interface{}{
		"request-header": map[string]interface{}{
			"json-web-token":   "eyJraWQiOvxxxxxxx",
			"message-id":       "1a8eb33e-9db2-11eb-a8b3-0242ac130003",
			"aws-user-name":    "soteuser",
			"organizations-id": 10003,
		},
		"client-company-id":  32,
		"trip-id":            10002,
		"amount":             102.90,
		"currency":           "USD",
		"fintrans-type":      "F147",
		"cost-is-unexpected": true,
		"load-name":          "1",
		"memo":               "This is a memo",
	})
	AssertEqual(t, err, nil)
	msg := sHelper.Msg{
		Data: data,
	}
	s := createSubscriber(t, addSchema)
	soteErr := addFintrans(&s, &msg)
	AssertEqual(t, soteErr.FmtErrMsg, "")
}

func TestRunInvalidCustomBody(t *testing.T) {
	soteError := "206200: Message doesn't match signature. Sender must provide the following parameter names: #/properties/client-company-id"
	s := createSubscriber(t, addSchema)
	s.PublishMessage = func(header sHelper.RequestHeaderSchema, soteErr sError.SoteError, message interface{}) sError.SoteError {
		AssertEqual(t, soteErr.FmtErrMsg, soteError)
		return sError.SoteError{}
	}
	data, err := json.Marshal(map[string]interface{}{
		"request-header": map[string]interface{}{
			"json-web-token":   "eyJraWQiOvxxxxxxx",
			"message-id":       "1a8eb33e-9db2-11eb-a8b3-0242ac130003",
			"aws-user-name":    "soteuser",
			"organizations-id": 10003,
		},
		"trip-id":            10002,
		"amount":             102.90,
		"currency":           "USD",
		"fintrans-type":      "F147",
		"cost-is-unexpected": true,
		"load-name":          "1",
		"memo":               "This is a memo",
	})
	AssertEqual(t, err, nil)
	msg := sHelper.Msg{
		Data: data,
	}
	soteErr := addFintrans(&s, &msg)
	AssertEqual(t, soteErr.FmtErrMsg, soteError)
}

func TestRunRemoveFintrans(t *testing.T) {
	var (
		id          int64 = 123
		removeGuard *sHelper.PatchGuard
	)
	removeGuard = sHelper.Patch(removeTripFinancialTransactions, func(*sHelper.Subscriber, FintransRemove, string) (int64, sError.SoteError) {
		removeGuard.Unpatch()
		return id, sError.SoteError{}
	})
	data, err := json.Marshal(map[string]interface{}{
		"request-header": map[string]interface{}{
			"json-web-token":   "eyJraWQiOvxxxxxxx",
			"message-id":       "1a8eb33e-9db2-11eb-a8b3-0242ac130003",
			"aws-user-name":    "soteuser",
			"organizations-id": 10003,
		},
		"transaction-id": id,
	})
	AssertEqual(t, err, nil)
	msg := sHelper.Msg{
		Data: data,
	}
	s := createSubscriber(t, removeSchema)
	s.PublishMessage = func(header sHelper.RequestHeaderSchema, soteErr sError.SoteError, message interface{}) sError.SoteError {
		AssertEqual(t, fmt.Sprint(message), "REMOVED")
		return sError.SoteError{}
	}
	soteErr := removeFintrans(&s, &msg)
	AssertEqual(t, soteErr.FmtErrMsg, "")
}

func TestRunRemoveFintransNotFoundError(t *testing.T) {
	var (
		removeGuard *sHelper.PatchGuard
	)
	s := createSubscriber(t, removeSchema)
	removeGuard = sHelper.Patch(removeTripFinancialTransactions, func(*sHelper.Subscriber, FintransRemove, string) (int64, sError.SoteError) {
		removeGuard.Unpatch()
		return 0, sError.SoteError{}
	})
	s.PublishMessage = func(header sHelper.RequestHeaderSchema, soteErr sError.SoteError, message interface{}) sError.SoteError {
		AssertEqual(t, soteErr.FmtErrMsg, "109999: tripfinancialtransactions_id=271 was/were not found")
		return sError.SoteError{}
	}
	data, err := json.Marshal(map[string]interface{}{
		"request-header": map[string]interface{}{
			"json-web-token":   "eyJraWQiOvxxxxxxx",
			"message-id":       "1a8eb33e-9db2-11eb-a8b3-0242ac130003",
			"aws-user-name":    "soteuser",
			"organizations-id": 10003,
		},
		"transaction-id": 271,
	})
	AssertEqual(t, err, nil)
	msg := sHelper.Msg{
		Data: data,
	}
	soteErr := removeFintrans(&s, &msg)
	AssertEqual(t, soteErr.FmtErrMsg, "")
}

func TestRunListFintrans(t *testing.T) {
	var (
		listGuard *sHelper.PatchGuard
	)
	listGuard = sHelper.Patch(listTripFinancialTransactions, func(*sHelper.Subscriber, FintransList) (sHelper.QueryResult, sError.SoteError) {
		listGuard.Unpatch()
		return sHelper.QueryResult{
			Items: []interface{}{"COL1", "COL2", "COL3"},
		}, sError.SoteError{}
	})
	data, err := json.Marshal(map[string]interface{}{
		"request-header": map[string]interface{}{
			"json-web-token":   "eyJraWQiOvxxxxxxx",
			"message-id":       "1a8eb33e-9db2-11eb-a8b3-0242ac130003",
			"aws-user-name":    "soteuser",
			"organizations-id": 10003,
		},
	})
	AssertEqual(t, err, nil)
	msg := sHelper.Msg{
		Data: data,
	}
	s := createSubscriber(t, listSchema)
	s.PublishMessage = func(header sHelper.RequestHeaderSchema, soteErr sError.SoteError, message interface{}) sError.SoteError {
		AssertEqual(t, soteErr.FmtErrMsg, "")
		AssertEqual(t, fmt.Sprint(message), "{[COL1 COL2 COL3] <nil>}")
		return sError.SoteError{}
	}
	soteErr := listFintrans(&s, &msg)
	AssertEqual(t, soteErr.FmtErrMsg, "")
}

func TestRunListFintransError(t *testing.T) {
	var (
		listGuard *sHelper.PatchGuard
	)
	listGuard = sHelper.Patch(listTripFinancialTransactions, func(*sHelper.Subscriber, FintransList) (sHelper.QueryResult, sError.SoteError) {
		listGuard.Unpatch()
		return sHelper.QueryResult{}, sHelper.NewError().SqlError("Invalid column")
	})
	data, err := json.Marshal(map[string]interface{}{
		"request-header": map[string]interface{}{
			"json-web-token":   "eyJraWQiOvxxxxxxx",
			"message-id":       "1a8eb33e-9db2-11eb-a8b3-0242ac130003",
			"aws-user-name":    "soteuser",
			"organizations-id": 10003,
		},
	})
	AssertEqual(t, err, nil)
	msg := sHelper.Msg{
		Data: data,
	}
	s := createSubscriber(t, listSchema)
	s.PublishMessage = func(header sHelper.RequestHeaderSchema, soteErr sError.SoteError, message interface{}) sError.SoteError {
		AssertEqual(t, soteErr.FmtErrMsg, "200999: SQL error - see Details ERROR DETAILS: >>Key: SQL ERROR Value: Invalid column")
		return sError.SoteError{}
	}
	soteErr := listFintrans(&s, &msg)
	AssertEqual(t, soteErr.FmtErrMsg, "")
}
