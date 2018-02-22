package types

// This file contains some code from  https://github.com/twitchtv/twirp/:
// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.  All rights reserved.
// https://github.com/twitchtv/twirp/

import (
	"bytes"
	"errors"
	"fmt"
	"go/token"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// BaseName the last path element of a slash-delimited name, with the last
// dotted suffix removed.
func BaseName(name string) string {
	// First, find the last element
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	// Now drop the suffix
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[0:i]
	}
	return name
}

// SnakeCase converts a string from CamelCase to snake_case.
func SnakeCase(s string) string {
	var buf bytes.Buffer
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			fmt.Fprintf(&buf, "_")
		}
		r = unicode.ToLower(r)
		fmt.Fprintf(&buf, "%c", r)
	}
	return buf.String()
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func CamelCase(s string) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// Need a capital letter; drop the '_'.
		t = append(t, 'X')
		i++
	}
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	//
	// That is, we process a word at a time, where words are marked by _ or upper
	// case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier. The next
		// word is a sequence of characters that must start upper case.
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}

var noneLiteralChars = regexp.MustCompile(`([^\w(),".&]{0,})`)

func identifierList(names []string) string {
	var text bytes.Buffer
	for _, name := range names {
		if text.Len() > 0 {
			text.WriteString(", ")
		}
		text.WriteString(Identifier(name))
	}
	return text.String()
}

func ValueList(names []string) string {
	return strings.Join(names, ", ")
}

func typeList(types []TypeReference) string {

	var text bytes.Buffer
	for _, typ := range types {
		if text.Len() > 0 {
			text.WriteString(", ")
		}
		text.WriteString(typ.GetName())
	}

	return text.String()
}

func paramList(parameters []*Parameter) string {

	var text bytes.Buffer
	for _, parameter := range parameters {
		if text.Len() > 0 {
			text.WriteString(", ")
		}

		text.WriteString(Identifier(parameter.NameOfParameter))
		text.WriteString(" ")
		text.WriteString(parameter.Typ.GetName())
	}

	return text.String()
}

func ExportedIdentifier(name string) string {
	return Identifier(strings.Title(name))
}

func UnexportedIdentifier(name string) string {
	return Identifier(UnTitle(name))
}

func Identifier(name string) string {
	variadic := strings.HasSuffix(name, "...")
	ident := noneLiteralChars.ReplaceAllString(name, "")
	if variadic {
		ident += "..."
	}
	return ident
}

func UnTitle(s string) string {

	if len(s) == 0 {
		return s
	}

	r, width := utf8.DecodeRuneInString(s)
	return fmt.Sprintf("%c", unicode.ToLower(r)) + s[width:]
}

func GoFileName(fileName string) string {
	return fmt.Sprintf("%s.go", SnakeCase(fileName))
}

