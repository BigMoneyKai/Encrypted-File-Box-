package service

import (
	"errors"
	"fmt"
	"github.com/Kaikai20040827/graduation/internal/model"
	"github.com/Kaikai20040827/graduation/internal/pkg"
	"time"

	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	fmt.Println("✓ Creating a new user service done")
	return &UserService{db: db}

}

func (s *UserService) CreateUser(username string, email string, password string) (*model.User, error) {
	//检查邮箱是否注册
	var count int64
	s.db.Model(&model.User{}).Where("email = ?", email).Count(&count)

	//邮箱已注册
	if count > 0 {
		return nil, errors.New("email already exists")
	}

	//密码哈希化
	hashedPwd, err := pkg.HashPassword(password)
	if err != nil {
		return nil, err
	}

	//创建用户
	user := &model.User{
		Email:     email,
		Username:  username,
		Password:  hashedPwd,
		CreatedAt: time.Now(),
	}

	//数据库创建用户
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	// 将返回给调用方的用户对象中的密码字段清空，
	// 以避免将密码（即使是哈希后的密码）暴露给外部调用者
	user.Password = ""

	return user, nil
}

func (s *UserService) DeleteUser(username string, inputPassword string) error {
	var count int64
	s.db.Model(&model.User{}).Where("Username = ?", username).Count(&count)

	if count == 0 {
		return errors.New("user does not exist")
	}

	var user *model.User
	s.db.Model(&model.User{}).Where("Username = ?", username).Find(&user)
	hashedPassword, _ := pkg.HashPassword(inputPassword)
	if user.Password != hashedPassword {
		return errors.New("incorrect password")
	}

	if err := s.db.Delete(user); err != nil {
		return errors.New("failed to delete")
	}

	return nil
}

// oldPassword should be input by users
func (s *UserService) ChangePassword(userID uint, oldPassword string, newPassword string) error {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user does not exist")
	}

	if err := pkg.CheckPassword(user.Password, oldPassword); err != nil {
		return errors.New("old password incorrect")
	}

	newHashedPassword, err := pkg.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = newHashedPassword
	user.UpdatedAt = time.Now()

	return s.db.Save(&user).Error
}

func (s *UserService) ChangeUsername(email string, newUsername string) error {
	var count int64
	s.db.Model(&model.User{}).Where("Username = ?", newUsername).Count(&count)

	if count == 0 {
		return errors.New("user does not exist")
	}

	var user *model.User
	s.db.Model(&model.User{}).Where("email = ?", email).Find(&user)

	user.Username = newUsername
	user.UpdatedAt = time.Now()

	return s.db.Save(&user).Error
}

func (s *UserService) Authenticate(email, password string) (*model.User, error) {
	var u model.User
	if err := s.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	if err := pkg.CheckPassword(u.Password, password); err != nil {
		return nil, errors.New("invalid credentials")
	}
	u.Password = ""
	return &u, nil
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	var u model.User
	if err := s.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	u.Password = ""
	return &u,
		nil
}

func (s *UserService) GetByUsername(username string) (*model.User, error) {
	var u model.User
	if err := s.db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	u.Password = ""
	return &u, nil
}

func (s *UserService) UpdateProfile(id uint, username string, email string) (*model.User, error) {
	var u model.User
	if err := s.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	u.Username = username
	if email != "" && email != u.Email {
		var existing model.User
		if err := s.db.Where("email = ?", email).First(&existing).Error; err == nil {
			return nil, errors.New("email already in use")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		u.Email = email
	}
	if err := s.db.Save(&u).Error; err != nil {
		return nil, err
	}
	u.Password = ""

	return &u, nil
}

func (s *UserService) UpdateAvatar(id uint, avatarPath string, avatarMime string) (*model.User, error) {
	var u model.User
	if err := s.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	u.AvatarPath = avatarPath
	u.AvatarMime = avatarMime
	u.AvatarUpdatedAt = &now
	if err := s.db.Save(&u).Error; err != nil {
		return nil, err
	}
	u.Password = ""
	return &u, nil
}
