#!/bin/zsh
source <(curl -s https://gitlab.com/soteapps/packages/-/raw/v2021/setup.sh)

echo `cat <<EOF
{
    "request-header": {
        "json-web-token": "eyJraWQiOvxxxxxxx",
        "message-id": "1a8eb33e-9db2-11eb-a8b3-0242ac130003",
        "aws-user-name": "$USERID",
        "organizations-id": $ORGID,
        "device-id": $DEVICE_ID,
        "role-list": [
            "CLIENT_ADMIN",
            "EXECUTIVE"
        ]
    },
    "filter-header": {
        "items": ["tripfinancialtransactions_id", "organizations_id", "client_company_id", "trips_id", "financialtransactions_type", "currency_type", 
        "transactions_amount", "load_name", "memo", "cost_is_unexpected"],
        "limit": 3,
        "offset": 0,
        "sort_asc": ["tripfinancialtransactions_id"],
        "sort_desc": ["client_company_id"],
        "eq": {
            "organizations_id": 10003
        },
        "gt": {
            "client_company_id": 10
        },
        "lt": {
            "transactions_amount": 200
        }
        }

}
EOF` | nats pub bsl.fin-trans.trip.list

wait

`
Return JSON:
{
        "message": {
            "items": [
                {
                        "client_company_id": 1000,
                        "cost_is_unexpected": true,
                        "currency_type": "USD",
                        "financialtransactions_type": "TEST_FINTRANS_TYPE",
                        "load_name": "Test Load",
                        "memo": "This is a memo",
                        "organizations_id": 10003,
                        "transactions_amount": 102.9,
                        "tripfinancialtransactions_id": 10000,
                        "trips_id": 10000
                },
                {
                        "client_company_id": 32,
                        "cost_is_unexpected": true,
                        "currency_type": "USD",
                        "financialtransactions_type": "F147",
                        "load_name": "",
                        "memo": "This is a memo",
                        "organizations_id": 10003,
                        "transactions_amount": 102.9,
                        "tripfinancialtransactions_id": 10003,
                        "trips_id": 10000
                }
            ],
            "pagination": {
                "total": 2,
                "limit": 3,
                "offset": 0
            }
        },
        "message-id": "1a8eb33e-9db2-11eb-a8b3-0242ac130003"
}

Error JSON:
{
        "error": {
            "ErrCode": 200999,
            "ErrType": "Process_Error",
            "ParamCount": 0,
            "ParamDescription": "None",
            "FmtErrMsg": "200999: SQL error - see Details ERROR DETAILS: \u003e\u003eKey: SQL ERROR Value: ERROR: column \"tripfinancialtransactions_ids\" does not exist (SQLSTATE 42703)",
            "ErrorDetails": {
                "SQL ERROR": "ERROR: column \"tripfinancialtransactions_ids\" does not exist (SQLSTATE 42703)"
            },
            "Loc": ""
        },
        "message-id": "1a8eb33e-9db2-11eb-a8b3-0242ac130003"
}
`