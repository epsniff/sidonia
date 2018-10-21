#! /bin/sh

# set the lio api  (chose env your are working on)
# export LIOAPI="http://localhost:5353"
# export LIOAPI="http://lio7:5353"
# export LIOAPI="http://bulk.lytics.io"
# export LIOAPI="https://lioapidev1.ngrok.com/"
# export LIOAPI="https://staging.lytics.io"


# Check status of Work, this will return list of work for given status
#   https://github.com/lytics/lio/blob/develop/src/api/rw/workmgmt.go#L220
#  
#  it currently requires a Lytics aid=12 Access Token (not sysadmin)
# 

http -v $LIOAPI/api/work/_status access_token==$LIOKEY #  status default = errorwait 
http -v $LIOAPI/api/work/_status access_token==$LIOKEY status==paused
http -v $LIOAPI/api/work/_status access_token==$LIOKEY status==killed
http -v $LIOAPI/api/work/_status access_token==$LIOKEY status==completed
http -v $LIOAPI/api/work/_status access_token==$LIOKEY status==runnable
http -v $LIOAPI/api/work/_status access_token==$LIOKEY status==sleeping
http -v $LIOAPI/api/work/_status access_token==$LIOKEY status==failed
http -v $LIOAPI/api/work/_status access_token==$LIOKEY status==failed acctid==


# all errorwait work items
curl -s "$LIOAPI/api/work/_status?access_token=EYBDi6MectBTMbm8ewX7rsIb" | jq -c '.data[] | {aid, statuscode, workflow, id}'
#  show aid, status of all runnable work
curl -s "$LIOAPI/api/work/_status?status=runnable&access_token=EYBDi6MectBTMbm8ewX7rsIb" | \
   jq -c '.data[] | {aid, statuscode, workflow, id}'
# show aid, status, workflow for all items in aid = 12
curl -s "$LIOAPI/api/work/_status?status=all&access_token=EYBDi6MectBTMbm8ewX7rsIb" | \
   jq '.data[] | {aid, statuscode, workflow, id} | select(.aid == 12)'
curl -s "$LIOAPI/api/work/_status?status=all&access_token=EYBDi6MectBTMbm8ewX7rsIb" | \
   jq -c '.data[] | {aid, statuscode, workflow, id} | select(.statuscode != "completed") | select(.aid == 12)'
# show all {lql,scoring} jobs
curl -s "$LIOAPI/api/work/_status?access_token=EYBDi6MectBTMbm8ewX7rsIb&status=all" | \
   jq -c '.data[] | select(.workflow == "lql_multi_select" or .workflow == "lql_multi_select_v2") | {aid, statuscode, workflow, id}'
curl -s "$LIOAPI/api/work/_status?access_token=EYBDi6MectBTMbm8ewX7rsIb&status=all" | \
   jq -c '.data[] | select(.workflow == "lytics_scores") | {aid, statuscode, workflow, id}'
# show all paused jobs
curl -s "$LIOAPI/api/work/_status?access_token=EYBDi6MectBTMbm8ewX7rsIb&status=all" | \
   jq -c '.data[] | select(.statuscode == "paused") | {aid, workflow, statuscode, id}'

# all errorwait and their state 
curl -s "$LIOAPI/api/work/_status?access_token=EYBDi6MectBTMbm8ewX7rsIb" | jq '.data[] | {aid, statuscode, workflow, id, state}'

# my .local_env has the following
alias loadaid='curl -sL "$LIOAPI/api/account/$AID/env?access_token=$LIOADMINKEY" > /tmp/lyticsenv && source /tmp/lyticsenv && http $LIOAPI/api/account key==$LIOKEY'
alias showwork='curl -s "$LIOAPI/api/work?access_token=$LIOKEY&showhidden=true" | jq -c ".data[] | {workflow, id, statuscode} "'
alias rebuildlql='http -v --timeout 120 POST $LIOAPI/api/work/_qryrebuild access_token==$LIOADMINKEY account_id==$LIOACCTID concurrent==true'

# bunch of little shortcuts for querying mgmt api's per account
export AID=12
curl -sL "$LIOAPI/api/account/$AID/env?access_token=$LIOADMINKEY" > /tmp/lyticsenv && source /tmp/lyticsenv && http $LIOAPI/api/account key==$LIOKEY
curl -s "$LIOAPI/api/query?access_token=$LIOKEY" | jq -c '.data[] | {alias}'
curl -s "$LIOAPI/api/work?access_token=$LIOKEY&showhidden=true" | jq -c '.data[] | {workflow, id, statuscode} '

