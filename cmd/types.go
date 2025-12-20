package cmd

import (
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/pkg/queue"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
	"github.com/Infinite-Locus-Product/thums_up_backend/vendors"
)

type Server struct {
	db             *gorm.DB
	cfg            *config.Config
	firebaseClient *vendors.FirebaseClient
	infobipClient  *vendors.InfobipClient
	gcsService     utils.GCSService
	workerPool     *queue.WorkerPool
	repositories   *Repositories
	handlers       *Handlers
}

type Repositories struct {
	user         repository.UserRepository
	otp          repository.OTPRepository
	refreshToken repository.RefreshTokenRepository
	address      repository.GenericRepository[entities.Address]
	avatar       repository.GenericRepository[entities.Avatar]
	state        repository.StateRepository
	city         repository.CityRepository
	pinCode      repository.PinCodeRepository
	question     repository.QuestionRepository
	thunderSeat  repository.ThunderSeatRepository
	winner       repository.WinnerRepository
}

type Handlers struct {
	auth        *handlers.AuthHandler
	profile     *handlers.ProfileHandler
	address     *handlers.AddressHandler
	avatar      *handlers.AvatarHandler
	question    *handlers.QuestionHandler
	thunderSeat *handlers.ThunderSeatHandler
	winner      *handlers.WinnerHandler
}
