

Api Shortcuts
=========================

* ``rebuild.sh`` contains bunch of copy/paste snippets for shortcuts ``$LIOAPI``.

```sh
export LIOAPI="https://bulk.lytics.io"
export LIOAPI="http://lio7:5353"
export LIOAPI="http://localhost:5353"


# Use this request to check if the account specified by key has the query specified by alias

http  -v GET $LIOAPI/api/query/user_lyris_subscribers key==$LIOKEY

#To add a public query to an account 

http  -v POST $LIOAPI/api/query/user_lyris_subscribers/use key==$LIOKEY

# query validation
http -v POST $LIOAPI/api/query/_validate access_token==$LIOKEY < entity_scores.lql

# adding a query manually
http -v POST $LIOAPI/api/query access_token==$LIOKEY < lytics_scores.lql

# account registration with welcome email suppression
echo '{"email" : "fredmeyerjewlers@eroi.com"
        , "name": "Fred Meyer Jewlers"
        , "fid": "fredmeyerjewelers.com"
        , "domain": "fredmeyerjewelers.com"
        , "password": "PZsOhTD2"
    }' | http -v POST $LIOAPI/api/register suppress==true


# user-creation
echo '{"email" : "aaron+2@lytics.io"
    , "password":"spin"
    , "name":"aaron"
    , "schemaid": 1
    , "roles": ["api","admin","sysadmin"]
}' | http -v POST $LIOAPI/api/user key==$LIOKEY

```

