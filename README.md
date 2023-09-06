# simple-web-apps

simple web apps for logging user activity & get the highest activity user and for censored valid sensitive information from user using Go Programming Language
<br>

## What to prepare ?
- Latest Go version (this apps using Go 1.18.1 for development)
- IDE (ex: VS Code, GoLand, etc.)
- Postman (for API testing)
<br>

## How to Run ?
1. Open your terminal (make sure you already installed Go)

2. Run this command on your terminal
```
go run main.go
```
3. Test the API/Endpoint using Postman with the following API Documentation
<br>

## API Documentation

### highest-activity Endpoint
- HTTP Method : 
```
POST
```
<br>

- URL :
```
localhost:8080/highest-activity
```
<br>

- Example Request Body :
```
[
    {
        "username": "Aiman",
        "action": "Select"
    },
    {
        "username": "Aiman",
        "action": "Search"
    },
    {
        "username": "Hisyam",
        "action": "Delete"
    }
]
```
<br>

- Example Success Response Body :
```
{
    "code": 200,
    "status": "OK",
    "data": "user Aiman with total records 2 action"
}
```

<br>

- Example Error Response Body :
```
{
    "code": 405,
    "status": "Method Not Allowed"
}
```
```
{
    "code": 400,
    "status": "Bad Request",
    "errors": [
        "json: cannot unmarshal object into Go value of type []main.UserActivity"
    ]
}
```
<br>

### user-info Endpoint
- HTTP Method : 
```
POST
```
<br>

- URL :
```
localhost:8080/user-info
```
<br>

- Example Request Body :
```
{
    "name":"Aiman Hisyam",
    "personal_email":"aiman_personal@gmail.com",
    "personal_number":"081345678932",
    "office_email":"aiman_officeyahoo",
    "office_Number":"+6281245631236"
}
```
<br>

- Example Success Response Body :
```
{
    "code": 200,
    "status": "OK",
    "data": {
        "name": "Aiman Hisyam",
        "personal_email": "*CENSORED*",
        "personal_number": "*CENSORED*",
        "office_email": "aiman_officeyahoo",
        "office_Number": "*CENSORED*"
    }
}
```
<br>

- Example Error Response Body :
```
{
    "code": 405,
    "status": "Method Not Allowed"
}
```
```
{
    "code": 400,
    "status": "Bad Request",
    "errors": [
        "json: cannot unmarshal array into Go value of type main.UserInfo"
    ]
}
```
