#!/usr/bin/env bash

usage() {
  cat << EOT

    This is the Lytics customer helper

    ./sync.sh loadenv 10   # load aid=10
    ./sync.sh loadenv lytics.io # load lytics.io domain account

    sync.sh [account]

    # usage
    "========================================="
    ./sync.sh public
    ./sync.sh athletepath


EOT
}

LENV=""  #$LIOENV               # LioEnv override
CMD=$1

loadFlags() {
    if [ $# -ne "0" ] ; then
        while [ $# -gt 0 ]; do
            case $1 in
                -e) LENV=$2;                                    shift 2 ;;
                --help)
                  usage
                  exit 0
                  ;;
                *)         return 0;
            esac
        done
    fi
}
loadFlags


case $LENV in
    "prod")
        # the reason to use bulk is timeouts....
        export LIOAPI="https://bulk.lytics.io"
        ;;
    "staging")
        export LIOAPI="https://staging.lytics.io"
        ;;
    "pie")
        export LIOAPI="http://lio7:5353"
        ;;
    "local")
        export LIOAPI="http://localhost:5353"
        ;;
esac

# echo "case lenv=$LENV LIOAPI=$LIOAPI"
# export LIOAPI="http://lio7:5353"
# export LIOAPI="http://localhost:5353"
# export LIOAPI="http://bulk.lytics.io"

# Check if a list contains a given element. Param 1 is the needle, param 2 is the haystack.
# Example: containsElement "needle" "x" "y" "needle" "z"
# Taken from http://stackoverflow.com/a/8574392
containsElement () {
    local e
    for e in "${@:2}"; do [[ "$e" == "$1" ]] && return 0; done
    return 1
}

# First param is directory name that is being considered for upload, second param is a list
# of directories to upload, or "all" to upload all directories. Returns boolean whether
# the directory should be uploaded.
shouldUploadDir() {
    if [ "$2" = "all" ]; then
        return 0
    fi
    if containsElement $1 $2 ; then
        return 0
    fi
    return 1
}

# loadenv athletepath.com
# loadenv 1329
loadenv() {
    echo "#  Loading Env for $1"
    # curl -sL "$LIOAPI/api/account/$1/env?access_token=$LIOADMINKEY" > /tmp/lyticsenv && source /tmp/lyticsenv
    # eval $(curl -sL "$LIOAPI/api/account/$1/env?access_token=$LIOADMINKEY" | tail -5)
    # source /dev/stdin < ./settings
    # declare `curl -sL "$LIOAPI/api/account/$1/env?access_token=$LIOADMINKEY" | tail -5`
    #curl -sL "$LIOAPI/api/account/$1/env?access_token=$LIOADMINKEY"  > /tmp/lyticsenv && eval `cat /tmp/lyticsenv`
    curl -sL "$LIOAPI/api/account/$1/env?access_token=$LIOADMINKEY"  > /tmp/lyticsenv && source /tmp/lyticsenv
}

# showenv   prints out env we expect
showenv() {
    echo "LIOAPI      = $LIOAPI"
    echo "LIOADMINKEY = $LIOADMINKEY"
    echo "LIODOMAIN   = $LIODOMAIN"
    echo "LIOAID      = $LIOAID"
    echo "LIOKEY      = $LIOKEY"
    echo "LIOINTAID   = $LIOINTAID"
}

# assuming we have already been CD'd into appropriate directory
syncLql() {
    FILES=*.lql
    for f in $FILES
    do
      if [ "$f" != "lytics_scores.lql" ]; then
        pwd
        echo "uploading $f file..."
        http -v --check-status POST $LIOAPI/api/query key==$LIOKEY < $f > /dev/null || exitError "Could not use upload $f query"
      fi
    done
    # we have to do lytics scores last so that the by fields have been added first
    FILES=*scores.lql
    for f in $FILES
    do
      if [ "$f" = "lytics_scores.lql" ]; then
        echo "uploading $f file..."
        http -v --check-status POST $LIOAPI/api/query key==$LIOKEY < $f > /dev/null || exitError "Could not use upload $f query"
      fi
    done
    # the worlds worse race-preventer
    sleep 1
}
# The first/only param is a string error message. This is useful when running a program
# that prints error messages to stdout instead of stderr and stdout is /dev/null'ed.
exitError() {
    echo $1 > /dev/stderr
    exit 1
}

