# Install Modules for Powershell Core (pwsh)

See the full article here: [Install Modules for Powershell Core (pwsh)](https://coolstercodes.com/installing-modules-for-powershell-core-pwsh/)

## What is Powershell Core (pwsh) ?
Powershell Core is a cross-platform version of Powershell, meant to run on Linux, Windows and MacOS

## Install Powershell Core (pwsh)
We will use Microsoft's pre-build docker image `mcr.microsoft.com/powershell:lts-nanoserver-1809` which already has Powershell Core on it

## Browse Powershell Core Modules
Go to [Powershell Gallery](https://www.powershellgallery.com/) and find a module that you want
- Here we use the module `DnsClient-PS`

## Install a Module
`docker build -t dns-module:v1.0.0 -f Dockerfile .`