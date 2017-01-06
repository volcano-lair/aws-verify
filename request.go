package main

import "errors"
import "log"
import "io/ioutil"
import "net/http"

import "encoding/pem"
import "crypto/x509"

import "github.com/fullsailor/pkcs7"

/**
 * Instantiate a new Request and read the body of the provided
 * http.Request instance
 */
func NewRequest(r *http.Request) (*Request, error) {
	request := &Request{}

	return request, request.Read(r)
}

/**
 * Encapsulate incoming HTTP request parameters and the resulting PKCS7 object
 */
type Request struct {
	P7        *pkcs7.PKCS7
	Signature []byte
}

/**
 * Read an incoming request's body into parameter fields
 */
func (request *Request) Read(r *http.Request) (err error) {
	body, err := ioutil.ReadAll(r.Body)

	if err == nil {
		request.Signature = body
	}

	return err
}

/**
 * Parse the Request's PEM encoded Signature field into a PKCS7 instance
 * and attach a set of x509 signing candidates for later verification
 */
func (request *Request) Parse(certificates []*x509.Certificate) (err error) {
	/*
	 * Catch panics from `pem.Decode` and `pkcs7.Parse`
	 *
	 * Send a generic error message back to the client, and log the panic message
	 * for later debugging.
	 */
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic parsing PKCS7 PEM block in request: %s", r)
			err = errors.New("Error parsing PKCS7 PEM block")
		}
	}()

	// Extract ANS1 from the PEM block
	block, _ := pem.Decode(request.Signature)

	// Parse the ASN1 object into a PKCS7 instance
	p7, err := pkcs7.Parse(block.Bytes)

	// Attach a certificate set for later verification
	if err == nil {
		p7.Certificates = certificates
		request.P7 = p7
	}

	return err
}

/**
 * Verify the signature's signer integrity against signing candidates then
 * validate the signed content against the provided request document
 */
func (request *Request) Verify() error {
	return request.P7.Verify()
}