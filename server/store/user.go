package store

import (
	"github.com/nst-sdc/github-activity-tracker/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStore interface {
	Save(u model.GitId) (model.GitId, error)
	GetByID(id string) (model.GitId, error)
	GetByGitHubID(githubID string) (model.GitId, error)
	GetAll() ([]model.GitId, error)
	Delete(id string) error
}

type InMemoryStore struct {
	data map[string]model.GitId
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]model.GitId),
	}
}

func (s *InMemoryStore) Save(u model.GitId) (model.GitId, error) {
	u.ID = uuid.New().String()
	s.data[u.ID] = u
	return u, nil
}

func (s *InMemoryStore) GetByID(id string) (model.GitId, error) {
	if user, exists := s.data[id]; exists {
		return user, nil
	}
	return model.GitId{}, nil
}

func (s *InMemoryStore) GetByGitHubID(githubID string) (model.GitId, error) {
	for _, user := range s.data {
		if user.GitHubID == githubID {
			return user, nil
		}
	}
	return model.GitId{}, nil
}

func (s *InMemoryStore) GetAll() ([]model.GitId, error) {
	users := make([]model.GitId, 0, len(s.data))
	for _, user := range s.data {
		users = append(users, user)
	}
	return users, nil
}

func (s *InMemoryStore) Delete(id string) error {
	delete(s.data, id)
	return nil
}

// PostgreSQL Store using GORM
type PostgreSQLStore struct {
	db *gorm.DB
}

func NewPostgreSQLStore(db *gorm.DB) *PostgreSQLStore {
	return &PostgreSQLStore{db: db}
}

func (s *PostgreSQLStore) Save(u model.GitId) (model.GitId, error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}

	err := s.db.Create(&u).Error
	return u, err
}

func (s *PostgreSQLStore) GetByID(id string) (model.GitId, error) {
	var user model.GitId
	err := s.db.Where("id = ?", id).First(&user).Error
	return user, err
}

func (s *PostgreSQLStore) GetByGitHubID(githubID string) (model.GitId, error) {
	var user model.GitId
	err := s.db.Where("git_hub_id = ?", githubID).First(&user).Error
	return user, err
}

func (s *PostgreSQLStore) GetAll() ([]model.GitId, error) {
	var users []model.GitId
	err := s.db.Find(&users).Error
	return users, err
}

func (s *PostgreSQLStore) Delete(id string) error {
	return s.db.Delete(&model.GitId{}, "id = ?", id).Error
}
