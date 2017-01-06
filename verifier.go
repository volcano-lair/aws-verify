package main

import "log"
import "net/http"

import "encoding/pem"
import "crypto/x509"
import "io/ioutil"

/**
 * The global public certificate used by Amazon to generate EC2
 * identity signatures
 */
var AMAZON_PUBLIC_CLOUD = []byte(`-----BEGIN CERTIFICATE-----
MIIC7TCCAq0CCQCWukjZ5V4aZzAJBgcqhkjOOAQDMFwxCzAJBgNVBAYTAlVTMRkw
FwYDVQQIExBXYXNoaW5ndG9uIFN0YXRlMRAwDgYDVQQHEwdTZWF0dGxlMSAwHgYD
VQQKExdBbWF6b24gV2ViIFNlcnZpY2VzIExMQzAeFw0xMjAxMDUxMjU2MTJaFw0z
ODAxMDUxMjU2MTJaMFwxCzAJBgNVBAYTAlVTMRkwFwYDVQQIExBXYXNoaW5ndG9u
IFN0YXRlMRAwDgYDVQQHEwdTZWF0dGxlMSAwHgYDVQQKExdBbWF6b24gV2ViIFNl
cnZpY2VzIExMQzCCAbcwggEsBgcqhkjOOAQBMIIBHwKBgQCjkvcS2bb1VQ4yt/5e
ih5OO6kK/n1Lzllr7D8ZwtQP8fOEpp5E2ng+D6Ud1Z1gYipr58Kj3nssSNpI6bX3
VyIQzK7wLclnd/YozqNNmgIyZecN7EglK9ITHJLP+x8FtUpt3QbyYXJdmVMegN6P
hviYt5JH/nYl4hh3Pa1HJdskgQIVALVJ3ER11+Ko4tP6nwvHwh6+ERYRAoGBAI1j
k+tkqMVHuAFcvAGKocTgsjJem6/5qomzJuKDmbJNu9Qxw3rAotXau8Qe+MBcJl/U
hhy1KHVpCGl9fueQ2s6IL0CaO/buycU1CiYQk40KNHCcHfNiZbdlx1E9rpUp7bnF
lRa2v1ntMX3caRVDdbtPEWmdxSCYsYFDk4mZrOLBA4GEAAKBgEbmeve5f8LIE/Gf
MNmP9CM5eovQOGx5ho8WqD+aTebs+k2tn92BBPqeZqpWRa5P/+jrdKml1qx4llHW
MXrs3IgIb6+hUIB+S8dz8/mmO0bpr76RoZVCXYab2CZedFut7qc3WUH9+EUAH5mw
vSeDCOUMYQR7R9LINYwouHIziqQYMAkGByqGSM44BAMDLwAwLAIUWXBlk40xTwSw
7HX32MxXYruse9ACFBNGmdX2ZBrVNGrN9N2f6ROk0k9K
-----END CERTIFICATE-----
`)

/**
 * Create a Verifier instance
 */
func CreateVerifier() *Verifier {
	return &Verifier{}
}

/**
 * A custom `http.Handler` for verifying AWS EC2 instance identity document
 * signatures
 */
type Verifier struct {
	certificates []*x509.Certificate
}

/**
 * Read a PEM-encoded x509 certificate from a file and add to signing candidates
 */
func (verify *Verifier) ReadPEMCertificate(path string) {
	log.Printf("Loading certificate from %s", path)

	data, err := ioutil.ReadFile(path)
	fatal(err)

	verify.AddPEMCertificate(data)
}

/**
 * Parse a PEM-encoded x509 certificate and add to signing candidates
 */
func (verify *Verifier) AddPEMCertificate(data []byte) {
	block, _ := pem.Decode(data)

	certificate, err := x509.ParseCertificate(block.Bytes)
	fatal(err)

	verify.AddCertificate(certificate)
}

/**
 * Add a certificate to the Verifier's signing candidates
 */
func (verify *Verifier) AddCertificate(certificate *x509.Certificate) {
	log.Printf("Adding certificate %s, %s, %s to singing candidates",
		certificate.Subject.Organization,
		certificate.Subject.Province,
		certificate.Subject.Country,
	)

	verify.certificates = append(verify.certificates, certificate)
}

/**
 * Check for an error value and respond immediately with the specified status
 * code and the error's message, returning `false`
 */
func (*Verifier) OK(err error, code int, w http.ResponseWriter) bool {
	if err == nil {
		return true
	}

	response := NewResponse(code, false)
	response.AddError(err)
	response.Send(w)

	return false
}

/**
 * `http.Handler` interface. Handle incoming requests.
 */
func (verify *Verifier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Load data and signature parameters from the incoming request
	request, err := NewRequest(r)

	if verify.OK(err, http.StatusBadRequest, w) &&

		// Parse the PKCS7 object
		verify.OK(request.Parse(verify.certificates), http.StatusBadRequest, w) &&

		// Verify the signature and content
		verify.OK(request.Verify(), http.StatusForbidden, w) {

		// Respond with a success status and the verified document
		response := NewResponse(http.StatusOK, true)

		if verify.OK(response.AddDocument(request.P7.Content), http.StatusBadRequest, w) {
			response.Send(w)
		}
	}
}