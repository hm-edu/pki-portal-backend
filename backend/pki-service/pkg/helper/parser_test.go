package helper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseCertificates(t *testing.T) {
	cert := `-----BEGIN CERTIFICATE-----
MIIEMjCCAxqgAwIBAgIBATANBgkqhkiG9w0BAQUFADB7MQswCQYDVQQGEwJHQjEb
MBkGA1UECAwSR3JlYXRlciBNYW5jaGVzdGVyMRAwDgYDVQQHDAdTYWxmb3JkMRow
GAYDVQQKDBFDb21vZG8gQ0EgTGltaXRlZDEhMB8GA1UEAwwYQUFBIENlcnRpZmlj
YXRlIFNlcnZpY2VzMB4XDTA0MDEwMTAwMDAwMFoXDTI4MTIzMTIzNTk1OVowezEL
MAkGA1UEBhMCR0IxGzAZBgNVBAgMEkdyZWF0ZXIgTWFuY2hlc3RlcjEQMA4GA1UE
BwwHU2FsZm9yZDEaMBgGA1UECgwRQ29tb2RvIENBIExpbWl0ZWQxITAfBgNVBAMM
GEFBQSBDZXJ0aWZpY2F0ZSBTZXJ2aWNlczCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAL5AnfRu4ep2hxxNRUSOvkbIgwadwSr+GB+O5AL686tdUIoWMQua
BtDFcCLNSS1UY8y2bmhGC1Pqy0wkwLxyTurxFa70VJoSCsN6sjNg4tqJVfMiWPPe
3M/vg4aijJRPn2jymJBGhCfHdr/jzDUsi14HZGWCwEiwqJH5YZ92IFCokcdmtet4
YgNW8IoaE+oxox6gmf049vYnMlhvB/VruPsUK6+3qszWY19zjNoFmag4qMsXeDZR
rOme9Hg6jc8P2ULimAyrL58OAd7vn5lJ8S3frHRNG5i1R8XlKdH5kBjHYpy+g8cm
ez6KJcfA3Z3mNWgQIJ2P2N7Sw4ScDV7oL8kCAwEAAaOBwDCBvTAdBgNVHQ4EFgQU
oBEKIz6W8Qfs4q8p74Klf9AwpLQwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQF
MAMBAf8wewYDVR0fBHQwcjA4oDagNIYyaHR0cDovL2NybC5jb21vZG9jYS5jb20v
QUFBQ2VydGlmaWNhdGVTZXJ2aWNlcy5jcmwwNqA0oDKGMGh0dHA6Ly9jcmwuY29t
b2RvLm5ldC9BQUFDZXJ0aWZpY2F0ZVNlcnZpY2VzLmNybDANBgkqhkiG9w0BAQUF
AAOCAQEACFb8AvCb6P+k+tZ7xkSAzk/ExfYAWMymtrwUSWgEdujm7l3sAg9g1o1Q
GE8mTgHj5rCl7r+8dFRBv/38ErjHT1r0iWAFf2C3BUrz9vHCv8S5dIa2LX1rzNLz
Rt0vxuBqw8M0Ayx9lt1awg6nCpnBBYurDC/zXDrPbDdVCYfeU0BsWO/8tqtlbgT2
G9w84FoVxp7Z8VlIMCFlA2zs6SFz7JsDoeA3raAVGI/6ugLOpyypEBMs1OUIJqsi
l2D4kF501KKaU73yqWjgom7C12yxow+ev+to51byrvLjKzg6CYG1a4XXvi3tPxq3
smPi9WIsgtRqAEFQ8TmDn5XpNpaYbg==
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIID0zCCArugAwIBAgIQVmcdBOpPmUxvEIFHWdJ1lDANBgkqhkiG9w0BAQwFADB7
MQswCQYDVQQGEwJHQjEbMBkGA1UECAwSR3JlYXRlciBNYW5jaGVzdGVyMRAwDgYD
VQQHDAdTYWxmb3JkMRowGAYDVQQKDBFDb21vZG8gQ0EgTGltaXRlZDEhMB8GA1UE
AwwYQUFBIENlcnRpZmljYXRlIFNlcnZpY2VzMB4XDTE5MDMxMjAwMDAwMFoXDTI4
MTIzMTIzNTk1OVowgYgxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpOZXcgSmVyc2V5
MRQwEgYDVQQHEwtKZXJzZXkgQ2l0eTEeMBwGA1UEChMVVGhlIFVTRVJUUlVTVCBO
ZXR3b3JrMS4wLAYDVQQDEyVVU0VSVHJ1c3QgRUNDIENlcnRpZmljYXRpb24gQXV0
aG9yaXR5MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEGqxUWqn5aCPnetUkb1PGWthL
q8bVttHmc3Gu3ZzWDGH926CJA7gFFOxXzu5dP+Ihs8731Ip54KODfi2X0GHE8Znc
JZFjq38wo7Rw4sehM5zzvy5cU7Ffs30yf4o043l5o4HyMIHvMB8GA1UdIwQYMBaA
FKARCiM+lvEH7OKvKe+CpX/QMKS0MB0GA1UdDgQWBBQ64QmG1M8ZwpZ2dEl23OA1
xmNjmjAOBgNVHQ8BAf8EBAMCAYYwDwYDVR0TAQH/BAUwAwEB/zARBgNVHSAECjAI
MAYGBFUdIAAwQwYDVR0fBDwwOjA4oDagNIYyaHR0cDovL2NybC5jb21vZG9jYS5j
b20vQUFBQ2VydGlmaWNhdGVTZXJ2aWNlcy5jcmwwNAYIKwYBBQUHAQEEKDAmMCQG
CCsGAQUFBzABhhhodHRwOi8vb2NzcC5jb21vZG9jYS5jb20wDQYJKoZIhvcNAQEM
BQADggEBABns652JLCALBIAdGN5CmXKZFjK9Dpx1WywV4ilAbe7/ctvbq5AfjJXy
ij0IckKJUAfiORVsAYfZFhr1wHUrxeZWEQff2Ji8fJ8ZOd+LygBkc7xGEJuTI42+
FsMuCIKchjN0djsoTI0DQoWz4rIjQtUfenVqGtF8qmchxDM6OW1TyaLtYiKou+JV
bJlsQ2uRl9EMC5MCHdK8aXdJ5htN978UeAOwproLtOGFfy/cQjutdAFI3tZs4RmY
CV4Ks2dH/hzg1cEo70qLRDEmBDeNiXQ2Lu+lIg+DdEmSx/cQwgwp+7e9un/jX9Wf
8qn0dNW44bOwgeThpWOjzOoEeJBuv/c=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDeTCCAv+gAwIBAgIRAOuOgRlxKfSvZO+BSi9QzukwCgYIKoZIzj0EAwMwgYgx
CzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpOZXcgSmVyc2V5MRQwEgYDVQQHEwtKZXJz
ZXkgQ2l0eTEeMBwGA1UEChMVVGhlIFVTRVJUUlVTVCBOZXR3b3JrMS4wLAYDVQQD
EyVVU0VSVHJ1c3QgRUNDIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MB4XDTIwMDIx
ODAwMDAwMFoXDTMzMDUwMTIzNTk1OVowRDELMAkGA1UEBhMCTkwxGTAXBgNVBAoT
EEdFQU5UIFZlcmVuaWdpbmcxGjAYBgNVBAMTEUdFQU5UIE9WIEVDQyBDQSA0MFkw
EwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEXYkvGrfrMs2IwdI5+IwpEwPh+igW/BOW
etmOwP/ZIXC8fNeC3/ZYPAAMyRpFS0v3/c55FDTE2xbOUZ5zeVZYQqOCAYswggGH
MB8GA1UdIwQYMBaAFDrhCYbUzxnClnZ0SXbc4DXGY2OaMB0GA1UdDgQWBBTttKAz
ahsIkba9+kGSvZqrq2P0UzAOBgNVHQ8BAf8EBAMCAYYwEgYDVR0TAQH/BAgwBgEB
/wIBADAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwOAYDVR0gBDEwLzAt
BgRVHSAAMCUwIwYIKwYBBQUHAgEWF2h0dHBzOi8vc2VjdGlnby5jb20vQ1BTMFAG
A1UdHwRJMEcwRaBDoEGGP2h0dHA6Ly9jcmwudXNlcnRydXN0LmNvbS9VU0VSVHJ1
c3RFQ0NDZXJ0aWZpY2F0aW9uQXV0aG9yaXR5LmNybDB2BggrBgEFBQcBAQRqMGgw
PwYIKwYBBQUHMAKGM2h0dHA6Ly9jcnQudXNlcnRydXN0LmNvbS9VU0VSVHJ1c3RF
Q0NBZGRUcnVzdENBLmNydDAlBggrBgEFBQcwAYYZaHR0cDovL29jc3AudXNlcnRy
dXN0LmNvbTAKBggqhkjOPQQDAwNoADBlAjAfs9nsM0qaJGVu6DpWVy4qojiOpwV1
h/MWZ5GJxy6CKv3+RMB3STkaFh0+Hifbk24CMQDRf/ujXAQ1b4nFpZGaSIKldygc
dCDAxbAd9tlxcN/+J534CJDblzd/40REzGWwS5k=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIE2DCCBH6gAwIBAgIQBIJfJ2dTQE57ATBBDsk/8TAKBggqhkjOPQQDAjBEMQsw
CQYDVQQGEwJOTDEZMBcGA1UEChMQR0VBTlQgVmVyZW5pZ2luZzEaMBgGA1UEAxMR
R0VBTlQgT1YgRUNDIENBIDQwHhcNMjIwNDI2MDAwMDAwWhcNMjMwNDI2MjM1OTU5
WjB0MQswCQYDVQQGEwJERTEPMA0GA1UECBMGQmF5ZXJuMTswOQYDVQQKDDJIb2No
c2NodWxlIGbDvHIgYW5nZXdhbmR0ZSBXaXNzZW5zY2hhZnRlbiBNw7xuY2hlbjEX
MBUGA1UEAxMOYWNtZS5obXRlc3QuZGUwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
AASCxog6gG6g7/j6xr+NWFIoPan1Fj+oLNC+blnRa7pbMYf86gdCHZkgfs3XbemC
bHvyS1hRdZvHqZQn/6bRYqMwo4IDIDCCAxwwHwYDVR0jBBgwFoAU7bSgM2obCJG2
vfpBkr2aq6tj9FMwHQYDVR0OBBYEFGZBwVmuD1ipdcdJFwX8HKdbaJu9MA4GA1Ud
DwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggr
BgEFBQcDAjBJBgNVHSAEQjBAMDQGCysGAQQBsjEBAgJPMCUwIwYIKwYBBQUHAgEW
F2h0dHBzOi8vc2VjdGlnby5jb20vQ1BTMAgGBmeBDAECAjA/BgNVHR8EODA2MDSg
MqAwhi5odHRwOi8vR0VBTlQuY3JsLnNlY3RpZ28uY29tL0dFQU5UT1ZFQ0NDQTQu
Y3JsMHUGCCsGAQUFBwEBBGkwZzA6BggrBgEFBQcwAoYuaHR0cDovL0dFQU5ULmNy
dC5zZWN0aWdvLmNvbS9HRUFOVE9WRUNDQ0E0LmNydDApBggrBgEFBQcwAYYdaHR0
cDovL0dFQU5ULm9jc3Auc2VjdGlnby5jb20wggF9BgorBgEEAdZ5AgQCBIIBbQSC
AWkBZwB2AK33vvp8/xDIi509nB4+GGq0Zyldz7EMJMqFhjTr3IKKAAABgGXBhBUA
AAQDAEcwRQIgR1mft+C7R+g7ZJ0+TBiwxMX77/SdXRtEifQyWWpm7I0CIQCXnOnr
8oHDkJcUfBeNTbq/+h/FU2KXcOW/2Z9sHx7AhgB1AHoyjFTYty22IOo44FIe6YQW
cDIThU070ivBOlejUutSAAABgGXBg/4AAAQDAEYwRAIgP0+6X4uWbSYGcuB/Fytt
nIUR3zA8nmEqqpyGObDi8ZICIFMf4faZ7kFRGp0grxJ1RXjufChW5K1ilNVIZUlL
cXUYAHYA6D7Q2j71BjUy51covIlryQPTy9ERa+zraeF3fW0GvW4AAAGAZcGDsQAA
BAMARzBFAiA/uSXiAcsdHlS36tqexD0E8haB0oTZCNWx96kEWZlMRwIhAJ1OIHS8
jLWi9ogVJpq+RhGRfZP2Yj/lYH7jEOOqZ4guMBkGA1UdEQQSMBCCDmFjbWUuaG10
ZXN0LmRlMAoGCCqGSM49BAMCA0gAMEUCIFUQcnoZg32tUrBAe8kgYiOzL3sDwrxi
P6q6wVTtStlKAiEApNAfxM117NR/FlnwMUOSJtKEWNwAqwpNLUzQ+4PlqVg=
-----END CERTIFICATE-----`

	certs, err := ParseCertificates([]byte(cert))
	assert.NoError(t, err)
	assert.Len(t, certs, 4)
	assert.Len(t, certs[0:len(certs)-1], 3)
	assert.Equal(t, "04825f276753404e7b0130410ec93ff1", fmt.Sprintf("%032x", certs[3].SerialNumber))
}