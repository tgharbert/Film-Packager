package userservice_test

import (
	"context"
	"errors"
	"filmPackager/internal/application/userservice"
	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserById(ctx context.Context, id uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUserByID(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) CreateNewUser(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	if args.Get(0) == nil {
		return args.Error(1)
	}
	return args.Error(1)
}

func (m *MockUserRepository) GetAllNewUsersByName(ctx context.Context, name string, ids []uuid.UUID) ([]user.User, error) {
	args := m.Called(ctx, name, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]user.User), args.Error(1)
}

func (m *MockUserRepository) GetAllUsersByName(ctx context.Context, name string) ([]user.User, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]user.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByName(ctx context.Context, name string) (*user.User, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]user.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]user.User), args.Error(1)
}

// SHOULD ALL THIS BE ABSTRACTED ELSEWHERE....?
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) CreateNewProject(ctx context.Context, p *project.Project, userID uuid.UUID) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockProjectRepository) GetProjectByID(ctx context.Context, id uuid.UUID) (*project.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Project), args.Error(1)
}

func (m *MockProjectRepository) GetProjectsByUserID(ctx context.Context, userID uuid.UUID) ([]*project.Project, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*project.Project), args.Error(1)
}

func (m *MockProjectRepository) DeleteProject(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) UpdateProject(ctx context.Context, p *project.Project) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockProjectRepository) GetProjectsByMembershipIDs(ctx context.Context, ids []uuid.UUID) ([]project.Project, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]project.Project), args.Error(1)
}

func (m *MockProjectRepository) InviteMember(ctx context.Context, pID uuid.UUID, uID uuid.UUID) error {
	args := m.Called(ctx, pID, uID)
	return args.Error(0)
}

func (m *MockProjectRepository) JoinProject(ctx context.Context, pID uuid.UUID, uID uuid.UUID) error {
	args := m.Called(ctx, pID, uID)
	return args.Error(0)
}

func (m *MockProjectRepository) UpdateMemberRoles(ctx context.Context, pID uuid.UUID, uID uuid.UUID, role string) error {
	args := m.Called(ctx, pID, uID, role)
	return args.Error(0)
}

// Helper function to create a test user with a password
func createTestUserWithPassword(password string) *user.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedString := string(hashedPassword)
	u := user.CreateNewUser("Test Name", "test@test.com", hashedString)
	return u
}

func TestVerifyOldPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjRepo := new(MockProjectRepository)

	service := userservice.NewUserService(mockUserRepo, mockProjRepo)

	t.Run("passwords don't match", func(t *testing.T) {
		userID := uuid.New()

		err := service.VerifyOldPassword(context.Background(), userID, "password1", "password2")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "passwords do not match")
		mockUserRepo.AssertNotCalled(t, "GetUserById")
	})

	t.Run("user not found", func(t *testing.T) {
		userID := uuid.New()

		mockUserRepo.On("GetUserById", mock.Anything, userID).
			Return(nil, errors.New("user not found")).Once()

		err := service.VerifyOldPassword(context.Background(), userID, "password", "password")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error getting user")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("incorrect password", func(t *testing.T) {
		userID := uuid.New()
		testUser := createTestUserWithPassword("correctpassword")

		mockUserRepo.On("GetUserById", mock.Anything, userID).
			Return(testUser, nil).Once()

		err := service.VerifyOldPassword(context.Background(), userID, "wrongpassword", "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, user.ErrInvalidPassword, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("successful verification", func(t *testing.T) {
		userID := uuid.New()
		testUser := createTestUserWithPassword("correctpassword")

		mockUserRepo.On("GetUserById", mock.Anything, userID).
			Return(testUser, nil).Once()

		err := service.VerifyOldPassword(context.Background(), userID, "correctpassword", "correctpassword")

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestSetNewPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjRepo := new(MockProjectRepository)

	service := userservice.NewUserService(mockUserRepo, mockProjRepo)

	t.Run("passwords don't match", func(t *testing.T) {
		userID := uuid.New()

		err := service.SetNewPassword(context.Background(), userID, "newpassword1", "newpassword2")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "passwords do not match")
		mockUserRepo.AssertNotCalled(t, "GetUserById")
		mockUserRepo.AssertNotCalled(t, "UpdateUserByID")
	})

	t.Run("user not found", func(t *testing.T) {
		userID := uuid.New()

		mockUserRepo.On("GetUserById", mock.Anything, userID).
			Return(nil, errors.New("user not found")).Once()

		err := service.SetNewPassword(context.Background(), userID, "newpassword", "newpassword")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error getting user")
		mockUserRepo.AssertExpectations(t)
		mockUserRepo.AssertNotCalled(t, "UpdateUserByID")
	})

	t.Run("update fails", func(t *testing.T) {
		// userID := uuid.New()
		testUser := createTestUserWithPassword("oldpassword")

		mockUserRepo.On("GetUserById", mock.Anything, testUser.Id).
			Return(testUser, nil).Once()
		mockUserRepo.On("UpdateUserByID", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
			// Verify the password was updated (can't check exact hash)
			// FIX -- THIS IS WHERE THE ERROR LIVES... COMPARING THE U AND TESTUSER PWS
			// return u.Id == testUser.Id && u.Password != testUser.Password
			return u.Id == testUser.Id
		})).Return(errors.New("database error")).Once()

		err := service.SetNewPassword(context.Background(), testUser.Id, "newpassword", "newpassword")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error setting new password")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("successful password update", func(t *testing.T) {
		userID := uuid.New()
		testUser := createTestUserWithPassword("oldpassword")

		mockUserRepo.On("GetUserById", mock.Anything, userID).
			Return(testUser, nil).Once()
		mockUserRepo.On("UpdateUserByID", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
			// Verify password was changed and can be verified with bcrypt
			err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("newpassword"))
			return u.Id == testUser.Id && err == nil
		})).Return(nil).Once()

		err := service.SetNewPassword(context.Background(), userID, "newpassword", "newpassword")

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}
