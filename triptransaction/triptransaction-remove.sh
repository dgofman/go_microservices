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
    "transaction-id": 10006
}
EOF` | nats pub bsl.fin-trans.trip.remove

wait

`
Return JSON:
{
    "message": "REMOVED",
    "message-id": "1a8eb33e-9db2-11eb-a8b3-0242ac130003"
}

Error JSON:
{
        "error": {
                "ErrCode": 206200,
                "ErrType": "NATS_Error",
                "ParamCount": 1,
                "ParamDescription": "List of required parameters",
                "FmtErrMsg": "206200: Message doesn't match signature. Sender must provide the following parameter names: #/properties/client-company-id",
                "ErrorDetails": {},
                "Loc": ""
        },
        "message-id": "1a8eb33e-9db2-11eb-a8b3-0242ac130003"
}

or

{
        "error": {
                "ErrCode": 109999,
                "ErrType": "User_Error",
                "ParamCount": 1,
                "ParamDescription": "Item name",
                "FmtErrMsg": "109999: tripfinancialtransactions_id=10006 was/were not found",
                "ErrorDetails": {},
                "Loc": ""
        },
        "message-id": "1a8eb33e-9db2-11eb-a8b3-0242ac130003"
}
`