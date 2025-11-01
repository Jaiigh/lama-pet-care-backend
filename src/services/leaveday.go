package services

import (
	"fmt"
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
	"time"
)

type leavedayService struct {
	repo repositories.ILeavedayRepository
}

type ILeavedayService interface {
	InsertLeaveday(staffId, role string, leaveday time.Time) (*entities.LeavedayModel, error)
}

func NewLeavedayService(repo repositories.ILeavedayRepository) ILeavedayService {
	return &leavedayService{
		repo: repo,
	}
}

func (sv *leavedayService) InsertLeaveday(staffId, role string, leaveday time.Time) (*entities.LeavedayModel, error) {
	switch role {
	case "caretaker":
		return sv.repo.InsertCaretakerLeaveday(staffId, leaveday)
	case "doctor":
		return sv.repo.InsertDoctorLeaveday(staffId, leaveday)
	default:
		return nil, fmt.Errorf("service layer -> invalid role")
	}
}
