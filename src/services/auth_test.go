package services

import (
    "errors"
    "testing"
    "time"

    "lama-backend/domain/entities"
    "lama-backend/src/services/mocks"
    "lama-backend/src/utils"

    "github.com/golang/mock/gomock"
    "github.com/shopspring/decimal"
)

func TestAuthService_Login(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockUsers := mocks.NewMockIUsersRepository(ctrl)
    sv := &authService{UsersRepository: mockUsers}

    // prepare hashed password to simulate stored hash
    const plain = "secret"
    hashed, err := utils.HashPassword(plain)
    if err != nil {
        t.Fatalf("hash password failed: %v", err)
    }

    tests := []struct {
        name     string
        email    string
        role     string
        pass     string
        mockResp *entities.LoginUserResponseModel
        mockErr  error
        wantErr  bool
        wantMsg  string
    }{
        {
            name:  "success (owner)", // <-- เปลี่ยนชื่อเล็กน้อย
            email: "a@b.com", role: "owner", pass: plain,
            mockResp: &entities.LoginUserResponseModel{UserID: "u1", Password: hashed},
            mockErr:  nil, wantErr: false,
        },
        
        // --- โค้ดที่เพิ่มเข้ามา ---
        {
            name:  "success (admin)",
            email: "admin@lama.com", role: "admin", pass: plain,
            mockResp: &entities.LoginUserResponseModel{UserID: "u-admin", Password: hashed},
            mockErr:  nil, wantErr: false,
        },
        {
            name:  "success (doctor)",
            email: "doc@lama.com", role: "doctor", pass: plain,
            mockResp: &entities.LoginUserResponseModel{UserID: "u-doc", Password: hashed},
            mockErr:  nil, wantErr: false,
        },
        {
            name:  "success (caretaker)",
            email: "care@lama.com", role: "caretaker", pass: plain,
            mockResp: &entities.LoginUserResponseModel{UserID: "u-care", Password: hashed},
            mockErr:  nil, wantErr: false,
        },
        // --- จบส่วนที่เพิ่ม ---

        {
            name:  "invalid password",
            email: "a@b.com", role: "owner", pass: "wrong",
            mockResp: &entities.LoginUserResponseModel{UserID: "u1", Password: hashed},
            mockErr:  nil, wantErr: true, wantMsg: "invalid password",
        },
        {
            name:    "repo error",
            email:   "a@b.com", role: "owner", pass: plain,
            mockResp: nil, mockErr: errors.New("db error"), wantErr: true,
        },
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            // Mock จะถูกตั้งค่าตามค่าใน tc (เช่น tc.email, tc.role) ของแต่ละรอบ
            mockUsers.EXPECT().
                FindByEmailAndRole(tc.email, tc.role).
                Return(tc.mockResp, tc.mockErr).
                Times(1)

            got, err := sv.Login(tc.role, entities.LoginUserRequestModel{
                Email:    tc.email,
                Password: tc.pass,
            })

            if tc.wantErr {
                if err == nil {
                    t.Fatalf("expected error but got nil")
                }
                if tc.wantMsg != "" && err.Error() != tc.wantMsg {
                    t.Fatalf("expected error %q, got %q", tc.wantMsg, err.Error())
                }
                return
            }
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got.UserID != tc.mockResp.UserID {
                t.Fatalf("unexpected user id: want %s got %s", tc.mockResp.UserID, got.UserID)
            }
        })
    }
}

