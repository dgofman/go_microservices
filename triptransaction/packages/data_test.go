package packages

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"gitlab.com/soteapps/packages/v2021/sDatabase"
	"gitlab.com/soteapps/packages/v2021/sError"
	"gitlab.com/soteapps/packages/v2021/sHelper"
)

func TestCreateTripFinancialTransactions(t *testing.T) {
	var (
		rowId      int64 = 123
		queryClose *sHelper.PatchGuard
		queryExec  *sHelper.PatchGuard
	)
	queryClose = sHelper.Patch(sHelper.Query.Close, func(sHelper.Query, sDatabase.SRows, *sError.SoteError) { queryClose.Unpatch() })
	queryExec = sHelper.Patch(sHelper.Query.Exec, func(q sHelper.Query, r *sHelper.Run) (sDatabase.SRows, sError.SoteError) {
		queryExec.Unpatch()
		rows := sDatabase.Rows{}
		index := 0
		rows.IScan = func(dest ...interface{}) error {
			val := dest[0].(*int64)
			*val = rowId
			return nil
		}
		rows.INext = func() bool {
			index++
			return index == 1
		}
		return rows, sError.SoteError{}
	})
	s := sHelper.Subscriber{}
	id, soteErr := createTripFinancialTransactions(&s, FintransAdd{})
	AssertEqual(t, id, rowId)
	AssertEqual(t, soteErr.FmtErrMsg, "")
}

func TestRemovetTripFinancialTransactions(t *testing.T) {
	var (
		companyId  int64 = 123
		queryClose *sHelper.PatchGuard
		queryExec  *sHelper.PatchGuard
	)
	queryClose = sHelper.Patch(sHelper.Query.Close, func(sHelper.Query, sDatabase.SRows, *sError.SoteError) { queryClose.Unpatch() })
	queryExec = sHelper.Patch(sHelper.Query.Exec, func(q sHelper.Query, r *sHelper.Run) (sDatabase.SRows, sError.SoteError) {
		queryExec.Unpatch()
		AssertEqual(t, q.Sql.String(), "DELETE FROM sote.tripfinancialtransactions")
		rows := sDatabase.Rows{}
		index := 0
		rows.IScan = func(dest ...interface{}) error {
			val := dest[0].(*int64)
			*val = companyId
			return nil
		}
		rows.INext = func() bool {
			index++
			return index == 1
		}
		return rows, sError.SoteError{}
	})
	s := sHelper.Subscriber{}
	id, soteErr := removeTripFinancialTransactions(&s, FintransRemove{}, "client_company_id=123")
	AssertEqual(t, id, companyId)
	AssertEqual(t, soteErr.FmtErrMsg, "")
}

func TestDataListTripFinancialTransactions(t *testing.T) {
	var (
		total      int64 = 1
		queryExec  *sHelper.PatchGuard
		queryClose *sHelper.PatchGuard
	)
	queryExec = sHelper.Patch(sHelper.Query.Exec, func(q sHelper.Query, r *sHelper.Run) (sDatabase.SRows, sError.SoteError) {
		queryExec.Unpatch()
		AssertEqual(t, q.Sql.String(), "SELECT count(*) OVER(), COL1, COL2, COL3")
		rows := sDatabase.Rows{}
		index := 0
		rows.IValues = func() ([]interface{}, error) {
			return []interface{}{total, 123, true, "Hello World"}, nil
		}
		rows.INext = func() bool {
			index++
			return index == 1
		}
		return rows, sError.SoteError{}
	})
	queryClose = sHelper.Patch(sHelper.Query.Close, func(sHelper.Query, sDatabase.SRows, *sError.SoteError) { queryClose.Unpatch() })
	body := FintransList{
		Filter: sHelper.FilterHeaderSchema{
			Items: []string{"COL1", "COL2", "COL3"},
		},
	}
	result, soteErr := listTripFinancialTransactions(&sHelper.Subscriber{}, body)
	AssertEqual(t, soteErr.FmtErrMsg, "")
	data, _ := json.MarshalIndent(result, "", "")
	re := regexp.MustCompile(`\r?\n`)
	AssertEqual(t, re.ReplaceAllString(string(data), ""), `{"items": [{"COL1": 123,"COL2": true,"COL3": "Hello World"}],"pagination": {"total": 1,"limit": 0,"offset": 0}}`)
}

func TestDataListTripFinancialTransactionsError(t *testing.T) {
	var (
		queryExec  *sHelper.PatchGuard
		queryClose *sHelper.PatchGuard
	)
	queryExec = sHelper.Patch(sHelper.Query.Exec, func(q sHelper.Query, r *sHelper.Run) (sDatabase.SRows, sError.SoteError) {
		queryExec.Unpatch()
		AssertEqual(t, q.Sql.String(), "SELECT count(*) OVER(), tripfinancialtransactions_id, organizations_id, client_company_id, trips_id, financialtransactions_type, currency_type, transactions_amount, cost_is_unexpected, load_name, memo, transactions_timestamp, created_by_requestor_username")
		rows := sDatabase.Rows{}
		index := 0
		rows.IValues = func() ([]interface{}, error) {
			return nil, errors.New("invalid column")
		}
		rows.INext = func() bool {
			index++
			return index == 1
		}
		return rows, sError.SoteError{}
	})
	queryClose = sHelper.Patch(sHelper.Query.Close, func(sHelper.Query, sDatabase.SRows, *sError.SoteError) { queryClose.Unpatch() })
	body := FintransList{}
	defer func() {
		r := recover()
		AssertEqual(t, fmt.Sprint(r), "200999: SQL error - see Details ERROR DETAILS: >>Key: SQL ERROR Value: invalid column")
	}()
	listTripFinancialTransactions(&sHelper.Subscriber{}, body)
}
