package repository

import (
	"awesomeProject/internal/config"
	"awesomeProject/internal/model"
	"errors"
	"sync"
)

type RepositoryI interface {
	AddFile(requestID string, url string) error
	Status(requestID string) (int, error)
	CreateTask(requestID string)
}

type Repository struct {
	db      map[string]*model.Task
	dbMutex sync.RWMutex
}

func NewRepository(database map[string]*model.Task) *Repository {
	return &Repository{
		db: database,
	}
}

func (r *Repository) AddFile(ID string, nameFile string) error {
	r.dbMutex.Lock()
	defer r.dbMutex.Unlock()
	if task, ok := r.db[ID]; ok {
		task.Files = append(task.Files, nameFile)
		return nil
	}
	return errors.New("task not founded")
}

func (r *Repository) Status(ID string) (int, error) {
	r.dbMutex.RLock()
	defer r.dbMutex.RUnlock()
	obj, ok := r.db[ID]

	if !ok {
		return 0, errors.New("task not founded")
	}
	return len(obj.Files), nil
}

func (r *Repository) CreateTask(ID string) {
	r.dbMutex.Lock()
	r.db[ID] = &model.Task{
		Files: make([]string, 0, config.MaxFilesInTask),
	}
	r.dbMutex.Unlock()
}
