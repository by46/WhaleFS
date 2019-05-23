POST / HTTP/1.1
Content-Type: multipart/form-data; boundary=9431149156168 
Content-Length: length

--9431149156168
Content-Disposition: form-data; name="key"

/benjamin/hello.jpg
--9431149156168
Content-Disposition: form-data; name="file"; filename="MyFilename.jpg"
Content-Type: image/jpeg

file_content
--9431149156168

PUT /benjamin/hello.jpg HTTP/1.1
Content-Type: text/html
Content-Length: 129

file content which upload

GET /benjamin/hello.jpg HTTP/1.1

POST /ObjectName?uploads HTTP/1.1 
Date: date

POST /ObjectName?partNumber=PartNumber&uploadId=UploadId HTTP/1.1 
Date: Date
Content-Length: Size

{
  "partNumber":1,
  "etag":"11231231"
}

POST /ObjectName?uploadId=UploadId HTTP/1.1 
Date: Date
Content-Length: Size

[{
"partNumber":1,
"etag":"11231231"
},
{
  "partNumber":1,
  "etag":"11231231"
}]