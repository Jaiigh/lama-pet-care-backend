package entities

import (
	"lama-backend/domain/prisma/db"
)

type LeavedayModel struct {
	StaffID   string
	StaffType string
	Leaveday  db.DateTime
}
