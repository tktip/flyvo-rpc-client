#TIP Flyvo RPC Client

Welcome to the TIP Flyvo-RPC Client! We hope you'll have a pleasant experience. It's very much a WIP, and things are subject to change. However, it should provide something to base initial work on @Visma's side.
Log output is written to the specified logFile, as well as to the windows service event log.

###Configuration

Below is an explanation of the config. The service expects an yml config that follows the below structure. The values are just examples.
```yaml
api:
  port: 8080 (rpc client http port, 8080)
  rpc: (details used by rpc client)
    serverAddress: "www.fake.com:50051" (rpc server address:port)
    certFile: Z:\server.crt (certificate file - public key, drop to not use TLS)
    pollFrequency: 100ms (sleeptime between checking for incoming requests*, should be low) #how often you should poll TIP for new requests (should be low)
    connFailSleep: 5s (sleeptime on failed connection)
    flyvo: (details on the Flyvo api)
      address: http://localhost:8081
      pathConvertAbsence: /flyvo/convertAbsence (flyvo endpoint for converting absence to sick leave)**
      pathGetCourses: /flyvo/getcourses (endpoint for getting courses for a teacher)
      pathRegisterAbsence: /flyvo/absence/register (endpoint for registering absence)
      pathGetUnauthorizedAbsences: /flyvo/absences (endpoint for getting absences)
      pathRegisterSickLeave: /flyvo/sickleave/register (endpoint for registering absences)
      pathGetSickLeaves: /flyvo/sickleave (endpoint for getting absences)
logFile: output.txt (specify a log output file)
logLevel: debug (Log level, one of [debug, info, warn, error, fatal]. Debug is very noisy as it outputs on server polling)

# *sleeptime indicates how long the client should wait before polling. 
#  On the TIP side, a user will perform a request, which the server will store 
#  until the rpc client grabs it. The rpc client responds to the request, 
#  which the server then passes on to the end user. As such, it should be low (e.g. 100ms).
#
#**How this will be performed is as of yet undefined.
```
###Running
To run, type `.\flyvo-rpc-client.exe configFile=[CFG].yml`

Example: `flyvo-rpc-client.exe configFile=Z:\config.yml`

###Starting a service
The writer of this document has only set this up on a normal Windows 10 VM instance, so this might not apply on a server. Who knows.

To create service:
 - 1: sc.exe create flyvo-client-test binPath="**exeLocation**\flyvo-rpc-client.exe configFile=**configFileFull**.yml. 
 - 2: sc.exe start flyvo-client-test

Example: `sc.exe create flyvo-client-test binPath=Z:\flyvo-rpc-client.exe configFile=Z:\config.yml`

To delete:
 - 1: sc.exe stop flyvo-client-test
 - 2: sc.exe delete flyvo-client-test
   

###Endpoints (to TIP)
Information about endpoints that are used to send requests to TIP can be found in the swagger.json documentation.
However, in short, there's 4 specific endpoints:
- **/generic \[post\]**:
  Sends a generic request (see swagger doc) with a path (not optional), headers, a body and a msg id (all optional). The path is essentially an endpoint specification.
- **/events \[POST/PATCH\]**:
  Accepts an event json (see swagger) as specified by Visma, and creates or updates it in TIP. Specific response data is as of yet not decided and is subject to change.
- **/events/:id \[DELETE\]**:
  Accepts an event id (path param) and deletes the specified event. Specific response data is as of yet not decided and is subject to change. 

###Endpoints (from TIP)
Regarding endpoints at Flyvo which the TIP RPC client will contact. All bodies will be in camelCase json, and contain the fields provided in the google docs file "[Datafelter API FlyVO/TIP](https://docs.google.com/document/d/1hZ6hT79Lmvknoh-5U-TKbzbOBwHKf4IcFlTAH1LJX3E/edit)". Below are all current request (By TIP) and response bodies (by Visma):
Note: Values that are enclosed in square brackets ([]) are expected to be arrays. 
Note: Values that are null have not been specified by Visma as of yet.

```json
{
    "GetCoursesRequest" : {
        "fromDate": "string",
        "toDate": "string"
    },

    "GetCoursesResponse":[{ 
        "vismaID": "string",
        "givenName": "string",
        "surName": "string",
        "courseID": "string",
        "calendarEventID": "string",
        "to": "string",
        "date": "string",
        "from": "string",
        "place": "string",
        "rom": "string"
    }],
"GetCoursesRequest" : {"fromDate": "2011-06-03T10:00:00-07:00","toDate": "2021-06-03T10:00:00-07:00"}
    "GetSickLeavesRequest": {
        "vismaId": "string",
        "toDate": "string"
    },
    
    "GetSickLeavesResponse": [{
        "vismaId": "string",
        "givenName": "string",
        "surname": "string",
        "sickLeaveCount": "int",
        "sickChildCount": "int"
    }],
    
    "GetUnauthorizedAbsencesRequest": {
        "vismaId": "string",
        "from": "string",
        "to": "string"
    },
    
    "GetUnauthorizedAbsencesResponse": [{
        "vismaId":"string",
        "givenName": "string",
        "surname": "string",
        "fromTime": "string",
        "toTime": "string",
        "date": "string"
    }],
    
    "RegisterAbsenceRequest": {
        "vismaId": "string",
        "absenceType": "string",
        "courseId": "string",
        "calendarEventId": "string",
        "date": "string",
        "hours": "string"
    },
    
    "RegisterAbsenceResponse": null,  
    
    "RegisterSickLeave": {
        "vismaId": "string",
        "code": "string",
        "fromDate": "string",
        "toDate": "string"
    },
    
    "RegisterSickLeaveResponse": null,
    
    "ConvertAbsenceToSickLeaveRequest": null,
    "ConvertAbsenceToSickLeaveResponse": null
}
``` 