if [ "$LIOAPI" = "" ]; then
    echo "LIOAPI is not set, this is required"
    exit 1
fi

INITWD=`pwd`


case $CMD in
    "loadenv")
        #echo "load env: $2"
        loadenv $2
        showenv
        exit 0
        ;;
esac


use_web_user() {
  http -v --check-status POST $LIOAPI/api/query/web_default/use key==$LIOKEY  > /dev/null || exitError "Could not use web_default public query"
  http -v --check-status POST $LIOAPI/api/query/web_default_events/use key==$LIOKEY  > /dev/null || exitError "Could not use web_default_events public query"
  http -v --check-status POST $LIOAPI/api/query/web_default_identify/use key==$LIOKEY  > /dev/null || exitError "Could not use web_default_identify public query"
  http -v --check-status POST $LIOAPI/api/query/web_pathfora/use key==$LIOKEY  > /dev/null || exitError "Could not use web_pathfora public query"
}
usetwitter() {
  http -v --check-status POST $LIOAPI/api/query/user_twitter/use key==$LIOKEY > /dev/null || exitError "Could not use twitter user query"
}
useappreciationengine() {
  http -v --check-status POST $LIOAPI/api/query/user_appreciationengine_users/use key==$LIOKEY > /dev/null || exitError "Could not use appreciationengine user query"
}
usetwitter_leadcard() {
  http -v --check-status POST $LIOAPI/api/query/user_twleadcard/use key==$LIOKEY > /dev/null || exitError "Could not use twitter lead card query"
}
usesendgrid() {
    http -v --check-status POST $LIOAPI/api/query/user_sendgrid_subscribers/use key==$LIOKEY > /dev/null || exitError "Could not use user_sendgrid_subscribers query"
    http -v --check-status POST $LIOAPI/api/query/user_sendgrid/use key==$LIOKEY > /dev/null || exitError "Could not use user_sendgrid query"
}
usemailgun() {
    http -v --check-status POST $LIOAPI/api/query/user_mailgun/use key==$LIOKEY > /dev/null || exitError "Could not use user_mailgun query"
}
userapleaf() {
    echo "Use Rapleaf query"
    http -v --check-status POST $LIOAPI/api/query/user_rapleaf/use key==$LIOKEY > /dev/null || exitError "Could not use user_rapleaf query"
}
usefullcontact() {
    http -v --check-status POST $LIOAPI/api/query/user_fullcontact/use key==$LIOKEY > /dev/null || exitError "Could not use user_fullcontact query"
}
usemailchimp() {
    http -v --check-status POST $LIOAPI/api/query/user_mailchimp_activity/use key==$LIOKEY > /dev/null || exitError "Could not use user_mailchimp_activity query"
    http -v --check-status POST $LIOAPI/api/query/user_mailchimp_subscribers/use key==$LIOKEY > /dev/null || exitError "Could not use user_mailchimp_subscribers query"
}
usecampaignmonitor() {
    http -v --check-status POST $LIOAPI/api/query/user_campaignmonitor_activity/use key==$LIOKEY > /dev/null || exitError "Could not use user_campaignmonitor_activity query"
    http -v --check-status POST $LIOAPI/api/query/user_campaignmonitor_users/use key==$LIOKEY > /dev/null || exitError "Could not use user_campaignmonitor_users query"
}
useexacttarget() {
    http -v --check-status POST $LIOAPI/api/query/user_exacttarget_events/use key==$LIOKEY > /dev/null || exitError "Could not use user_exacttarget_events query"
    http -v --check-status POST $LIOAPI/api/query/user_exacttarget_subscribers/use key==$LIOKEY > /dev/null || exitError "Could not use user_exacttarget_subscribers query"
}
useklout() {
    http -v --check-status POST $LIOAPI/api/query/user_klout/use key==$LIOKEY > /dev/null || exitError "Could not use user_klout query"
}
usemarketo() {
    http -v --check-status POST $LIOAPI/api/query/user_marketo_activity/use key==$LIOKEY > /dev/null || exitError "Could not use user_marketo_activity query"
    http -v --check-status POST $LIOAPI/api/query/user_marketo_leads/use key==$LIOKEY > /dev/null || exitError "Could not use user_marekto_leads query"
}
usemixpanel() {
    http -v --check-status POST $LIOAPI/api/query/user_mixpanel_profiles/use key==$LIOKEY > /dev/null || exitError "Could not use user_mixpanel_profile query"
    http -v --check-status POST $LIOAPI/api/query/user_mixpanel_events/use key==$LIOKEY > /dev/null || exitError "Could not use user_mixpanel_events query"
}
uselinkedin() {
    http -v --check-status POST $LIOAPI/api/query/user_linkedin_data/use key==$LIOKEY > /dev/null || exitError "Could not use user_linkedin_data query"
}
uselyris() {
    http -v --check-status POST $LIOAPI/api/query/user_lyris_NAMEME/use key==$LIOKEY > /dev/null || exitError "Could not use user_lyris_NAMEME query"
}
usemaropost() {
    http -v --check-status POST $LIOAPI/api/query/user_maropost_users/use key==$LIOKEY > /dev/null || exitError "Could not use user_maropost_users query"
    http -v --check-status POST $LIOAPI/api/query/user_maropost_activity/use key==$LIOKEY > /dev/null || exitError "Could not use user_maropost_activity query"
}
userpardot() {
    http -v --check-status POST $LIOAPI/api/query/user_pardot_visitors/use key==$LIOKEY > /dev/null || exitError "Could not use user_pardot_visitors query"
    http -v --check-status POST $LIOAPI/api/query/user_pardot_visitor_activity/use key==$LIOKEY > /dev/null || exitError "Could not use user_pardot_visitor_activity query"
    http -v --check-status POST $LIOAPI/api/query/user_pardot_prospects/use key==$LIOKEY > /dev/null || exitError "Could not use user_pardot_prospects query"
}
usebluehornet() {
    http -v --check-status POST $LIOAPI/api/query/user_bluehornet_activity/use key==$LIOKEY > /dev/null || exitError "Could not use user_bluehornet_activity query"
    http -v --check-status POST $LIOAPI/api/query/user_bluehornet_subscribers/use key==$LIOKEY > /dev/null || exitError "Could not use user_bluehornet_subscribers query"
}
usezendesk() {
    http -v --check-status POST $LIOAPI/api/query/user_zendesktickets/use key==$LIOKEY > /dev/null || exitError "Could not use user_zendesktickets query"
    http -v --check-status POST $LIOAPI/api/query/user_zendeskusers/use key==$LIOKEY > /dev/null || exitError "Could not use user_zendeskusers query"
}
useretailnext() {
    http -v --check-status POST $LIOAPI/api/query/retailnext_events/use key==$LIOKEY > /dev/null || exitError "Could not use retailnext_events query"
    http -v --check-status POST $LIOAPI/api/query/retailnext_users/use key==$LIOKEY > /dev/null || exitError "Could not use retailnext_users query"
}
usesailthru() {
    http -v --check-status POST $LIOAPI/api/query/user_sailthru_users/use key==$LIOKEY > /dev/null || exitError "Could not use user_sailthru_users query"
}
usesalesforce() {
    http -v --check-status POST $LIOAPI/api/query/user_salesforce_users/use key==$LIOKEY > /dev/null || exitError "Could not use user_salesforce_users query"
}
usejanrain() {
    http -v --check-status POST $LIOAPI/api/query/user_janrain_users/use key==$LIOKEY > /dev/null || exitError "Could not use janrain_users query"
}
useswrve() {
    http -v --check-status POST $LIOAPI/api/query/user_swrve_users/use key==$LIOKEY > /dev/null || exitError "Could not use user_swrve_users query"
    http -v --check-status POST $LIOAPI/api/query/user_swrve_activty/use key==$LIOKEY > /dev/null || exitError "Could not use user_swrve_activty query"
}
useresponsys() {
    http -v --check-status POST $LIOAPI/api/query/user_responsys_activity/use key==$LIOKEY > /dev/null || exitError "Could not use user_responsys_activity query"
}
useclearbit() {
    http -v --check-status POST $LIOAPI/api/query/user_clearbit_users/use key==$LIOKEY > /dev/null || exitError "Could not use user_clearbit_users query"
}
usesendinblue() {
    http -v --check-status POST $LIOAPI/api/query/user_sendinblue_users/use key==$LIOKEY > /dev/null || exitError "Could not use user_sendinblue_users query"
    http -v --check-status POST $LIOAPI/api/query/user_sendinblue_activity/use key==$LIOKEY > /dev/null || exitError "Could not use user_sendinblue_activity query"
}
usesilverpop() {
    http -v --check-status POST $LIOAPI/api/query/user_silverpop_users/use key==$LIOKEY > /dev/null || exitError "Could not use user_silverpop_users query"
    http -v --check-status POST $LIOAPI/api/query/user_silverpop_activity/use key==$LIOKEY > /dev/null || exitError "Could not use user_silverpop_activity query"
}
usezuora() {
    http -v --check-status POST $LIOAPI/api/query/user_zuora_users/use key==$LIOKEY > /dev/null || exitError "Could not use user_zuora_users query"
    http -v --check-status POST $LIOAPI/api/query/user_zuora_activity/use key==$LIOKEY > /dev/null || exitError "Could not use user_zuora_activity query"
}


