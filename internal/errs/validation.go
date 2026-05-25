package errs

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FieldViolation struct {
	Field       string
	Description string
}

func NewInvalidArgumentError(
	message string,
	violations []FieldViolation,
) error {
	st := status.New(codes.InvalidArgument, message)

	badRequest := &errdetails.BadRequest{
		FieldViolations: make([]*errdetails.BadRequest_FieldViolation, 0, len(violations)),
	}

	for _, violation := range violations {
		badRequest.FieldViolations = append(
			badRequest.FieldViolations,
			&errdetails.BadRequest_FieldViolation{
				Field:       violation.Field,
				Description: violation.Description,
			},
		)
	}

	stWithDetails, err := st.WithDetails(badRequest)
	if err != nil {
		return st.Err()
	}

	return stWithDetails.Err()
}
