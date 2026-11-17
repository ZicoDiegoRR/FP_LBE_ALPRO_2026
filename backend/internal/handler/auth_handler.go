package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pmo-backend/internal/config"
	"pmo-backend/internal/model"
	"pmo-backend/internal/repository"
	"pmo-backend/internal/security"
)

// AuthHandler menangani alur HTTP untuk autentikasi pengguna
type AuthHandler struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

// NewAuthHandler membuat instance baru AuthHandler
func NewAuthHandler(userRepo *repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// RegisterInput mendefinisikan skema validasi request registrasi
type RegisterInput struct {
	Username string `json:"username" binding:"required,min=4,max=50,alphanum"`
	Password string `json:"password" binding:"required,min=8,max=100"`
	Role     string `json:"role" binding:"omitempty,oneof=user admin"`
}

// LoginInput mendefinisikan skema validasi request login
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register menangani pembuatan akun pengguna baru
func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterInput

	// 1. Validasi format dasar input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validasi gagal: format input tidak sesuai",
			"details": err.Error(),
		})
		return
	}

	// 1b. Validasi kekuatan password (NIST Standard)
	if err := security.ValidatePasswordStrength(input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 2. Cek apakah username sudah pernah terdaftar
	existingUser, err := h.userRepo.GetByUsername(input.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa ketersediaan username"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username sudah digunakan, silakan pilih username lain"})
		return
	}

	// 3. Hash password menggunakan Bcrypt (One-Way Salted Hash)
	hashedPassword, err := security.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses keamanan password"})
		return
	}

	// 4. Default role ke 'user' jika tidak dispesifikasikan
	role := "user"
	if input.Role != "" {
		role = input.Role
	}

	// 5. Simpan pengguna baru ke database
	newUser := model.User{
		Username: input.Username,
		Password: hashedPassword,
		Role:     role,
	}

	if err := h.userRepo.CreateUser(&newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan akun pengguna"})
		return
	}

	// Berhasil (Password otomatis tersembunyi karena tag json:"-" di struct User)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registrasi akun berhasil!",
		"user": gin.H{
			"id":       newUser.ID,
			"username": newUser.Username,
			"role":     newUser.Role,
		},
	})
}

// Login memvalidasi kredensial dan menerbitkan token JWT
func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username dan password wajib diisi"})
		return
	}

	// 1. Ambil data pengguna dari database berdasarkan username
	user, err := h.userRepo.GetByUsername(input.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan internal server"})
		return
	}

	// Mencegah Username Enumeration Attack: Jika user tidak ada atau password salah,
	// berikan respon kesalahan yang generik ("Kredensial tidak valid")
	if user == nil || !security.CheckPassword(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kombinasi username atau password salah"})
		return
	}

	// 2. Terbitkan token JWT bertanda tangan digital
	token, err := security.GenerateToken(user.ID, user.Username, user.Role, h.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menerbitkan token autentikasi"})
		return
	}

	// Kembalikan token ke klien
	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil!",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// GetProfile mengembalikan data profil pengguna saat ini berdasarkan token JWT yang valid
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pengguna tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		},
	})
}