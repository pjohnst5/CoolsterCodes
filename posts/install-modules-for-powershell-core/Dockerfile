# Use a readily available powershell core docker image by Microsoft as base
FROM mcr.microsoft.com/powershell:lts-nanoserver-1809

# Run as admin
USER ContainerAdministrator

# Make default shell powershell
SHELL ["pwsh", "-command"]

# Label powershell gallery as trusted
RUN Set-PSRepository -Name PSGallery -InstallationPolicy Trusted

# Add ContainerAdministrator's powershell module location to PSModulePath
RUN setx /M PSModulePath $($env:PSModulePath + 'C:\Users\ContainerAdministrator\Documents\PowerShell\Modules')

# Install powershell core dns module DnsClient-PS
RUN Install-Module -Name DnsClient-PS -SkipPublisherCheck -RequiredVersion 1.0.0

# Run a simple command on startup
ENTRYPOINT [ "Resolve-DNS", "www.microsoft.com" ]
