package packages

import (
	"gitlab.com/soteapps/packages/v2021/sError"
	"gitlab.com/soteapps/packages/v2021/sHelper"
	"gitlab.com/soteapps/packages/v2021/sLogger"
)

func createTripFinancialTransactions(s *sHelper.Subscriber, body FintransAdd) (int64, sError.SoteError) {
	sLogger.DebugMethod()
	var id int64
	query := sHelper.Query{
		Table: "tripfinancialtransactions",
		Columns: []string{"created_by_requestor_username", "organizations_id", "client_company_id", "trips_id",
			"financialtransactions_type", "currency_type", "transactions_amount", "cost_is_unexpected", "load_name", "memo"},
		Values: []interface{}{body.Header.AwsUserName, body.Header.OrganizationId, body.ClientCompanyId, body.TripId,
			body.FintransType, body.Currency, body.Amount, body.CostIsUnexpected, body.LoadName, body.Memo},
	}
	tRows, soteErr := query.Insert("tripfinancialtransactions_id").Exec(s.Run)
	if soteErr.ErrCode == nil {
		for tRows.Next() {
			tRows.Scan(&id)
		}
	}
	query.Close(tRows, &soteErr)
	return id, soteErr
}

func removeTripFinancialTransactions(s *sHelper.Subscriber, body FintransRemove, whereClause string) (int64, sError.SoteError) {
	sLogger.DebugMethod()
	var id int64
	query := sHelper.Query{
		Table: "tripfinancialtransactions",
		Where: whereClause,
	}
	tRows, soteErr := query.Delete("tripfinancialtransactions_id").Exec(s.Run)
	if soteErr.ErrCode == nil {
		for tRows.Next() {
			tRows.Scan(&id)
		}
	}
	query.Close(tRows, &soteErr)
	return id, soteErr
}

func listTripFinancialTransactions(s *sHelper.Subscriber, body FintransList) (sHelper.QueryResult, sError.SoteError) {
	sLogger.DebugMethod()
	query := sHelper.Query{
		Table:  "tripfinancialtransactions",
		Filter: &body.Filter,
	}.Pagination()
	if len(body.Filter.Items) == 0 {
		query.Filter.Items = []string{"tripfinancialtransactions_id", "organizations_id", "client_company_id", "trips_id", "financialtransactions_type",
			"currency_type", "transactions_amount", "cost_is_unexpected", "load_name", "memo", "transactions_timestamp", "created_by_requestor_username"} // display default columns
	}
	tRows, soteErr := query.Select().Exec(s.Run)
	if soteErr.ErrCode == nil {
		for tRows.Next() {
			if _, soteErr := query.Scan(tRows); soteErr.ErrCode != nil {
				s.Run.PanicService(soteErr)
			}
		}
	}
	query.Close(tRows, &soteErr)
	return query.Result, soteErr
}