var statusCode = map[int]string{
	http.StatusContinue:           "http.StatusContinue",
	http.StatusSwitchingProtocols: "http.StatusSwitchingProtocols",
	http.StatusProcessing:         "http.StatusProcessing",

	http.StatusOK:                   "http.StatusOK",
	http.StatusCreated:              "http.StatusCreated",
	http.StatusAccepted:             "http.StatusAccepted",
	http.StatusNonAuthoritativeInfo: "http.StatusNonAuthoritativeInfo",
	http.StatusNoContent:            "http.StatusNoContent",
	http.StatusResetContent:         "http.StatusResetContent",
	http.StatusPartialContent:       "http.StatusPartialContent",
	http.StatusMultiStatus:          "http.StatusMultiStatus",
	http.StatusAlreadyReported:      "http.StatusAlreadyReported",
	http.StatusIMUsed:               "http.StatusIMUsed",

	http.StatusMultipleChoices:   "http.StatusMultipleChoices",
	http.StatusMovedPermanently:  "http.StatusMovedPermanently",
	http.StatusFound:             "http.StatusFound",
	http.StatusSeeOther:          "http.StatusSeeOther",
	http.StatusNotModified:       "http.StatusNotModified",
	http.StatusUseProxy:          "http.StatusUseProxy",
	http.StatusTemporaryRedirect: "http.StatusTemporaryRedirect",
	http.StatusPermanentRedirect: "http.StatusPermanentRedirect",

	http.StatusBadRequest:                   "http.StatusBadRequest",
	http.StatusUnauthorized:                 "http.StatusUnauthorized",
	http.StatusPaymentRequired:              "http.StatusPaymentRequired",
	http.StatusForbidden:                    "http.StatusForbidden",
	http.StatusNotFound:                     "http.StatusNotFound",
	http.StatusMethodNotAllowed:             "http.StatusMethodNotAllowed",
	http.StatusNotAcceptable:                "http.StatusNotAcceptable",
	http.StatusProxyAuthRequired:            "http.StatusProxyAuthRequired",
	http.StatusRequestTimeout:               "http.StatusRequestTimeout",
	http.StatusConflict:                     "http.StatusConflict",
	http.StatusGone:                         "http.StatusGone",
	http.StatusLengthRequired:               "http.StatusLengthRequired:",
	http.StatusPreconditionFailed:           "http.StatusPreconditionFailed",
	http.StatusRequestEntityTooLarge:        "http.StatusRequestEntityTooLarge",
	http.StatusRequestURITooLong:            "http.StatusRequestURITooLong",
	http.StatusUnsupportedMediaType:         "http.StatusUnsupportedMediaType",
	http.StatusRequestedRangeNotSatisfiable: "http.StatusRequestedRangeNotSatisfiable",
	http.StatusExpectationFailed:            "http.StatusExpectationFailed",
	http.StatusTeapot:                       "http.StatusTeapot",
	http.StatusUnprocessableEntity:          "http.StatusUnprocessableEntity",
	http.StatusLocked:                       "http.StatusLocked",
	http.StatusFailedDependency:             "http.StatusFailedDependency",
	http.StatusUpgradeRequired:              "http.StatusUpgradeRequired",
	http.StatusPreconditionRequired:         "http.StatusPreconditionRequired",
	http.StatusTooManyRequests:              "http.StatusTooManyRequests",
	http.StatusRequestHeaderFieldsTooLarge:  "http.StatusRequestHeaderFieldsTooLarge",
	http.StatusUnavailableForLegalReasons:   "http.StatusUnavailableForLegalReasons",

	http.StatusInternalServerError:           "http.StatusInternalServerError",
	http.StatusNotImplemented:                "http.StatusNotImplemented",
	http.StatusBadGateway:                    "http.StatusBadGateway",
	http.StatusServiceUnavailable:            "http.StatusServiceUnavailable",
	http.StatusGatewayTimeout:                "http.StatusGatewayTimeout",
	http.StatusHTTPVersionNotSupported:       "http.StatusHTTPVersionNotSupported",
	http.StatusVariantAlsoNegotiates:         "http.StatusVariantAlsoNegotiates",
	http.StatusInsufficientStorage:           "http.StatusInsufficientStorage",
	http.StatusLoopDetected:                  "http.StatusLoopDetected",
	http.StatusNotExtended:                   "http.StatusNotExtended",
	http.StatusNetworkAuthenticationRequired: "http.StatusNetworkAuthenticationRequired",
}

// StatusCodeVar returns a var for the HTTP status code. It returns the empty
// string if the code is unknown.
func StatusCodeVar(code int) string {
	return statusCode[code]
}

func StatusText(statusCode int) string {
	status := StatusCodeVar(statusCode)
	if status == "" {
		return ""
	}
	return strings.Replace(status, "http.", "", 1)
}

func ValidateIdent(ident string) error {
	if t := token.Lookup(ident); t != token.IDENT {
		return fmt.Errorf("identifier is a key word (%s)", ident)
	}
	return nil
}

func ValidateIdents(idents []string) error {
	for _, ident := range idents {
		if t := token.Lookup(ident); t != token.IDENT {
			return fmt.Errorf("identifier is a key word (%s)", ident)
		}
	}
	return nil
}

func ValidateParameters(parameters []*Parameter) error {
	for _, parameter := range parameters {

		if parameter == nil {
			return errors.New("TypeReference is missing")
		}

		if parameter.NameOfParameter == "" {
			return errors.New("name of parameter is missing")
		}

		if parameter.Typ.GetName() == "" {
			return errors.New("value of parameter is missing")
		}

		if err := ValidateIdent(parameter.NameOfParameter); err != nil {
			return err
		}
	}

	return nil
}

func ValidateOperation(op token.Token) error {
	if op.Precedence() == token.LowestPrec {
		return fmt.Errorf("operation is invalid in operation scope (%s)", op.String())
	}
	return nil
}

func ValidateIfOperation(op token.Token) error {

	switch op.Precedence() {
	case 1:
		fallthrough
	case 2:
		fallthrough
	case 3:
		return nil
	}

	return fmt.Errorf("operation is invalid in if scope (%s)", op.String())
}

func NewGeneratorErrorString(generator interface{}, s string) *GeneratorErrorString {
	return &GeneratorErrorString{
		s:         s,
		generator: generator,
	}
}

type GeneratorErrorString struct {
	s         string
	generator interface{}
}

func (e *GeneratorErrorString) Error() string {
	return fmt.Sprintf("%s: %s", reflect.TypeOf(e.generator).String(), e.s)
}

func NewGeneratorError(generator interface{}, err error) *GeneratorError {
	return &GeneratorError{
		err:       err,
		generator: generator,
	}
}

type GeneratorError struct {
	err       error
	generator interface{}
}

func (e *GeneratorError) Error() string {
	return fmt.Sprintf("%s: %s", reflect.TypeOf(e.generator).String(), e.err)
}