if shouldUploadDir public $1; then
    cd $INITWD/public
    echo `pwd`
    source .lytics
    echo LIOKEY=$LIOKEY public

    # NodeStore/Content Enrichment
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < content.lql > /dev/null || exitError "Could not upload content.lql public query"

    # SegmentML
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < segmentml.lql > /dev/null || exitError "Could not updated segmentml.lql public query"

    # appreciation engine
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < appreciationengine_users.lql > /dev/null || exitError "Could not upload appreciationengine_users.lql public query"

    # bluehornet
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < bluehornet_activity.lql  > /dev/null || exitError "Could not upload bluehornet_activity.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < bluehornet_subscribers.lql  > /dev/null || exitError "Could not upload bluehornet_subscribers.lql public query"

    # campaignmonitor
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < campaignmonitor_activity.lql  > /dev/null || exitError "Could not upload campaignmonitor_activity.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < campaignmonitor_users.lql  > /dev/null || exitError "Could not upload campaignmonitor_users.lql public query"

    # clearbit
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < clearbit_users.lql  > /dev/null || exitError "Could not upload clearbit_users.lql public query"

    # customerio
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < customerio_events.lql  > /dev/null || exitError "Could not upload customerio_events.lql public query"

    # exacttarget
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < exacttarget_events.lql  > /dev/null || exitError "Could not upload exacttarget_events.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < exacttarget_subscribers.lql  > /dev/null || exitError "Could not upload exacttarget_subscribers.lql public query"

    # fullcontact
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < fullcontact.lql  > /dev/null || exitError "Could not upload fullcontact.lql public query"

    # gigya
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < gigya_users.lql > /dev/null || exitError "Could not upload gigya_users.lql public query"

    # icontact
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < icontact_users.lql > /dev/null || exitError "Could not upload icontact_users.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < icontact_activity.lql > /dev/null || exitError "Could not upload icontact_activity.lql public query"

    # janrain
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < janrain_users.lql > /dev/null || exitError "Could not upload janrain_users.lql public query"

    # lyris
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < lyris_import.lql  > /dev/null || exitError "Could not upload lyris_import.lql public query"

    # lytics web default

    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < web_default.lql  > /dev/null || exitError "Could not upload web_default public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < web_default_events.lql  > /dev/null || exitError "Could not upload web_default_events public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < web_default_identify.lql  > /dev/null || exitError "Could not upload web_default_identify public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < web_pathfora.lql  > /dev/null || exitError "Could not use upload public query"

    # mailchimp
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < mailchimp_activity.lql  > /dev/null || exitError "Could not upload mailchimp_events.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < mailchimp_subscribers.lql  > /dev/null || exitError "Could not upload mailchimp_subscribers.lql public query"

    # mandrill
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < mandrill_events.lql  > /dev/null || exitError "Could not upload mandrill_events.lql public query"

    # mailgun
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < mailgun_events.lql  > /dev/null || exitError "Could not upload mailgun_events.lql public query"

    # maropost
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < maropost_activity.lql  > /dev/null || exitError "Could not upload maropost_events.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < maropost_users.lql  > /dev/null || exitError "Could not upload maropost_users.lql public query"

    # mixpanel
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < mixpanel_profiles.lql  > /dev/null || exitError "Could not upload mixpanel_profiles.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < mixpanel_events.lql  > /dev/null || exitError "Could not upload mixpanel_events.lql public query"

    # pardot
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < pardot_prospects.lql  > /dev/null || exitError "Could not upload pardot_prospects.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < pardot_visitor_activity.lql  > /dev/null || exitError "Could not upload pardot_visitor_activity.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < pardot_visitors.lql  > /dev/null || exitError "Could not upload pardot_visitors.lql public query"

    # marketo
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < marketo_leads.lql > /dev/null || exitError "Could not upload marketo_leads.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < marketo_activity.lql > /dev/null || exitError "Could not upload marketo_activity.lql public query"

    # rapleaf/towerdata
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < rapleaf.lql  > /dev/null || exitError "Could not upload rapleaf.lql public query"

    # responsys
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < responsys_events.lql  > /dev/null || exitError "Could not upload responsys_events.lql public query"

    # retailnext
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < retailnext_users.lql  > /dev/null || exitError "Could not upload retailnext_users.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < retailnext_events.lql  > /dev/null || exitError "Could not upload retailnext_events.lql public query"

    # sailthru
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < sailthru_users.lql  > /dev/null || exitError "Could not upload sailthru_users.lql public query"

    # salesforce
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < salesforce_users.lql  > /dev/null || exitError "Could not upload salesforce_users.lql public query"

    # segment.io/segment.com
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < segment_legacy.lql  > /dev/null || exitError "Could not upload segment_legacy.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < segment_users.lql  > /dev/null || exitError "Could not upload segment_users.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < segment_events.lql  > /dev/null || exitError "Could not upload segment_events.lql public query"

    # sendgrid
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < sendgrid_subscribers.lql > /dev/null || exitError "Could not upload sendgrid_subscribers.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < sendgrid_events.lql > /dev/null || exitError "Could not upload sendgrid_events.lql public query"

    # sendinblue
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < sendinblue_activity.lql  > /dev/null || exitError "Could not upload sendinblue_activity.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < sendinblue_users.lql  > /dev/null || exitError "Could not upload sendinblue_users.lql public query"

    # silverpop
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < silverpop_users.lql  > /dev/null || exitError "Could not upload silverpop_users.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < silverpop_activity.lql  > /dev/null || exitError "Could not upload silverpop_activity.lql public query"

    # sparkpost
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < sparkpost_activity.lql > /dev/null || exitError "Could not upload sparkpost_activity.lql public query"

    # swrve
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < swrve_users.lql > /dev/null || exitError "Could not upload swrve_users.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < swrve_activity.lql > /dev/null || exitError "Could not upload swrve_activty.lql public query"

    # twitter
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < twitter_stream.lql > /dev/null || exitError "Could not upload twitter_stream.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < twitter_leadcard.lql > /dev/null || exitError "Could not upload twitter_leadcard.lql public query"

    # urban-airship
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < ua_events.lql > /dev/null || exitError "Could not upload ua_events.lql public query"

    # urban airship
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < ua_events.lql  > /dev/null || exitError "Could not upload ua_events.lql public query"

    # zendesk
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < zendesk_tickets.lql  > /dev/null || exitError "Could not upload zendesk_tickets.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < zendesk_users.lql  > /dev/null || exitError "Could not upload zendesk_users.lql public query"

    # zuora
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < zuora_contacts.lql  > /dev/null || exitError "Could not upload zuora_contacts.lql public query"
    http --check-status $LIOAPI/api/query key==$LIOKEY disabled==true share_mode==public < zuora_activity.lql  > /dev/null || exitError "Could not upload zuora_activity.lql public query"

