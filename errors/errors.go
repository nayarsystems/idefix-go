package errors

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	ErrUnknown          = IdefixError{Code: 1, Message: "Unknown error"}
	ErrInternal         = IdefixError{Code: 2, Message: "Internal error"}
	ErrNotImplemented   = IdefixError{Code: 3, Message: "Not implemented"}
	ErrNotAuthorized    = IdefixError{Code: 4, Message: "Not authorized"}
	ErrPermissionDenied = IdefixError{Code: 5, Message: "Permission denied"}
	ErrInvalidParams    = IdefixError{Code: 6, Message: "Invalid parameters"}
	ErrTimeout          = IdefixError{Code: 7, Message: "Timeout"}
	ErrAlreadyExists    = IdefixError{Code: 8, Message: "Already exists"}
	ErrNotFound         = IdefixError{Code: 9, Message: "Not found"}
	ErrContextClosed    = IdefixError{Code: 11, Message: "Context closed"}
	ErrChannelClosed    = IdefixError{Code: 12, Message: "Channel closed"}
	ErrMarshal          = IdefixError{Code: 13, Message: "Marshal error"}
	ErrParse            = IdefixError{Code: 14, Message: "Parse error"}

	ErrInvalidCommand = IdefixError{Code: 100, Message: "Invalid command"}
	ErrEmptyTopic     = IdefixError{Code: 101, Message: "Empty topic"}
	ErrInvalidSession = IdefixError{Code: 102, Message: "Invalid session"}

	ErrInvalidToken         = IdefixError{Code: 201, Message: "Invalid token"}
	ErrInvalidDataType      = IdefixError{Code: 202, Message: "Invalid data type"}
	ErrMissingDomain        = IdefixError{Code: 203, Message: "Missing domain"}
	ErrMissingAddress       = IdefixError{Code: 204, Message: "Missing address"}
	ErrInvalidAddress       = IdefixError{Code: 205, Message: "Invalid address"}
	ErrInvalidAddressSyntax = IdefixError{Code: 206, Message: "Invalid address syntax"}
	ErrInvalidDomainSyntax  = IdefixError{Code: 207, Message: "Invalid domain syntax"}
	ErrDomainNotFound       = IdefixError{Code: 208, Message: "Domain not found"}
	ErrAddressNotFound      = IdefixError{Code: 209, Message: "Address not found"}
	ErrInvalidSchemaSyntax  = IdefixError{Code: 210, Message: "Invalid schema syntax"}
	ErrAddressNotAssigned   = IdefixError{Code: 211, Message: "Address not assigned to a domain"}
	ErrSchemaNotFound       = IdefixError{Code: 212, Message: "Schema not found"}
)

type IdefixError struct {
	Code    int
	Message string
	Extra   string
}

func (ie IdefixError) Error() string {
	s := fmt.Sprintf("[%d]", ie.Code)

	if ie.Message != "" {
		s = fmt.Sprintf("%s %s", s, ie.Message)
	}

	if ie.Extra != "" {
		s = fmt.Sprintf("%s (%s)", s, ie.Extra)
	}

	return s
}

func (ie IdefixError) Is(e error) bool {
	var ep *IdefixError
	switch et := e.(type) {
	case IdefixError:
		ep = &et
	case *IdefixError:
		ep = et
	default:
		var err error
		ep, err = Parse(e.Error())
		if err != nil {
			return false
		}
	}

	return ie.Code == ep.Code
}

func (ie IdefixError) With(extra string) IdefixError {
	ie.Extra = extra
	return ie
}

func (ie IdefixError) WithErr(err error) IdefixError {
	ie.Extra = err.Error()
	return ie
}

func (ie IdefixError) Withf(format string, a ...any) IdefixError {
	ie.Extra = fmt.Sprintf(format, a...)
	return ie
}

func Parse(e string) (*IdefixError, error) {
	r := regexp.MustCompile(`^\[([-]?[0-9]+)\] ([^()]*)(?:[ ]\((.+)\))?$`)
	ret := r.FindStringSubmatch(e)

	if len(ret) < 3 {
		return nil, ErrParse
	}

	code, err := strconv.ParseInt(ret[1], 10, 64)
	if err != nil {
		return nil, err
	}

	ie := &IdefixError{Code: int(code), Message: ret[2]}

	if len(ret) == 4 {
		ie.Extra = ret[3]
	}

	return ie, nil
}
