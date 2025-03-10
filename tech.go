package wappalyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

// Wappalyze is a client for working with tech detection
type Wappalyze struct {
	fingerprints *CompiledFingerprints
}

// New creates a new tech detection instance
func New(fingerfilepath string) (*Wappalyze, error) {
	wappalyze := &Wappalyze{
		fingerprints: &CompiledFingerprints{
			Apps: make(map[string]*CompiledFingerprint),
		},
	}

	err := wappalyze.loadFingerprints(fingerfilepath)
	if err != nil {
		return nil, err
	}
	return wappalyze, nil
}

// loadFingerprints loads the fingerprints and compiles them
func (s *Wappalyze) loadFingerprints(fingerfilepath string) error {
	var fingerprintsStruct Fingerprints
	//读取指纹库文件
	f, err := ioutil.ReadFile(fingerfilepath)
	if err != nil {
		fmt.Printf("读取本地文件失败，%v", err)
		return nil
	}
	err = json.Unmarshal(f, &fingerprintsStruct)
	if err != nil {
		return err
	}

	for i, fingerprint := range fingerprintsStruct.Apps {
		s.fingerprints.Apps[i] = compileFingerprint(fingerprint)
	}
	return nil
}

// Fingerprint identifies technologies on a target based on
// headers and response recieved.
//
// Body should not be mutated while this function is being called or it may
// lead to unexpected things.
func (s *Wappalyze) Fingerprint(headers map[string][]string, body []byte) map[string]struct{} {
	uniqueFingerprints := newUniqueFingerprints()

	// Lowercase everything that we have recieved to check
	normalizedBody := bytes.ToLower(body)
	normalizedHeaders := s.normalizeHeaders(headers)

	// Run header based fingerprinting if the number
	// of header checks if more than 0.
	for _, application := range s.checkHeaders(normalizedHeaders) {
		uniqueFingerprints.setIfNotExists(application)
	}

	cookies := s.findSetCookie(normalizedHeaders)
	// Run cookie based fingerprinting if we have a set-cookie header
	if len(cookies) > 0 {
		for _, application := range s.checkCookies(cookies) {
			uniqueFingerprints.setIfNotExists(application)
		}
	}

	// Check for stuff in the body finally
	bodyTech := s.checkBody(normalizedBody)
	for _, application := range bodyTech {
		uniqueFingerprints.setIfNotExists(application)
	}
	return uniqueFingerprints.getValues()
}

type uniqueFingerprints struct {
	values map[string]struct{}
}

func newUniqueFingerprints() uniqueFingerprints {
	return uniqueFingerprints{
		values: make(map[string]struct{}),
	}
}

func (u uniqueFingerprints) getValues() map[string]struct{} {
	return u.values
}

const versionSeparator = ":"

// separateAppVersion returns app name and version
func separateAppVersion(value string) (string, string) {
	if strings.Contains(value, versionSeparator) {
		if parts := strings.Split(value, versionSeparator); len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return value, ""
}

func (u uniqueFingerprints) setIfNotExists(value string) {
	app, version := separateAppVersion(value)
	if _, ok := u.values[app]; ok {
		// Handles case when we get additional version information next
		if version != "" {
			delete(u.values, app)
			u.values[strings.Join([]string{app, version}, versionSeparator)] = struct{}{}
		}
		return
	}

	// Handle duplication for : based values
	for k := range u.values {
		if strings.Contains(k, versionSeparator) {
			if parts := strings.Split(k, versionSeparator); len(parts) == 2 && parts[0] == value {
				return
			}
		}
	}
	u.values[value] = struct{}{}
}

// FingerprintWithTitle identifies technologies on a target based on
// headers and response recieved. It also returns the title of the page.
//
// Body should not be mutated while this function is being called or it may
// lead to unexpected things.
func (s *Wappalyze) FingerprintWithTitle(headers map[string][]string, body []byte) (map[string]struct{}, string) {
	uniqueFingerprints := newUniqueFingerprints()

	// Lowercase everything that we have recieved to check
	normalizedBody := bytes.ToLower(body)
	normalizedHeaders := s.normalizeHeaders(headers)

	// Run header based fingerprinting if the number
	// of header checks if more than 0.
	for _, application := range s.checkHeaders(normalizedHeaders) {
		uniqueFingerprints.setIfNotExists(application)
	}

	cookies := s.findSetCookie(normalizedHeaders)
	// Run cookie based fingerprinting if we have a set-cookie header
	if len(cookies) > 0 {
		for _, application := range s.checkCookies(cookies) {
			uniqueFingerprints.setIfNotExists(application)
		}
	}

	// Check for stuff in the body finally
	if strings.Contains(normalizedHeaders["content-type"], "text/html") {
		bodyTech := s.checkBody(normalizedBody)
		for _, application := range bodyTech {
			uniqueFingerprints.setIfNotExists(application)
		}
		title := s.getTitle(body)
		return uniqueFingerprints.getValues(), title
	}
	return uniqueFingerprints.getValues(), ""
}
