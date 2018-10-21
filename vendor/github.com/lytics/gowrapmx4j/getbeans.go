package gowrapmx4j

import (
	"encoding/xml"
	"io"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
)

//Handles reading of the http.Body and passes bytes of io.ReadCloser
//to getAttrUnmarshal() for unmarshaling XML.
func getBeans(httpBody io.ReadCloser, unmarshalFunc func([]byte) (*Bean, error)) (*Bean, error) {
	xmlBytes, err := ioutil.ReadAll(httpBody)
	if err != nil {
		log.Errorf("Failed to read http response: %#v", err)
		return nil, err
	}

	return unmarshalFunc(xmlBytes)
}

//Unmarshals XML and returns an Bean struct
func beanUnmarshal(xmlBytes []byte) (*Bean, error) {
	var mb Bean
	err := xml.Unmarshal([]byte(xmlBytes), &mb)
	if err != nil {
		log.Errorf("Failed to Unmarshal xml: %#v", err)
		log.Errorf("Bytes failed to be unmarshalled: \n%s", xmlBytes)
		return nil, err
	}
	return &mb, nil
}
