
# flyvo-rpc-client

  

**Introduction**

Flyvo-rpc-client is the RPC client communicating with FlyVo calendar integration and the RPC server.
This will receive push events from visma and send them to RPC server, and it will receive requests from RPC server and retrieve data from the FlyVo calendar integration API.
  

**How to build**

  

1. Build as a runable file

  

We use make to build our projects. You can define what system to build for by configuring the GOOS environment variable.

  

  

>\> GOOS=windows make clean build

  

  

>\> GOOS=linux make clean build

  

  

These commands will build either a runable linux or windows file in the /bin/amd64 folder

  

  

2. Build a docker container

  

First you have to define the docker registry you are going to use in "envfile". Replace the REGISTRY variable with your selection.

  

Run the following command

  

  

>\> GOOS=linux make clean build container push

  

  

This will build the container and push it to your docker registry.

  

  

**How to run**

  

1. Executable

  

If you want to run it as a executable (Windows service etc) you will need to configure the correct environment variable. When you start the application set the CONFIG environment to **file::\<location\>** for linux or run it as a argument for windows

  

  

Windows example: **.\flyvo-rpc-client.exe configFile=Z:/folder/config.yml**

  

Linux example: **CONFIG=file::../folder/cfg.yml ./flyvo-rpc-client**

  

  

2. Docker

  

You have to first mount the cfg file into the docker container, and then set the config variable to point to that location before running the service/container

**Configuration file**

**api.port:** Exported API port. This is the API that FlyVo pushes events. The endpoints exposed are defined in internal/api/api.go

**api.timeout:** Request timeout

**api.rpc.connTimeout:** Connection timeout

**api.rpc.serverAddress:** Address and port to the RPC server.

**api.rpc.certFile:** Path to certificate if you want to encrypt the content that is sent over RPC. This needs to be the public file to the certificate you set up on the server

**api.rpc.pollFrequency:** How often it should check for new messages on the rpc server

**api.rpc.connFailSleep:** How long it should sleep if it loose connection to the RPC server before it tries to reconnect.

**api.rpc.flyvo.address:** Address and port to the FlyVo calendar integration API

**logFile:** Path to logfile(Windows)

**logLevel:** Lowest loglevel; debug, info, error, panic