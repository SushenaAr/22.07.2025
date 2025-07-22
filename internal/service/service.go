package service

import (
	"archive/zip"
	"awesomeProject/internal/config"
	"awesomeProject/internal/repository"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ошибки, логи и ответы систематизировать
type ServiceI interface {
	CreateTaskForZIPAchieve(c *gin.Context)
	AddFileIntoTask(c *gin.Context)
	GetStatusTask(c *gin.Context)
	GetZip(c *gin.Context)
}

type Service struct {
	repo      repository.RepositoryI
	semaphore chan struct{}
	cfg       *config.Config
}

func NewService(repo repository.RepositoryI, cfg *config.Config) *Service {
	return &Service{repo: repo,
		semaphore: make(chan struct{}, config.MaxTasks),
		cfg:       cfg,
	}
}

func (s *Service) GetZip(c *gin.Context) {
	filename := c.Param("filename")

	// extend checker
	if filepath.Ext(filename) != ".zip" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .zip files are allowed"})
		return
	}

	//abs to zip folder
	basePath := filepath.Join("resources", "zip")
	filePath := filepath.Join(basePath, filename)

	// file is not exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.FileAttachment(filePath, filename)
}

// CreateTaskForZIPAchieve Create task for pushing url filesw
func (s *Service) CreateTaskForZIPAchieve(c *gin.Context) {
	logger := c.Request.Context().Value("logger").(*slog.Logger)
	logger.Info("Creating task")

	select {
	case s.semaphore <- struct{}{}:
		ID := uuid.New().String()
		s.repo.CreateTask(ID)
		logger.Info("Task created")
		c.JSON(http.StatusCreated, gin.H{"ID": ID})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "server is busy"})
	}
}

// AddFileIntoTask Add url into task
func (s *Service) AddFileIntoTask(c *gin.Context) {
	logger := c.Request.Context().Value("logger").(*slog.Logger)
	logger.Info("creating file")

	ID := c.Query("id")
	if ID == "" {
		logger.Debug("id is not received")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is not received"})
		return
	}

	URL := c.Query("url")
	if URL == "" {
		logger.Debug("url is not received")
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is not received"})
		return
	}

	//status checker
	status, err := s.repo.Status(ID)
	if err != nil {
		logger.Debug(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if status == config.MaxFilesInTask {
		logger.Debug("max count files")
		c.JSON(http.StatusBadRequest, gin.H{"error": "count is full"})
		return
	}

	//create in storage
	fileName, err := createFile(URL, ID)
	if err != nil {
		if err.Error() != "network error" {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		logger.Warn("network error")
	}

	//create in db
	if err := s.repo.AddFile(ID, fileName); err != nil {
		logger.Debug("not fonud")
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	//response
	if fileName != "" {
		logger.Info("file created")
		c.JSON(http.StatusOK, gin.H{"ok": "file created"})
		return
	}
	logger.Info("file is not available")
	c.JSON(http.StatusOK, gin.H{"ok": "file is not available"})

}

// GetStatusTask Get status once task
func (s *Service) GetStatusTask(c *gin.Context) {
	logger := c.Request.Context().Value("logger").(*slog.Logger)
	logger.Info("status task")

	ID := c.Query("id")
	if ID == "" {
		logger.Debug("id is not received")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is not received"})
		return
	}
	//check status
	status, err := s.repo.Status(ID)
	if err != nil {
		logger.Debug(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	if status < config.MaxFilesInTask {
		logger.Debug("status: < maxfilesintask")
		c.JSON(http.StatusOK, gin.H{"status": status})
		return
	}

	//create zip
	path, err := createZip(ID)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	//releasing the semaphore
	<-s.semaphore

	c.JSON(http.StatusOK, gin.H{"status": config.MaxFilesInTask,
		"url": s.cfg.Protocol + "://" + s.cfg.AdvertisedAddr + "/resources" + "/zip/" + path})
	logger.Info("status is ok")
}

// isAllowedExtension format checker
func isAllowedExtension(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	return ext == ".jpeg" || ext == ".pdf"
}

// createFile create file .jpeg|.pdf in resources/downloads
// and concatenate unique suffix with file name
func createFile(url string, id string) (string, error) {
	//format path
	fileName := filepath.Base(url)

	// extension check
	if !isAllowedExtension(fileName) {
		return "", fmt.Errorf("file extension %q not allowed", filepath.Ext(fileName))
	}

	baseDir := filepath.Join("resources", "downloads", id)
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		return "", err
	}
	targetPath := getUniqueFilePath(baseDir, fileName)

	//receive file from other server
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.New("network error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "network err", err
	}

	//empty file
	out, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	//copy stream in file
	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}

	return filepath.Base(targetPath), nil
}

// getUniqueFilePath append suffix _n
func getUniqueFilePath(dir, fileName string) string {
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	path := filepath.Join(dir, fileName)
	counter := 1
	for {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path // don't have file - ok
		}
		// else append suffix
		newName := fmt.Sprintf("%s_%d%s", base, counter, ext)
		path = filepath.Join(dir, newName)
		counter++
	}
}

// createZip archive downloads/{id} to zip/{id}.zip
func createZip(id string) (string, error) {
	// Paths
	sourceDir := filepath.Join("resources", "downloads", id)
	zipDir := filepath.Join("resources", "zip")
	zipFilePath := filepath.Join(zipDir, id+".zip")

	// creating zip-file
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create a zip file: %w", err)
	}
	defer zipFile.Close()

	// creating zip-archive
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk error: %w", err)
		}
		if info.IsDir() || !info.Mode().IsRegular() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("error path: %w", err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("error header: %w", err)
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("error writer: %w", err)
		}

		// Читаю вручную буферами
		file, err := os.OpenFile(path, os.O_RDONLY, 0)
		if err != nil {
			return fmt.Errorf("failed open file: %w", err)
		}
		defer file.Close()

		buf := make([]byte, 32*1024)
		for {
			n, readErr := file.Read(buf)
			if n > 0 {
				if _, writeErr := writer.Write(buf[:n]); writeErr != nil {
					return fmt.Errorf("failed write in file: %w", writeErr)
				}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return fmt.Errorf("failed read: %w", readErr)
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed create zip: %w", err)
	}

	return id + ".zip", nil
}
