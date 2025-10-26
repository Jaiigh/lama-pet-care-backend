package repositories

import (
	"context"
	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
	"time"

	"fmt"
)

type leavedayRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type ILeavedayRepository interface {
	InsertCaretakerLeaveday(cid string, leaveday time.Time) (*entities.LeavedayModel, error)
	InsertDoctorLeaveday(did string, leaveday time.Time) (*entities.LeavedayModel, error)
}

func NewLeavedayRepository(db *ds.PrismaDB) ILeavedayRepository {
	return &leavedayRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *leavedayRepository) InsertCaretakerLeaveday(cid string, leaveday time.Time) (*entities.LeavedayModel, error) {
	createdData, err := repo.Collection.Leaveday.CreateOne(
		db.Leaveday.Leaveday.Set(leaveday),
		db.Leaveday.Cid.Set(cid),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("leaveday -> InsertLeaveday: %v", err)
	}

	staffId, ok := createdData.Cid()
	if !ok {
		return nil, fmt.Errorf("leaveday -> InsertLeaveday: cannot insert/get caretaker_id")
	}

	return &entities.LeavedayModel{
		StaffID:   staffId,
		StaffType: "caretaker",
		Leaveday:  createdData.Leaveday,
	}, nil
}

func (repo *leavedayRepository) InsertDoctorLeaveday(did string, leaveday time.Time) (*entities.LeavedayModel, error) {
	createdData, err := repo.Collection.Leaveday.CreateOne(
		db.Leaveday.Leaveday.Set(leaveday),
		db.Leaveday.Did.Set(did),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("leaveday -> InsertLeaveday: %v", err)
	}

	staffId, ok := createdData.Did()
	if !ok {
		return nil, fmt.Errorf("leaveday -> InsertLeaveday: cannot insert/get doctor_id")
	}

	return &entities.LeavedayModel{
		StaffID:   staffId,
		StaffType: "doctor",
		Leaveday:  createdData.Leaveday,
	}, nil
}
