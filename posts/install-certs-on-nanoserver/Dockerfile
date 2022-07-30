# Use Nanoserver as a base image (with powershell installed)
FROM mcr.microsoft.com/powershell:lts-nanoserver-1809

# Run as admin
USER ContainerAdministrator

# Make default shell powershell
SHELL ["pwsh", "-command"]

# Make a directory to store copy of cert
WORKDIR C:\\certs

# Copy cert to container
COPY certsonnanoserverkv-dummycert-20220730.pfx .

# Copy cert installation script to container
COPY NanoserverInstallCert.ps1 .

# Run cert installation script in container
RUN .\NanoserverInstallCert.ps1 -PfxFile .\\certsonnanoserverkv-dummycert-20220730.pfx -StoreName """LocalMachine""" -StoreLoc """My"""

# Verify that the cert was installed on the container
RUN Get-Childitem Cert:\LocalMachine\My

# Remove copy of cert from container
RUN Remove-Item certsonnanoserverkv-dummycert-20220730.pfx