curl -s "$LIOAPI/api/segment?access_token=$LIOKEY" | jq -c '.data[] | {name, id, slug_name}'
curl -s "$LIOAPI/api/account?access_token=$LIOADMINKEY" | jq -c '.data[] | {name, aid, id}'
curl -s "$LIOAPI/api/account?access_token=$LIOKEY" | jq '.data[] '
curl -s "$LIOAPI/api/auth?access_token=$LIOKEY" | jq -c '.data[] | {pn: .provider_name, modified_ms, id, user_id} | select(.pn == "Facebook") '
curl -s "$LIOAPI/api/user?access_token=$LIOKEY" | jq -c '.data[] | {name, email, id}'


# query validation  and updating/push
http -v POST $LIOAPI/api/query/_validate access_token==$LIOKEY < lytics_scores.lql
http -v POST $LIOAPI/api/query access_token==$LIOKEY < lytics_scores.lql

# work resume, pause, bounce
http -v POST $LIOAPI/api/work/$WID/resume access_token==$LIOKEY
http -v POST $LIOAPI/api/work/$WID/pause access_token==$LIOKEY
http -v POST $LIOAPI/api/work/$WID/bounce access_token==$LIOKEY

# create script to pause all currently running work
http $LIOAPI/api/work/_status workflow_id==$WFID status==runnable key==$LIOADMINKEY | jq -r '.data[] | .id ' | sed 's/^\(.*\)$/\$LIOAPI\/api\/work\/\1\/pause key==\$LIOADMINKEY/'

# generic rebuild

http -v --timeout 120 POST $LIOAPI/api/work/_qryrebuild access_token==$LIOADMINKEY account_id==$LIOACCTID table==user
http -v POST $LIOAPI/api/work/_streamstats access_token==$LIOADMINKEY rebuild==true account_id==$LIOACCTID

# use a query
http -v POST $LIOAPI/api/query/user_campaignmonitor_activity/use key==$LIOKEY
http -v POST $LIOAPI/api/query/user_campaignmonitor_users/use key==$LIOKEY


# Sync Queries for accounts and start data processing
./sync.sh public  #you must first load public queries
./sync.sh athletepath  # 10

# scoring should always be started with both scoring workflows
echo '{
    "workflow_id": "lytics_scores_v2",
    "tag":"lytics_scores",
    "name": "Scoring v2",
    "config": {}
}' | http -v POST $LIOAPI/api/work key==$LIOKEY


echo '{
    "features": { "audience-attribution": true  }
}' |  http -v PUT $LIOAPI/api/account/$LIOACCTID key==$LIOKEY

echo '{
    "features": { "lookalike-modeling": true  }
}' |  http -v PUT $LIOAPI/api/account/$LIOACCTID key==$LIOKEY

# Settings are NOT visible in API, internal db only
echo '{
    "settings": { "lql_version": "linkgrid_v4"  }
}' |  http -v PUT $LIOAPI/api/account/$LIOACCTID key==$LIOKEY

echo '{
    "settings": { "store_version": 1  }
}' |  http -v PUT $LIOAPI/api/account/$LIOACCTID key==$LIOKEY

# boolean flag for content enrichment
echo '{
    "settings": { "enrich_content": true  }
}' |  http -v PUT $LIOAPI/api/account/$LIOACCTID key==$LIOKEY

# Tags are NOT visible in API, internal db only
#   they may not be deleted from api
echo '{
    "tags": ["lqlv2"]
}' |  http -v PUT $LIOAPI/api/account/$LIOACCTID key==$LIOKEY

# set the folder path for finding github lql statements for an account 
# IF IT DOES NOT match:
#   /customers/:name    or /customers/:aid 
echo '{
    "settings": { "github_path": "praetorian/policeone"  }
}' |  http -v PUT $LIOAPI/api/account/$LIOACCTID key==$LIOKEY

# collect account_state api
echo '{
  "flag_is_customer": true 
}' |  http -v PUT $LIOAPI/collect/bulk/account_state key==$LIODATAKEY