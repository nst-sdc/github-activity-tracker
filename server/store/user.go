package store

import (
	"github-activity-tracker/models"

	"gorm.io/gorm"
)

type UserStore interface {
	Save(u models.User) (models.User, error)
	GetByID(id string) (models.User, error)
	GetByGitHubID(githubID string) (models.User, error)
	GetAll() ([]models.User, error)
	Delete(id string) error
}

type InMemoryStore struct {
	data map[string]models.User
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]models.User),
	}
}

func (s *InMemoryStore) Save(u models.User) (models.User, error) {
	if u.ID == 0 {
		u.ID = uint(len(s.data) + 1) // Simple ID generation
	}
	s.data[string(rune(u.ID))] = u
	return u, nil
}

func (s *InMemoryStore) GetByID(id string) (models.User, error) {
	if user, exists := s.data[id]; exists {
		return user, nil
	}
	return models.User{}, nil
}

func (s *InMemoryStore) GetByGitHubID(githubID string) (models.User, error) {
	for _, user := range s.data {
		if user.GithubUser == githubID {
			return user, nil
		}
	}
	return models.User{}, nil
}

func (s *InMemoryStore) GetAll() ([]models.User, error) {
	users := make([]models.User, 0, len(s.data))
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

func (s *PostgreSQLStore) Save(u models.User) (models.User, error) {
	err := s.db.Create(&u).Error
	return u, err
}

func (s *PostgreSQLStore) GetByID(id string) (models.User, error) {
	var user models.User
	err := s.db.Where("id = ?", id).First(&user).Error
	return user, err
}

func (s *PostgreSQLStore) GetByGitHubID(githubID string) (models.User, error) {
	var user models.User
	err := s.db.Where("github_user = ?", githubID).First(&user).Error
	return user, err
}

func (s *PostgreSQLStore) GetAll() ([]models.User, error) {
	var users []models.User
	err := s.db.Find(&users).Error
	return users, err
}

func (s *PostgreSQLStore) Delete(id string) error {
	return s.db.Delete(&models.User{}, "id = ?", id).Error
}