fi

if shouldUploadDir vadio $1; then
    cd $INITWD/vadio && loadenv 1246 && showenv
    syncLql
fi

if shouldUploadDir optimizely $1; then
    cd $INITWD/optimizely && loadenv 1397 && showenv
    syncLql
fi


if shouldUploadDir campaignmonitor $1; then
    cd $INITWD/campaignmonitor
    echo "not implemented"
    exit 1
    #syncLql
fi

if shouldUploadDir athletepath $1; then
    loadenv 10
    showenv
    cd $INITWD/athletepath
    syncLql
fi

if shouldUploadDir lytics $1; then
    loadenv getlytics.com
    showenv
    cd $INITWD/lytics
    syncLql
    # now for dev 1286 account
    if containsElement "api.lytics.io" $LIOAPI ; then
        #source lyticsdev/.lytics
        #lytics query sync . || exitError "lytics upload failed"
        return 0
    fi
fi

if shouldUploadDir 132 $1; then
    cd $INITWD/lytics
    loadenv 132 && showenv
    syncLql
fi


if shouldUploadDir directv $1; then
    cd $INITWD/directv && loadenv 1267 && showenv
    syncLql
fi
if shouldUploadDir directvdev $1; then
    cd $INITWD/directv && loadenv 1296 && showenv
    syncLql
