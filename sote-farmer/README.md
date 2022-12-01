### Update dependencies
go mod download

### Clear cache
go clean -cache -modcache -i -r

### Clean Table(s)
sote-farmer --targetEnv=custom:development --configFile=sample.json --clean-tables

### Export Data
sote-farmer --targetEnv=staging --configFile=sample.json --export=raw

sote-farmer --targetEnv=staging --configFile=sample.json --export=csv

sote-farmer --targetEnv=staging --configFile=sample.json --export=json --obscure

sote-farmer --targetEnv=custom:production --configFile=sample.json --export=json --obscure

### Import Data
sote-farmer --targetEnv=development --bulkFile=organizations.raw --import

sote-farmer --targetEnv=staging --bulkFile=organizations.json --import

sote-farmer --targetEnv=demo --bulkFile=organizations-csv.zip --import

sote-farmer --targetEnv=custom:staging --configFile=sample.json --bulkFile=organizations.raw --import --clean-tables

### Screenshots
https://sote.myjetbrains.com/youtrack/issue/BE21-239

### Note
Setup ssh tunnel for the staging environment. See: https://gitlab.com/getsote/Wiki/-/wikis/Hosting/Database-Access

vi ~/.ssh/config

Host sote_remote_dbs
        HostName 63.33.91.244
        User YOUR_USERNAME
        IdentityFile ~/.ssh/id_rsa
        LocalForward 5432 sote-backend-staging.cyqkmy6t8tbb.eu-west-1.rds.amazonaws.com:5432

### Obscure pattern

[**PATTERN**]{**MIN**,**MAX**}  or [**PATTERN**]{**LENGTH**}

**MIN** - The preceding item is matched at least n times (If zero randomly exit or continue)

**MAX** - The preceding item is matched not more than n times.

**LENGTH** - The preceding item is matched exactly n times.

**PATTERN**:

> A-Z - All upper case

>a-z - All lower case

>A-z - Capitalize

>0-9 - Numbers

>[any characters]

#### Example:
_**[A-z]**{3,50} **[A-Z]**{0,50}_
```
    Hello WORLD
    Hello
```


 _**[a-z]**{5,20}**[-]**{0,1}**[A-Z]**{0,20}_
```
    hello-WORLD
    hello
```


_**[+]**{0,1}**[0-9]**{1,3}_
```
    (nil)
    +1
    +231
```


_**[A]**{0,1}**[0-9]**{9}_
```
    (nil)
    A123456789
```

