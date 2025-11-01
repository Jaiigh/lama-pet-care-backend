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
	FindByCaretakerID(cid string) (*entities.LeavedayModel, error)
	FindByDoctorID(did string) (*entities.LeavedayModel, error)
	FindByLeaveday(leaveday time.Time) (*[]entities.LeavedayModel, error)
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
		db.Leaveday.Caretaker.Link(db.Caretaker.UserID.Equals(cid)),
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
		Leaveday:  []db.DateTime{createdData.Leaveday},
	}, nil
}

func (repo *leavedayRepository) InsertDoctorLeaveday(did string, leaveday time.Time) (*entities.LeavedayModel, error) {
	createdData, err := repo.Collection.Leaveday.CreateOne(
		db.Leaveday.Leaveday.Set(leaveday),
		db.Leaveday.Doctor.Link(db.Doctor.UserID.Equals(did)),
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
		Leaveday:  []db.DateTime{createdData.Leaveday},
	}, nil
}

func (repo *leavedayRepository) FindByCaretakerID(cid string) (*entities.LeavedayModel, error) {
	data, err := repo.Collection.Leaveday.FindMany(
		db.Leaveday.Cid.Equals(cid),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("leaveday -> FindLeavedayByStaffID: %v", err)
	}

	var leavedays []db.DateTime
	for i := range data {
		leavedays = append(leavedays, data[i].Leaveday)
	}

	return &entities.LeavedayModel{
		StaffID:   cid,
		StaffType: "caretaker",
		Leaveday:  leavedays,
	}, nil
}

func (repo *leavedayRepository) FindByDoctorID(did string) (*entities.LeavedayModel, error) {
	data, err := repo.Collection.Leaveday.FindMany(
		db.Leaveday.Did.Equals(did),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("leaveday -> FindLeavedayByStaffID: %v", err)
	}

	var leavedays []db.DateTime
	for i := range data {
		leavedays = append(leavedays, data[i].Leaveday)
	}

	return &entities.LeavedayModel{
		StaffID:   did,
		StaffType: "doctor",
		Leaveday:  leavedays,
	}, nil
}

func (repo *leavedayRepository) FindByLeaveday(leaveday time.Time) (*[]entities.LeavedayModel, error) {
	data, err := repo.Collection.Leaveday.FindMany(
		db.Leaveday.Leaveday.Equals(leaveday),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("leaveday -> FindLeavedayByStaffID: %v", err)
	}

	var results []entities.LeavedayModel
	for i := range data {
		var staffId, staffType string
		var ok bool
		if staffId, ok = data[i].Cid(); !ok {
			staffId, _ = data[i].Did()
			staffType = "doctor"
		} else {
			staffType = "caretaker"
		}
		results = append(results, entities.LeavedayModel{
			StaffID:   staffId,
			StaffType: staffType,
			Leaveday:  []db.DateTime{data[i].Leaveday},
		})
	}

	return &results, nil
}