fi

if shouldUploadDir johndao $1; then
    cd $INITWD/johndao
    source .lytics
    syncLql
fi

if shouldUploadDir purina $1; then
    cd $INITWD/purina/beyond-1430 && loadenv 1430 && showenv
    syncLql
    cd $INITWD/purina/justright-1432 && loadenv 1432 && showenv
    syncLql
    cd $INITWD/purina/proplan-1433 && loadenv 1433 && showenv
    syncLql
    cd $INITWD/purina/rollup-1421 && loadenv 1421 && showenv
    syncLql
    cd $INITWD/purina/rollupdev-1836 && loadenv 1836 && showenv
    syncLql
    cd $INITWD/purina/purinaone-1431 && loadenv 1431 && showenv
    syncLql
fi

if shouldUploadDir 1430 $1; then
    cd $INITWD/purina/beyond-1430  && loadenv 1430 && showenv
    syncLql
fi
if shouldUploadDir 1421 $1; then
    cd $INITWD/purina/rollup-1421  && loadenv 1421 && showenv
    syncLql
fi
if shouldUploadDir 1431 $1; then
    cd $INITWD/purina/purinaone-1431  && loadenv 1431 && showenv
    syncLql
fi
if shouldUploadDir 1433 $1; then
    cd $INITWD/purina/proplan-1433  && loadenv 1433 && showenv
    syncLql
