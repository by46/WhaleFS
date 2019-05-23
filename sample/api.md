POST /ObjectName?uploads HTTP/1.1 
Date: date

POST /ObjectName?partNumber=PartNumber&uploadId=UploadId HTTP/1.1 
Date: Date
Content-Length: Size
Authorization: authorization string

{
  "partNumber":1,
  "etag":"11231231"
}

POST /ObjectName?uploadId=UploadId HTTP/1.1 
Date: Date
Content-Length: Size
Authorization: authorization string

[{
"partNumber":1,
"etag":"11231231"
},
{
  "partNumber":1,
  "etag":"11231231"
}]