package model

type User struct {
	BaseModel

	Name     string `gorm:"column:name;type:varchar(50);not null"`
	Email    string `gorm:"column:email;type:varchar(100);uniqueIndex;not null"`
	Password string `gorm:"column:password;type:varchar(255);not null"`
	Avatar   string `gorm:"column:avatar;type:varchar(255)"`
}

func (User) TableName() string {
	return "users"
}
