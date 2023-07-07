package errors

import (
	"errors"

	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/internal/jettisonpb"
)

func ErrorFromProto(wrappedError *jettisonpb.WrappedError) (*JettisonError, error) {
	meta, err := internal.MetadataFromProto(wrappedError.Metadata)
	if err != nil {
		return nil, err
	}
	je := &JettisonError{
		Message: wrappedError.Message,
	}
	if meta != nil {
		je.Metadata = *meta
	}
	if len(wrappedError.JoinedErrors) > 0 {
		var errs []error
		for _, joinErr := range wrappedError.JoinedErrors {
			subErr, err := ErrorFromProto(joinErr)
			if err != nil {
				return nil, err
			}
			errs = append(errs, subErr)
		}
		je.Err = errors.Join(errs...)
	} else if wrappedError.WrappedError != nil {
		subErr, err := ErrorFromProto(wrappedError.WrappedError)
		if err != nil {
			return nil, err
		}
		je.Err = subErr
	}
	return je, nil
}

func ErrorToProto(err error) (*jettisonpb.WrappedError, error) {
	if err == nil {
		return nil, nil
	}
	var we jettisonpb.WrappedError
	je, ok := err.(*JettisonError)
	if ok {
		we.Message = je.Message
		if !je.Metadata.IsZero() {
			mdProto, err := internal.MetadataToProto(&je.Metadata)
			if err != nil {
				return nil, err
			}
			we.Metadata = mdProto
		}
	}
	switch unw := err.(type) {
	case unwrapper:
		subWe, err := ErrorToProto(unw.Unwrap())
		if err != nil {
			return nil, err
		}
		we.WrappedError = subWe
	case unwrapperList:
		for _, e := range unw.Unwrap() {
			subWe, err := ErrorToProto(e)
			if err != nil {
				return nil, err
			}
			we.JoinedErrors = append(we.JoinedErrors, subWe)
		}
	default:
		we.Message = err.Error()
	}
	return &we, nil
}
