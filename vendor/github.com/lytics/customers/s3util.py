#!/usr/bin/python
"""
Lytics AWS Provisioning utilities

usage:

  s3util.py adduser 
"""
import os
import glob 
import sys
import argparse
import boto
 
partct=32
s3topic = "prod-repartition"
kafkadir = "/vol/kafkalogs"


# First create a connection to the IAM service
#  this uses keys in ~/.ec2/
#
iam = boto.connect_iam()
 
def create_group():
    # This is a group policy to allow members to use just one s3 bucket 
    group_policy = """{
      "Statement": [
        {
          "Sid": "AllowGroupToSeeBucketListAndAlsoAllowGetBucketLocationRequiredForListBucket",
          "Action": [ "s3:ListAllMyBuckets", "s3:GetBucketLocation" ],
          "Effect": "Allow",
          "Resource": [ "arn:aws:s3:::*"  ]
        },
        {
          "Sid": "AllowRootLevelListingOfCompanyBucket",
          "Action": ["s3:ListBucket"],
          "Effect": "Allow",
          "Resource": ["arn:aws:s3:::lyticspublic"],
          "Condition":{ 
                "StringEquals":{
                        "s3:prefix":[""], "s3:delimiter":["/"]
                               }
                     }
        }
      ]
    }  
    """
    response = iam.create_group('s3_lyticspublic')
    response = iam.put_group_policy('s3_lyticspublic', 's3_lyticspublic', group_policy)
    

# Now create a user and place him in the EC2 group.
def adduser(name):
    if name == "":
      print("needs user name")
      return
    response = iam.create_user(name)
    user = response.user
    response = iam.add_user_to_group('s3_lyticspublic', name)
     
    # Create AccessKey/SecretKey pair for user
    response = iam.create_access_key(name)
    access_key = response.access_key_id
    secret_key = response.secret_access_key
    print("Name: %s access_key: %s secret_key: %s" %(name,access_key, secret_key))
 
#
# create connection to EC2 as user Bob
#
#ec2 = boto.connect_ec2(access_key, secret_key)
 

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Manage Lytics AWS Provisioning')
    parser.add_argument('cmd', metavar='Command', type=str, nargs=1, default="down",
                       help='Name of command to run (adduser,addgroup,addserver)')
    #parser.add_argument('-x','--xxx', type=int, default=32, help='Number of Partitions')
    parser.add_argument('-u','--user', type=str, default="", help='Username to add')
    #parser.add_argument('-d','--kafkadir', type=str, default=kafkadir, help='Kafka directory')
    args = parser.parse_args()
    # s3topic = args.s3topic 
    # kafkadir = args.kafkadir 

    print "cmd=%s user=%s" % ( args.cmd[0], args.user )
    if args.cmd[0] == "adduser":
        adduser(args.user)

