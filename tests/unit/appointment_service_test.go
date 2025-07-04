package unit

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	apiModels "citynext/internal/api/models"
	"citynext/internal/database"
	dbModels "citynext/internal/database/models"
	"citynext/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mock implementation of AppointmentRepository
type MockAppointmentRepository struct {
	mock.Mock
}

func (m *MockAppointmentRepository) Create(ctx context.Context, appointment *dbModels.Appointment) error {
	args := m.Called(ctx, appointment)
	return args.Error(0)
}

func (m *MockAppointmentRepository) GetByDate(ctx context.Context, date apiModels.Date) (*dbModels.Appointment, error) {
	args := m.Called(ctx, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dbModels.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) ExistsByDate(ctx context.Context, date apiModels.Date) (bool, error) {
	args := m.Called(ctx, date)
	return args.Bool(0), args.Error(1)
}

// mock implementation of HolidayServiceInterface
type MockHolidayService struct {
	mock.Mock
}

func (m *MockHolidayService) IsPublicHoliday(ctx context.Context, date apiModels.Date) (bool, error) {
	args := m.Called(ctx, date)
	return args.Bool(0), args.Error(1)
}

func (m *MockHolidayService) ValidateDate(ctx context.Context, date apiModels.Date) error {
	args := m.Called(ctx, date)
	return args.Error(0)
}

func TestAppointmentService_CreateAppointment(t *testing.T) {

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	createDate := func(daysFromNow int) apiModels.Date {
		return apiModels.Date{Time: time.Now().AddDate(0, 0, daysFromNow).UTC()}
	}

	tests := []struct {
		name           string
		request        *services.CreateAppointmentRequest
		setupMocks     func(*MockAppointmentRepository, *MockHolidayService)
		expectedError  error
		expectedResult *dbModels.Appointment
	}{
		{
			name: "Success",
			request: &services.CreateAppointmentRequest{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: createDate(7), // 7 days from now
			},
			setupMocks: func(repo *MockAppointmentRepository, holiday *MockHolidayService) {
				holiday.On("ValidateDate", mock.Anything, mock.Anything).Return(nil)
				repo.On("ExistsByDate", mock.Anything, mock.Anything).Return(false, nil)
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
			expectedResult: &dbModels.Appointment{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: createDate(7),
			},
		},
		{
			name: "Invalid Input - Missing First Name",
			request: &services.CreateAppointmentRequest{
				FirstName: "",
				LastName:  "Doe",
				VisitDate: createDate(7),
			},
			setupMocks:     func(repo *MockAppointmentRepository, holiday *MockHolidayService) {},
			expectedError:  services.ErrInvalidInput,
			expectedResult: nil,
		},
		{
			name: "Invalid Input - Missing Last Name",
			request: &services.CreateAppointmentRequest{
				FirstName: "John",
				LastName:  "",
				VisitDate: createDate(7),
			},
			setupMocks:     func(repo *MockAppointmentRepository, holiday *MockHolidayService) {},
			expectedError:  services.ErrInvalidInput,
			expectedResult: nil,
		},
		{
			name: "Date in Past",
			request: &services.CreateAppointmentRequest{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: createDate(-1), // yesterday
			},
			setupMocks: func(repo *MockAppointmentRepository, holiday *MockHolidayService) {
				holiday.On("ValidateDate", mock.Anything, mock.Anything).Return(services.ErrDateInPast)
			},
			expectedError:  services.ErrDateInPast,
			expectedResult: nil,
		},
		{
			name: "Date is Holiday",
			request: &services.CreateAppointmentRequest{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: createDate(7),
			},
			setupMocks: func(repo *MockAppointmentRepository, holiday *MockHolidayService) {
				holiday.On("ValidateDate", mock.Anything, mock.Anything).Return(services.ErrDateIsHoliday)
			},
			expectedError:  services.ErrDateIsHoliday,
			expectedResult: nil,
		},
		{
			name: "Date is Weekend",
			request: &services.CreateAppointmentRequest{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: createDate(7),
			},
			setupMocks: func(repo *MockAppointmentRepository, holiday *MockHolidayService) {
				holiday.On("ValidateDate", mock.Anything, mock.Anything).Return(services.ErrDateIsWeekend)
			},
			expectedError:  services.ErrDateIsWeekend,
			expectedResult: nil,
		},
		{
			name: "Duplicate Appointment",
			request: &services.CreateAppointmentRequest{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: createDate(7),
			},
			setupMocks: func(repo *MockAppointmentRepository, holiday *MockHolidayService) {
				holiday.On("ValidateDate", mock.Anything, mock.Anything).Return(nil)
				repo.On("ExistsByDate", mock.Anything, mock.Anything).Return(true, nil)
			},
			expectedError:  database.ErrDuplicateAppointment,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(MockAppointmentRepository)
			mockHoliday := new(MockHolidayService)
			tt.setupMocks(mockRepo, mockHoliday)

			service := services.NewAppointmentService(mockRepo, mockHoliday, logger)

			result, err := service.CreateAppointment(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.FirstName, result.FirstName)
				assert.Equal(t, tt.expectedResult.LastName, result.LastName)
				assert.Equal(t, tt.expectedResult.VisitDate.String(), result.VisitDate.String())
			}

			mockRepo.AssertExpectations(t)
			mockHoliday.AssertExpectations(t)
		})
	}
}
