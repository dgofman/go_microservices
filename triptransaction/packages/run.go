package packages

import (
	"fmt"

	"gitlab.com/soteapps/packages/v2021/sError"
	"gitlab.com/soteapps/packages/v2021/sHelper"
	"gitlab.com/soteapps/packages/v2021/sLogger"
)

type FintransAdd struct {
	Header           sHelper.RequestHeaderSchema `json:"request-header"`
	ClientCompanyId  int64                       `json:"client-company-id"`
	TripId           int64                       `json:"trip-id"`
	FintransType     string                      `json:"fintrans-type"`
	Currency         string                      `json:"currency"`
	Amount           float64                     `json:"amount"`
	CostIsUnexpected bool                        `json:"cost-is-unexpected"`
	Memo             string                      `json:"memo"`
	LoadName         string                      `json:"load-name"`
}

type FintransRemove struct {
	Header sHelper.RequestHeaderSchema `json:"request-header"`
	Id     int64                       `json:"transaction-id"`
}

type FintransList struct {
	Header sHelper.RequestHeaderSchema `json:"request-header"`
	Filter sHelper.FilterHeaderSchema  `json:"filter-header"`
}

var (
	natsConsumerName = "bsl-fin-trans-trip-wildcard"
	natsSubject      = "bsl.fin-trans.trip.>"
	url              = "https://gitlab.com/soteapps/messages/-/raw/master/fin-trans-trip-%[1]s/request/fin-trans-trip-%[1]s.json"
	addSchema        = sHelper.Schema{
		FileName:  fmt.Sprintf(url, "add"),
		StructRef: &FintransAdd{},
	}
	removeSchema = sHelper.Schema{
		FileName:  fmt.Sprintf(url, "remove"),
		StructRef: &FintransRemove{},
	}
	listSchema = sHelper.Schema{
		FileName:  fmt.Sprintf(url, "list"),
		StructRef: &FintransList{},
	}
)

func Run(env sHelper.Environment) (soteErr sError.SoteError) {
	sLogger.DebugMethod()
	helper := sHelper.NewHelper(env)
	soteErr = addSchema.Validate()
	if soteErr.ErrCode == nil {
		soteErr = removeSchema.Validate()
	}
	if soteErr.ErrCode == nil {
		soteErr = listSchema.Validate()
	}
	if soteErr.ErrCode == nil {
		if soteErr = helper.AddSubscriber(natsConsumerName, natsSubject, messageListener, nil); soteErr.ErrCode == nil {
			helper.Run(true) //asynchronously using goroutine
		}
	}
	return
}

func messageListener(s *sHelper.Subscriber, message *sHelper.Msg) (soteErr sError.SoteError) {
	sLogger.DebugMethod()
	switch message.Subject {
	case "bsl.fin-trans.trip.add":
		return addFintrans(s, message)
	case "bsl.fin-trans.trip.remove":
		return removeFintrans(s, message)
	case "bsl.fin-trans.trip.list":
		return listFintrans(s, message)
	default:
		soteErr = sHelper.NewError().ItemNotFound(message.Subject)
	}
	return
}

func addFintrans(s *sHelper.Subscriber, message *sHelper.Msg) (soteErr sError.SoteError) {
	sLogger.DebugMethod()
	var (
		header sHelper.RequestHeaderSchema
		id     int64
	)
	body := FintransAdd{}
	header, soteErr = addSchema.ParseAndValidate(s.Run.Env, message.Data, &body)
	if soteErr.ErrCode == nil {
		id, soteErr = createTripFinancialTransactions(s, body)
		soteErr = s.PublishMessage(header, soteErr, map[string]int64{
			"transaction-id": id,
		})
	} else {
		s.PublishMessage(header, soteErr, nil)
	}
	return
}

func removeFintrans(s *sHelper.Subscriber, message *sHelper.Msg) (soteErr sError.SoteError) {
	sLogger.DebugMethod()
	var (
		header sHelper.RequestHeaderSchema
		id     int64
		status string
	)
	body := FintransRemove{}
	header, soteErr = removeSchema.ParseAndValidate(s.Run.Env, message.Data, &body)
	if soteErr.ErrCode == nil {
		whereClause := fmt.Sprintf("tripfinancialtransactions_id=%v", body.Id)
		id, soteErr = removeTripFinancialTransactions(s, body, whereClause)
		if soteErr.ErrCode == nil {
			if id != body.Id {
				soteErr = sHelper.NewError().ItemNotFound(whereClause)
			} else {
				status = "REMOVED"
			}
		}
	}
	soteErr = s.PublishMessage(header, soteErr, status)
	return
}

func listFintrans(s *sHelper.Subscriber, message *sHelper.Msg) (soteErr sError.SoteError) {
	sLogger.DebugMethod()
	var (
		header sHelper.RequestHeaderSchema
		result sHelper.QueryResult
	)
	body := FintransList{}
	header, soteErr = listSchema.ParseAndValidate(s.Run.Env, message.Data, &body)
	if soteErr.ErrCode == nil {
		result, soteErr = listTripFinancialTransactions(s, body)
	}
	soteErr = s.PublishMessage(header, soteErr, result)
	return
}