fi
if shouldUploadDir 1836 $1; then
    cd $INITWD/purina/rollupdev-1836  && loadenv 1836 && showenv
    syncLql
fi

if shouldUploadDir theclymb $1; then
    cd $INITWD/directv && loadenv 1329 && showenv
    syncLql
fi

if shouldUploadDir 1327 $1; then
    cd $INITWD/unigo && loadenv 1327 && showenv
    syncLql
fi

if shouldUploadDir unigo $1; then
    cd $INITWD/unigo && loadenv 1404 && showenv
    syncLql
fi

if shouldUploadDir teamroboboogie.com $1; then
    cd $INITWD/teamroboboogie.com
    source .lytics
    syncLql
    use_web_user
fi

if shouldUploadDir daveweis $1; then
    cd $INITWD/daveweis
    source .lytics
    syncLql
    #use_web_user
fi

if shouldUploadDir robshields $1; then
    cd $INITWD/robshields
    source .lytics
    #syncLql
    use_web_user
    cd $INITWD/robshields/1357
    source .lytics
    #syncLql
    use_web_user
fi

## Access Intelligence
if shouldUploadDir powermag $1; then
    cd $INITWD/accessintelligence/powermag && loadenv 1340 && showenv
    syncLql
fi
if shouldUploadDir oilcomm $1; then
    cd $INITWD/accessintelligence/oilcomm  && loadenv 1346  && showenv
    syncLql
fi
if shouldUploadDir studiodaily $1; then
    cd $INITWD/accessintelligence/studiodaily && loadenv 1361 && showenv
    syncLql
fi
if shouldUploadDir cablefax $1; then
    cd $INITWD/accessintelligence/cablefax && loadenv 1364 && showenv
    syncLql
fi
if shouldUploadDir eventmarketer $1; then
    cd $INITWD/accessintelligence/eventmarketer && loadenv 1365 && showenv
    syncLql
fi
if shouldUploadDir satellitetoday $1; then
    cd $INITWD/accessintelligence/satellitetoday && loadenv 1366 && showenv
    syncLql
fi
if shouldUploadDir aviationtoday $1; then
    cd $INITWD/accessintelligence/aviationtoday && loadenv 1367 && showenv
    syncLql
fi
if shouldUploadDir 1367 $1; then
    cd $INITWD/accessintelligence/aviationtoday && loadenv 1367 && showenv
    syncLql
fi
if shouldUploadDir chemeng $1; then
    cd $INITWD/accessintelligence/chemeng  && loadenv 1371  && showenv
    syncLql
fi
if shouldUploadDir chiefmarketer $1; then
    cd $INITWD/accessintelligence/chiefmarketer  && loadenv 1372  && showenv
    syncLql
