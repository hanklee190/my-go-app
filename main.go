package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// User 模型
type User struct {
    gorm.Model
    Username     string `json:"username" gorm:"column:username"`
    Email        string `json:"email" gorm:"unique;column:email"`
    PasswordHash string `json:"password_hash" gorm:"column:password_hash"`
    IsActive     bool   `json:"is_active" gorm:"column:is_active"`
}

// 指定表名
func (User) TableName() string {
    return "app_go_users"
}

var db *gorm.DB

func main() {
    // 加载 .env 文件
    err := godotenv.Load()
    if err != nil {
        panic("failed to load .env file")
    }

    // 从环境变量构造 DSN
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_SSLMODE"),
        os.Getenv("DB_TIMEZONE"),
    )

    // 初始化数据库，启用日志并调整慢查询阈值
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.New(
            log.New(os.Stdout, "\r\n", log.LstdFlags),
            logger.Config{
                SlowThreshold: time.Second,
                LogLevel:      logger.Info,
                Colorful:      true,
            },
        ),
    })
    
    if err != nil {
        panic("failed to connect database")
    }

    // 自动迁移（可注释掉以提高启动速度）
    db.AutoMigrate(&User{})

    // 配置连接池并确保关闭
    sqlDB, err := db.DB()
    if err != nil {
        panic("failed to get sql.DB")
    }
    defer sqlDB.Close()
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)

    // 初始化 Gin 路由
    r := gin.Default()
    r.SetTrustedProxies([]string{"127.0.0.1"})

    // 根路径
    r.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "欢迎使用用户 API"})
    })

    // 定义 CRUD 路由
    r.POST("/users", createUser)
    r.GET("/users", getUsers)
    r.GET("/users/:id", getUser)
    r.PUT("/users/:id", updateUser)
    r.DELETE("/users/:id", deleteUser)

    // 启动服务
    r.Run(":8080")
}

func createUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if user.Username == "" || user.Email == "" || user.PasswordHash == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "username, email, and password_hash are required"})
        return
    }

    if err := db.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, user)
}

func getUsers(c *gin.Context) {
    var users []User
    if err := db.Find(&users).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
        return
    }
    c.JSON(http.StatusOK, users)
}

func getUser(c *gin.Context) {
    id := c.Param("id")
    var user User

    if err := db.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, user)
}

func updateUser(c *gin.Context) {
    id := c.Param("id")
    var user User

    if err := db.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    var updatedUser User
    if err := c.ShouldBindJSON(&updatedUser); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user.Username = updatedUser.Username
    user.Email = updatedUser.Email
    user.PasswordHash = updatedUser.PasswordHash
    user.IsActive = updatedUser.IsActive

    if err := db.Save(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
        return
    }

    c.JSON(http.StatusOK, user)
}

func deleteUser(c *gin.Context) {
    id := c.Param("id")
    var user User

    if err := db.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    if err := db.Delete(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}