package controllers

import (
	"net/http"
	"os"
	"time"
	"trashure/internal/domain/entity"
	"trashure/internal/infra/postgresql"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Signup(c *gin.Context) {
	//Get the email/pass off req body
	var body struct {
		Name  string `gorm:"not null"` 
 	Email  string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Province    string
    City        string
    District    string
    SubDistrict string
    Address     string
    Points      int `gorm:"default:0"` // Jumlah poin
	DateOfBirth  string `gorm:"not null"`      // Tanggal lahir (format YYYY-MM-DD)        
    BankAccount  string `gorm:"unique;not null"`
	Role        string `gorm:"not null;default:user"`
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	//Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}
	//Create a new user
	user := entity.User{Email: body.Email, Password: string(hash)}
	result := postgresql.DB.Create(&user) // pass pointer of data to Create

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})
		return
	}
	//respond
	c.JSON(http.StatusOK, gin.H{})
}

func Login(c *gin.Context) {
	//Get the email and password off req body
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}
	//Look up the user by email
	var user entity.User
	postgresql.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password",
		})
		return
	}
	//Compare sent in pass with saved user pass hash
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password",
		})
		return
	}
	//Generate a Jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create token",
		})
		return

	}
	// Send it back
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{})
}

func Validate(c *gin.Context){
	user, _ :=c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"message": user,
	})
}
func CreateUser(c *gin.Context) {
    var body struct {
        Name     string `json:"name" binding:"required"`
        Email    string `json:"email" binding:"required"`
        Password string `json:"password" binding:"required"`
        Role     string `json:"role" binding:"required"` // Misalnya: admin, user, banksampah
    }

    // Validasi input
    if err := c.Bind(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    // Buat pengguna baru
    user := entity.User{
        Name:     body.Name,
        Email:    body.Email,
        Password: body.Password, // Pastikan hashing kata sandi dilakukan sebelum disimpan
        Role:     body.Role,
    }

    // Simpan ke database
    if err := postgresql.DB.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "data": user})
}
func UpdateUserByID(c *gin.Context) {
    id := c.Param("id")
    var user entity.User

    // Temukan user berdasarkan ID
    if err := postgresql.DB.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Ambil data dari body request
    var body struct {
        Name        string `json:"name"`
        Province    string `json:"province"`
        City        string `json:"city"`
        District    string `json:"district"`
        SubDistrict string `json:"sub_district"`
        Address     string `json:"address"`
    }

    if err := c.Bind(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    // Update data user
    user.Name = body.Name
    user.Province = body.Province
    user.City = body.City
    user.District = body.District
    user.SubDistrict = body.SubDistrict
    user.Address = body.Address

    if err := postgresql.DB.Save(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "data": user})
}
func GetUserByID(c *gin.Context) {
    id := c.Param("id")
    var user entity.User

    // Cari pengguna berdasarkan ID
    if err := postgresql.DB.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": user})
}
func GetAllUsers(c *gin.Context) {
    var users []entity.User

    // Ambil semua data pengguna dari database
    if err := postgresql.DB.Find(&users).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": users})
}
func DeleteUserByID(c *gin.Context) {
    id := c.Param("id")
    var user entity.User

    // Cari pengguna berdasarkan ID
    if err := postgresql.DB.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Hapus pengguna dari database
    if err := postgresql.DB.Delete(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully", "id": id})
}
// GetUserHistory - Mendapatkan riwayat pengguna
func GetUserHistory(c *gin.Context) {
    id := c.Param("id")
    var histories []entity.UserHistory

    // Cari riwayat berdasarkan UserID
    if err := postgresql.DB.Where("user_id = ?", id).Find(&histories).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user history"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": histories})
}

// AddUserHistory - Menambahkan riwayat pengguna
func AddUserHistory(userID uint, address, day string) error {
    history := entity.UserHistory{
        UserID:  userID,
        Address: address,
        Time:    time.Now(),
        Day:     day,
    }

    if err := postgresql.DB.Create(&history).Error; err != nil {
        return err
    }

    return nil
}
// GetUserRanking - Mendapatkan peringkat pengguna berdasarkan poin
func GetUserRanking(c *gin.Context) {
    var users []entity.User

    // Ambil semua pengguna dan urutkan berdasarkan poin secara descending
    if err := postgresql.DB.Order("points DESC").Find(&users).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user rankings"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": users})
}
