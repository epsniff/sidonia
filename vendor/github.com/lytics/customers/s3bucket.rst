
http://docs.aws.amazon.com/AWSJavaScriptSDK/guide/browser-configuring.html

1.  Create user in AWS user
2.  Copy this policy (made here: http://awspolicygen.s3.amazonaws.com/policygen.html)
   but edit the "target"

{
  "Statement": [
    {
      "Sid": "Stmt1369171504227",
      "Action": [
        "s3:DeleteObject",
        "s3:GetObject",
        "s3:GetObjectTorrent",
        "s3:GetObjectVersion",
        "s3:ListBucket",
        "s3:PutObject"
      ],
      "Effect": "Allow",
      "Resource": [
      	"arn:aws:s3:::lytics-target/*",
      	"arn:aws:s3:::lytics-target"
      ],
      "Principal": {
        "AWS": [
          "arn:aws:iam::358991168639:user/edplus"
        ]
      }
    }
  ]
}





