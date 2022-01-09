# Docker Nanoserver base image Windows

See the full article here: [Docker Nanoserver base image Windows]()

## Create the .Net app
`dotnet new console -o App -n DotNet.Docker`

## Compile .Net app into binaries (from App directory)
`dotnet publish -c Release`

## Build docker image
`docker build -t dotnetexample:v1.0.0 -f .\Dockerfile .`

## Run docker image locally
`docker run dotnetexample:v1.0.0`