fi
if shouldUploadDir cynopsis $1; then
    cd $INITWD/accessintelligence/cynopsis  && loadenv 1373  && showenv
    syncLql
fi
if shouldUploadDir defensedaily $1; then
    cd $INITWD/accessintelligence/defensedaily  && loadenv 1374  && showenv
    syncLql
fi
if shouldUploadDir edpa $1; then
    cd $INITWD/accessintelligence/edpa  && loadenv 1375  && showenv
    syncLql
fi
if shouldUploadDir exchangemonitor $1; then
    cd $INITWD/accessintelligence/exchangemonitor  && loadenv 1916  && showenv
    syncLql
fi
if shouldUploadDir 1916 $1; then
    cd $INITWD/accessintelligence/exchangemonitor  && loadenv 1916 && showenv
    syncLql
fi
if shouldUploadDir folio $1; then
    cd $INITWD/accessintelligence/folio  && loadenv 1376  && showenv
    syncLql
fi
if shouldUploadDir 1376 $1; then
    cd $INITWD/accessintelligence/folio  && loadenv 1376  && showenv
    syncLql
fi
if shouldUploadDir min $1; then
    cd $INITWD/accessintelligence/min  && loadenv 1378  && showenv
    syncLql
fi
if shouldUploadDir prnews $1; then
    cd $INITWD/accessintelligence/prnews  && loadenv 1379  && showenv
    syncLql
fi
if shouldUploadDir ormanager $1; then
    cd $INITWD/accessintelligence/ormanager  && loadenv 1390  && showenv
    syncLql
fi
if shouldUploadDir leadscon $1; then
    cd $INITWD/accessintelligence/leadscon  && loadenv 1347  && showenv
    syncLql
fi
if shouldUploadDir aisandbox $1; then
    cd $INITWD/accessintelligence/sandbox  && loadenv 1711  && showenv
    syncLql
fi
if shouldUploadDir mcm $1; then
    cd $INITWD/accessintelligence/mcm  && loadenv 1377  && showenv
    syncLql
fi
if shouldUploadDir admonsters $1; then
    cd $INITWD/accessintelligence/admonsters  && loadenv 1841 && showenv
    syncLql
fi


if shouldUploadDir sauceyapp $1; then
    cd $INITWD/sauceyapp  && loadenv 1393  && showenv
    syncLql
fi

if shouldUploadDir siliconflorist $1; then
    cd $INITWD/siliconflorist  && loadenv 1415  && showenv
    syncLql
fi

if shouldUploadDir hylete $1; then
    cd $INITWD/hylete && loadenv 1414  && showenv
    syncLql
fi

if shouldUploadDir coldandsharky $1; then
    cd $INITWD/coldandsharky  && loadenv 1413  && showenv
    syncLql
fi

if shouldUploadDir betabrand $1; then
    cd $INITWD/betabrand && loadenv 1418 && showenv
    syncLql
fi

if shouldUploadDir rga $1; then
    cd $INITWD/rga  && loadenv 1417  && showenv
    syncLql
fi

if shouldUploadDir westwardleaning $1; then
    cd $INITWD/westwardleaning  && loadenv 1423  && showenv
    syncLql
fi

if shouldUploadDir outdoorproject $1; then
    cd $INITWD/outdoorproject  && loadenv 1424  && showenv
    syncLql
fi

if shouldUploadDir krrb $1; then
    cd $INITWD/krrb && loadenv 1474 && showenv
    syncLql
fi

if shouldUploadDir westfield $1; then
    cd $INITWD/westfield  && loadenv 1405  && showenv
    syncLql
fi

if shouldUploadDir westfielddev $1; then
    cd $INITWD/westfield  && loadenv 1483  && showenv
    syncLql
fi

if shouldUploadDir freshpet $1; then
    cd $INITWD/freshpet  && loadenv 1422  && showenv
    syncLql
fi

if shouldUploadDir infinitywireless $1; then
    cd $INITWD/infinitywireless  && loadenv 1494  && showenv
    syncLql
fi

if shouldUploadDir safetysmart $1; then
    cd $INITWD/safetysmart  && loadenv 1452  && showenv
    syncLql
fi

if shouldUploadDir dknewmedia $1; then
    cd $INITWD/dknewmedia  && loadenv 1500  && showenv
    syncLql
