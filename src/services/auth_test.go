package services

import (
    "errors"
    "testing"

    "lama-backend/domain/entities"
    "lama-backend/src/services/mocks"
    "lama-backend/src/utils"

    "github.com/golang/mock/gomock"
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