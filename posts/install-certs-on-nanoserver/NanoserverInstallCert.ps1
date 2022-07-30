param(
    [Parameter(Mandatory,Position=0)]
    [string]$PfxFile,
    [securestring]$PfxPass,
    [string]$StoreName = 'LocalMachine',
    [string]$StoreLoc = 'My'
)

Write-Host ("Installing cert file {0} to Cert:\{1}\{2}" -f $PfxFile, $StoreName, $StoreLoc)
$pfx = New-Object Security.Cryptography.X509Certificates.X509Certificate2($PfxFile, $PfxPass)
$store = New-Object Security.Cryptography.X509Certificates.X509Store($StoreLoc,$StoreName)
$store.Open("MaxAllowed")
$store.Add($pfx)
$store.Close()
