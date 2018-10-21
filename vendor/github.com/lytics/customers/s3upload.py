#!/usr/bin/python
"""
./s3upload.py mindjet text/csv name.csv 

curl -s https://gist.github.com/araddon/5548285/raw/echo.sh | sh


"""
import os, sys
from datetime import datetime, timedelta
import base64
import hmac, hashlib

args = [arg for arg in sys.argv[1:] if arg not in ["--help","help"]]
AWS_ACCESS_KEY = os.environ['AWS_ACCESS_KEY']
AWS_SECRET_KEY = os.environ['AWS_SECRET_KEY']

print "args = %s" % (args)
file_path = "mindjet"
file_name = "myfile.csv"
mime_type = "text/csv"
if len(args) < 3:
	print("must have   customer  mimetype filename.csv")
	sys.exit(1)
if len(AWS_SECRET_KEY) < 2 || len(AWS_ACCESS_KEY) < 2:
	print("must have env variables:   AWS_ACCESS_KEY, AWS_SECRET_KEY")
	sys.exit(1)

file_path, mime_type, file_name = args[0] + "/", args[1], args[2]

# http://aws.amazon.com/articles/1434?_encoding=UTF8&jiveRedirect=1
# http://raamdev.com/2008/using-curl-to-upload-files-via-post-to-amazon-s3/
td = timedelta(days=14)
expire_data = datetime.now() + td
expires = expire_data.strftime('%Y-%m-%dT%H:00:00Z')

#"2009-01-01T00:00:00Z"
policy_document = """{"expiration": "%s",
  "conditions": [ 
    {"bucket": "lytics-uploads"}, 
    ["starts-with", "$key", "%s"],
    {"acl": "private"},
    ["starts-with", "$Content-Type", "%s"],
  ]
}""" % (expires, file_path, mime_type)

policy = base64.b64encode(policy_document)

signature = base64.b64encode(hmac.new(AWS_SECRET_KEY, policy, hashlib.sha1).digest())

# print policy_document
# print expires
# print(policy)
# print(signature)

print("""
# run this curl command
curl \\
	-F "key=%s%s" \\
	-F "acl=private" \\
	-F "AWSAccessKeyId=15SYRRBKY62WCAHVGXR2" \\
	-F "policy=%s" \\
	-F "signature=%s" \\
	-F "Content-Type=%s" \\
	-F "file=@yourfile.csv" \\
https://lytics-uploads.s3.amazonaws.com
""" % (file_path,file_name, policy,signature,mime_type)
)