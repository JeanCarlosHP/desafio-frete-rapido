// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package querier

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Quote struct {
	ID          int32
	CarrierName string
	Service     string
	Price       float64
	Deadline    int
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
}
