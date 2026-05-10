package http_server

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/Nobodywinsbutme/mangahub/internal/auth"
	"github.com/Nobodywinsbutme/mangahub/internal/database"
	"github.com/Nobodywinsbutme/mangahub/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type libraryRequest struct {
	MangaID        string `json:"manga_id" binding:"required"`
	Status         string `json:"status" binding:"omitempty,oneof=reading completed plan-to-read on-hold dropped"`
	CurrentChapter int    `json:"current_chapter" binding:"min=0"`
}

type progressRequest struct {
	MangaID        string `json:"manga_id" binding:"required"`
	CurrentChapter int    `json:"current_chapter" binding:"required,min=0"`
	Status         string `json:"status" binding:"omitempty,oneof=reading completed plan-to-read on-hold dropped"`
}

func Start(port string) {
	r := gin.Default()

	auth.RegisterRoutes(r)
	registerMangaRoutes(r)
	registerUserRoutes(r)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "mangahub-http",
			"protocols": []string{"HTTP", "TCP", "UDP", "gRPC", "WebSocket"},
		})
	})

	if err := r.Run(":" + port); err != nil {
		panic(err)
	}
}

func registerMangaRoutes(r *gin.Engine) {
	r.GET("/manga", func(c *gin.Context) {
		query := strings.TrimSpace(c.Query("query"))
		genre := strings.TrimSpace(c.Query("genre"))
		status := strings.TrimSpace(c.Query("status"))

		sqlQuery := `SELECT id, title, author, genres, status, total_chapters, description, cover_url FROM manga WHERE 1=1`
		args := []any{}

		if query != "" {
			sqlQuery += ` AND title LIKE ?`
			args = append(args, "%"+query+"%")
		}
		if genre != "" {
			sqlQuery += ` AND genres LIKE ?`
			args = append(args, "%"+genre+"%")
		}
		if status != "" {
			sqlQuery += ` AND status = ?`
			args = append(args, status)
		}
		sqlQuery += ` ORDER BY title LIMIT 100`

		rows, err := database.DB.Query(sqlQuery, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		manga, err := scanMangaRows(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"count": len(manga), "results": manga})
	})

	r.GET("/manga/:id", func(c *gin.Context) {
		manga, err := getMangaByID(c.Param("id"))
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, manga)
	})
}

func registerUserRoutes(r *gin.Engine) {
	users := r.Group("/users")
	users.Use(jwtMiddleware())

	users.POST("/library", func(c *gin.Context) {
		var req libraryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Status == "" {
			req.Status = "reading"
		}

		if _, err := getMangaByID(req.MangaID); err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetString("user_id")
		_, err := database.DB.Exec(`
			INSERT INTO user_progress (user_id, manga_id, current_chapter, status, updated_at)
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(user_id, manga_id) DO UPDATE SET
				current_chapter = excluded.current_chapter,
				status = excluded.status,
				updated_at = CURRENT_TIMESTAMP
		`, userID, req.MangaID, req.CurrentChapter, req.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "manga added to library"})
	})

	users.GET("/library", func(c *gin.Context) {
		rows, err := database.DB.Query(`
			SELECT m.id, m.title, m.author, m.genres, m.status, m.total_chapters, m.description, m.cover_url,
			       p.current_chapter, p.status, p.updated_at
			FROM user_progress p
			JOIN manga m ON m.id = p.manga_id
			WHERE p.user_id = ?
			ORDER BY p.updated_at DESC
		`, c.GetString("user_id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		items := []gin.H{}
		for rows.Next() {
			var manga models.Manga
			var currentChapter int
			var readingStatus string
			var updatedAt string
			if err := rows.Scan(&manga.ID, &manga.Title, &manga.Author, &manga.Genres, &manga.Status, &manga.TotalChapters, &manga.Description, &manga.CoverURL, &currentChapter, &readingStatus, &updatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, gin.H{
				"manga":           manga,
				"current_chapter": currentChapter,
				"reading_status":  readingStatus,
				"updated_at":      updatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"count": len(items), "library": items})
	})

	users.PUT("/progress", func(c *gin.Context) {
		var req progressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Status == "" {
			req.Status = "reading"
		}

		result, err := database.DB.Exec(`
			UPDATE user_progress
			SET current_chapter = ?, status = ?, updated_at = CURRENT_TIMESTAMP
			WHERE user_id = ? AND manga_id = ?
		`, req.CurrentChapter, req.Status, c.GetString("user_id"), req.MangaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "library item not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "progress updated"})
	})
}

func jwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return []byte(auth.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing user_id claim"})
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func scanMangaRows(rows *sql.Rows) ([]models.Manga, error) {
	result := []models.Manga{}
	for rows.Next() {
		var manga models.Manga
		if err := rows.Scan(&manga.ID, &manga.Title, &manga.Author, &manga.Genres, &manga.Status, &manga.TotalChapters, &manga.Description, &manga.CoverURL); err != nil {
			return nil, err
		}
		result = append(result, manga)
	}
	return result, rows.Err()
}

func getMangaByID(id string) (*models.Manga, error) {
	var manga models.Manga
	err := database.DB.QueryRow(`
		SELECT id, title, author, genres, status, total_chapters, description, cover_url
		FROM manga
		WHERE id = ?
	`, id).Scan(&manga.ID, &manga.Title, &manga.Author, &manga.Genres, &manga.Status, &manga.TotalChapters, &manga.Description, &manga.CoverURL)
	return &manga, err
}
