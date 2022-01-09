FROM mcr.microsoft.com/powershell:lts-nanoserver-1909

# Make default shell powershell
SHELL ["pwsh", "-command"]

# Dotnet 6.0.101 install
WORKDIR Users\\Example\\dotnetinstall\\6.0.101
RUN Invoke-WebRequest -OutFile dotnet-install.ps1 -URI https://dotnet.microsoft.com/download/dotnet/scripts/v1/dotnet-install.ps1
RUN .\dotnet-install.ps1 -Version 6.0.101 -InstallDir """C:\\Users\\Example\\\\dotnetinstall\\6.0.101"""

# Copy the application binaries into the docker image
COPY App/bin/Release/net6.0/publish/ C:\\Users\\Example\\App

# Run the application binaries as the main container program
WORKDIR C:\\Users\\Example\\App
ENTRYPOINT ["C:\\Users\\Example\\dotnetinstall\\6.0.101\\dotnet.exe", "C:\\Users\\Example\\App\\DotNet.Docker.dll"]