func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockIUsersRepository(ctrl)
	mockOwner := mocks.NewMockIOwnerRepository(ctrl)
	mockCaretaker := mocks.NewMockICaretakerRepository(ctrl)
	mockDoctor := mocks.NewMockIDoctorRepository(ctrl)

	sv := &authService{
		UsersRepository:     mockUsers,
		OwnerRepository:     mockOwner,
		CaretakerRepository: mockCaretaker,
		DoctorRepository:    mockDoctor,
	}

	baseBirth := time.Date(1990, time.March, 9, 0, 0, 0, 0, time.UTC)
	ownerSpend := decimal.NewFromInt(45)

	baseUserReq := entities.CreatedUserModel{
		Email:           "base@lama.com",
		Password:        "P@ss",
		Name:            "Base User",
		BirthDate:       baseBirth,
		TelephoneNumber: "0123456789",
		Address:         "Base Address",
	}
	caretakerReq := baseUserReq
	caretakerReq.Specialization = "grooming"
	doctorReq := baseUserReq
	doctorReq.LicenseNumber = "LIC-007"

	tests := []struct {
		name    string
		role    string
		req     entities.CreatedUserModel
		userRes *entities.UserDataModel
		userErr error
		setup   func(id string)
		wantErr string
		check   func(t *testing.T, got *entities.UserDataModel)
	}{
		{
			name:   "admin success (roleData path)",
			role:   "admin",
			req:    baseUserReq,
			userRes: &entities.UserDataModel{UserID: "admin-1"},
			check: func(t *testing.T, got *entities.UserDataModel) {
				if got.UserID != "admin-1" {
					t.Fatalf("unexpected id: %s", got.UserID)
				}
			},
		},
		{
			name: "owner success",
			role: "owner",
			req:  baseUserReq,
			userRes: &entities.UserDataModel{UserID: "owner-1"},
			setup: func(id string) {
				mockOwner.EXPECT().
					InsertOwner(id).
					Return(&entities.UserDataModel{UserID: id, TotalSpending: ownerSpend}, nil)
			},
			check: func(t *testing.T, got *entities.UserDataModel) {
				if !got.TotalSpending.Equal(ownerSpend) {
					t.Fatalf("expected spending %s, got %s", ownerSpend, got.TotalSpending)
				}
			},
		},
		{
			name: "doctor success",
			role: "doctor",
			req:  doctorReq,
			userRes: &entities.UserDataModel{UserID: "doc-1"},
			setup: func(id string) {
				mockDoctor.EXPECT().
					InsertDoctor(id, "LIC-007").
					Return(&entities.UserDataModel{UserID: id, LicenseNumber: "LIC-007"}, nil)
			},
		},
		{
			name: "caretaker success",
			role: "caretaker",
			req:  caretakerReq,
			userRes: &entities.UserDataModel{UserID: "care-1"},
			setup: func(id string) {
				mockCaretaker.EXPECT().
					InsertCaretaker(id, "grooming").
					Return(&entities.UserDataModel{UserID: id, Specialization: "grooming"}, nil)
			},
		},
		{
			name: "users repo error",
			role: "owner",
			req:  baseUserReq,
			userErr: errors.New("insert user failed"),
			wantErr: "insert user failed",
		},
		{
			name: "doctor repo error bubbles up",
			role: "doctor",
			req:  doctorReq,
			userRes: &entities.UserDataModel{UserID: "doc-1"},
			setup: func(id string) {
				mockDoctor.EXPECT().
					InsertDoctor(id, "LIC-007").
					Return(nil, errors.New("doctor repo failed"))
			},
			wantErr: "doctor repo failed",
		},
		{
			name: "caretaker repo error bubbles up",
			role: "caretaker",
			req:  caretakerReq,
			userRes: &entities.UserDataModel{UserID: "care-1"},
			setup: func(id string) {
				mockCaretaker.EXPECT().
					InsertCaretaker(id, "grooming").
					Return(nil, errors.New("caretaker repo failed"))
			},
			wantErr: "caretaker repo failed",
		},
		{
			name: "owner repo error bubbles up",
			role: "owner",
			req:  baseUserReq,
			userRes: &entities.UserDataModel{UserID: "owner-1"},
			setup: func(id string) {
				mockOwner.EXPECT().
					InsertOwner(id).
					Return(nil, errors.New("owner repo failed"))
			},
			wantErr: "owner repo failed",
		},
		{
			name: "invalid role",
			role: "guest",
			req:  baseUserReq,
			userRes: &entities.UserDataModel{UserID: "guest-1"},
			wantErr: "role is required",
		},
		{
			name: "foreign key mismatch",
			role: "doctor",
			req:  doctorReq,
			userRes: &entities.UserDataModel{UserID: "doc-1"},
			setup: func(id string) {
				mockDoctor.EXPECT().
					InsertDoctor(id, "LIC-007").
					Return(&entities.UserDataModel{UserID: "other-id"}, nil)
			},
			wantErr: "invalid foreign key user_id",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockUsers.EXPECT().
				InsertUser(tc.role, tc.req).
				Return(tc.userRes, tc.userErr).
				Times(1)

			if tc.userErr == nil && tc.setup != nil {
				tc.setup(tc.userRes.UserID)
			}

			got, err := sv.Register(tc.role, tc.req)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("want err %q, got %v", tc.wantErr, err)
				}
				if got != nil {
					t.Fatalf("expected nil result, got %#v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got == nil {
				t.Fatalf("expected user data but got nil")
			}
			if got.UserID != tc.userRes.UserID {
				t.Fatalf("want user id %s got %s", tc.userRes.UserID, got.UserID)
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}