fi

if shouldUploadDir wemeandigital $1; then
    cd $INITWD/wemeandigital  && loadenv 1520  && showenv
    syncLql
fi

if shouldUploadDir bludot $1; then
    cd $INITWD/bludot  && loadenv 1434  && showenv
    syncLql
fi

if shouldUploadDir religionnews $1; then
    cd $INITWD/religionnews  && loadenv 1537  && showenv
    syncLql
    loadenv 1799  && showenv
    syncLql
fi

if shouldUploadDir bungaloow $1; then
    cd $INITWD/bungaloow  && loadenv 1557 && showenv
    syncLql
fi

if shouldUploadDir panjo $1; then
    # cd $INITWD/panjo  && loadenv 1559 && showenv
    # syncLql
    cd $INITWD/panjo  && loadenv 1843 && showenv
    syncLql
fi

if shouldUploadDir gatheredtable $1; then
    cd $INITWD/gatheredtable  && loadenv 1572 && showenv
    syncLql
fi

if shouldUploadDir 77west $1; then
    cd $INITWD/77west  && loadenv 1583 && showenv
    syncLql
fi

if shouldUploadDir sandboxadvisors $1; then
    cd $INITWD/sandboxadvisors  && loadenv 1584 && showenv
    syncLql
fi

if shouldUploadDir freshnation $1; then
    cd $INITWD/freshnation  && loadenv 1635 && showenv
    syncLql
fi

if shouldUploadDir boostedboards $1; then
    cd $INITWD/boostedboards  && loadenv 1638 && showenv
    syncLql
fi

if shouldUploadDir lifelock $1; then
    cd $INITWD/1690  && loadenv 1690 && showenv
    syncLql
fi
if shouldUploadDir 1690 $1; then
    cd $INITWD/1690  && loadenv 1690 && showenv
    syncLql
fi

if shouldUploadDir dotandbo $1; then
    cd $INITWD/dotandbo  && loadenv 1736 && showenv
    syncLql
fi

if shouldUploadDir pathfora $1; then
    cd $INITWD/pathfora  && loadenv 1762 && showenv
    syncLql
fi

if shouldUploadDir 1834 $1; then
    cd $INITWD/1834  && loadenv 1834 && showenv
    syncLql
fi

if shouldUploadDir 1917 $1; then
    cd $INITWD/1917  && loadenv 1917 && showenv
    syncLql
fi

if shouldUploadDir 1931 $1; then
    cd $INITWD/1931  && loadenv 1931 && showenv
    syncLql
fi

if shouldUploadDir 1935 $1; then
    cd $INITWD/1935  && loadenv 1935 && showenv
    syncLql
fi

if shouldUploadDir 2000 $1; then
    cd $INITWD/2000  && loadenv 2000 && showenv
    syncLql
fi

if shouldUploadDir 2002 $1; then
    cd $INITWD/2002  && loadenv 2002 && showenv
    syncLql
fi

if shouldUploadDir 2004 $1; then
    cd $INITWD/2004  && loadenv 2004 && showenv
    syncLql
fi

if shouldUploadDir 2005 $1; then
    cd $INITWD/2005  && loadenv 2005 && showenv
    syncLql
fi

if shouldUploadDir 2011 $1; then
    cd $INITWD/2011  && loadenv 2011 && showenv
    syncLql
fi

if shouldUploadDir 2133 $1; then
    cd $INITWD/2133 && loadenv 2133 && showenv
    syncLql
fi

if shouldUploadDir 2140 $1; then
    cd $INITWD/2140  && loadenv 2140 && showenv
    syncLql
fi

if shouldUploadDir cbsi $1; then
    cd $INITWD/cbsi
    source .lytics
    echo LIOKEY=$LIOKEY cbsi
    lytics query sync . || exitError "cbsi query upload failed"
    usebluehornet
fi

if shouldUploadDir racingpost $1; then
    cd $INITWD/racingpost  && loadenv 1356 && showenv
    syncLql
    loadenv 1401 && showenv
    syncLql
fi

if shouldUploadDir segmentio $1; then
    cd $INITWD/segmentio
    source .lytics
    echo LIOKEY=$LIOKEY segmentio
    #lytics query sync . || exitError "segmentio upload failed"
    #usesendgrid
    #usetwitter
    use_web_user
fi
