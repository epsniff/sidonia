#!/usr/bin/env bash

# s3link.sh edplus/scholarshipexperts/ScholarshipExpertsBatchData.json

# http://cuz.cx/vince_riv/2013/04/create-signed-s3-url-using-s3cmd/
if [ $# -ne "1" ] ; then
  echo "not enough args"
  exit 1
fi

bucket="lytics-uploads"

PATHTOFILE=$1

echo " path/to/file=$PATHTOFILE"


# 1-day expiration
ts_exp=$((`date +%s` + 3600 * 24))
# string to sign: GET + expiration-time + bucket/object
can_string="GET\n\n\n$ts_exp\n/$bucket/$PATHTOFILE"
# generate the signature
sig=$(s3cmd sign "$(echo -e "$can_string")" | sed -n 's/^Signature: //p')
# extract access key from .s3cfg
s3_access_key=$(sed -n 's/^access_key = //p' ~/.s3cfg)

# sanity check
if [ -z "$s3_access_key" -o -z "$sig" ]; then
  echo "Failed to created signed URL for s3://$PATHTOFILE" >&2
  exit 1
fi

value="$(perl -MURI::Escape -e 'print uri_escape($ARGV[0]);' "$sig")"
base_url="https://s3.amazonaws.com/$bucket/$PATHTOFILE"
params="AWSAccessKeyId=$s3_access_key&Expires=$ts_exp&Signature=$value"
echo "$base_url?$params"